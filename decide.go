package robots

import "strings"

// Decision is the audit result for one query.
type Decision struct {
	Allowed  bool
	RuleLine *int  // selected rule line, nil when no rule matched
	Groups   []int // start lines of participating groups
}

// matchPattern reports whether pat matches path. '*' matches any number
// of characters; a trailing '$' anchors the match to the end of path.
// Matching is case-sensitive and starts at the first path character.
func matchPattern(pat, path string) bool {
	anchor := strings.HasSuffix(pat, "$")
	pat = strings.TrimSuffix(pat, "$")
	ends := map[int]bool{}
	matchFrom(pat, path, 0, 0, ends)
	if anchor {
		return ends[len(path)]
	}
	return len(ends) > 0
}

// matchFrom records all possible end positions of pat[pi:] matched
// against path starting at si.
func matchFrom(pat, path string, pi, si int, ends map[int]bool) {
	for pi < len(pat) {
		if pat[pi] == '*' {
			rest := pat[pi+1:]
			if rest == "" {
				ends[len(path)] = true
				return
			}
			for k := si; k <= len(path); k++ {
				matchFrom(rest, path, 0, k, ends)
			}
			return
		}
		if si >= len(path) || pat[pi] != path[si] {
			return
		}
		pi++
		si++
	}
	ends[si] = true
}

// patternLen counts characters excluding '*' and a trailing '$'.
func patternLen(pat string) int {
	pat = strings.TrimSuffix(pat, "$")
	n := 0
	for i := 0; i < len(pat); i++ {
		if pat[i] != '*' {
			n++
		}
	}
	return n
}

// selectGroups returns groups matching crawler (case-insensitive full
// match); if none match, groups containing the '*' agent.
func (p *Parsed) selectGroups(crawler string) []*Group {
	var named, star []*Group
	for _, g := range p.Groups {
		for _, a := range g.Agents {
			if a == "*" {
				star = append(star, g)
			} else if strings.EqualFold(a, crawler) {
				named = append(named, g)
			}
		}
	}
	if len(named) > 0 {
		return named
	}
	return star
}

// Decide evaluates whether crawler may fetch path.
func (p *Parsed) Decide(crawler, path string) Decision {
	d := Decision{}
	if path == "/robots.txt" {
		d.Allowed = true
		return d
	}
	groups := p.selectGroups(crawler)
	for _, g := range groups {
		d.Groups = append(d.Groups, g.Start)
	}
	var best *Rule
	bestLen := -1
	for _, g := range groups {
		for i := range g.Rules {
			r := &g.Rules[i]
			if !matchPattern(r.Pattern, path) {
				continue
			}
			l := patternLen(r.Pattern)
			if best == nil || l > bestLen ||
				(l == bestLen && r.Allow && !best.Allow) ||
				(l == bestLen && r.Allow == best.Allow && r.Line < best.Line) {
				best, bestLen = r, l
			}
		}
	}
	if best == nil {
		d.Allowed = true
		return d
	}
	d.Allowed = best.Allow
	line := best.Line
	d.RuleLine = &line
	return d
}
