// Package cmd contains Cassor's command-line interface.
package cmd

import "github.com/spf13/cobra"

const version = "0.1.0-dev"

// NewRootCommand constructs the cassor command. Commands are attached here as
// their supporting application services become available.
func NewRootCommand() *cobra.Command {
	command := &cobra.Command{
		Use:           "cassor",
		Short:         "Plan, approve, and track repository work",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	command.Version = version
	command.AddCommand(newInitCommand())
	addWorkflowCommands(command)
	command.AddCommand(newContextCommand(), newSiteCommand())
	return command
}
