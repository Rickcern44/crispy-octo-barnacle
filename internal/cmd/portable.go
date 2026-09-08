package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/rickcern44/cassor/internal/store"
)

func newExportCommand() *cobra.Command {
	var path string
	command := &cobra.Command{Use: "export", Short: "Export deterministic Cassor logical state", RunE: func(command *cobra.Command, _ []string) error {
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		data, err := store.ExportState(database)
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, data, 0o600); err != nil {
			return fmt.Errorf("write state export: %w", err)
		}
		_, err = fmt.Fprintf(command.OutOrStdout(), "Exported Cassor state to %s (%d bytes)\n", path, len(data))
		return err
	}}
	command.Flags().StringVar(&path, "file", "", "state export path")
	_ = command.MarkFlagRequired("file")
	return command
}

func newImportCommand() *cobra.Command {
	var path string
	command := &cobra.Command{Use: "import", Short: "Import a versioned Cassor state export", RunE: func(command *cobra.Command, _ []string) error {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read state export: %w", err)
		}
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		if err := store.ImportState(database, data); err != nil {
			return err
		}
		_, err = fmt.Fprintf(command.OutOrStdout(), "Imported Cassor state from %s\n", path)
		return err
	}}
	command.Flags().StringVar(&path, "file", "", "state export path")
	_ = command.MarkFlagRequired("file")
	return command
}
