# hdiutil

[![Go Reference](https://pkg.go.dev/badge/al.essio.dev/pkg/tools/hdiutil.svg)](https://pkg.go.dev/al.essio.dev/pkg/tools/hdiutil)
[![Go Report Card](https://goreportcard.com/badge/al.essio.dev/pkg/tools)](https://goreportcard.com/report/al.essio.dev/pkg/tools)
[![License](https://img.shields.io/github/license/alessio/unixtools.svg)](https://github.com/alessio/unixtools/blob/main/LICENSE)

Package `hdiutil` provides a Go wrapper around the macOS `hdiutil` command-line tool for creating, manipulating, and signing DMG disk images.

## Features

- **Image formats** — UDZO (zlib, default), UDBZ (bzip2), ULFO (lzfse), ULMO (lzma).
- **Filesystems** — HFS+ (default, with tuned allocation parameters) and APFS.
- **Workflow orchestration** — `Runner` manages the full lifecycle: create, mount, modify, convert, sign, notarize.
- **Sandbox-safe images** — produce DMGs openable by sandboxed macOS applications.
- **Code signing and notarization** — integrated codesign and Apple notarytool/stapler support.
- **JSON configuration** — load/save `Config` from JSON files for CI/CD pipelines.
- **Input sanitization** — rejects null bytes and dash-prefixed paths to prevent argument injection.
- **Dry-run mode** — preview all hdiutil invocations without executing them.
- **Testable** — typed `CommandExecutor` interface with `WithExecutor` option for mock injection.

## Installation

```sh
go get "al.essio.dev/pkg/tools/hdiutil"
```

## Usage

```go
package main

import (
	"log"

	"al.essio.dev/pkg/tools/hdiutil"
)

func main() {
	cfg := &hdiutil.Config{
		SourceDir:  "./dist",
		OutputPath: "MyApp.dmg",
		VolumeName: "My App",
	}

	runner := hdiutil.New(cfg)
	defer runner.Cleanup()

	// 1. Validate config, create temp directory.
	if err := runner.Setup(); err != nil {
		log.Fatal(err)
	}

	// 2. Create a writable temporary image populated from SourceDir.
	if err := runner.Start(); err != nil {
		log.Fatal(err)
	}

	// 3. (Optional) Mount, modify contents, unmount.
	// if err := runner.AttachDiskImage(); err != nil { log.Fatal(err) }
	// ... copy files, customise .DS_Store, etc.
	// _ = runner.Bless()           // mark bootable (no-op unless Config.Bless is set)
	// _ = runner.DetachDiskImage() // fixes permissions and unmounts

	// 4. Convert to final compressed DMG.
	if err := runner.FinalizeDMG(); err != nil {
		log.Fatal(err)
	}

	// 5. (Optional) Sign and notarize — no-ops when credentials are empty.
	if err := runner.Codesign(); err != nil {
		log.Fatal(err)
	}
	if err := runner.Notarize(); err != nil {
		log.Fatal(err)
	}
}
```

### JSON configuration

Config can be loaded from a JSON file, which is useful for CI/CD:

```json
{
  "source_dir": "./dist",
  "output_path": "MyApp.dmg",
  "volume_name": "My App",
  "filesystem": "HFS+",
  "image_format": "UDZO",
  "signing_identity": "Developer ID Application: Example Inc",
  "notarize_credentials": "AC_PASSWORD"
}
```

```go
cfg, err := hdiutil.LoadConfig("dmg.json")
```

### Sandbox-safe images

```go
cfg := &hdiutil.Config{
	SourceDir:   "./dist",
	OutputPath:  "MyApp.dmg",
	SandboxSafe: true,
	FileSystem:  "HFS+", // APFS is not supported in sandbox-safe mode
}
```

## License

This package is part of the [unixtools](https://github.com/alessio/unixtools) project and is released under the same license.
