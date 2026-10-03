package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeInput(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "input.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunEndToEnd(t *testing.T) {
	input := `{
	  "rules": "User-agent: bot\nDisallow: /admin\nAllow: /admin/public$\nUser-agent: *\nDisallow: /tmp",
	  "queries": [
	    {"crawler": "Bot", "path": "/admin/panel"},
	    {"crawler": "bot", "path": "/admin/public"},
	    {"crawler": "stranger", "path": "/tmp/x"},
	    {"crawler": "stranger", "path": "/robots.txt"}
	  ]
	}`
	var out bytes.Buffer
	if err := run([]string{writeInput(t, input)}, &out); err != nil {
		t.Fatal(err)
	}
	var got outputFile
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Results) != 4 {
		t.Fatalf("got %d results, want 4", len(got.Results))
	}
	wantAllowed := []bool{false, true, false, true}
	wantLines := []int{2, 3, 5, 0}
	for i, r := range got.Results {
		if r.Allowed != wantAllowed[i] {
			t.Errorf("result %d: allowed=%v, want %v", i, r.Allowed, wantAllowed[i])
		}
		line := 0
		if r.RuleLine != nil {
			line = *r.RuleLine
		}
		if line != wantLines[i] {
			t.Errorf("result %d: ruleLine=%d, want %d", i, line, wantLines[i])
		}
	}
	if len(got.Results[0].Groups) != 1 || got.Results[0].Groups[0] != 1 {
		t.Errorf("result 0 groups = %v, want [1]", got.Results[0].Groups)
	}
	if len(got.Results[2].Groups) != 1 || got.Results[2].Groups[0] != 4 {
		t.Errorf("result 2 groups = %v, want [4]", got.Results[2].Groups)
	}
}

func TestRunRejectsInvalidInput(t *testing.T) {
	cases := map[string]string{
		"bad crawler":   `{"rules":"","queries":[{"crawler":"bot 1","path":"/"}]}`,
		"empty crawler": `{"rules":"","queries":[{"crawler":"","path":"/"}]}`,
		"relative path": `{"rules":"","queries":[{"crawler":"bot","path":"x"}]}`,
		"percent path":  `{"rules":"","queries":[{"crawler":"bot","path":"/a%20"}]}`,
		"fragment path": `{"rules":"","queries":[{"crawler":"bot","path":"/a#f"}]}`,
		"space path":    `{"rules":"","queries":[{"crawler":"bot","path":"/a b"}]}`,
		"not json":      `{"rules":`,
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			var out bytes.Buffer
			if err := run([]string{writeInput(t, input)}, &out); err == nil {
				t.Fatal("expected error, got nil")
			}
			if out.Len() != 0 {
				t.Errorf("stdout not empty on error: %q", out.String())
			}
		})
	}
	if err := run(nil, &bytes.Buffer{}); err == nil {
		t.Error("expected usage error for missing argument")
	}
}
