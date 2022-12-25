package main // import "github.com/etu/llr"

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func main() {
	// Parse the command line flags
	width := flag.Int("width", getTerminalWidth(), "maximum width of the output")
	flag.IntVar(width, "w", *width, "maximum width of the output")
	debug := flag.Bool("debug", false, "enable debug output")
	flag.BoolVar(debug, "d", *debug, "enable debug output")
	flag.Parse()

	// Get the filename argument
	filename := "-"
	if flag.NArg() > 0 {
		filename = flag.Arg(0)
	}

	// Open the file or stdin
	var reader io.Reader
	if filename == "-" {
		reader = os.Stdin
	} else {
		file, err := os.Open(filename)
		if err != nil {
			log.Fatal("Error: ", err)
		}
		defer file.Close()
		reader = file
	}

	// Read the file or stdin
	contents, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	// Split the contents of the file into lines
	lines := strings.Split(string(contents), "\n")

	// Print the lines
	if *debug {
		fmt.Fprintln(os.Stderr, "Flags:")
		fmt.Fprintf(os.Stderr, "  width: %d\n", *width)
		fmt.Fprintf(os.Stderr, "  debug: %t\n", *debug)
	}

	printLines(os.Stdout, *width, lines)
}
