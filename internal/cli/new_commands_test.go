package cli

import (
	"os"
	"strings"
	"testing"
)

func TestLogAdd(t *testing.T) {
	dir := setupRepo(t)
	run(t, "--root", dir, "new", "Alpha")
	if _, err := run(t, "--root", dir, "log", "0001", "Max 3 retries", "--rationale", "Simplicity"); err != nil {
		t.Fatal(err)
	}
	b := readSpec(t, dir, "0001-alpha.md")
	if !strings.Contains(b, "| Max 3 retries | Simplicity |") {
		t.Fatalf("decision row missing:\n%s", b)
	}
	if !strings.Contains(b, "| Date | Decision | Rationale |") {
		t.Fatalf("table header missing:\n%s", b)
	}
	// second row appends, header not duplicated
	if _, err := run(t, "--root", dir, "log", "0001", "Second call"); err != nil {
		t.Fatal(err)
	}
	b = readSpec(t, dir, "0001-alpha.md")
	if strings.Count(b, "| Date | Decision | Rationale |") != 1 {
		t.Fatalf("header duplicated:\n%s", b)
	}
}

func TestVerifyRunsBacktickedCommands(t *testing.T) {
	dir := setupRepo(t)
	run(t, "--root", dir, "new", "Alpha")
	run(t, "--root", dir, "ac", "add", "0001", "`true` passes")
	run(t, "--root", dir, "ac", "add", "0001", "`false` passes")
	run(t, "--root", dir, "ac", "add", "0001", "no command here")

	old := runShell
	n := 0
	runShell = func(c string) (string, int, error) {
		n++
		if c == "false" {
			return "", 1, nil
		}
		return "", 0, nil
	}
	defer func() { runShell = old }()

	out, err := run(t, "--root", dir, "verify", "0001")
	if err == nil || n != 2 {
		t.Fatalf("want failure for 2 runs, got err=%v n=%d out=%s", err, n, out)
	}
	if !strings.Contains(out, "PASS AC-1") || !strings.Contains(out, "FAIL AC-2") {
		t.Fatalf("out:\n%s", out)
	}
	b := readSpec(t, dir, "0001-alpha.md")
	if !strings.Contains(b, "- [x] **AC-1**") || strings.Contains(b, "- [x] **AC-2**") {
		t.Fatalf("check state wrong:\n%s", b)
	}
}

func TestVerifyAllGreen(t *testing.T) {
	dir := setupRepo(t)
	run(t, "--root", dir, "new", "Alpha")
	run(t, "--root", dir, "ac", "add", "0001", "`true` passes")
	out, err := run(t, "--root", dir, "verify", "0001")
	if err != nil || !strings.Contains(out, "1 passed, 0 failed") {
		t.Fatalf("out=%q err=%v", out, err)
	}
}

func TestSearch(t *testing.T) {
	dir := setupRepo(t)
	run(t, "--root", dir, "new", "Retry budget")
	run(t, "--root", dir, "req", "add", "0001", "The client SHALL retry 3 times")
	run(t, "--root", dir, "new", "Unrelated")

	out, _ := run(t, "--root", dir, "search", "retry")
	if strings.Contains(out, "Unrelated") || !strings.Contains(out, "REQ-1") {
		t.Fatalf("search:\n%s", out)
	}
	out, _ = run(t, "--root", dir, "--json", "search", "retry 3")
	if !strings.Contains(out, `"item": "REQ-1"`) {
		t.Fatalf("search json:\n%s", out)
	}
	out, _ = run(t, "--root", dir, "search", "zzz-nothing")
	if !strings.Contains(out, "no matches") {
		t.Fatalf("search empty:\n%s", out)
	}
}

func TestExportCSVAndJSON(t *testing.T) {
	dir := setupRepo(t)
	run(t, "--root", dir, "new", "Alpha")
	run(t, "--root", dir, "req", "add", "0001", "SHALL do x")
	run(t, "--root", dir, "ac", "add", "0001", "`true` works")

	out, _ := run(t, "--root", dir, "--json", "export")
	if !strings.Contains(out, `"type": "req"`) || !strings.Contains(out, `"type": "ac"`) {
		t.Fatalf("export json:\n%s", out)
	}
	out, _ = run(t, "--root", dir, "export", "--csv")
	if !strings.Contains(out, "spec,status,type,id,text,done,rel,target") {
		t.Fatalf("export csv header:\n%s", out)
	}
	if !strings.Contains(out, "0001-alpha,draft,req,REQ-1,SHALL do x") {
		t.Fatalf("export csv row:\n%s", out)
	}
}

func TestIndex(t *testing.T) {
	dir := setupRepo(t)
	run(t, "--root", dir, "new", "Alpha")
	run(t, "--root", dir, "new", "Beta")

	out, _ := run(t, "--root", dir, "index")
	if !strings.Contains(out, "| [0001](0001-alpha.md) | draft | Alpha |") {
		t.Fatalf("index table:\n%s", out)
	}
	if _, err := run(t, "--root", dir, "index", "--write"); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(dir + "/specs/INDEX.md")
	if err != nil || !strings.Contains(string(b), "0002-beta.md") {
		t.Fatalf("INDEX.md: %v\n%s", err, b)
	}
	// dry-run does not write
	os.Remove(dir + "/specs/INDEX.md")
	if _, err := run(t, "--root", dir, "--dry-run", "index", "--write"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir + "/specs/INDEX.md"); err == nil {
		t.Fatal("dry-run wrote INDEX.md")
	}
}

func TestImportConvertsSDDFormat(t *testing.T) {
	dir := setupRepo(t)
	// existing specmg file must be skipped (create BEFORE the broken legacy file,
	// since store.List parses everything and NextID would fail on the legacy file)
	run(t, "--root", dir, "new", "Modern")
	legacy := `# 0007 — Retry budget

- **Status:** Approved
- **Owner:** alice

## Purpose

Legacy prose.

## Requirements

- **REQ-1**: The client SHALL retry.
- The client SHALL log failures.

## Acceptance Criteria

- [ ] AC1: ` + "`true`" + ` exits 0
- [x] AC2: already done
`
	if err := os.WriteFile(dir+"/specs/0007-retry-budget.md", []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := run(t, "--root", dir, "import")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "1 converted, 1 skipped") {
		t.Fatalf("import summary:\n%s", out)
	}
	b := readSpec(t, dir, "0007-retry-budget.md")
	for _, want := range []string{"---\n", "status: approved", "owner: alice", "**REQ-1**: The client SHALL retry.",
		"[x] **AC-2**: already done", "Legacy prose."} {
		if !strings.Contains(b, want) {
			t.Fatalf("imported file missing %q:\n%s", want, b)
		}
	}
	// re-import skips (already converted)
	out, _ = run(t, "--root", dir, "import")
	if !strings.Contains(out, "0 converted, 2 skipped") {
		t.Fatalf("re-import:\n%s", out)
	}
	// imported spec must parse via the normal path
	if _, err := run(t, "--root", dir, "get", "0007"); err != nil {
		t.Fatalf("get after import: %v", err)
	}
}
