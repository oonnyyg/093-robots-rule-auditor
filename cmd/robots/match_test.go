package main

import "testing"

func TestMatchPattern(t *testing.T) {
	cases := []struct {
		pattern, path string
		anchor        bool
		want          bool
	}{
		{"/a", "/a/b", false, true},  // suffix allowed without $
		{"/a", "/a/b", true, false},  // $ requires full match
		{"/a/b", "/a/b", true, true}, //
		{"/a/*/c", "/a/x/y/c", false, true},
		{"/*", "/anything", false, true},
		{"/a", "/A", false, false}, // case-sensitive
		{"/a?x=1", "/a?x=1&y=2", false, true},
		{"/b", "/a", false, false},
		{"/a$c", "/a$c", false, true}, // parser rejects mid-pattern '$'; here it is a literal
	}
	for _, c := range cases {
		if got := matchPattern(c.pattern, c.path, c.anchor); got != c.want {
			t.Errorf("matchPattern(%q, %q, %v) = %v, want %v",
				c.pattern, c.path, c.anchor, got, c.want)
		}
	}
}

func eval(t *testing.T, rules, crawler, path string) result {
	t.Helper()
	return evaluate(parseRobots(rules), query{Crawler: crawler, Path: path})
}

func TestEvaluateLongestMatchAndTies(t *testing.T) {
	rules := "" +
		"User-agent: bot\n" + // group starts line 1
		"Disallow: /a\n" + // line 2, length 2
		"Allow: /a/b\n" + // line 3, length 4 wins
		"Disallow: /a/b\n" // line 4, tie on length: Allow wins
	res := eval(t, rules, "bot", "/a/b/c")
	if !res.Allowed || res.RuleLine == nil || *res.RuleLine != 3 {
		t.Errorf("got %+v, want allowed via line 3", res)
	}
	if len(res.Groups) != 1 || res.Groups[0] != 1 {
		t.Errorf("groups = %v, want [1]", res.Groups)
	}
}

func TestEvaluateTieBreaksByLine(t *testing.T) {
	rules := "" +
		"User-agent: bot\n" +
		"Disallow: /x\n" +
		"User-agent: bot\n" + // second group, same agent: merged
		"Disallow: /x\n"
	res := eval(t, rules, "BOT", "/x") // agent match is case-insensitive
	if res.Allowed || res.RuleLine == nil || *res.RuleLine != 2 {
		t.Errorf("got %+v, want disallowed via line 2", res)
	}
	if len(res.Groups) != 2 || res.Groups[0] != 1 || res.Groups[1] != 3 {
		t.Errorf("groups = %v, want [1 3]", res.Groups)
	}
}

func TestEvaluateWildcardFallbackAndDefault(t *testing.T) {
	rules := "" +
		"User-agent: other\n" +
		"Disallow: /\n" +
		"User-agent: *\n" +
		"Disallow: /private\n"
	res := eval(t, rules, "bot", "/private/x")
	if res.Allowed || res.RuleLine == nil || *res.RuleLine != 4 {
		t.Errorf("got %+v, want disallowed via line 4", res)
	}
	res = eval(t, rules, "bot", "/public")
	if !res.Allowed || res.RuleLine != nil {
		t.Errorf("got %+v, want allowed with nil rule line", res)
	}
	res = eval(t, rules, "other", "/anything")
	if res.Allowed || res.RuleLine == nil || *res.RuleLine != 2 {
		t.Errorf("got %+v, want disallowed via line 2", res)
	}
	res = eval(t, "", "bot", "/anything")
	if !res.Allowed || res.RuleLine != nil || len(res.Groups) != 0 {
		t.Errorf("got %+v, want allowed with no groups", res)
	}
}

func TestEvaluateRobotsTxtAlwaysAllowed(t *testing.T) {
	rules := "User-agent: *\nDisallow: /\n"
	res := eval(t, rules, "bot", "/robots.txt")
	if !res.Allowed || res.RuleLine != nil {
		t.Errorf("got %+v, want allowed with nil rule line", res)
	}
}

func TestEvaluateDollarAndStar(t *testing.T) {
	rules := "" +
		"User-agent: bot\n" +
		"Disallow: /x$\n" + // line 2
		"Allow: /*.php$\n" // line 3
	res := eval(t, rules, "bot", "/x")
	if res.Allowed || res.RuleLine == nil || *res.RuleLine != 2 {
		t.Errorf("got %+v, want disallowed via line 2", res)
	}
	res = eval(t, rules, "bot", "/x/y")
	if !res.Allowed {
		t.Errorf("got %+v, want allowed: /x$ must not match /x/y", res)
	}
	res = eval(t, rules, "bot", "/a.php")
	if !res.Allowed || res.RuleLine == nil || *res.RuleLine != 3 {
		t.Errorf("got %+v, want allowed via line 3", res)
	}
	res = eval(t, rules, "bot", "/a.php?x=1")
	if !res.Allowed || res.RuleLine != nil {
		t.Errorf("got %+v, want allowed with nil rule: /*.php$ must not match query", res)
	}
}
