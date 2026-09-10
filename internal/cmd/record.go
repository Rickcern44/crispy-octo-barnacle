package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rickcern44/cassor/internal/store"
)

func recordPlanCommand() *cobra.Command {
	var path, approvalNote string
	var approve, approvedByUser, allowOpenQuestions, asJSON bool
	command := &cobra.Command{Use: "record", Short: "Atomically record an approved plan packet", RunE: func(command *cobra.Command, _ []string) error {
		if !approve || !approvedByUser {
			return fmt.Errorf("recording a plan requires both --approve and --approved-by-user")
		}
		packet, err := loadPlanPacket(path)
		if err != nil {
			return err
		}
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		if len(packet.OpenQuestions) > 0 && !allowOpenQuestions {
			return fmt.Errorf("plan has unresolved open questions; use --allow-open-questions only with explicit approval")
		}
		if allowOpenQuestions && len(packet.OpenQuestions) > 0 {
			_, err = fmt.Fprintln(command.ErrOrStderr(), "WARNING: recording this approved plan with unresolved open questions; they remain in immutable plan content.")
			if err != nil {
				return err
			}
		}
		var recorded store.RecordedPlan
		if allowOpenQuestions {
			recorded, err = store.RecordApprovedPlanAllowOpenQuestions(database, packet, approvalNote)
		} else {
			recorded, err = store.RecordApprovedPlan(database, packet, approvalNote)
		}
		if err != nil {
			return err
		}
		if asJSON {
			return output(command, recorded, true)
		}
		_, err = fmt.Fprintf(command.OutOrStdout(), "Plan %d revision %d approved and active for item %d; tasks: %d; next: cassor context --item %d --role implementation --max-bytes 12000\n", recorded.Plan.ID, recorded.Plan.Revision, recorded.Plan.ItemID, len(recorded.Tasks), recorded.Plan.ItemID)
		return err
	}}
	command.Flags().StringVar(&path, "file", "", "path to a plan-packet JSON file")
	command.Flags().BoolVar(&approve, "approve", false, "record this approved plan revision")
	command.Flags().BoolVar(&approvedByUser, "approved-by-user", false, "confirm explicit user approval")
	command.Flags().BoolVar(&allowOpenQuestions, "allow-open-questions", false, "explicitly approve recording unresolved open questions")
	command.Flags().StringVar(&approvalNote, "approval-note", "", "optional user approval note")
	command.Flags().BoolVar(&asJSON, "json", false, "emit JSON")
	_ = command.MarkFlagRequired("file")
	return command
}
