package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rickcern44/cassor/internal/config"
	"github.com/rickcern44/cassor/internal/project"
	"github.com/rickcern44/cassor/internal/store"
)

func newInitCommand() *cobra.Command {
	var name string
	command := &cobra.Command{Use: "init", Short: "Initialize Cassor for the current Git repository", RunE: func(command *cobra.Command, _ []string) error {
		workingDirectory, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}
		root, err := project.FindRepositoryRoot(workingDirectory)
		if err != nil {
			return err
		}
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("--name is required")
		}
		stateDir := filepath.Join(root, project.StateDirectory)
		if _, err := os.Stat(stateDir); err == nil {
			return fmt.Errorf("Cassor is already initialized at %s", stateDir)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("inspect Cassor state directory: %w", err)
		}
		if err := os.Mkdir(stateDir, 0o755); err != nil {
			return fmt.Errorf("create Cassor state directory: %w", err)
		}
		if err := config.Write(config.Path(stateDir), config.Config{Name: name}); err != nil {
			return fmt.Errorf("write configuration: %w", err)
		}
		if err := store.Initialize(filepath.Join(stateDir, store.DatabaseFileName)); err != nil {
			return err
		}
		_, err = fmt.Fprintf(command.OutOrStdout(), "Initialized Cassor for %q at %s\n", name, stateDir)
		return err
	}}
	command.Flags().StringVar(&name, "name", "", "project name")
	return command
}
