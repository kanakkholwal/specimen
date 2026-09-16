// Command fixtures regenerates testdata/fixtures/*.result.json from a saved flight payload.
package main

import (
	"fmt"
	"os"

	"github.com/kanak/design-supply/internal/flight"
	"github.com/kanak/design-supply/internal/sites/refero"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: fixtures <flight-file> <out.json>")
		os.Exit(2)
	}
	b, err := os.ReadFile(os.Args[1])
	check(err)
	doc, err := flight.Parse(b)
	check(err)
	out, err := refero.RawResultJSON(doc)
	check(err)
	check(os.WriteFile(os.Args[2], out, 0o644))
	fmt.Printf("wrote %s (%d bytes)\n", os.Args[2], len(out))
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
