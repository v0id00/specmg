package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// resetFlag restores a flag to its default between test invocations.
// Slice flags must be set to "" (DefValue "[]" would parse as the literal ["[]"]).
func resetFlag(f *pflag.Flag) {
	if f.Value.Type() == "stringSlice" {
		f.Value.Set("")
		return
	}
	f.Value.Set(f.DefValue)
}

// run executes the root command with args, returning combined output.
func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	// cobra flag vars persist across Execute calls in one process — reset everything
	rootCmd.PersistentFlags().VisitAll(resetFlag)
	var visitAll func(*cobra.Command)
	visitAll = func(c *cobra.Command) {
		c.Flags().VisitAll(resetFlag)
		for _, sub := range c.Commands() {
			visitAll(sub)
		}
	}
	visitAll(rootCmd)
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

// setupRepo creates a temp repo and chdirs into it.
func setupRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	if _, err := run(t, "--root", dir, "init"); err != nil {
		t.Fatal(err)
	}
	return dir
}

func readSpec(t *testing.T, dir, name string) string {
	t.Helper()
	b, err := os.ReadFile(dir + "/specs/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestLifecycle(t *testing.T) {
	dir := setupRepo(t)

	out, err := run(t, "--root", dir, "new", "Retry budget", "--owner", "alice", "--tag", "backend")
	if err != nil || !strings.Contains(out, "created specs/0001-retry-budget.md") {
		t.Fatalf("new: %q %v", out, err)
	}
	if _, err := run(t, "--root", dir, "new", "Backoff"); err != nil {
		t.Fatal(err)
	}

	if _, err := run(t, "--root", dir, "req", "add", "0001", "SHALL retry 3 times"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "--root", dir, "ac", "add", "0001", "`specmg list` exits 0"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "--root", dir, "ac", "check", "0001", "AC-1"); err != nil {
		t.Fatal(err)
	}
	if b := readSpec(t, dir, "0001-retry-budget.md"); !strings.Contains(b, "- [x] **AC-1**") {
		t.Fatalf("ac check not persisted:\n%s", b)
	}
	if _, err := run(t, "--root", dir, "ac", "uncheck", "0001", "AC-1"); err != nil {
		t.Fatal(err)
	}
	if b := readSpec(t, dir, "0001-retry-budget.md"); !strings.Contains(b, "- [ ] **AC-1**") {
		t.Fatal("ac uncheck not persisted")
	}

	if _, err := run(t, "--root", dir, "link", "add", "0001", "--rel", "implements", "--to", "0002"); err != nil {
		t.Fatal(err)
	}
	if b := readSpec(t, dir, "0001-retry-budget.md"); !strings.Contains(b, "**L-1**: implements -> 0002-backoff") {
		t.Fatal("link not persisted")
	}
	if _, err := run(t, "--root", dir, "link", "add", "0001", "--rel", "bogus", "--to", "0002"); err == nil {
		t.Fatal("want invalid rel error")
	}
	if _, err := run(t, "--root", dir, "link", "add", "0001", "--rel", "relates", "--to", "9999"); err == nil {
		t.Fatal("want target-not-found error")
	}

	if _, err := run(t, "--root", dir, "q", "add", "0001", "Use backoff?", "--blocking"); err != nil {
		t.Fatal(err)
	}
	if b := readSpec(t, dir, "0001-retry-budget.md"); !strings.Contains(b, "**Q-1** [blocking]:") {
		t.Fatal("blocking tag missing")
	}

	// approved + unresolved blocking question → lint must fail
	if _, err := run(t, "--root", dir, "set", "0001", "--status", "approved"); err != nil {
		t.Fatal(err)
	}
	out, err = run(t, "--root", dir, "lint")
	if err == nil || !strings.Contains(out, "blocking Q-1") {
		t.Fatalf("lint should fail: %q err=%v", out, err)
	}
	if _, err := run(t, "--root", dir, "q", "resolve", "0001", "Q-1"); err != nil {
		t.Fatal(err)
	}
	if out, err := run(t, "--root", dir, "lint"); err != nil || !strings.Contains(out, "lint: ok") {
		t.Fatalf("lint after resolve: %q %v", out, err)
	}

	// invalid status rejected
	if _, err := run(t, "--root", dir, "set", "0001", "--status", "bogus"); err == nil {
		t.Fatal("want invalid status error")
	}

	// filters
	if out, _ := run(t, "--root", dir, "list", "--status", "approved"); strings.Contains(out, "0002") {
		t.Fatalf("status filter broken:\n%s", out)
	}
	if out, _ := run(t, "--root", dir, "--json", "list"); !strings.Contains(out, `"id": "0001"`) {
		t.Fatalf("json list broken:\n%s", out)
	}

	// dry-run writes nothing
	if _, err := run(t, "--root", dir, "--dry-run", "new", "Third"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir + "/specs/0003-third.md"); err == nil {
		t.Fatal("dry-run wrote a file")
	}

	// slug rename renames file
	if _, err := run(t, "--root", dir, "set", "0001", "--slug", "renamed"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir + "/specs/0001-renamed.md"); err != nil {
		t.Fatal("renamed file missing")
	}
	if _, err := os.Stat(dir + "/specs/0001-retry-budget.md"); err == nil {
		t.Fatal("old file still present")
	}

	// supersede + rm
	if _, err := run(t, "--root", dir, "supersede", "0002", "--by", "0001"); err != nil {
		t.Fatal(err)
	}
	if b := readSpec(t, dir, "0002-backoff.md"); !strings.Contains(b, "status: superseded") {
		t.Fatalf("supersede not persisted:\n%s", b)
	}
	if _, err := run(t, "--root", dir, "--dry-run", "rm", "0002"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir + "/specs/0002-backoff.md"); err != nil {
		t.Fatal("dry-run removed file")
	}
	if _, err := run(t, "--root", dir, "rm", "0002"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir + "/specs/0002-backoff.md"); err == nil {
		t.Fatal("file still present")
	}
}

func TestEvidenceFailClosed(t *testing.T) {
	dir := setupRepo(t)
	run(t, "--root", dir, "new", "Alpha")
	run(t, "--root", dir, "ac", "add", "0001", "boom passes")

	old := runShell
	runShell = func(c string) (string, int, error) { return "boom failed", 3, nil }
	defer func() { runShell = old }()

	if _, err := run(t, "--root", dir, "evidence", "0001", "--ac", "AC-1", "--cmd", "boom"); err == nil {
		t.Fatal("want error on non-zero exit")
	}
	if b := readSpec(t, dir, "0001-alpha.md"); strings.Contains(b, "- [x] **AC-1**") {
		t.Fatal("fail-closed violated: AC checked despite failure")
	}
	entries, _ := os.ReadDir(dir + "/specs/evidence")
	if len(entries) != 1 {
		t.Fatalf("evidence file not written on failure: %v", entries)
	}
}

func TestEvidenceSuccessChecksAC(t *testing.T) {
	dir := setupRepo(t)
	run(t, "--root", dir, "new", "Alpha")
	run(t, "--root", dir, "ac", "add", "0001", "`go test` passes")

	old := runShell
	runShell = func(c string) (string, int, error) { return "ok output", 0, nil }
	defer func() { runShell = old }()

	if _, err := run(t, "--root", dir, "evidence", "0001", "--ac", "AC-1", "--cmd", "go test"); err != nil {
		t.Fatal(err)
	}
	if b := readSpec(t, dir, "0001-alpha.md"); !strings.Contains(b, "- [x] **AC-1**") {
		t.Fatal("AC not checked on success")
	}
}

func TestEvidenceDryRunDoesNotExecute(t *testing.T) {
	dir := setupRepo(t)
	run(t, "--root", dir, "new", "Alpha")
	run(t, "--root", dir, "ac", "add", "0001", "x")

	called := false
	old := runShell
	runShell = func(c string) (string, int, error) { called = true; return "", 0, nil }
	defer func() { runShell = old }()

	if _, err := run(t, "--root", dir, "--dry-run", "evidence", "0001", "--ac", "AC-1", "--cmd", "x"); err != nil || called {
		t.Fatalf("dry-run executed shell: called=%v err=%v", called, err)
	}
}

func TestStatsAndTreeAndSkill(t *testing.T) {
	dir := setupRepo(t)
	run(t, "--root", dir, "new", "Alpha")
	run(t, "--root", dir, "new", "Beta")
	run(t, "--root", dir, "ac", "add", "0001", "c1")
	run(t, "--root", dir, "q", "add", "0001", "q1", "--blocking")
	run(t, "--root", dir, "link", "add", "0001", "--rel", "relates", "--to", "0002")

	out, _ := run(t, "--root", dir, "--json", "stats")
	if !strings.Contains(out, `"open_acceptance_criteria": 1`) || !strings.Contains(out, `"blocking_questions": 1`) {
		t.Fatalf("stats json: %s", out)
	}
	out, _ = run(t, "--root", dir, "tree")
	if !strings.Contains(out, "relates -> 0002-beta") {
		t.Fatalf("tree: %s", out)
	}
	out, _ = run(t, "--root", dir, "tree", "0002")
	if !strings.Contains(out, "incoming:") || !strings.Contains(out, "0001-alpha") {
		t.Fatalf("tree backlinks: %s", out)
	}
	out, _ = run(t, "--root", dir, "skill")
	if !strings.Contains(out, "NEVER edit") || !strings.Contains(out, "specmg") {
		t.Fatalf("skill output: %s", out)
	}
	// skill must document the real command surface (signatures caught drift before)
	for _, want := range []string{
		"specmg new \"<title>\"", "specmg req add <ref>", "check|uncheck <ref> AC-N",
		"specmg link add <ref> --rel <rel> --to <target>", "specmg q add <ref>",
		"specmg verify <ref>", "specmg report [<ref>...]", "specmg log <ref>",
		"--dry-run", "status: draft",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("skill missing %q", want)
		}
	}
}
