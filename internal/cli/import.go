package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/v0id00/specmg/internal/spec"
	"github.com/v0id00/specmg/internal/store"
)

var importDir string

var (
	headingRE    = regexp.MustCompile(`^#\s+(\S+)\s+[—-]\s+(.+)$`) // "# 0001 — Title"
	statusRE     = regexp.MustCompile(`\*\*Status:\*\*\s*(\S+)`)
	ownerRE      = regexp.MustCompile(`\*\*Owner:\*\*\s*(\S+)`)
	h2RE         = regexp.MustCompile(`^## (.+)$`)
	sectionNumRE = regexp.MustCompile(`^\d+\.\s*(.+)$`) // "7. Acceptance Criteria"
	lastUpdRE    = regexp.MustCompile(`^-\s+\*\*Last updated:\*\*\s*(\S+)`)
	dateRE       = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	acBoxRE      = regexp.MustCompile(`^\s*\|?\s*[-*]\s+\[( |x|X)\]\s*(.*)$`)
	bulletRE     = regexp.MustCompile(`^\s*\|?\s*[-*]\s+(.*)$`)
	numItemRE    = regexp.MustCompile(`^\s*(\d+)\.\s+(.*)$`)
	itemIDRE     = regexp.MustCompile(`^(REQ|AC|Q|OQ)-?(\d+)\*{0,2}\s*(?::|—|–|-)`)
)

