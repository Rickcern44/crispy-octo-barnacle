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
	var updateAgent, updateScope string
	var updateJSON bool
	update := &cobra.Command{Use: "update", RunE: func(command *cobra.Command, _ []string) error {
		root, state, err := currentProject()
		if err != nil {
			return err
		}
		result, err := install.Update(root, state, updateAgent, updateScope)
		if err != nil {
			return err
		}
		return output(command, result, updateJSON)
	}}
	update.Flags().StringVar(&updateAgent, "agent", "codex", "runtime agent")
	update.Flags().StringVar(&updateScope, "scope", "project", "installation scope")
	update.Flags().BoolVar(&updateJSON, "json", false, "emit JSON")
	var uninstallAgent, uninstallScope string
	var uninstallJSON bool
	uninstall := &cobra.Command{Use: "uninstall", RunE: func(command *cobra.Command, _ []string) error {
		root, state, err := currentProject()
		if err != nil {
			return err
		}
		result, err := install.Uninstall(root, state, uninstallAgent, uninstallScope)
		if err != nil {
			return err
		}
		return output(command, result, uninstallJSON)
	}}
	uninstall.Flags().StringVar(&uninstallAgent, "agent", "codex", "runtime agent")
	uninstall.Flags().StringVar(&uninstallScope, "scope", "project", "installation scope")
	uninstall.Flags().BoolVar(&uninstallJSON, "json", false, "emit JSON")
	var doctorJSON bool
	doctor := &cobra.Command{Use: "doctor", RunE: func(command *cobra.Command, _ []string) error {
		root, state, err := currentProject()
		if err != nil {
			return err
		}
		report, err := install.Doctor(root, state)
		if err != nil {
			return err
		}
		return output(command, report, doctorJSON)
	}}
	doctor.Flags().BoolVar(&doctorJSON, "json", false, "emit JSON")
	command.AddCommand(update, uninstall, doctor)
	return command
}
