package spec

import "regexp"

// reSpecFile matches the file-name contract "NNNN-slug.md" that Filename()
// produces. Anything else in specs/ is a document, not a spec (README.md,
// INDEX.md, notes): it must never be parsed as one.
var reSpecFile = regexp.MustCompile(`^\d{4}-.+\.md$`)

// IsSpecFilename reports whether name is a spec file under the format contract.
func IsSpecFilename(name string) bool {
	return reSpecFile.MatchString(name)
}
