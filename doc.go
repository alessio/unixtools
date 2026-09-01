// Package tools provides a suite of UNIX convenience utilities implemented in Go.
//
// The suite contains command-line utilities designed to streamline everyday shell workflows:
//
//   - elvoke: Runs or postpones command execution based on elapsed time since the last successful invocation.
//   - mcd: Creates directories including missing parents, emitting a "cd" command for shell evaluation.
//   - refiles: Performs batch file renaming across directories using regular expressions and capture groups.
//   - seq: Generates integer sequences with support for custom intervals, separators, and zero-padding.
//
// Each utility is independently installable via standard Go tooling:
//
//	go install al.essio.dev/cmd/<tool>@latest
package cmd

