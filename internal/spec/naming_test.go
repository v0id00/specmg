package spec

import "testing"

func TestIsSpecFilename(t *testing.T) {
	cases := map[string]bool{
		"0001-core-architecture.md":  true,
		"0018-inventory.md":          true,
		"README.md":                  false,
		"INDEX.md":                   false,
		"TEMPLATE.md":                false,
		"1-core.md":                  false,
		"0001.md":                    false,
		"0001-core-architecture.txt": false,
	}
	for name, want := range cases {
		if got := IsSpecFilename(name); got != want {
			t.Errorf("IsSpecFilename(%q) = %v, want %v", name, got, want)
		}
	}
}
