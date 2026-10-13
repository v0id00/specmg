package cli

import (
	"os"
	"strings"
	"testing"
)

// legacySpec is a sdd-format spec in the shape real projects wrote: numbered
// section headings, wrapped list items, a resolved question and a header block
// above the first heading.
const legacySpec = `# 0007 — Retry budget

- **Status:** Approved
- **Last updated:** 2026-09-28
- **Owner spec for:** retry counting
- **Amends:** 0004 §3

## 1. Purpose

Legacy prose over
two lines.

## 2. Out of Scope

Code fences survive:

` + "```" + `
# not a heading
` + "```" + `

## 6. Open Questions

No blocking question remains.

- ~~OQ1 — Ship it?~~ **Resolved (2026-09-21): yes** because reasons.
- OQ2 — What next?

## 7. Acceptance Criteria

- [x] AC1 — ` + "`pnpm -r test`" + ` passes the
  engine suite.
- [ ] AC13: the gap is preserved

## 9. Decision Log

| Date | Decision |
|------|----------|
| 2026-09-21 | Ship |
`

func TestImportLegacyNumberedSections(t *testing.T) {
	dir := setupRepo(t)
	if err := os.WriteFile(dir+"/specs/0007-retry-budget.md", []byte(legacySpec), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := run(t, "--root", dir, "import"); err != nil {
		t.Fatal(err)
	}

	b := readSpec(t, dir, "0007-retry-budget.md")
	for _, want := range []string{
		"status: approved",
		"created: \"2026-09-28\"",
		"updated: \"2026-09-28\"",
		// "## 1. Purpose" -> Purpose, wrapping preserved
		"## Purpose\n\nLegacy prose over\ntwo lines.",
		// wrapped AC text joined, original number and checkbox kept
		"- [x] **AC-1**: `pnpm -r test` passes the engine suite.",
		"- [ ] **AC-13**: the gap is preserved",
		// "## 6. Open Questions" -> items, resolution kept as one text
		"- **Q-1** [resolved]: Ship it? **Resolved (2026-09-21): yes** because reasons.",
		"- **Q-2**: What next?",
		// non-item lines survive under the same Extra heading
		"## Open Questions\n\nNo blocking question remains.",
		"## Out of Scope\n\nCode fences survive:",
		// header block above the first heading survives as Metadata
		"## Metadata",
		"- **Owner spec for:** retry counting",
		"- **Amends:** 0004 §3",
		// decision log body untouched
		"| 2026-09-21 | Ship |",
	} {
		if !strings.Contains(b, want) {
			t.Fatalf("imported file missing %q:\n%s", want, b)
		}
	}

	// the structured items must be readable through the normal commands
	out, err := run(t, "--root", dir, "ac", "list", "0007")
	if err != nil {
		t.Fatalf("ac list: %v", err)
	}
	for _, want := range []string{"AC-1", "AC-13", "☑"} {
		if !strings.Contains(out, want) {
			t.Fatalf("ac list missing %q:\n%s", want, out)
		}
	}
	out, err = run(t, "--root", dir, "q", "list", "0007")
	if err != nil {
		t.Fatalf("q list: %v", err)
	}
	if !strings.Contains(out, "Q-1") || !strings.Contains(out, "resolved") {
		t.Fatalf("q list:\n%s", out)
	}
	if out, err = run(t, "--root", dir, "lint"); err != nil {
		t.Fatalf("lint: %v\n%s", err, out)
	}
	// re-import is a no-op
	out, _ = run(t, "--root", dir, "import")
	if !strings.Contains(out, "0 converted, 1 skipped") {
		t.Fatalf("re-import:\n%s", out)
	}
}

func TestSpecsDirDocumentsAreNotSpecs(t *testing.T) {
	dir := setupRepo(t)
	readme := "# Specs index\n\nNot a spec, but it lives in specs/.\n"
	if err := os.WriteFile(dir+"/specs/README.md", []byte(readme), 0o644); err != nil {
		t.Fatal(err)
	}

	// every directory-wide command must ignore the document, not fail on it
	if _, err := run(t, "--root", dir, "new", "Alpha"); err != nil {
		t.Fatalf("new: %v", err)
	}
	for _, cmd := range [][]string{{"list"}, {"lint"}, {"stats"}, {"index"}, {"export"}} {
		out, err := run(t, append([]string{"--root", dir}, cmd...)...)
		if err != nil {
			t.Fatalf("%s: %v\n%s", cmd[0], err, out)
		}
		if strings.Contains(out, "README.md") {
			t.Fatalf("%s mentions the document:\n%s", cmd[0], out)
		}
	}

	// import leaves the document alone (the spec created above is the one skip)
	out, err := run(t, "--root", dir, "--verbose", "import")
	if err != nil {
		t.Fatalf("import: %v\n%s", err, out)
	}
	if !strings.Contains(out, "skip README.md: not a spec file") {
		t.Fatalf("import did not report the document:\n%s", out)
	}
	if !strings.Contains(out, "0 converted, 1 skipped") {
		t.Fatalf("import summary:\n%s", out)
	}
	got, err := os.ReadFile(dir + "/specs/README.md")
	if err != nil || string(got) != readme {
		t.Fatalf("README.md changed: %v\n%s", err, got)
	}
}
