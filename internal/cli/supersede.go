package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var supersedeCmd = &cobra.Command{
	Use:   "supersede <ref> --by <other-ref>",
	Short: "Mark a spec as superseded by another",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]
		if flagSupersedeBy == "" {
			return usageErr{fmt.Errorf("--by is required")}
		}

		root, s, err := mustLoad(ref)
		if err != nil {
			return err
		}
		_, other, err := mustLoad(flagSupersedeBy)
		if err != nil {
			return err
		}

		s.FM.Status = "superseded"
		s.FM.SupersededBy = other.FM.ID
		today := time.Now().UTC().Format("2006-01-02")
		entry := fmt.Sprintf("| %s | Superseded by %s | via specmg |\n", today, other.FM.ID)
		if s.DecisionLog != "" {
			s.DecisionLog += "\n" + entry
		} else {
			s.DecisionLog = entry
		}

		if flagDryRun {
			fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] would supersede %s by %s\n", s.FM.ID, other.FM.ID)
			return nil
		}

		if err := saveOrDry(root, s); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "superseded %s by %s\n", s.FM.ID, other.FM.ID)
		return nil
	},
}

var flagSupersedeBy string

func init() {
	supersedeCmd.Flags().StringVarP(&flagSupersedeBy, "by", "b", "", "reference of the overriding spec")
	supersedeCmd.MarkFlagRequired("by")
	rootCmd.AddCommand(supersedeCmd)
}
