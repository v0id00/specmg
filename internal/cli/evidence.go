package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/v0id00/specmg/internal/spec"
)

// runShell is a seam for tests; sh -c is deliberate: the user supplies a full shell command.
var runShell = func(c string) (string, int, error) {
	cmd := exec.Command("/bin/sh", "-c", c)
	out, err := cmd.CombinedOutput()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		return "", -1, err
	}
	return string(out), code, nil
}

var evidenceCmd = &cobra.Command{
	Use:   "evidence <ref> --ac <AC-N> --cmd <string>",
	Short: "Run a shell command and record evidence against an acceptance criterion",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]
		acFlag := cmd.Flag("ac").Value.String()
		cmdFlag := cmd.Flag("cmd").Value.String()

		if acFlag == "" || cmdFlag == "" {
			return usageErr{fmt.Errorf("--ac and --cmd are required")}
		}

		typ, num, ok := spec.ParseItemID(acFlag)
		if !ok || typ != "AC" {
			return usageErr{fmt.Errorf("invalid or non-AC item ID: %s", acFlag)}
		}

		root, s, err := mustLoad(ref)
		if err != nil {
			return err
		}

		var crit *spec.Criterion
		for i := range s.Criteria {
			if s.Criteria[i].Num == num {
				crit = &s.Criteria[i]
				break
			}
		}
		if crit == nil {
			return fmt.Errorf("no AC-%d found in spec", num)
		}

		evDir := filepath.Join(root, flagDir, "evidence")
		ts := strconv.FormatInt(time.Now().Unix(), 10)
		evFile := fmt.Sprintf("%s-AC-%d-%s.md", s.FM.ID, num, ts)

		if flagDryRun {
			fmt.Fprintf(cmd.OutOrStdout(), "would run %s, evidence: %s/%s\n", cmdFlag, evDir, evFile)
			return nil
		}

		output, code, err := runShell(cmdFlag)
		writeEvidence(evDir, evFile, s.FM.ID, crit.Text, cmdFlag, code, output)
		if err != nil {
			return fmt.Errorf("command failed: %v (AC not checked)", err)
		}
		if code != 0 {
			return fmt.Errorf("command failed with exit %d (AC not checked)", code)
		}
		crit.Done = true
		if err := saveOrDry(root, s); err != nil {
			return err
		}

		evidenceRel := filepath.Join("evidence", evFile)
		fmt.Fprintf(cmd.OutOrStdout(), "AC-%d verified (evidence: %s)\n", num, evidenceRel)
		return nil
	},
}

func writeEvidence(dir, filename, specID, acText, cmdStr string, code int, output string) {
	os.MkdirAll(dir, 0o755)
	path := filepath.Join(dir, filename)
	var b strings.Builder
	fmt.Fprintf(&b, "---\nspec: %s\n---\n\n", specID)
	fmt.Fprintf(&b, "**AC**: %s\n\n", acText)
	fmt.Fprintf(&b, "**Command**: `%s`\n\n", cmdStr)
	fmt.Fprintf(&b, "**Exit code**: %d\n\n", code)
	fmt.Fprintf(&b, "```\n%s\n```\n", output)
	os.WriteFile(path, []byte(b.String()), 0o644)
}

func init() {
	evidenceCmd.Flags().StringP("ac", "a", "", "acceptance criterion ID (AC-N)")
	evidenceCmd.Flags().StringP("cmd", "c", "", "shell command to execute")
	evidenceCmd.MarkFlagsRequiredTogether("ac", "cmd")
	rootCmd.AddCommand(evidenceCmd)
}
