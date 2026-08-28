package cmd

import (
	"github.com/spf13/cobra"

	install "github.com/rickcern44/cassor/internal/skills"
)

func newSkillsCommand() *cobra.Command {
	command := &cobra.Command{Use: "skills", Short: "Install and inspect Cassor runtime skills"}
	var agent, scope string
	var dryRun, asJSON bool
	installCommand := &cobra.Command{Use: "install", RunE: func(command *cobra.Command, _ []string) error {
		root, state, err := currentProject()
		if err != nil {
			return err
		}
		result, err := install.Install(root, state, install.InstallOptions{Agent: agent, Scope: scope, DryRun: dryRun})
		if err != nil {
			return err
		}
		return output(command, result, asJSON)
	}}
	installCommand.Flags().StringVar(&agent, "agent", "", "runtime agent (currently codex)")
	installCommand.Flags().StringVar(&scope, "scope", "project", "installation scope")
	installCommand.Flags().BoolVar(&dryRun, "dry-run", false, "show changes without writing")
	installCommand.Flags().BoolVar(&asJSON, "json", false, "emit JSON")
	_ = installCommand.MarkFlagRequired("agent")
	var listJSON bool
	list := &cobra.Command{Use: "list", RunE: func(command *cobra.Command, _ []string) error { return output(command, []string{"codex"}, listJSON) }}
	list.Flags().BoolVar(&listJSON, "json", false, "emit JSON")
	var statusJSON bool
	status := &cobra.Command{Use: "status", RunE: func(command *cobra.Command, _ []string) error {
		root, state, err := currentProject()
		if err != nil {
			return err
		}
		manifest, err := install.Status(root, state)
		if err != nil {
			return err
		}
		return output(command, manifest, statusJSON)
	}}
	status.Flags().BoolVar(&statusJSON, "json", false, "emit JSON")
	command.AddCommand(installCommand, list, status)
	return command
}
