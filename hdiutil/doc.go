// Package hdiutil provides a Go wrapper around macOS hdiutil command-line tool
// for creating, manipulating, and signing DMG disk images.
//
// The package supports various DMG formats (UDZO, UDBZ, ULFO, ULMO) and filesystems
// (HFS+, APFS), as well as optional code signing and Apple notarization.
//
// It provides a high-level Runner that orchestrates the process of creating
// a writable image, mounting it, and converting it to a final compressed format.
//
// Example usage:
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
//	if err := runner.Setup(); err != nil {
//		log.Fatal(err)
//	}
//
//	if err := runner.Start(); err != nil {
//		log.Fatal(err)
//	}
//
//	// Optionally attach, modify, bless, and detach
//	// ...
//
//	if err := runner.FinalizeDMG(); err != nil {
//		log.Fatal(err)
//	}
package hdiutil
