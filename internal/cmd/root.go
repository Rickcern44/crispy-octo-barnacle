// Package cmd contains Cassor's command-line interface.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version = "0.0.0-dev"
	Commit  = "unknown"
	Date    = "unknown"
)

// NewRootCommand constructs the cassor command. Commands are attached here as
// their supporting application services become available.
func NewRootCommand() *cobra.Command {
	command := &cobra.Command{
		Use:           "cassor",
		Short:         "Plan, approve, and track repository work",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	command.Version = BuildVersion()
	command.AddCommand(&cobra.Command{Use: "version", Short: "Show build version metadata", Run: func(command *cobra.Command, _ []string) {
		fmt.Fprintf(command.OutOrStdout(), "cassor %s\ncommit: %s\nbuilt: %s\n", BuildVersion(), Commit, Date)
	}})
	command.AddCommand(newInitCommand())
	command.AddCommand(newMigrateCommand())
	addWorkflowCommands(command)
	command.AddCommand(lifecycleCommand())
	command.AddCommand(newContextCommand(), newSiteCommand())
	command.AddCommand(newAuditCommand())
	command.AddCommand(newCheckCommand())
	command.AddCommand(newSkillsCommand())
	command.AddCommand(newRunReportCommand())
	return command
}

func BuildVersion() string { return Version }
