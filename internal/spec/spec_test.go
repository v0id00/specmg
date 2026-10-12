package spec

import (
	"bytes"
	"reflect"
	"testing"
)

func sample() *Spec {
	return &Spec{
		FM: Frontmatter{
			ID: "0001", Slug: "retry-budget", Title: "Retry budget",
			Status: StatusDraft, Owner: "alice",
			Created: "2026-10-12", Updated: "2026-10-12",
			Tags: []string{"backend", "reliability"},
		},
		Purpose: "Prose text that is preserved verbatim.",
		Reqs:    []Requirement{{Num: 1, Text: "The client SHALL retry at most 3 times."}},
		Links:   []Link{{Num: 1, Rel: "implements", Target: "0002-other-slug"}},
		Criteria: []Criterion{
			{Num: 1, Text: "`specmg list` exits 0."},
			{Num: 2, Text: "Already verified.", Done: true},
		},
		Questions: []Question{
			{Num: 1, Text: "Should backoff be exponential?", Blocking: true},
			{Num: 2, Text: "What is the max retry count?", Resolved: true},
		},
		DecisionLog: "| Date | Decision | Rationale |\n|------|----------|-----------|\n| 2026-10-12 | Max 3 retries | Simplicity |",
	}
}

func TestRenderParseRoundTrip(t *testing.T) {
	s := sample()
	got, err := Parse(Render(s))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(s.FM, got.FM) {
		t.Fatalf("frontmatter:\n got %+v\nwant %+v", got.FM, s.FM)
	}
	if !reflect.DeepEqual(s, got) {
		t.Fatalf("spec:\n got %+v\nwant %+v", got, s)
	}
	// second render must be byte-identical (canonical form)
	if a, b := Render(s), Render(got); !bytes.Equal(a, b) {
		t.Fatalf("render not deterministic:\n%s\n---\n%s", a, b)
	}
}

func TestParseErrors(t *testing.T) {
	if _, err := Parse([]byte("no frontmatter")); err == nil {
		t.Fatal("want error for missing frontmatter")
	}
	if _, err := Parse([]byte("---\nid: x\nslug: y\n---\n\n## Purpose\n\np\n")); err == nil {
		t.Fatal("want error for missing/invalid status")
	}
	if _, err := Parse([]byte("---\nid: x\nslug: y\nstatus: bogus\n---\n\n## Purpose\n\np\n")); err == nil {
		t.Fatal("want error for invalid status")
	}
}

func TestParseItemID(t *testing.T) {
	cases := []struct {
		in  string
		typ string
		num int
		ok  bool
	}{
		{"REQ-1", "REQ", 1, true}, {"ac-12", "AC", 12, true},
		{"Q-3", "Q", 3, true}, {"L-7", "L", 7, true},
		{"BOGUS", "", 0, false}, {"REQ-", "", 0, false}, {"REQ-x", "", 0, false},
	}
	for _, c := range cases {
		typ, num, ok := ParseItemID(c.in)
		if typ != c.typ || num != c.num || ok != c.ok {
			t.Errorf("ParseItemID(%q) = (%q,%d,%v)", c.in, typ, num, ok)
		}
	}
}

func TestNextNums(t *testing.T) {
	s := sample()
	if got := NextReqNum(s.Reqs); got != 2 {
		t.Errorf("NextReqNum = %d", got)
	}
	if got := NextACNum(s.Criteria); got != 3 {
		t.Errorf("NextACNum = %d", got)
	}
	if got := NextQNum(s.Questions); got != 3 {
		t.Errorf("NextQNum = %d", got)
	}
	if got := NextLNum(s.Links); got != 2 {
		t.Errorf("NextLNum = %d", got)
	}
}
