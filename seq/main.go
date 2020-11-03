package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/alessio/unixtools/internal/seq"
)

var (
	helpMode    bool
	versionMode bool

	separator string
	width     uint
)

func init() {
	flag.BoolVar(&helpMode, "help", false, "display this help and exit.")
	flag.BoolVar(&versionMode, "version", false, "output version information and exit.")
	flag.StringVar(&separator, "separator", `\n`, "use STRING to separate numbers.")
	flag.UintVar(&width, "width", 0, "equalize width by padding with leading zeroes.")
	flag.Usage = usage
	flag.ErrHelp = nil
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("seq: ")
	log.SetOutput(os.Stderr)
	flag.Parse()

	handleHelpAndVersionModes()

	separator, err := strconv.Unquote(`"` + separator + `"`)
	if err != nil {
		log.Fatal(err)
	}

	var (
		start = 1
		end   = 0
		incr  = 1
	)

	switch flag.NArg() {
	case 0:
		log.Fatal("missing operand")
	case 1:
		end = parseIntArg(0)
		if end < 0 {
			start = 1
		}
	case 2:
		start, end = parseIntArg(0), parseIntArg(1)
	case 3:
		start, incr, end = parseIntArg(0), parseIntArg(1), parseIntArg(2)
		if incr < 0 {
			log.Fatalf("%d is not a valid unsigned integer", incr)
		}
	default:
		log.Fatal("too many operands")
	}

	bldr := strings.Builder{}
	sequence := seq.NewInt(start, uint(incr), end, width)

	for item := range sequence.Items() {
		if bldr.Len() > 0 {
			fmt.Printf("%s%s", bldr.String(), separator)
			bldr.Reset()
		}

		bldr.WriteString(item)
	}

	fmt.Println(bldr.String())

	if sequence.WidthExceeded() {
		log.Fatal("width exceeded")
	}
}

func handleHelpAndVersionModes() {
	if helpMode {
		usage()
		os.Exit(0)
	}

	if versionMode {
		version()
		os.Exit(0)
	}
}

func parseIntArg(i int) int {
	out, err := strconv.Atoi(flag.Arg(i))
	if err != nil {
		log.Fatalf("%q is not a valid integer", flag.Arg(i))
	}
	return out
}

func usage() {
	usageString := `Usage: seq [OPTION]... LAST
  or:  seq [OPTION]... FIRST LAST
  or:  seq [OPTION]... FIRST INCREMENT LAST
Print numbers from FIRST to LAST, in steps of INCREMENT.
`
	_, _ = fmt.Fprintln(os.Stderr, usageString)
	flag.PrintDefaults()
}

func version() {
	_, _ = fmt.Fprintln(os.Stderr, "alessio's seq program, version 1.0.0")
	_, _ = fmt.Fprintln(os.Stderr, "Copyright (C) 2020 Alessio Treglia <alessio@debian.org>")
}
