package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/v0id00/specmg/internal/spec"
	"github.com/v0id00/specmg/internal/store"
)

var importDir string

var (
	headingRE = regexp.MustCompile(`^#\s+(\S+)\s+[—-]\s+(.+)$`) // "# 0001 — Title"
	statusRE  = regexp.MustCompile(`\*\*Status:\*\*\s*(\S+)`)
	ownerRE   = regexp.MustCompile(`\*\*Owner:\*\*\s*(\S+)`)
	h2RE      = regexp.MustCompile(`^## (.+)$`)
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
files (YAML frontmatter) are skipped. Run specmg lint afterwards.`,
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
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || e.Name() == "INDEX.md" {
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
	flush := func() {
		body := strings.TrimSpace(strings.Join(buf, "\n"))
		buf = nil
		switch section {
		case "Purpose":
			s.Purpose = body
		case "Requirements":
			for _, ln := range strings.Split(body, "\n") {
				if text := itemText(ln, "REQ"); text != "" {
					s.Reqs = append(s.Reqs, spec.Requirement{Num: spec.NextReqNum(s.Reqs), Text: text})
				}
			}
		case "Acceptance Criteria":
			for _, ln := range strings.Split(body, "\n") {
				text, done := acText(ln)
				if text != "" {
					s.Criteria = append(s.Criteria, spec.Criterion{Num: spec.NextACNum(s.Criteria), Text: text, Done: done})
				}
			}
		case "Open Questions":
			for _, ln := range strings.Split(body, "\n") {
				if text := itemText(ln, "Q"); text != "" {
					s.Questions = append(s.Questions, spec.Question{Num: spec.NextQNum(s.Questions), Text: text})
				}
			}
		case "Decision Log":
			s.DecisionLog = body
		default:
			if body != "" {
				if s.Extra != "" {
					s.Extra += "\n\n"
				}
				s.Extra += "## " + section + "\n\n" + body
			}
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
		if m := h2RE.FindStringSubmatch(ln); m != nil {
			flush()
			section = strings.TrimSpace(m[1])
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
	s.FM.Created = today()
	s.FM.Updated = s.FM.Created
	return s, nil
}

func itemText(line, typ string) string {
	ln := strings.TrimSpace(line)
	if !strings.HasPrefix(ln, "- ") {
		return ""
	}
	rest := ln[2:]
	// "- **REQ-1**: text" or "- REQ1: text" or "- **AC1**: text"
	for _, prefix := range []string{"**" + typ + "-", typ + "-"} {
		if strings.HasPrefix(rest, prefix) {
			rest = rest[len(prefix):]
			if i := strings.Index(rest, "**"); i >= 0 {
				rest = rest[i+2:]
			}
			return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(rest), ":"))
		}
	}
	return ""
}

func acText(line string) (string, bool) {
	ln := strings.TrimSpace(line)
	done := strings.HasPrefix(ln, "- [x]") || strings.HasPrefix(ln, "- [X]")
	if idx := strings.Index(ln, "]"); idx > 0 && (strings.HasPrefix(ln, "- [") || strings.HasPrefix(ln, "- [")) {
		ln = ln[idx+1:]
		rest := strings.TrimSpace(ln)
		// strip "**AC-1**:" style id
		if strings.HasPrefix(rest, "**") {
			if i := strings.Index(rest, "**:"); i > 0 {
				rest = rest[i+3:]
			} else if i := strings.Index(rest[2:], "**"); i >= 0 {
				rest = rest[2+i+2:]
			}
		}
		rest = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(rest), ":"))
		if rest != "" {
			return stripItemID(rest), done
		}
	}
	return "", false
}

// stripItemID removes a leading "AC-1:" / "AC1:" style id from criterion text.
func stripItemID(s string) string {
	if m := regexp.MustCompile(`^A[Cc]-?\d+\s*:\s*`).FindString(s); m != "" {
		return strings.TrimPrefix(s, m)
	}
	return s
}

func init() {
	importCmd.Flags().StringVar(&importDir, "dir", "specs", "directory (under root) to scan")
	rootCmd.AddCommand(importCmd)
}