// normalizeStatus maps sdd status words to specmg statuses.
func normalizeStatus(s string) string {
	m := map[string]string{
		"draft": spec.StatusDraft, "in_review": spec.StatusInReview,
		"review": spec.StatusInReview, "approved": spec.StatusApproved,
		"implemented": spec.StatusImplemented, "verified": spec.StatusVerified,
		"superseded": spec.StatusSuperseded, "rejected": spec.StatusRejected,
	}
	if v, ok := m[strings.ToLower(strings.TrimSuffix(s, "**"))]; ok {
		return v
	}
	return strings.ToLower(strings.TrimSuffix(s, "**"))
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Convert sdd-style spec files (heading + **Status:** front) into specmg format",
	Long: `Import converts legacy sdd-format files in place: files with "# NNNN — Title"
and a "**Status:**" line become specmg YAML-frontmatter files. Already-converted
files (YAML frontmatter) are skipped, and so are documents that are not specs
(specs/README.md, specs/INDEX.md).

The legacy body is kept, not rewritten: numbered section headings ("## 7.
Acceptance Criteria") are mapped to their specmg section, list items are lifted
into Requirements / Acceptance Criteria / Open Questions (wrapped lines are
joined, original item numbers are preserved), and everything else — out of
scope, architecture, references, the header block — survives verbatim under
Extra headings. Run specmg lint afterwards.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := resolveRoot()
		if err != nil {
			return err
		}
		dir := flagDir
		if importDir != "" {
			dir = importDir
		}
		scanDir := filepath.Join(root, dir)
		entries, err := os.ReadDir(scanDir)
		if err != nil {
			return err
		}
		converted, skipped := 0, 0
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if !spec.IsSpecFilename(e.Name()) {
				if e.Name() != "INDEX.md" && strings.HasSuffix(e.Name(), ".md") {
					verbosef(cmd, "skip %s: not a spec file (NNNN-slug.md)", e.Name())
				}
				continue
			}
			path := filepath.Join(scanDir, e.Name())
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if strings.HasPrefix(string(b), "---\n") {
				skipped++ // already specmg format
				continue
			}
			s, err := convertSDD(e.Name(), b)
			if err != nil {
				return fmt.Errorf("%s: %w", e.Name(), err)
			}
			if flagDryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "would convert %s -> %s (%s)\n", e.Name(), s.Filename(), s.FM.Status)
				converted++
				continue
			}
			if err := os.WriteFile(path, spec.Render(s), 0o644); err != nil {
				return err
			}
			if s.Filename() != e.Name() {
				os.Rename(path, filepath.Join(scanDir, s.Filename()))
			}
			fmt.Fprintf(cmd.OutOrStdout(), "converted %s -> %s\n", e.Name(), s.Filename())
			converted++
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%d converted, %d skipped\n", converted, skipped)
		return nil
	},
}

// convertSDD parses a legacy sdd spec (heading, **Status:** line, ## sections)
// into a specmg Spec.
func convertSDD(filename string, data []byte) (*spec.Spec, error) {
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	s := &spec.Spec{}
	var id, title string
	section := ""
	var buf []string

	// appendExtra keeps the given body under an Extra heading so the next
	// Parse reads it back. The header block that precedes the first heading
	// has no name of its own; it lands under "## Metadata".
	appendExtra := func(name, body string) {
		body = strings.TrimSpace(body)
		if body == "" {
			return
		}
		if name == "" {
			name = "Metadata"
		}
		if s.Extra != "" {
			s.Extra += "\n\n"
		}
		s.Extra += "## " + name + "\n\n" + body
	}
	appendExtraLines := func(name string, lines []string) {
		if len(lines) == 0 {
			return
		}
		appendExtra(name, strings.Join(lines, "\n"))
	}

	flush := func() {
		body := strings.TrimSpace(strings.Join(buf, "\n"))
		buf = nil
		if body == "" {
			return
		}
		switch section {
		case "Purpose":
			if s.Purpose == "" {
				s.Purpose = body
			} else {
				s.Purpose += "\n\n" + body
			}
		case "Requirements":
			items, left := readItems(body, func(l string) bool {
				_, _, ok := parseReqLine(l)
				return ok
			})
			for _, it := range items {
				if n, text, ok := parseReqLine(it); ok {
					s.Reqs = append(s.Reqs, spec.Requirement{Num: n, Text: text})
				}
			}
			appendExtraLines(section, left)
		case "Acceptance Criteria":
			items, left := readItems(body, func(l string) bool {
				_, _, _, ok := parseACLine(l)
				return ok
			})
			for _, it := range items {
				n, text, done, _ := parseACLine(it)
				if n == 0 {
					n = spec.NextACNum(s.Criteria)
				}
				s.Criteria = append(s.Criteria, spec.Criterion{Num: n, Text: text, Done: done})
			}
			appendExtraLines(section, left)
		case "Open Questions":
			items, left := readItems(body, func(l string) bool {
				if _, _, _, ok := parseQLine(l); ok {
					return true
				}
				return numItemRE.MatchString(l)
			})
			for _, it := range items {
				if n, text, resolved, ok := parseQLine(it); ok {
					s.Questions = append(s.Questions, spec.Question{Num: n, Text: text, Resolved: resolved})
					continue
				}
				if m := numItemRE.FindStringSubmatch(strings.TrimSpace(it)); m != nil {
					n, _ := strconv.Atoi(m[1])
					s.Questions = append(s.Questions, spec.Question{Num: n, Text: strings.TrimSpace(m[2])})
				}
			}
			appendExtraLines(section, left)
		case "Decision Log":
			if s.DecisionLog == "" {
				s.DecisionLog = body
			} else {
				s.DecisionLog += "\n\n" + body
			}
		default:
			appendExtra(section, body)
		}
	}

	for _, ln := range lines {
		if m := headingRE.FindStringSubmatch(ln); m != nil {
			flush()
			id, title = strings.TrimSpace(m[1]), strings.TrimSpace(m[2])
			continue
		}
		if m := statusRE.FindStringSubmatch(ln); m != nil {
			s.FM.Status = normalizeStatus(m[1])
			continue
		}
		if m := ownerRE.FindStringSubmatch(ln); m != nil {
			s.FM.Owner = strings.TrimSpace(m[1])
			continue
		}
		if m := lastUpdRE.FindStringSubmatch(ln); m != nil && dateRE.MatchString(m[1]) {
			// Kept in the body too: any later specmg write bumps "updated".
			s.FM.Updated = m[1]
		}
		if m := h2RE.FindStringSubmatch(ln); m != nil {
			flush()
			section = normalizeSection(m[1])
			continue
		}
		buf = append(buf, ln)
	}
	flush()

	if id == "" {
		// fall back to filename NNNN prefix
		id = strings.SplitN(strings.TrimSuffix(filename, ".md"), "-", 2)[0]
	}
	s.FM.ID = id
	s.FM.Slug = store.Slugify(title)
	if s.FM.Slug == "" {
		s.FM.Slug = store.Slugify(strings.TrimSuffix(filename, ".md"))
	}
	if title == "" {
		title = s.FM.Slug
	}
	s.FM.Title = title
	if s.FM.Status == "" {
		s.FM.Status = spec.StatusDraft
	}
	if !spec.ValidStatus(s.FM.Status) {
		return nil, fmt.Errorf("unknown status %q", s.FM.Status)
	}
	if s.Purpose == "" {
		s.Purpose = "TODO"
	}
	if s.DecisionLog == "" {
		s.DecisionLog = "| Date | Decision | Rationale |\n|------|----------|-----------|"
	}
	// The legacy files carry one date ("Last updated"); it is the only date
	// known, so it fills both frontmatter fields.
	if s.FM.Updated == "" {
		s.FM.Updated = today()
	}
	s.FM.Created = s.FM.Updated
	return s, nil
}

// normalizeSection strips a leading section number: "7. Acceptance Criteria"
// and "Acceptance Criteria" are the same section to specmg.
func normalizeSection(name string) string {
	name = strings.TrimSpace(name)
	if m := sectionNumRE.FindStringSubmatch(name); m != nil {
		return strings.TrimSpace(m[1])
	}
	return name
}

// readItems splits a legacy section body into logical list items. An item is a
// line accepted by start plus the indented lines wrapping it; a blank line ends
// the current item. Every other line comes back as leftover, in order, so no
// section content is dropped.
func readItems(body string, start func(string) bool) (items []string, leftovers []string) {
	var cur []string
	pendingBlank := false
	flush := func() {
		if len(cur) > 0 {
			items = append(items, strings.TrimSpace(strings.Join(cur, " ")))
			cur = nil
		}
	}
	for _, ln := range strings.Split(body, "\n") {
		if strings.TrimSpace(ln) == "" {
			flush()
			pendingBlank = true
			continue
		}
		if start(ln) {
			flush()
			pendingBlank = false
			cur = append(cur, strings.TrimSpace(ln))
			continue
		}
		if len(cur) > 0 && (strings.HasPrefix(ln, " ") || strings.HasPrefix(ln, "\t")) {
			cur = append(cur, strings.TrimSpace(ln))
			continue
		}
		flush()
		if pendingBlank && len(leftovers) > 0 {
			leftovers = append(leftovers, "")
		}
		pendingBlank = false
		leftovers = append(leftovers, ln)
	}
	flush()
	return items, leftovers
}

// splitItemID removes a leading "AC1 — ", "AC-2: " or "**REQ-3**: " id from a
// legacy bullet, returning the item type, its number and the remaining text.
func splitItemID(s string) (typ string, num int, rest string, ok bool) {
	t := strings.TrimSpace(s)
	t = strings.TrimPrefix(t, "**")
	m := itemIDRE.FindStringSubmatch(t)
	if m == nil {
		return "", 0, "", false
	}
	n, err := strconv.Atoi(m[2])
	if err != nil {
		return "", 0, "", false
	}
	return strings.ToLower(m[1]), n, strings.TrimSpace(t[len(m[0]):]), true
}

// parseReqLine reads "- **REQ-1**: text" (or "- REQ1: text").
func parseReqLine(line string) (num int, text string, ok bool) {
	m := bulletRE.FindStringSubmatch(strings.TrimSpace(line))
	if m == nil {
		return 0, "", false
	}
	typ, n, body, matched := splitItemID(m[1])
	if !matched || typ != "req" || body == "" {
		return 0, "", false
	}
	return n, body, true
}

// parseACLine reads "- [x] AC1 — text": checkbox state, original number, text.
func parseACLine(line string) (num int, text string, done bool, ok bool) {
	m := acBoxRE.FindStringSubmatch(line)
	if m == nil {
		return 0, "", false, false
	}
	done = m[1] == "x" || m[1] == "X"
	rest := strings.TrimSpace(m[2])
	if rest == "" {
		return 0, "", false, false
	}
	if typ, n, body, matched := splitItemID(rest); matched && typ == "ac" && body != "" {
		return n, body, done, true
	}
	return 0, rest, done, true
}

// parseQLine reads "- OQ1 — question" and the resolved form
// "- ~~OQ1 — question~~ **Resolved (date): answer**".
func parseQLine(line string) (num int, text string, resolved bool, ok bool) {
	m := bulletRE.FindStringSubmatch(strings.TrimSpace(line))
	if m == nil {
		return 0, "", false, false
	}
	rest := strings.TrimSpace(m[1])
	resolved = strings.Contains(rest, "~~") && strings.Contains(rest, "Resolved")
	rest = strings.TrimPrefix(rest, "~~")
	typ, n, body, matched := splitItemID(rest)
	if !matched || (typ != "q" && typ != "oq") || body == "" {
		return 0, "", false, false
	}
	if resolved {
		// "question~~ **Resolved …**" -> "question **Resolved …**"
		if i := strings.Index(body, "~~"); i >= 0 {
			body = strings.TrimSpace(strings.TrimSpace(body[:i]) + " " + strings.TrimSpace(body[i+2:]))
		}
	}
	return n, body, resolved, true
}

func verbosef(cmd *cobra.Command, format string, args ...any) {
	if flagVerbose {
		fmt.Fprintf(cmd.OutOrStdout(), format+"\n", args...)
	}
}

func init() {
	importCmd.Flags().StringVar(&importDir, "dir", "specs", "directory (under root) to scan")
	rootCmd.AddCommand(importCmd)
}
