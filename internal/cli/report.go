package cli

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/spf13/cobra"
	"github.com/v0id00/specmg/internal/spec"
	"github.com/v0id00/specmg/internal/store"
)

var (
	reportFormat string
	reportOut    string
	reportTitle  string
)

var reportCmd = &cobra.Command{
	Use:   "report [<ref>...]",
	Short: "Generate a HTML or PDF report for one or more specs (default: all)",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := resolveRoot()
		if err != nil {
			return err
		}
		var specs []*spec.Spec
		if len(args) == 0 {
			if specs, err = store.List(root, flagDir); err != nil {
				return err
			}
		} else {
			for _, ref := range args {
				s, err := store.Load(root, flagDir, ref)
				if err != nil {
					return err
				}
				specs = append(specs, s)
			}
		}
		if len(specs) == 0 {
			return fmt.Errorf("no specs to report")
		}
		out := reportOut
		if out == "" {
			out = "specmg-report." + reportFormat
		}
		if flagDryRun {
			fmt.Fprintf(cmd.OutOrStdout(), "would write %s (%d specs)\n", out, len(specs))
			return nil
		}
		switch reportFormat {
		case "html":
			err = writeHTMLReport(out, reportTitle, specs)
		case "pdf":
			err = writePDFReport(out, reportTitle, specs)
		default:
			return usageErr{fmt.Errorf("invalid --format %q (html or pdf)", reportFormat)}
		}
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "wrote %s (%d specs)\n", out, len(specs))
		return nil
	},
}

// --- HTML ---

var tmplFuncs = template.FuncMap{
	"statusColor": func(s string) string {
		if c, ok := statusColors[s]; ok {
			return c
		}
		return "#777"
	},
	"join": func(xs []string) string { return strings.Join(xs, ", ") },
}

