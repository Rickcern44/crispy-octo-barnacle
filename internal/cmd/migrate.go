package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/rickcern44/cassor/internal/config"
	"github.com/rickcern44/cassor/internal/store"
)

func newMigrateCommand() *cobra.Command {
	return &cobra.Command{Use: "migrate", Short: "Apply Cassor schema and default configuration migrations", RunE: func(command *cobra.Command, _ []string) error {
		_, stateDir, err := currentProject()
		if err != nil {
			return err
		}
		if err := store.Migrate(filepath.Join(stateDir, store.DatabaseFileName)); err != nil {
			return err
		}
		if _, err := os.Stat(config.AgentsPath(stateDir)); errors.Is(err, os.ErrNotExist) {
			if err := config.WriteDefaultAgents(config.AgentsPath(stateDir)); err != nil {
				return fmt.Errorf("write agent configuration: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("inspect agent configuration: %w", err)
		}
		_, err = fmt.Fprintln(command.OutOrStdout(), "Cassor migrations applied")
		return err
	}}
}
