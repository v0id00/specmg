package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm <ref>",
	Short: "Remove a spec file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]

		root, s, err := mustLoad(ref)
		if err != nil {
			return err
		}

		fname := filepath.Join(flagDir, s.Filename())
		if flagDryRun {
			fmt.Fprintf(cmd.OutOrStdout(), "would remove %s\n", fname)
			return nil
		}

		path := filepath.Join(root, fname)
		if err := os.Remove(path); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "removed %s\n", fname)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}
