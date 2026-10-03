package robots

import (
	"reflect"
	"testing"
)

func TestValidCrawler(t *testing.T) {
	for _, ok := range []string{"Googlebot", "my_bot", "a-b", "B"} {
		if !ValidCrawler(ok) {
			t.Errorf("ValidCrawler(%q) = false, want true", ok)
		}
	}
	for _, bad := range []string{"", "bot1", "a b", "a/b", "*", "a.b"} {
		if ValidCrawler(bad) {
			t.Errorf("ValidCrawler(%q) = true, want false", bad)
		}
	}
}

func TestValidPath(t *testing.T) {
	for _, ok := range []string{"/", "/a/b?x=1&y=2", "/robots.txt"} {
		if !ValidPath(ok) {
			t.Errorf("ValidPath(%q) = false, want true", ok)
		}
	}
	for _, bad := range []string{"", "a/", "/a b", "/a%20b", "/a#frag", "/é"} {
		if ValidPath(bad) {
			t.Errorf("ValidPath(%q) = true, want false", bad)
		}
	}
}

func TestParseGrouping(t *testing.T) {
	p, err := Parse(`User-agent: a
User-agent: b
Disallow: /x

User-agent: c
Allow: /y
User-agent: d
Disallow: /z
`)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Groups) != 3 {
		t.Fatalf("got %d groups, want 3", len(p.Groups))
	}
	if got := p.Groups[0].Agents; !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Errorf("group0 agents = %v", got)
	}
	if p.Groups[0].Start != 1 || p.Groups[1].Start != 5 || p.Groups[2].Start != 7 {
		t.Errorf("starts = %d,%d,%d", p.Groups[0].Start, p.Groups[1].Start, p.Groups[2].Start)
	}
}

func TestParseCommentsCaseAndUnknown(t *testing.T) {
	p, err := Parse("USER-AGENT: Bot # trailing\nDISALLOW: /x # c\nSitemap: http://x\nCrawl-delay: 5\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Groups) != 1 || len(p.Groups[0].Rules) != 1 {
		t.Fatalf("groups=%d rules=%d", len(p.Groups), len(p.Groups[0].Rules))
	}
	if p.Groups[0].Rules[0].Pattern != "/x" {
		t.Errorf("pattern = %q", p.Groups[0].Rules[0].Pattern)
	}
}

func TestParseIgnoresOutOfGroupAndEmptyRules(t *testing.T) {
	p, err := Parse("Disallow: /orphan\nUser-agent: a\nDisallow:\nAllow: \n")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Groups) != 1 || len(p.Groups[0].Rules) != 0 {
		t.Fatalf("groups=%d rules=%d", len(p.Groups), len(p.Groups[0].Rules))
	}
}

func TestParseInvalid(t *testing.T) {
	for _, doc := range []string{
		"User-agent: bad name\n",
		"User-agent: a\nDisallow: x\n",
		"User-agent: a\nDisallow: /a$b\n",
	} {
		if _, err := Parse(doc); err == nil {
			t.Errorf("Parse(%q) succeeded, want error", doc)
		}
	}
}
