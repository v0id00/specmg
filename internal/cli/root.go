package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/v0id00/specmg/internal/output"
	"github.com/v0id00/specmg/internal/spec"
	"github.com/v0id00/specmg/internal/store"
)

var (
	flagJSON    bool
	flagVerbose bool
	flagDryRun  bool
	flagPlain   bool
	flagRoot    string
	flagDir     = "specs"
)

var rootCmd = &cobra.Command{
	Use:           "specmg",
	Short:         "Manage specs, requirements and acceptance criteria from the CLI",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "JSON output")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "verbose logging")
	rootCmd.PersistentFlags().BoolVar(&flagDryRun, "dry-run", false, "print changes without writing")
	rootCmd.PersistentFlags().BoolVar(&flagPlain, "plain", false, "disable icons/status badges in human output")
	rootCmd.PersistentFlags().StringVar(&flagRoot, "root", "", "repo root (default: search upwards for specs/)")
}

func Execute() error { return rootCmd.Execute() }

// usageErr marks flag/usage errors (exit code 2).
type usageErr struct{ error }

func IsUsageErr(err error) bool {
	var u usageErr
	return errors.As(err, &u)
}

func resolveRoot() (string, error) {
	if flagRoot != "" {
		return flagRoot, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	root, err := store.FindRoot(cwd)
	if err != nil {
		return "", fmt.Errorf("%w (run inside a repo with specs/, or pass --root)", err)
	}
	return root, nil
}

// mustLoad resolves an id-or-slug reference into (root, spec).
func mustLoad(ref string) (string, *spec.Spec, error) {
	root, err := resolveRoot()
	if err != nil {
		return "", nil, err
	}
	s, err := store.Load(root, flagDir, ref)
	return root, s, err
}

// saveOrDry persists unless --dry-run.
func saveOrDry(root string, s *spec.Spec) error {
	if flagDryRun {
		return nil
	}
	return store.Save(root, flagDir, s)
}

// badge wraps a status with its icon unless --plain or --json is set.
func badge(status string) string {
	if flagPlain || flagJSON {
		return status
	}
	return output.StatusBadge(status)
}

// check renders a criterion checkbox icon unless --plain/--json.
func check(done bool) string {
	if flagPlain || flagJSON {
		if done {
			return "x"
		}
		return " "
	}
	return output.CriterionIcon(done)
}
