package version

import (
	"fmt"
	"strings"

	_ "embed"
)

//go:generate sh generate_version.sh
//go:embed version.txt
var Version string

func Short() string {
	return fmt.Sprintf("unixtools %s", Version)
}

func PrintWithCopyright() {
	_, _ = fmt.Println(longWithCopyright())
}

func longWithCopyright() string {
	return fmt.Sprintf("alessio's unixtools, Version %s\nCopyright (C) 2020, 2021, 2022, 2023 Alessio Treglia <alessio@debian.org>", strings.TrimSpace(Version))
}
