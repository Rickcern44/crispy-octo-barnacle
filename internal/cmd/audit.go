package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rickcern44/cassor/internal/store"
)

func newAuditCommand() *cobra.Command {
	var asJSON bool
	command := &cobra.Command{Use: "audit", Short: "Report roadmap lifecycle reconciliation candidates", RunE: func(command *cobra.Command, _ []string) error {
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		candidates, err := store.CompletionCandidates(database)
		if err != nil {
			return err
		}
		if asJSON {
			return output(command, candidates, true)
		}
		if len(candidates) == 0 {
			_, err = fmt.Fprintln(command.OutOrStdout(), "No roadmap completion reconciliation candidates")
			return err
		}
		_, err = fmt.Fprintln(command.OutOrStdout(), "Roadmap completion reconciliation candidates:")
		if err != nil {
			return err
		}
		for _, candidate := range candidates {
			if _, err := fmt.Fprintf(command.OutOrStdout(), "- RM-%d · %s (%s; %d/%d approved-plan tasks done)\n", candidate.ItemID, candidate.Title, candidate.ItemStatus, candidate.TaskCount, candidate.TaskCount); err != nil {
				return err
			}
		}
		return nil
	}}
	command.Flags().BoolVar(&asJSON, "json", false, "emit JSON")
	return command
}
