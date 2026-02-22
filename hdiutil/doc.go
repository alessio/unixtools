// Package hdiutil provides a Go wrapper around the macOS hdiutil command-line tool
// for creating, manipulating, and signing DMG disk images.
//
// # Supported formats and filesystems
//
// The following compressed image formats are supported via [Config.ImageFormat]:
//
//   - UDZO — zlib compression (level 9). This is the default and the most
//     widely compatible format.
//   - UDBZ — bzip2 compression (level 9). Better compression ratio than UDZO
//     at the cost of slower creation and extraction.
//   - ULFO — lzfse compression. Apple's modern codec; fast with good ratios,
//     but only supported on macOS 10.11+.
//   - ULMO — lzma compression. Highest compression ratio, slowest speed.
//
// The following filesystem types are supported via [Config.FileSystem]:
//
//   - HFS+ — the default; includes tuned allocation parameters
//     (-fsargs -c c=64,a=16,e=16).
//   - APFS — Apple File System. Cannot be combined with [Config.SandboxSafe].
//
// # Configuration
//
// A [Config] struct holds all settings for image creation. It can be built
// programmatically or loaded from a JSON file with [LoadConfig]:
//
//	cfg, err := hdiutil.LoadConfig("dmg.json")
//
// Configs can also be serialized and deserialized with [Config.ToJSON] and
// [Config.FromJSON] for round-tripping through pipelines or storage.
//
// [Config.Validate] must be called (either directly or implicitly through
// [Runner.Setup]) before the lazy option functions ([Config.FilesystemOpts],
// [Config.ImageFormatOpts], [Config.VolumeSizeOpts], [Config.VolumeNameOpt])
// become usable. Calling them before validation panics.
//
// Required fields:
//
//   - [Config.SourceDir] — directory whose contents are copied into the DMG.
//   - [Config.OutputPath] — destination path; must end in ".dmg".
//
// Optional fields with defaults:
//
//   - [Config.VolumeName] — defaults to the output filename without extension
//     (e.g. "MyApp.dmg" → "MyApp").
//   - [Config.VolumeSizeMb] — when zero, hdiutil sizes the image automatically.
//   - [Config.ImageFormat] — defaults to "UDZO".
//   - [Config.FileSystem] — defaults to "HFS+".
//
// # Runner lifecycle
//
// [New] creates a [Runner] from a [Config]. The [Runner] must go through a
// fixed sequence of steps; calling methods out of order returns an error
// (typically [ErrNeedInit]).
//
//	cfg := &hdiutil.Config{
//		SourceDir:  "path/to/source",
//		OutputPath: "output.dmg",
//		VolumeName: "MyVolume",
//	}
//
//	runner := hdiutil.New(cfg)
//	defer runner.Cleanup()
//
//	// 1. Validate config, create temp directory.
//	if err := runner.Setup(); err != nil {
//		log.Fatal(err)
//	}
//
//	// 2. Create a writable temporary image populated from SourceDir.
//	if err := runner.Start(); err != nil {
//		log.Fatal(err)
//	}
//
//	// 3. (Optional) Mount the image, modify contents, mark bootable, unmount.
//	if err := runner.AttachDiskImage(); err != nil {
//		log.Fatal(err)
//	}
//	// ... copy additional files into runner.MountDir, customise .DS_Store, etc.
//	_ = runner.Bless()           // mark as bootable (no-op unless Config.Bless is set)
//	_ = runner.DetachDiskImage() // fixes permissions and unmounts
//
//	// 4. Convert the writable image to the final compressed format.
//	if err := runner.FinalizeDMG(); err != nil {
//		log.Fatal(err)
//	}
//
//	// 5. (Optional) Sign and notarize.
//	if err := runner.Codesign(); err != nil {  // no-op when SigningIdentity is empty
//		log.Fatal(err)
//	}
//	if err := runner.Notarize(); err != nil {  // no-op when NotarizeCredentials is empty
//		log.Fatal(err)
//	}
//
// [Runner.Cleanup] removes the temporary working directory and is safe to call
// multiple times.
//
// # Sandbox-safe images
//
// Setting [Config.SandboxSafe] uses a two-step process (hdiutil makehybrid +
// convert) that produces images openable by sandboxed macOS applications.
// APFS cannot be used in this mode; attempting it returns [ErrSandboxAPFS].
// The [Runner.Bless] step is also skipped for sandbox-safe images.
//
// # Code signing and notarization
//
// When [Config.SigningIdentity] is set, [Runner.Codesign] signs the final DMG
// and verifies the signature with --deep --strict. When
// [Config.NotarizeCredentials] is set to a keychain profile name,
// [Runner.Notarize] submits the DMG via xcrun notarytool and staples the
// ticket with xcrun stapler. Both methods are no-ops when their respective
// config fields are empty.
//
// # Verbosity
//
// [Config.HDIUtilVerbosity] controls the flags passed to hdiutil:
//
//   - 0 — no flag (default).
//   - 1 — -quiet.
//   - 2 — -verbose.
//   - 3+ — -debug.
//
// Negative values are treated as 0.
//
// # Logging
//
// Internal log messages are discarded by default. Call [SetLogWriter] with
// [os.Stderr] (or any [io.Writer]) to enable them.
//
// # Dry-run mode
//
// Setting [Config.Simulate] logs every external command without executing it,
// which is useful for previewing the hdiutil invocations that would be made.
//
// # Input sanitization
//
// [Config.Validate] rejects values that could lead to OS command argument
// injection:
//
//   - Null bytes in any string field (SourceDir, OutputPath, VolumeName,
//     SigningIdentity, NotarizeCredentials).
//   - Paths (SourceDir, OutputPath) that start with a dash after
//     [filepath.Clean], which could be misinterpreted as flags by external
//     commands.
//
// # Error handling
//
// Sentinel errors are defined for every category of failure and can be
// matched with [errors.Is]:
//
//   - [ErrUnsafeArg] — config value contains null bytes or unsafe characters.
//   - [ErrInvSourceDir] — empty or missing source directory.
//   - [ErrImageFileExt] — output path does not end in ".dmg".
//   - [ErrInvFormatOpt] — unsupported image format.
//   - [ErrInvFilesystemOpt] — unsupported filesystem.
//   - [ErrVolumeSize] — negative volume size.
//   - [ErrSandboxAPFS] — sandbox-safe mode with APFS.
//   - [ErrNeedInit] — [Runner.Setup] was not called.
//   - [ErrCreateDir] — failed to create temporary directory.
//   - [ErrMountImage] — attach/mount failed.
//   - [ErrCodesignFailed] — signing or verification failed.
//   - [ErrNotarizeFailed] — notarization or stapling failed.
//
// # Testing
//
// The [CommandExecutor] interface and the [WithExecutor] functional option
// allow injecting a mock executor into [New], so tests can verify command
// arguments and simulate failures without invoking real binaries.
// [CommandExecutor] uses typed methods (Hdiutil, Codesign, Xcrun, Chmod,
// Bless) rather than a generic Run(name, args...) to ensure that only
// known commands can be executed and that static analysis tools see
// literal command names in each [exec.Command] call.
package hdiutil