const htmlBody = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>{{.Title}}</title>
<style>
body{font-family:system-ui,sans-serif;margin:2rem auto;max-width:60rem;color:#1a1a1a;line-height:1.5}
h1{border-bottom:2px solid #ddd;padding-bottom:.3rem}
article{border:1px solid #ddd;border-radius:8px;padding:1rem 1.4rem;margin:1.4rem 0}
h2{margin:.2rem 0 1rem;font-size:1.25rem}
.meta{color:#666;font-size:.9rem}
.badge{display:inline-block;padding:.1rem .55rem;border-radius:99px;color:#fff;font-size:.8rem}
ul{padding-left:1.2rem}li{margin:.25rem 0}
code{background:#f2f2f2;padding:.05rem .3rem;border-radius:4px}
</style>
</head>
<body>
<h1>{{.Title}}</h1>
<p>{{len .Specs}} spec(s) · generated {{.Date}} by specmg</p>
{{range .Specs}}
<article>
<h2><code>{{.FM.ID}}-{{.FM.Slug}}</code> {{.FM.Title}}
  <span class="badge" style="background:{{statusColor .FM.Status}}">{{.FM.Status}}</span></h2>
<p class="meta">{{if .FM.Owner}}owner: {{.FM.Owner}} · {{end}}created {{.FM.Created}} · updated {{.FM.Updated}}{{if .FM.Tags}} · tags: {{join .FM.Tags}}{{end}}</p>
{{if .Purpose}}<h3>Purpose</h3><p>{{.Purpose}}</p>{{end}}
{{if .Reqs}}<h3>Requirements</h3><ul>{{range .Reqs}}<li><strong>REQ-{{.Num}}:</strong> {{.Text}}</li>{{end}}</ul>{{end}}
{{if .Criteria}}<h3>Acceptance Criteria</h3><ul>{{range .Criteria}}<li>{{if .Done}}☑{{else}}☐{{end}} <strong>AC-{{.Num}}:</strong> {{.Text}}</li>{{end}}</ul>{{end}}
{{if .Questions}}<h3>Open Questions</h3><ul>{{range .Questions}}<li>{{if .Resolved}}✅{{else if .Blocking}}❗{{else}}❓{{end}} <strong>Q-{{.Num}}:</strong> {{.Text}}</li>{{end}}</ul>{{end}}
{{if .Links}}<h3>Links</h3><ul>{{range .Links}}<li><strong>L-{{.Num}}:</strong> {{.Rel}} → {{.Target}}</li>{{end}}</ul>{{end}}
{{if .DecisionLog}}<h3>Decision Log</h3><pre>{{.DecisionLog}}</pre>{{end}}
</article>
{{end}}
</body></html>`

var statusColors = map[string]string{
	spec.StatusDraft:       "#e0b400",
	spec.StatusInReview:    "#2f7fdb",
	spec.StatusApproved:    "#2e9e44",
	spec.StatusImplemented: "#7b4fd3",
	spec.StatusVerified:    "#1f9d55",
	spec.StatusSuperseded:  "#555555",
	spec.StatusRejected:    "#dd3333",
}

func writeHTMLReport(path, title string, specs []*spec.Spec) error {
	t := template.Must(template.New("report").Funcs(tmplFuncs).Parse(htmlBody))
	data := struct {
		Title string
		Date  string
		Specs []*spec.Spec
	}{title, today(), specs}
	var b strings.Builder
	if err := t.Execute(&b, data); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// --- PDF ---

// fontCandidates are UTF-8 capable TTF fonts in common locations.
var fontCandidates = []string{
	"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
	"/usr/share/fonts/TTF/DejaVuSans.ttf",
	"/usr/share/fonts/dejavu/DejaVuSans.ttf",
	"/System/Library/Fonts/Supplemental/Arial Unicode.ttf",
	"C:/Windows/Fonts/arial.ttf",
}

func findPDFFont() string {
	for _, p := range fontCandidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func writePDFReport(path, title string, specs []*spec.Spec) error {
	p := fpdf.New("P", "mm", "A4", "")
	p.SetMargins(18, 18, 18)
	p.AddPage()
	if font := findPDFFont(); font != "" {
		p.SetFontLocation(filepath.Dir(font))
		p.AddUTF8Font("body", "", filepath.Base(font))
	} else {
		p.SetFont("helvetica", "", 11) // ASCII-only fallback
	}
	p.SetTitle(title, true)
	p.SetFont("body", "", 16)
	p.Cell(0, 10, title)
	p.Ln(8)
	p.SetFont("body", "", 9)
	p.SetTextColor(110, 110, 110)
	p.Cell(0, 6, fmt.Sprintf("%d spec(s) - generated %s by specmg", len(specs), today()))
	p.Ln(10)
	p.SetTextColor(0, 0, 0)

	for _, s := range specs {
		p.SetFont("body", "", 13)
		p.MultiCell(0, 7, fmt.Sprintf("%s-%s  %s", s.FM.ID, s.FM.Slug, s.FM.Title), "", "L", false)
		// status chip
		c := statusColors[s.FM.Status]
		var r, g, b int
		if c != "" {
			fmt.Sscanf(c, "#%02x%02x%02x", &r, &g, &b)
		} else {
			r, g, b = 0x77, 0x77, 0x77
		}
		p.SetFillColor(r, g, b)
		p.SetTextColor(255, 255, 255)
		p.SetFont("body", "", 9)
		p.CellFormat(38, 6, " "+s.FM.Status+" ", "", 1, "L", true, 0, "")
		p.SetTextColor(0, 0, 0)
		if s.FM.Owner != "" {
			p.SetFont("body", "", 9)
			p.CellFormat(0, 6, "owner: "+s.FM.Owner, "", 1, "L", false, 0, "")
		}
		pdfSection(p, "Purpose", s.Purpose)
		var lines []string
		for _, req := range s.Reqs {
			lines = append(lines, fmt.Sprintf("REQ-%d: %s", req.Num, req.Text))
		}
		pdfSection(p, "Requirements", strings.Join(lines, "\n"))
		lines = nil
		for _, c := range s.Criteria {
			mark := "[ ]"
			if c.Done {
				mark = "[x]"
			}
			lines = append(lines, fmt.Sprintf("%s AC-%d: %s", mark, c.Num, c.Text))
		}
		pdfSection(p, "Acceptance Criteria", strings.Join(lines, "\n"))
		lines = nil
		for _, q := range s.Questions {
			mark := "[open]"
			if q.Resolved {
				mark = "[resolved]"
			} else if q.Blocking {
				mark = "[blocking]"
			}
			lines = append(lines, fmt.Sprintf("Q-%d %s: %s", q.Num, mark, q.Text))
		}
		pdfSection(p, "Open Questions", strings.Join(lines, "\n"))
		lines = nil
		for _, l := range s.Links {
			lines = append(lines, fmt.Sprintf("L-%d: %s -> %s", l.Num, l.Rel, l.Target))
		}
		pdfSection(p, "Links", strings.Join(lines, "\n"))
		pdfSection(p, "Decision Log", strings.ReplaceAll(s.DecisionLog, "|", " "))
		p.Ln(4)
	}
	return p.OutputFileAndClose(path)
}

func pdfSection(p *fpdf.Fpdf, heading, body string) {
	if strings.TrimSpace(body) == "" {
		return
	}
	p.SetFont("body", "", 11)
	p.Ln(3)
	p.MultiCell(0, 6, heading, "", "L", false)
	p.SetFont("body", "", 9)
	p.SetTextColor(60, 60, 60)
	p.MultiCell(0, 5, body, "", "L", false)
	p.SetTextColor(0, 0, 0)
}

func init() {
	reportCmd.Flags().StringVarP(&reportFormat, "format", "f", "html", "output format: html or pdf")
	reportCmd.Flags().StringVarP(&reportOut, "out", "o", "", "output file (default specmg-report.<format>)")
	reportCmd.Flags().StringVar(&reportTitle, "title", "Spec Report", "report title")
	rootCmd.AddCommand(reportCmd)
}
