// Command evaluate runs a comprehensive project completion analysis and outputs
// a scored report with weighted categories, milestone tracking, and detailed breakdowns.
//
// Usage:
//
//	go run ./cmd/evaluate -dir . -format terminal
//	go run ./cmd/evaluate -dir . -format json -output report.json
//	go run ./cmd/evaluate -dir . -format both -output report.json
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/markbfromdc/cloudcode/internal/evaluate"
)

func main() {
	rootDir := flag.String("dir", ".", "Project root directory")
	format := flag.String("format", "terminal", "Output format: terminal, json, both")
	output := flag.String("output", "", "Output file path (empty for stdout)")
	skipBuild := flag.Bool("skip-build", false, "Skip running build commands")
	skipTests := flag.Bool("skip-tests", false, "Skip running test suites")
	flag.Parse()

	opts := evaluate.Options{
		SkipBuild: *skipBuild,
		SkipTests: *skipTests,
	}

	eval, err := evaluate.Evaluate(*rootDir, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "evaluation failed: %v\n", err)
		os.Exit(1)
	}

	// Terminal output.
	if *format == "terminal" || *format == "both" {
		fmt.Print(evaluate.FormatTerminal(eval))
	}

	// JSON output.
	if *format == "json" || *format == "both" {
		jsonData, err := evaluate.FormatJSON(eval)
		if err != nil {
			fmt.Fprintf(os.Stderr, "JSON formatting failed: %v\n", err)
			os.Exit(1)
		}

		if *output != "" {
			if err := os.WriteFile(*output, jsonData, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "failed to write output file: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Report written to %s\n", *output)
		} else if *format == "json" {
			fmt.Println(string(jsonData))
		}
	}
}
