package cli

import (
	"github.com/spf13/cobra"
	"github.com/v0id00/specmg/internal/output"
	"github.com/v0id00/specmg/internal/spec"
)

var getCmd = &cobra.Command{
	Use:   "get <id|slug>",
	Short: "Get a spec by id or slug",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, s, err := mustLoad(args[0])
		if err != nil {
			return err
		}

		if flagJSON {
			return output.JSON(cmd.OutOrStdout(), s)
		}
		_, err = cmd.OutOrStdout().Write(spec.Render(s))
		return err
	},
}

func init() { rootCmd.AddCommand(getCmd) }
