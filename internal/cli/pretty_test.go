package cli

import (
	"strings"
	"testing"
)

func TestPrettyOutput(t *testing.T) {
	dir := setupRepo(t)
	run(t, "--root", dir, "new", "Alpha")
	run(t, "--root", dir, "set", "0001", "--status", "approved")
	run(t, "--root", dir, "ac", "add", "0001", "`true` ok")
	run(t, "--root", dir, "ac", "check", "0001", "AC-1")
	run(t, "--root", dir, "q", "add", "0001", "open one", "--blocking")

	// pretty default: status badge in list
	out, _ := run(t, "--root", dir, "list")
	if !strings.Contains(out, "🟢 approved") {
		t.Fatalf("list missing status badge:\n%s", out)
	}
	// --plain disables badges
	out, _ = run(t, "--root", dir, "--plain", "list")
	if strings.Contains(out, "🟢") || !strings.Contains(out, "approved") {
		t.Fatalf("--plain should be icon-free:\n%s", out)
	}
	// --json never has icons
	out, _ = run(t, "--root", dir, "--json", "list")
	if strings.Contains(out, "🟢") {
		t.Fatalf("json must be icon-free:\n%s", out)
	}

	// ac list checkbox icons
	out, _ = run(t, "--root", dir, "ac", "list", "0001")
	if !strings.Contains(out, "☑") {
		t.Fatalf("ac list missing checked icon:\n%s", out)
	}
	// q list blocking icon
	out, _ = run(t, "--root", dir, "q", "list", "0001")
	if !strings.Contains(out, "❗ blocking") {
		t.Fatalf("q list missing blocking icon:\n%s", out)
	}
	// stats per-status badge
	out, _ = run(t, "--root", dir, "stats")
	if !strings.Contains(out, "🟢 approved") {
		t.Fatalf("stats missing badge:\n%s", out)
	}
	// search item + status icons
	out, _ = run(t, "--root", dir, "search", "open one")
	if !strings.Contains(out, "❗ Q-1") || !strings.Contains(out, "🟢 approved") {
		t.Fatalf("search missing icons:\n%s", out)
	}
	// search json stays clean
	out, _ = run(t, "--root", dir, "--json", "search", "open one")
	if strings.Contains(out, "❓") {
		t.Fatalf("search json has icon:\n%s", out)
	}
	// verify PASS icon
	out, _ = run(t, "--root", dir, "ac", "add", "0001", "`true` passes")
	out, _ = run(t, "--root", dir, "verify", "0001")
	if !strings.Contains(out, "✅ PASS") {
		t.Fatalf("verify missing pass icon:\n%s", out)
	}
}
