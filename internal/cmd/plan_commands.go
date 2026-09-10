package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rickcern44/cassor/internal/store"
)

// planTemplateCommand emits a complete packet scaffold for an existing item.
// It only reads the item; the resulting packet is suitable for editing and
// validation before recording.
func planTemplateCommand() *cobra.Command {
	var itemID int64
	command := &cobra.Command{
		Use:   "template",
		Short: "Emit a plan-packet template for an existing roadmap item",
		RunE: func(command *cobra.Command, _ []string) error {
			database, err := databaseForCommand()
			if err != nil {
				return err
			}
			defer database.Close()
			item, err := store.GetItem(database, itemID)
			if err != nil {
				return err
			}
			packet := store.PlanPacket{
				RoadmapItem: store.PacketRoadmapItem{
					ID:          &item.ID,
					Title:       item.Title,
					Description: item.Description,
					Category:    item.Category,
					Horizon:     item.Horizon,
					Rationale:   item.Rationale,
				},
				Goal:               fmt.Sprintf("Implement %s", item.Title),
				Scope:              store.PacketScope{Included: []string{"The approved implementation scope for this item"}, Excluded: []string{"Unrelated roadmap work"}},
				Decisions:          []string{"Confirm implementation details during planning"},
				AcceptanceCriteria: []store.PacketCriterion{{ID: "C1", Title: fmt.Sprintf("%s is implemented and verified", item.Title), Description: "The requested outcome is complete and evidence is recorded."}},
				Constraints:        []string{"Preserve existing behavior outside this item"},
				Tasks:              []store.PacketTask{{Title: fmt.Sprintf("Implement %s", item.Title), Description: "Complete the implementation and verify the acceptance criteria.", Verification: []string{"go test ./..."}}},
				Risks:              []string{"Implementation assumptions may need refinement"},
				OpenQuestions:      []string{},
			}
			encoder := json.NewEncoder(command.OutOrStdout())
			return encoder.Encode(packet)
		},
	}
	command.Flags().Int64Var(&itemID, "item", 0, "roadmap item ID")
	_ = command.MarkFlagRequired("item")
	return command
}

func planValidateCommand() *cobra.Command {
	var path string
	command := &cobra.Command{
		Use:   "validate",
		Short: "Validate a plan-packet JSON file",
		RunE: func(command *cobra.Command, _ []string) error {
			packet, err := loadPlanPacket(path)
			if err != nil {
				return err
			}
			if packet.RoadmapItem.ID != nil {
				_, err = fmt.Fprintf(command.OutOrStdout(), "Plan packet valid for item %d\n", *packet.RoadmapItem.ID)
				return err
			}
			_, err = fmt.Fprintln(command.OutOrStdout(), "Plan packet valid")
			return err
		},
	}
	command.Flags().StringVar(&path, "file", "", "path to a plan-packet JSON file")
	_ = command.MarkFlagRequired("file")
	return command
}
