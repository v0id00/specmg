package cli

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var verifyAll bool

var verifyCmd = &cobra.Command{
	Use:   "verify <ref>",
	Short: "Run backticked commands in all open ACs and check them on exit 0",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, s, err := mustLoad(args[0])
		if err != nil {
			return err
		}
		ran, failed := 0, 0
		passIcon, failIcon := "PASS", "FAIL"
		if !flagPlain {
			passIcon, failIcon = "✅ PASS", "❌ FAIL"
		}
		for i := range s.Criteria {
			c := &s.Criteria[i]
			if c.Done {
				continue
			}
			command := backtickCommand(c.Text)
			if command == "" {
				if verifyAll {
					fmt.Fprintf(cmd.OutOrStdout(), "SKIP AC-%d (no command in text)\n", c.Num)
				}
				continue
			}
			ran++
			out, code, err := runShell(command)
			if err != nil || code != 0 {
				failed++
				fmt.Fprintf(cmd.OutOrStdout(), "%s AC-%d: %s (exit %d)\n", failIcon, c.Num, command, code)
				if flagVerbose {
					fmt.Fprintln(cmd.OutOrStdout(), strings.TrimSpace(out))
				}
				continue
			}
			c.Done = true
			fmt.Fprintf(cmd.OutOrStdout(), "%s AC-%d: %s\n", passIcon, c.Num, command)
		}
		if ran == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "nothing to verify")
			return nil
		}
		if flagDryRun {
			return nil
		}
		if err := saveOrDry(root, s); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "verified %d AC(s): %d passed, %d failed\n", ran, ran-failed, failed)
		if failed > 0 {
			return fmt.Errorf("%d AC(s) failed", failed)
		}
		return nil
	},
}

var reBacktick = regexp.MustCompile("`([^`]+)`")

// backtickCommand extracts the first backtick-wrapped command from AC text.
func backtickCommand(text string) string {
	m := reBacktick.FindStringSubmatch(text)
	if m == nil {
		return ""
	}
	return m[1]
}

// today is a var for tests.
var today = func() string { return time.Now().UTC().Format("2006-01-02") }

func init() {
	verifyCmd.Flags().BoolVar(&verifyAll, "all", false, "also list ACs that have no command")
	rootCmd.AddCommand(verifyCmd)
}
