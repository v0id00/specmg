package spec

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Render deterministically regenerates the full file from the model.
func Render(s *Spec) []byte {
	var b bytes.Buffer
	b.WriteString("---\n")
	yaml.NewEncoder(&b).Encode(s.FM)
	b.WriteString("---\n\n")

	section := func(name, body string) {
		b.WriteString("## " + name + "\n\n")
		b.WriteString(strings.TrimSpace(body))
		b.WriteString("\n\n")
	}
	section("Purpose", s.Purpose)
	b.WriteString("## Requirements\n\n")
	for _, r := range s.Reqs {
		fmt.Fprintf(&b, "- **REQ-%d**: %s\n", r.Num, r.Text)
	}
	b.WriteString("\n## Links\n\n")
	for _, l := range s.Links {
		fmt.Fprintf(&b, "- **L-%d**: %s -> %s\n", l.Num, l.Rel, l.Target)
	}
	b.WriteString("\n## Acceptance Criteria\n\n")
	for _, c := range s.Criteria {
		mark := " "
		if c.Done {
			mark = "x"
		}
		fmt.Fprintf(&b, "- [%s] **AC-%d**: %s\n", mark, c.Num, c.Text)
	}
	b.WriteString("\n## Open Questions\n\n")
	for _, q := range s.Questions {
		tag := ""
		switch {
		case q.Resolved:
			tag = " [resolved]"
		case q.Blocking:
			tag = " [blocking]"
		}
		fmt.Fprintf(&b, "- **Q-%d**%s: %s\n", q.Num, tag, q.Text)
	}
	b.WriteString("\n## Decision Log\n\n")
	b.WriteString(strings.TrimSpace(s.DecisionLog))
	b.WriteString("\n")
	if strings.TrimSpace(s.Extra) != "" {
		b.WriteString("\n")
		b.WriteString(strings.TrimSpace(s.Extra))
		b.WriteString("\n")
	}
	return b.Bytes()
}
