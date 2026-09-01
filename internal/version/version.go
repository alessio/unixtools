// Package version provides version and build information for the unixtools suite.
package version

import (
	"fmt"
	"strings"

	_ "embed"
)

// Version contains the current semantic version of unixtools, embedded at build time.
//
//go:generate bash generate_version.sh
//go:embed version.txt
var Version string

// Short returns a concise single-line version identifier (e.g. "unixtools v1.0.0").
func Short() string {
	return fmt.Sprintf("unixtools %s", strings.TrimSpace(Version))
}

// PrintWithCopyright prints the long version string along with copyright notices to standard output.
func PrintWithCopyright() {
	_, _ = fmt.Println(longWithCopyright())
}

func longWithCopyright() string {
	return fmt.Sprintf("alessio's unixtools, Version %s\nCopyright (C) 2020, 2021, 2022, 2023 Alessio Treglia <alessio@debian.org>", strings.TrimSpace(Version))
}

