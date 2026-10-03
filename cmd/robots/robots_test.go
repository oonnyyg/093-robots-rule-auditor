package main

import "testing"

func TestValidCrawler(t *testing.T) {
	valid := []string{"Googlebot", "bing_bot", "my-crawler", "A", "_-", "a_b-C"}
	for _, s := range valid {
		if !validCrawler(s) {
			t.Errorf("validCrawler(%q) = false, want true", s)
		}
	}
	invalid := []string{"", "bot1", "bot*", "bot bot", "bot/x", "机器人"}
	for _, s := range invalid {
		if validCrawler(s) {
			t.Errorf("validCrawler(%q) = true, want false", s)
		}
	}
}

func TestValidPath(t *testing.T) {
	valid := []string{"/", "/a/b", "/a?x=1&y=2", "/a-b_c.d~e!", "/*$"}
	for _, s := range valid {
		if !validPath(s) {
			t.Errorf("validPath(%q) = false, want true", s)
		}
	}
	invalid := []string{"", "a", "/a b", "/a%20b", "/a#frag", "/\t", "/é", "/a\nb"}
	for _, s := range invalid {
		if validPath(s) {
			t.Errorf("validPath(%q) = true, want false", s)
		}
	}
}

func TestParseRobotsGrouping(t *testing.T) {
	rules := "" +
		"User-agent: one\n" + // line 1
		"User-agent: two\n" + // line 2, same group
		"\n" + // blank line does not split
		"User-agent: three\n" + // line 4, still same group
		"Disallow: /a\n" + // line 5
		"User-agent: four\n" + // line 6, new group after a rule
		"Disallow: /b\n" // line 7
	groups := parseRobots(rules)
	if len(groups) != 2 {
		t.Fatalf("got %d groups, want 2", len(groups))
	}
	if groups[0].start != 1 || len(groups[0].agents) != 3 {
		t.Errorf("group 0: start=%d agents=%v", groups[0].start, groups[0].agents)
	}
	if groups[1].start != 6 || len(groups[1].agents) != 1 {
		t.Errorf("group 1: start=%d agents=%v", groups[1].start, groups[1].agents)
	}
}

func TestParseRobotsIgnores(t *testing.T) {
	rules := "" +
		"Disallow: /outside\n" + // no group yet: ignored
		"Unknown: x\n" +
		"not a directive line\n" +
		"User-agent: bad name\n" + // invalid agent: ignored
		"User-agent: bot # inline comment\n" +
		"ALLOW: /ok\n" + // directive name case-insensitive
		"Disallow: relative\n" + // must start with '/': ignored
		"Disallow: /mid$dle\n" + // '$' only at end: ignored
		"Disallow:   \n" // empty rule: ignored
	groups := parseRobots(rules)
	if len(groups) != 1 {
		t.Fatalf("got %d groups, want 1", len(groups))
	}
	g := groups[0]
	if len(g.agents) != 1 || g.agents[0] != "bot" {
		t.Errorf("agents = %v, want [bot]", g.agents)
	}
	if len(g.rules) != 1 || !g.rules[0].allow || g.rules[0].pattern != "/ok" {
		t.Errorf("rules = %+v, want single Allow /ok", g.rules)
	}
}

func TestParseRuleAnchorAndLength(t *testing.T) {
	r, ok := parseRule(false, "/a/*/b$", 3)
	if !ok || !r.anchor || r.pattern != "/a/*/b" || r.length != 5 {
		t.Errorf("got %+v ok=%v", r, ok)
	}
	if _, ok := parseRule(true, "", 1); ok {
		t.Error("empty rule should be rejected")
	}
	if _, ok := parseRule(true, "noslash", 1); ok {
		t.Error("rule without leading slash should be rejected")
	}
}
