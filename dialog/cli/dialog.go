package cli

import "github.com/spf13/cobra"

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dialog",
		Short: "Dialog utilities",
		Long:  `Utilities for working with Infinity Engine dialogs.`,
	}

	cmd.AddCommand(NewLsCommand())
	return cmd
}
