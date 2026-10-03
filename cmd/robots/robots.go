package main

import "strings"

// validCrawler reports whether name is non-empty and contains only ASCII
// letters, underscores, or hyphens.
func validCrawler(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' || c == '-' {
			continue
		}
		return false
	}
	return true
}

// validPath reports whether p starts with '/', contains only visible ASCII
// characters (no percent signs, no whitespace), and has no fragment.
func validPath(p string) bool {
	if !strings.HasPrefix(p, "/") {
		return false
	}
	for i := 0; i < len(p); i++ {
		c := p[i]
		if c < 0x21 || c > 0x7e || c == '%' || c == '#' {
			return false
		}
	}
	return true
}

type rule struct {
	allow   bool
	pattern string // without trailing '$'
	anchor  bool   // pattern ended with '$'
	length  int    // chars excluding '*' and trailing '$'
	line    int    // 1-based source line
}

type group struct {
	start  int // 1-based line of the group's first user-agent line
	agents []string
	rules  []rule
}

// parseRule validates an Allow/Disallow value. Empty rules and rules not
// starting with '/' or containing '$' away from the end never match.
func parseRule(allow bool, value string, line int) (rule, bool) {
	if value == "" || !strings.HasPrefix(value, "/") {
		return rule{}, false
	}
	r := rule{allow: allow, pattern: value, line: line}
	if strings.HasSuffix(value, "$") {
		r.anchor = true
		r.pattern = value[:len(value)-1]
	}
	if strings.Contains(r.pattern, "$") {
		return rule{}, false
	}
	for i := 0; i < len(r.pattern); i++ {
		if r.pattern[i] != '*' {
			r.length++
		}
	}
	return r, true
}

// parseRobots splits robots.txt content into groups. Consecutive user-agent
// lines share a group; a user-agent line after an Allow/Disallow starts a
// new group. Blank lines never split groups. Unknown directives, comment
// text, and rules outside any group are ignored.
func parseRobots(content string) []*group {
	var groups []*group
	var cur *group
	hasRules := false
	lines := strings.Split(content, "\n")
	for i, raw := range lines {
		lineNo := i + 1
		if idx := strings.IndexByte(raw, '#'); idx >= 0 {
			raw = raw[:idx]
		}
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		colon := strings.IndexByte(raw, ':')
		if colon < 0 {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(raw[:colon]))
		value := strings.TrimSpace(raw[colon+1:])
		switch name {
		case "user-agent":
			if value != "*" && !validCrawler(value) {
				continue
			}
			if cur == nil || hasRules {
				cur = &group{start: lineNo}
				groups = append(groups, cur)
				hasRules = false
			}
			cur.agents = append(cur.agents, value)
		case "allow", "disallow":
			if cur == nil {
				continue
			}
			if r, ok := parseRule(name == "allow", value, lineNo); ok {
				cur.rules = append(cur.rules, r)
			}
			hasRules = true
		}
	}
	return groups
}
