//go:build ignore

package main

import (
	"basket-cost/internal/ticket"
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var reTotal = regexp.MustCompile(`TOTAL \(€\)\s*[\n\r]+([\d,]+)`)
var reTotalInline = regexp.MustCompile(`TOTAL \(€\)\s+([\d,]+)`)

func main() {
	paths, _ := filepath.Glob(os.Args[1])
	if len(paths) == 0 {
		paths = os.Args[1:]
	}

	ext := ticket.NewExtractor()
	parser := ticket.NewMercadonaParser()

	var problems []string
	var ok int
	for _, path := range paths {
		name := filepath.Base(path)
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v\n", name, err)
			continue
		}

		text, err := ext.Extract(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			fmt.Fprintf(os.Stderr, "extract %s: %v\n", name, err)
			continue
		}

		t, err := parser.Parse(text)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse %s: %v\n", name, err)
			continue
		}

		var m []string
		m = reTotal.FindStringSubmatch(text)
		if m == nil {
			m = reTotalInline.FindStringSubmatch(text)
		}
		if m == nil {
			continue
		}
		declaredStr := strings.ReplaceAll(m[1], ",", ".")
		declared, _ := strconv.ParseFloat(declaredStr, 64)

		var computed float64
		for _, l := range t.Lines {
			computed += l.UnitPrice * float64(l.Quantity)
		}

		diff := math.Abs(declared - computed)
		if diff > 0.05 {
			problems = append(problems, fmt.Sprintf("DIFF %.2f  declared=%.2f computed=%.2f lines=%d  %s", diff, declared, computed, len(t.Lines), name))
		} else {
			ok++
		}
	}

	fmt.Printf("OK: %d tickets cuadran\n", ok)
	if len(problems) == 0 {
		fmt.Println("Sin diferencias.")
	} else {
		for _, p := range problems {
			fmt.Println(p)
		}
	}
}
