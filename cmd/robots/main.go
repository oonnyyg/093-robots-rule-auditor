// Command robots audits robots.txt rules offline and reports, for each
// crawler path query, whether fetching is allowed and why.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type inputFile struct {
	Rules   string  `json:"rules"`
	Queries []query `json:"queries"`
}

type query struct {
	Crawler string `json:"crawler"`
	Path    string `json:"path"`
}

type result struct {
	Allowed  bool  `json:"allowed"`
	RuleLine *int  `json:"ruleLine"`
	Groups   []int `json:"groups"`
}

type outputFile struct {
	Results []result `json:"results"`
}

func run(args []string, stdout io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: robots <input.json>")
	}
	data, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}
	var in inputFile
	if err := json.Unmarshal(data, &in); err != nil {
		return fmt.Errorf("parse input JSON: %w", err)
	}
	for i, q := range in.Queries {
		if !validCrawler(q.Crawler) {
			return fmt.Errorf("query %d: invalid crawler name %q", i, q.Crawler)
		}
		if !validPath(q.Path) {
			return fmt.Errorf("query %d: invalid path %q", i, q.Path)
		}
	}
	groups := parseRobots(in.Rules)
	out := outputFile{Results: make([]result, 0, len(in.Queries))}
	for _, q := range in.Queries {
		out.Results = append(out.Results, evaluate(groups, q))
	}
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "robots:", err)
		os.Exit(1)
	}
}
