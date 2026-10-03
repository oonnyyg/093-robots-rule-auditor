// Command robots audits robots.txt rules against crawler path queries.
// Usage: go run ./cmd/robots input.json
package main

import (
	"encoding/json"
	"fmt"
	"os"

	robots "example.com/robots-rule-auditor"
)

type query struct {
	Crawler string `json:"crawler"`
	Path    string `json:"path"`
}

type input struct {
	Robots  string  `json:"robots"`
	Queries []query `json:"queries"`
}

type result struct {
	Crawler  string `json:"crawler"`
	Path     string `json:"path"`
	Allowed  bool   `json:"allowed"`
	RuleLine *int   `json:"ruleLine"`
	Groups   []int  `json:"groups"`
}

type output struct {
	Results []result `json:"results"`
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "robots:", err)
	os.Exit(1)
}

func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: robots <input.json>")
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		return err
	}
	var in input
	if err := json.Unmarshal(data, &in); err != nil {
		return fmt.Errorf("invalid input JSON: %w", err)
	}
	parsed, err := robots.Parse(in.Robots)
	if err != nil {
		return fmt.Errorf("invalid robots rules: %w", err)
	}
	out := output{Results: []result{}}
	for i, q := range in.Queries {
		if !robots.ValidCrawler(q.Crawler) {
			return fmt.Errorf("query %d: invalid crawler name %q", i, q.Crawler)
		}
		if !robots.ValidPath(q.Path) {
			return fmt.Errorf("query %d: invalid path %q", i, q.Path)
		}
		d := parsed.Decide(q.Crawler, q.Path)
		groups := d.Groups
		if groups == nil {
			groups = []int{}
		}
		out.Results = append(out.Results, result{
			Crawler:  q.Crawler,
			Path:     q.Path,
			Allowed:  d.Allowed,
			RuleLine: d.RuleLine,
			Groups:   groups,
		})
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func main() {
	if err := run(); err != nil {
		fail(err)
	}
}
