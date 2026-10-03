// Package robots implements an offline robots.txt rule auditor.
package robots

import (
	"fmt"
	"strings"
)

// Rule is a single Allow/Disallow directive.
type Rule struct {
	Allow   bool
	Pattern string
	Line    int // 1-based original line number
}

// Group is a set of user-agents and the rules that follow them.
type Group struct {
	Agents []string
	Rules  []Rule
	Start  int // line number of the group's first user-agent line
}

// Parsed holds the parsed groups of a robots.txt document.
type Parsed struct {
	Groups []*Group
}

var validAgentChars = func() map[rune]bool {
	m := map[rune]bool{'_': true, '-': true}
	for c := 'a'; c <= 'z'; c++ {
		m[c] = true
	}
	for c := 'A'; c <= 'Z'; c++ {
		m[c] = true
	}
	return m
}()

// ValidCrawler reports whether name is a legal crawler name:
// non-empty, ASCII letters, underscore and hyphen only.
func ValidCrawler(name string) bool {
	if name == "" {
		return false
	}
	for _, c := range name {
		if !validAgentChars[c] {
			return false
		}
	}
	return true
}

// ValidPath reports whether p is a legal query path: starts with '/',
// visible ASCII only, no '%', no whitespace, no fragment '#'.
func ValidPath(p string) bool {
	if !strings.HasPrefix(p, "/") {
		return false
	}
	for _, c := range p {
		if c < 0x21 || c > 0x7e || c == '%' || c == '#' {
			return false
		}
	}
	return true
}

// validatePattern checks a non-empty rule pattern.
func validatePattern(pat string) error {
	if !strings.HasPrefix(pat, "/") {
		return fmt.Errorf("rule pattern %q must start with '/'", pat)
	}
	body := strings.TrimSuffix(pat, "$")
	if strings.Contains(body, "$") {
		return fmt.Errorf("rule pattern %q contains '$' in a non-final position", pat)
	}
	return nil
}

// Parse parses robots.txt content into groups.
// Unknown directives and rules outside any group are ignored.
func Parse(content string) (*Parsed, error) {
	p := &Parsed{}
	var cur *Group
	hasRule := false
	lines := strings.Split(content, "\n")
	for i, raw := range lines {
		lineNo := i + 1
		line := raw
		if idx := strings.IndexByte(line, '#'); idx >= 0 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		name, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)
		switch strings.ToLower(name) {
		case "user-agent":
			if cur == nil || hasRule {
				cur = &Group{Start: lineNo}
				p.Groups = append(p.Groups, cur)
				hasRule = false
			}
			if value != "*" && !ValidCrawler(value) {
				return nil, fmt.Errorf("line %d: invalid user-agent value %q", lineNo, value)
			}
			cur.Agents = append(cur.Agents, value)
		case "allow", "disallow":
			if cur == nil {
				continue // rule outside any group
			}
			hasRule = true
			if value == "" {
				continue // empty rule does not participate
			}
			if err := validatePattern(value); err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNo, err)
			}
			cur.Rules = append(cur.Rules, Rule{
				Allow:   strings.EqualFold(name, "allow"),
				Pattern: value,
				Line:    lineNo,
			})
		default:
			// unknown directive ignored
		}
	}
	return p, nil
}
