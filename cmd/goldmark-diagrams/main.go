// Command goldmark-diagrams renders Markdown read from stdin to HTML on stdout.
// It is the container entry point the end-to-end fixtures run through, so it
// exercises the same public API a host integrates.
package main

import (
	"fmt"
	"io"
	"os"

	diagrams "github.com/frinknet/goldmark-diagrams"
	"github.com/yuin/goldmark"
)

func main() {
	source, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "goldmark-diagrams: read stdin: %v\n", err)
		os.Exit(1)
	}
	md := goldmark.New(goldmark.WithExtensions(diagrams.New()))
	if err := md.Convert(source, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "goldmark-diagrams: convert: %v\n", err)
		os.Exit(1)
	}
}
