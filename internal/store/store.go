package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/v0id00/specmg/internal/spec"
)

// FindRoot walks up from dir looking for a "specs" dir.
func FindRoot(dir string) (string, error) {
	d, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		if st, err := os.Stat(filepath.Join(d, "specs")); err == nil && st.IsDir() {
			return d, nil
		}
		p := filepath.Dir(d)
		if p == d {
			return "", errors.New("no specs/ directory found above " + dir)
		}
		d = p
	}
}

// timeNow is a var so tests can stub it.
var timeNow = func() string { return time.Now().UTC().Format("2006-01-02") }

// Save writes atomically (temp + rename) and bumps "updated".
func Save(root, dir string, s *spec.Spec) error {
	s.FM.Updated = timeNow()
	path := filepath.Join(root, dir, s.Filename())
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, spec.Render(s), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Load resolves "0001", "retry-budget" or "0001-retry-budget".
func Load(root, dir, ref string) (*spec.Spec, error) {
	p := filepath.Join(root, dir)
	entries, err := os.ReadDir(p)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() || !spec.IsSpecFilename(e.Name()) {
			continue
		}
		stem := strings.TrimSuffix(e.Name(), ".md")
		id := strings.SplitN(stem, "-", 2)[0]
		slug := strings.TrimPrefix(stem, id+"-")
		if stem == ref || id == ref || slug == ref {
			b, err := os.ReadFile(filepath.Join(p, e.Name()))
			if err != nil {
				return nil, err
			}
			return spec.Parse(b)
		}
	}
	return nil, fmt.Errorf("spec %q not found in %s", ref, p)
}

// List parses every spec file in dir, sorted by ID. Subdirs are ignored.
func List(root, dir string) ([]*spec.Spec, error) {
	p := filepath.Join(root, dir)
	entries, err := os.ReadDir(p)
	if err != nil {
		return nil, err
	}
	var out []*spec.Spec
	for _, e := range entries {
		if e.IsDir() || !spec.IsSpecFilename(e.Name()) {
			continue
		}
		b, err := os.ReadFile(filepath.Join(p, e.Name()))
		if err != nil {
			return nil, err
		}
		s, err := spec.Parse(b)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].FM.ID < out[j].FM.ID })
	return out, nil
}

// NextID returns the next zero-padded "NNNN".
func NextID(root, dir string) (string, error) {
	all, err := List(root, dir)
	if err != nil {
		return "", err
	}
	max := 0
	for _, s := range all {
		n, err := strconv.Atoi(s.FM.ID)
		if err == nil && n > max {
			max = n
		}
	}
	return fmt.Sprintf("%04d", max+1), nil
}

// Slugify makes a lowercase-hyphen slug, dropping non-ASCII.
func Slugify(title string) string {
	s := strings.ToLower(title)
	s = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			return r
		case r == ' ', r == '_', r == '-', r == '/':
			return '-'
		default:
			return -1
		}
	}, s)
	s = strings.Join(strings.Fields(s), "-")
	return strings.Trim(collapse(s, '-'), "-")
}

// collapse squeezes runs of sep into one.
func collapse(s string, sep byte) string {
	var b strings.Builder
	prev := byte(0)
	for i := 0; i < len(s); i++ {
		if s[i] == sep && prev == sep {
			continue
		}
		b.WriteByte(s[i])
		prev = s[i]
	}
	return b.String()
}
