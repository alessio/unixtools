# hdiutil

[![GoDoc](https://pkg.go.dev/badge/al.essio.dev/cmd/mkdmg/pkg/hdiutil.svg)](https://pkg.go.dev/al.essio.dev/cmd/mkdmg/pkg/hdiutil)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/gh/alessio/mkdmg/pkg/hdiutil)](https://www.codacy.com/gh/alessio/mkdmg/pkg/hdiutil/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=alessio/mkdmg&amp;utm_campaign=Badge_Grade)
[![codecov](https://codecov.io/gh/alessio/mkdmg/pkg/hdiutil/branch/main/graph/badge.svg)](https://codecov.io/gh/alessio/mkdmg/pkg/hdiutil)
[![Go Report Card](https://goreportcard.com/badge/github.com/alessio/mkdmg/pkg/hdiutil)](https://goreportcard.com/report/github.com/alessio/mkdmg/pkg/hdiutil)
[![License](https://img.shields.io/github/license/alessio/mkdmg.svg)](https://github.com/alessio/mkdmg/blob/main/LICENSE)

Package `hdiutil` provides a robust Go wrapper around the macOS `hdiutil` command-line tool, simplifying the creation, manipulation, and signing of DMG disk images.

It allows programmatic control over the entire DMG lifecycle, from creating writable temporary images to finalizing them into compressed, read-only distribution artifacts.

## ✨ Features

*   **Format Support:** Create DMGs in various formats: `UDZO` (zlib), `UDBZ` (bzip2), `ULFO` (lzfse), and `ULMO` (lzma).
*   **Filesystems:** Support for both legacy **HFS+** and modern **APFS** filesystems.
*   **Workflow Orchestration:** High-level `Runner` struct to manage the complex sequence of creating, mounting, copying, and converting images.
*   **Security:** Integrated support for **Codesigning** and **Apple Notarization** (including stapling).
*   **Sandboxing:** Option to create sandbox-safe disk images.

## 📦 Installation

```sh
go get al.essio.dev/cmd/mkdmg/pkg/hdiutil
```

## 🚀 Usage

Here is a simple example of how to use the package to create a DMG:

```go
package main

import (
	"log"

	"al.essio.dev/cmd/mkdmg/pkg/hdiutil"
)

func main() {
	cfg := &hdiutil.Config{
		SourceDir:  "./dist",
		OutputPath: "MyApp.dmg",
		VolumeName: "My App",
		FileSystem: "HFS+",
	}

	runner := hdiutil.New(cfg)
	// Ensure temporary files are cleaned up
	defer runner.Cleanup()

	// Initialize the runner
	if err := runner.Setup(); err != nil {
		log.Fatal(err)
	}

	// 1. Create temporary writable image
	if err := runner.Start(); err != nil {
		log.Fatal(err)
	}

	// 2. Attach (mount) the image (Optional, if you need to modify it)
	// if err := runner.AttachDiskImage(); err != nil {
	// 	log.Fatal(err)
	// }
	// ... perform operations on mount ...
	// if err := runner.DetachDiskImage(); err != nil {
	// 	log.Fatal(err)
	// }

	// 3. Convert to final compressed DMG
	if err := runner.FinalizeDMG(); err != nil {
		log.Fatal(err)
	}

	// 4. Sign and Notarize (Optional)
	// if err := runner.Codesign(); err != nil { ... }
	// if err := runner.Notarize(); err != nil { ... }
}
```

## 📄 License

This package is part of the `mkdmg` project and is released under the same license.
