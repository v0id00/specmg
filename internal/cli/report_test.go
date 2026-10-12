package cli

import (
	"os"
	"strings"
	"testing"
)

func setupReportRepo(t *testing.T) string {
	dir := setupRepo(t)
	run(t, "--root", dir, "new", "Retry budget", "--owner", "alice", "--tag", "backend")
	run(t, "--root", dir, "req", "add", "0001", "SHALL retry 3 times")
	run(t, "--root", dir, "ac", "add", "0001", "`true` exits 0")
	run(t, "--root", dir, "ac", "check", "0001", "AC-1")
	run(t, "--root", dir, "q", "add", "0001", "Use backoff?", "--blocking")
	run(t, "--root", dir, "set", "0001", "--status", "approved")
	return dir
}

func TestReportHTML(t *testing.T) {
	dir := setupReportRepo(t)
	out, err := run(t, "--root", dir, "report")
	if err != nil || !strings.Contains(out, "wrote specmg-report.html") {
		t.Fatalf("out=%q err=%v", out, err)
	}
	b, err := os.ReadFile(dir + "/specmg-report.html")
	if err != nil {
		t.Fatal(err)
	}
	html := string(b)
	for _, want := range []string{"<!doctype html>", "Retry budget", "approved", "REQ-1",
		"☑", "❗", "Use backoff?", "owner: alice"} {
		if !strings.Contains(html, want) {
			t.Fatalf("html missing %q", want)
		}
	}
	// selected specs only
	if _, err := run(t, "--root", dir, "--out", dir+"/one.html", "report", "0001"); err != nil {
		t.Fatal(err)
	}
	// unknown ref fails
	if _, err := run(t, "--root", dir, "report", "9999"); err == nil {
		t.Fatal("want error for unknown ref")
	}
	// dry-run writes nothing
	if _, err := run(t, "--root", dir, "--dry-run", "--out", dir+"/dry.html", "report"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir + "/dry.html"); err == nil {
		t.Fatal("dry-run wrote a file")
	}
}

func TestReportPDF(t *testing.T) {
	dir := setupReportRepo(t)
	out, err := run(t, "--root", dir, "report", "--format", "pdf", "--out", dir+"/r.pdf")
	if err != nil || !strings.Contains(out, "wrote") {
		t.Fatalf("out=%q err=%v", out, err)
	}
	b, err := os.ReadFile(dir + "/r.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if len(b) < 1000 || !strings.HasPrefix(string(b), "%PDF-") {
		t.Fatalf("not a real PDF: %d bytes, prefix %q", len(b), string(b[:min(8, len(b))]))
	}
	if !strings.Contains(string(b[len(b)-32:]), "%"+"%EOF") {
		t.Fatal("PDF missing EOF trailer")
	}
	// invalid format is a usage error
	if _, err := run(t, "--root", dir, "report", "--format", "docx"); err == nil {
		t.Fatal("want error for unknown format")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
