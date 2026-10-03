package robots

import (
	"reflect"
	"testing"
)

func TestMatchPattern(t *testing.T) {
	cases := []struct {
		pat, path string
		want      bool
	}{
		{"/a", "/a", true},
		{"/a", "/ab", true},   // suffix allowed without $
		{"/a$", "/ab", false}, // anchored
		{"/a$", "/a", true},
		{"/a", "/A", false}, // case-sensitive
		{"/*/x", "/q/w/x", true},
		{"*/b$", "/x/b/b", true},
		{"/a*", "/a", true},
		{"/a*b", "/axyb", true},
		{"/a*b", "/abx", true},
		{"/", "/anything", true},
		{"/x", "/y", false},
	}
	for _, c := range cases {
		if got := matchPattern(c.pat, c.path); got != c.want {
			t.Errorf("matchPattern(%q,%q)=%v want %v", c.pat, c.path, got, c.want)
		}
	}
}

func decide(t *testing.T, doc, crawler, path string) Decision {
	t.Helper()
	p, err := Parse(doc)
	if err != nil {
		t.Fatal(err)
	}
	return p.Decide(crawler, path)
}

func TestDecideLongestMatchAndAllowTie(t *testing.T) {
	doc := `User-agent: bot
Disallow: /a
Allow: /a/b
Disallow: /a/b
`
	d := decide(t, doc, "bot", "/a/b/c")
	if !d.Allowed || d.RuleLine == nil || *d.RuleLine != 3 {
		t.Errorf("got %+v, want allowed via line 3 (Allow wins tie)", d)
	}
	d = decide(t, doc, "bot", "/a/c")
	if d.Allowed || *d.RuleLine != 2 {
		t.Errorf("got %+v, want disallowed via line 2", d)
	}
}

func TestDecideSameKindTiePicksSmallestLine(t *testing.T) {
	doc := `User-agent: bot
Disallow: /x$
Disallow: /x$
`
	d := decide(t, doc, "bot", "/x")
	if d.Allowed || *d.RuleLine != 2 {
		t.Errorf("got %+v, want line 2", d)
	}
}

func TestDecideGroupSelection(t *testing.T) {
	doc := `User-agent: Other
Disallow: /o
User-agent: BOT
Disallow: /n
User-agent: *
Disallow: /s
`
	d := decide(t, doc, "bot", "/n") // case-insensitive name match
	if d.Allowed || *d.RuleLine != 4 {
		t.Errorf("got %+v, want disallow line 4", d)
	}
	if !reflect.DeepEqual(d.Groups, []int{3}) {
		t.Errorf("groups=%v", d.Groups)
	}
	d = decide(t, doc, "unknown", "/s") // falls back to *
	if d.Allowed || *d.RuleLine != 6 || !reflect.DeepEqual(d.Groups, []int{5}) {
		t.Errorf("got %+v", d)
	}
	d = decide(t, doc, "unknown", "/n") // * group does not cover /n
	if !d.Allowed || d.RuleLine != nil {
		t.Errorf("got %+v, want default allow", d)
	}
}

func TestDecideMergesMultipleGroups(t *testing.T) {
	doc := `User-agent: bot
Disallow: /a
User-agent: bot
Disallow: /b
`
	if d := decide(t, doc, "bot", "/b"); d.Allowed || *d.RuleLine != 4 {
		t.Errorf("got %+v", d)
	}
	if d := decide(t, doc, "bot", "/a"); d.Allowed {
		t.Errorf("got %+v", d)
	}
}

func TestDecideNoGroupsAndRobotsTxt(t *testing.T) {
	d := decide(t, "", "bot", "/anything")
	if !d.Allowed || d.RuleLine != nil || len(d.Groups) != 0 {
		t.Errorf("got %+v", d)
	}
	doc := "User-agent: *\nDisallow: /\n"
	d = decide(t, doc, "bot", "/robots.txt")
	if !d.Allowed || d.RuleLine != nil {
		t.Errorf("robots.txt must always be allowed, got %+v", d)
	}
	d = decide(t, doc, "bot", "/x")
	if d.Allowed {
		t.Errorf("got %+v, want disallow", d)
	}
}

func TestPatternLenIgnoresStarAndDollar(t *testing.T) {
	doc := `User-agent: bot
Disallow: /*/tmp
Allow: /a*tmp$
`
	// both length 5 ('/','/','t','m','p' vs '/','a','t','m','p'); Allow wins tie
	d := decide(t, doc, "bot", "/a/tmp")
	if !d.Allowed || *d.RuleLine != 3 {
		t.Errorf("got %+v", d)
	}
}
