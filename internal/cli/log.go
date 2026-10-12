package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var logRationale string

var logCmd = &cobra.Command{
	Use:   "log <ref> <decision>",
	Short: "Append a row to the Decision Log",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, s, err := mustLoad(args[0])
		if err != nil {
			return err
		}
		row := fmt.Sprintf("| %s | %s | %s |", today(), args[1], logRationale)
		if s.DecisionLog == "" {
			s.DecisionLog = "| Date | Decision | Rationale |\n|------|----------|-----------|\n" + row
		} else {
			s.DecisionLog = s.DecisionLog + "\n" + row
		}
		if flagDryRun {
			fmt.Fprintf(cmd.OutOrStdout(), "would add to %s: %s\n", s.Filename(), row)
			return nil
		}
		if err := saveOrDry(root, s); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "added decision to %s\n", s.Filename())
		return nil
	},
}

func init() {
	logCmd.Flags().StringVar(&logRationale, "rationale", "", "rationale column")
	rootCmd.AddCommand(logCmd)
}
