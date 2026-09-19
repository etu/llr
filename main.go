package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	// Parse the command line flags
	width := flag.Int("width", getTerminalWidth(), "maximum width of the output")
	flag.IntVar(width, "w", *width, "maximum width of the output")
	debug := flag.Bool("debug", false, "enable debug output")
	flag.BoolVar(debug, "d", *debug, "enable debug output")
	compact := flag.Bool("compact", false, "compact wide columns to reduce wasted padding (reads all input before printing)")
	flag.BoolVar(compact, "c", *compact, "compact wide columns to reduce wasted padding (reads all input before printing)")
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.BoolVar(showVersion, "v", *showVersion, "print the version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	// Get the filename argument
	filename := "-"
	if flag.NArg() > 0 {
		filename = flag.Arg(0)
	}

	// Print the lines
	if *debug {
		fmt.Fprintln(os.Stderr, "Flags:")
		fmt.Fprintf(os.Stderr, "  debug: %t\n", *debug)
		fmt.Fprintf(os.Stderr, "  width: %d\n", *width)
		fmt.Fprintf(os.Stderr, "  compact: %t\n", *compact)
		fmt.Fprintf(os.Stderr, "  filename: %s\n", filename)
	}

	// Compacting needs every row to work out the column widths, so it has
	// to read the whole input first. Without it, lines are truncated and
	// printed as they arrive, which keeps `journalctl -f | llr` live.
	if *compact {
		// Read the file or stdin
		contents, err := readFileOrStdin(filename)
		if err != nil {
			log.Fatal("Error: ", err)
		}

		// Split the contents of the file into lines
		lines := compactColumns(strings.Split(string(contents), "\n"))

		if err := printLines(os.Stdout, *width, lines); err != nil {
			log.Fatal("Error: ", err)
		}

		return
	}

	input, err := openInput(filename)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	defer input.Close()

	if err := streamLines(input, os.Stdout, *width); err != nil {
		log.Fatal("Error: ", err)
	}
}
