package cli

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestE2EHappyPath exercises the real binary end-to-end in a temp dir.
func TestE2EHappyPath(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	bin := filepath.Join(t.TempDir(), "specmg")
	if out, err := exec.Command("go", "build", "-o", bin, "github.com/v0id00/specmg/cmd/specmg").CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, out)
	}
	dir := t.TempDir()
	run := func(args ...string) string {
		cmd := exec.Command(bin, args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%v: %v %s", args, err, out)
		}
		return string(out)
	}
	run("init")
	run("new", "Retry budget", "--owner", "alice")
	run("req", "add", "0001", "SHALL retry 3 times")
	run("ac", "add", "0001", "`specmg list` exits 0")
	run("ac", "check", "0001", "AC-1")
	out := run("list")
	if !strings.Contains(out, "0001") || !strings.Contains(out, "Retry budget") {
		t.Fatalf("e2e list: %s", out)
	}
	out = run("--json", "get", "0001")
	if !strings.Contains(out, `"Status": "draft"`) {
		t.Fatalf("e2e get json: %s", out)
	}
	out = run("lint")
	if !strings.Contains(out, "lint: ok") {
		t.Fatalf("e2e lint: %s", out)
	}
}
