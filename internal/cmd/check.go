package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rickcern44/cassor/internal/check"
)

func newCheckCommand() *cobra.Command {
	return &cobra.Command{Use: "check", Short: "Validate Cassor state and roadmap generation", RunE: func(command *cobra.Command, _ []string) error {
		_, stateDir, err := currentProject()
		if err != nil {
			return err
		}
		if err := check.Run(stateDir); err != nil {
			return err
		}
		_, err = fmt.Fprintln(command.OutOrStdout(), "Cassor check passed")
		return err
	}}
}
