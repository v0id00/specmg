package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/v0id00/specmg/internal/spec"
)

func mkRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestFindRoot(t *testing.T) {
	root := mkRepo(t)
	sub := filepath.Join(root, "a", "b")
	os.MkdirAll(sub, 0o755)
	got, err := FindRoot(sub)
	if err != nil || got != root {
		t.Fatalf("FindRoot = %q, %v; want %q", got, err, root)
	}
	if _, err := FindRoot(t.TempDir()); err == nil {
		t.Fatal("want error outside repo")
	}
}

func TestSaveLoadListNextID(t *testing.T) {
	root := mkRepo(t)
	s := &spec.Spec{FM: spec.Frontmatter{
		ID: "0001", Slug: "hello", Title: "Hello", Status: spec.StatusDraft,
		Created: "2026-10-12", Updated: "2026-10-12",
	}, Purpose: "p"}
	if err := Save(root, "specs", s); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "specs", "0001-hello.md")); err != nil {
		t.Fatal(err)
	}
	for _, ref := range []string{"0001", "hello", "0001-hello"} {
		got, err := Load(root, "specs", ref)
		if err != nil || got.FM.Title != "Hello" {
			t.Fatalf("Load(%q): %+v %v", ref, got, err)
		}
	}
	if _, err := Load(root, "specs", "9999"); err == nil {
		t.Fatal("want not-found error")
	}
	all, err := List(root, "specs")
	if err != nil || len(all) != 1 {
		t.Fatalf("List: %d, %v", len(all), err)
	}
	id, err := NextID(root, "specs")
	if err != nil || id != "0002" {
		t.Fatalf("NextID = %q, %v", id, err)
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Retry budget":    "retry-budget",
		"Şubat Release":   "ubat-release",
		"Foo_Bar/Baz 1.0": "foo-bar-baz-10",
		"!!!":             "",
		"  spaced  out  ": "spaced-out",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
}
