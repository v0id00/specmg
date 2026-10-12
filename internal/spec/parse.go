package spec

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	reCheck = regexp.MustCompile(`^[-*] \[( |x|X)\] \*\*AC-(\d+)\*\*: ?(.*)$`)
	reLink  = regexp.MustCompile(`^[-*] \*\*L-(\d+)\*\*: ?(\w+) -> (\S+)$`)
	reItem  = regexp.MustCompile(`^[-*] \*\*(REQ|AC|Q|L)-(\d+)\*\*(?: \[(blocking|resolved)\])?: ?(.*)$`)
	reHead  = regexp.MustCompile(`^## (.+)$`)
)

func Parse(data []byte) (*Spec, error) {
	txt := strings.ReplaceAll(string(data), "\r\n", "\n")
	if !strings.HasPrefix(txt, "---\n") {
		return nil, errors.New("missing YAML frontmatter")
	}
	end := strings.Index(txt[4:], "\n---")
	if end < 0 {
		return nil, errors.New("unterminated frontmatter")
	}
	var fm Frontmatter
	if err := yaml.Unmarshal([]byte(txt[4:4+end+1]), &fm); err != nil {
		return nil, fmt.Errorf("frontmatter: %w", err)
	}
	if fm.ID == "" || fm.Slug == "" || fm.Status == "" {
		return nil, errors.New("frontmatter requires id, slug, status")
	}
	if !ValidStatus(fm.Status) {
		return nil, fmt.Errorf("invalid status %q", fm.Status)
	}
	s := &Spec{FM: fm}
	rest := txt[4+end+5:] // after closing "---\n"

	section, lineBuf := "", []string{}
	flush := func() {
		body := strings.TrimSpace(strings.Join(lineBuf, "\n"))
		switch section {
		case "Purpose":
			s.Purpose = body
		case "Decision Log":
			s.DecisionLog = body
		case "":
		default:
			if body != "" {
				if s.Extra != "" {
					s.Extra += "\n\n"
				}
				s.Extra += "## " + section + "\n\n" + body
			}
		}
		lineBuf = nil
	}
	for _, line := range strings.Split(rest, "\n") {
		if m := reHead.FindStringSubmatch(line); m != nil {
			flush()
			section = strings.TrimSpace(m[1])
			continue
		}
		if strings.TrimSpace(line) == "" && len(lineBuf) == 0 {
			continue
		}
		switch section {
		case "Requirements":
			if m := reItem.FindStringSubmatch(line); m != nil && m[1] == "REQ" {
				n, _ := strconv.Atoi(m[2])
				s.Reqs = append(s.Reqs, Requirement{Num: n, Text: m[4]})
				continue
			}
		case "Acceptance Criteria":
			if m := reCheck.FindStringSubmatch(line); m != nil {
				n, _ := strconv.Atoi(m[2])
				s.Criteria = append(s.Criteria, Criterion{Num: n, Text: m[3], Done: m[1] == "x" || m[1] == "X"})
				continue
			}
		case "Links":
			if m := reLink.FindStringSubmatch(line); m != nil {
				n, _ := strconv.Atoi(m[1])
				s.Links = append(s.Links, Link{Num: n, Rel: m[2], Target: m[3]})
				continue
			}
		case "Open Questions":
			if m := reItem.FindStringSubmatch(line); m != nil && m[1] == "Q" {
				n, _ := strconv.Atoi(m[2])
				s.Questions = append(s.Questions, Question{
					Num: n, Text: m[4],
					Blocking: m[3] == "blocking", Resolved: m[3] == "resolved",
				})
				continue
			}
		}
		lineBuf = append(lineBuf, line)
	}
	flush()
	return s, nil
}
