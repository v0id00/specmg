package spec

import (
	"strconv"
	"strings"
)

const (
	StatusDraft       = "draft"
	StatusInReview    = "in_review"
	StatusApproved    = "approved"
	StatusImplemented = "implemented"
	StatusVerified    = "verified"
	StatusSuperseded  = "superseded"
	StatusRejected    = "rejected"
)

var Statuses = []string{
	StatusDraft, StatusInReview, StatusApproved, StatusImplemented,
	StatusVerified, StatusSuperseded, StatusRejected,
}

func ValidStatus(s string) bool {
	for _, v := range Statuses {
		if v == s {
			return true
		}
	}
	return false
}

var linkRels = map[string]bool{
	"implements": true, "relates": true, "blocks": true,
	"depends_on": true, "supersedes": true,
}

func ValidRel(r string) bool { return linkRels[r] }

// ParseItemID parses "REQ-1" -> ("REQ", 1, true).
func ParseItemID(s string) (typ string, num int, ok bool) {
	parts := strings.SplitN(strings.ToUpper(s), "-", 2)
	if len(parts) != 2 {
		return "", 0, false
	}
	typ = parts[0]
	switch typ {
	case "REQ", "AC", "Q", "L":
	default:
		return "", 0, false
	}
	n, err := strconv.Atoi(parts[1])
	if err != nil || n < 1 {
		return "", 0, false
	}
	return typ, n, true
}

type Frontmatter struct {
	ID             string   `yaml:"id"`
	Slug           string   `yaml:"slug"`
	Title          string   `yaml:"title"`
	Status         string   `yaml:"status"`
	Owner          string   `yaml:"owner"`
	Created        string   `yaml:"created"`
	Updated        string   `yaml:"updated"`
	Tags           []string `yaml:"tags"`
	SupersededBy   string   `yaml:"superseded_by,omitempty"`
	RejectedReason string   `yaml:"rejected_reason,omitempty"`
}

type Requirement struct {
	Num  int
	Text string
}
type Criterion struct {
	Num  int
	Text string
	Done bool
}
type Question struct {
	Num      int
	Text     string
	Blocking bool
	Resolved bool
}
type Link struct {
	Num    int
	Rel    string
	Target string
}

type Spec struct {
	FM          Frontmatter
	Purpose     string // raw markdown under "## Purpose"
	Extra       string // unknown sections, preserved at the end
	Reqs        []Requirement
	Links       []Link
	Criteria    []Criterion
	Questions   []Question
	DecisionLog string // raw under "## Decision Log"
}

// Filename returns "NNNN-slug.md" per the format contract.
func (s *Spec) Filename() string { return s.FM.ID + "-" + s.FM.Slug + ".md" }

func maxNum(n int, nums ...int) int {
	for _, v := range nums {
		if v > n {
			n = v
		}
	}
	return n
}

func NextReqNum(items []Requirement) int {
	n := 0
	for _, it := range items {
		n = maxNum(n, it.Num)
	}
	return n + 1
}

func NextACNum(items []Criterion) int {
	n := 0
	for _, it := range items {
		n = maxNum(n, it.Num)
	}
	return n + 1
}

func NextQNum(items []Question) int {
	n := 0
	for _, it := range items {
		n = maxNum(n, it.Num)
	}
	return n + 1
}

func NextLNum(items []Link) int {
	n := 0
	for _, it := range items {
		n = maxNum(n, it.Num)
	}
	return n + 1
}
