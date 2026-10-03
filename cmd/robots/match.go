package main

import "strings"

// matchPattern reports whether pattern matches the beginning of path.
// '*' matches any run of characters; when anchor is true the match must
// consume the whole path, otherwise any suffix may remain.
func matchPattern(pattern, path string, anchor bool) bool {
	if !anchor {
		pattern += "*"
	}
	return globMatch(pattern, path)
}

// globMatch is a full-string wildcard match where '*' matches any run.
func globMatch(pattern, s string) bool {
	p, t := 0, 0
	starP, starT := -1, 0
	for t < len(s) {
		switch {
		case p < len(pattern) && pattern[p] == s[t]:
			p++
			t++
		case p < len(pattern) && pattern[p] == '*':
			starP = p
			starT = t
			p++
		case starP >= 0:
			p = starP + 1
			starT++
			t = starT
		default:
			return false
		}
	}
	for p < len(pattern) && pattern[p] == '*' {
		p++
	}
	return p == len(pattern)
}

// evaluate merges the groups matching the crawler (falling back to '*'
// groups) and picks the deciding rule: longest match wins, Allow beats
// Disallow on ties, then the lowest source line wins.
func evaluate(groups []*group, q query) result {
	if q.Path == "/robots.txt" {
		return result{Allowed: true, Groups: []int{}}
	}
	matched := selectGroups(groups, q.Crawler)
	res := result{Allowed: true, Groups: make([]int, 0, len(matched))}
	var best *rule
	for _, g := range matched {
		res.Groups = append(res.Groups, g.start)
		for i := range g.rules {
			r := &g.rules[i]
			if !matchPattern(r.pattern, q.Path, r.anchor) {
				continue
			}
			if best == nil || better(r, best) {
				best = r
			}
		}
	}
	if best != nil {
		line := best.line
		res.RuleLine = &line
		res.Allowed = best.allow
	}
	return res
}

func better(a, b *rule) bool {
	if a.length != b.length {
		return a.length > b.length
	}
	if a.allow != b.allow {
		return a.allow
	}
	return a.line < b.line
}

func selectGroups(groups []*group, crawler string) []*group {
	var named, wild []*group
	for _, g := range groups {
		isNamed, isWild := false, false
		for _, a := range g.agents {
			if a == "*" {
				isWild = true
			} else if strings.EqualFold(a, crawler) {
				isNamed = true
			}
		}
		if isNamed {
			named = append(named, g)
		} else if isWild {
			wild = append(wild, g)
		}
	}
	if len(named) > 0 {
		return named
	}
	return wild
}
