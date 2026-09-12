package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rickcern44/cassor/internal/project"
	"github.com/rickcern44/cassor/internal/store"
)

func addWorkflowCommands(root *cobra.Command) {
	root.AddCommand(categoryCommand(), epicCommand(), itemCommand(), planCommand(), taskCommand())
}
func databaseForCommand() (*sql.DB, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	stateDir, err := project.FindStateDirectory(workingDirectory)
	if err != nil {
		return nil, err
	}
	return store.Open(filepath.Join(stateDir, store.DatabaseFileName))
}
func id(argument string) (int64, error) {
	value, err := strconv.ParseInt(argument, 10, 64)
	if err != nil || value < 1 {
		return 0, fmt.Errorf("invalid ID %q", argument)
	}
	return value, nil
}
func output(command *cobra.Command, value any, asJSON bool) error {
	if asJSON {
		encoder := json.NewEncoder(command.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(value)
	}
	_, err := fmt.Fprintln(command.OutOrStdout(), value)
	return err
}
func required(value, name string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

func categoryCommand() *cobra.Command {
	command := &cobra.Command{Use: "category", Short: "Manage roadmap categories"}
	var name string
	add := &cobra.Command{Use: "add", RunE: func(command *cobra.Command, _ []string) error {
		if err := required(name, "--name"); err != nil {
			return err
		}
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		value, err := store.AddCategory(database, name)
		if err != nil {
			return err
		}
		return output(command, value, false)
	}}
	add.Flags().StringVar(&name, "name", "", "category name")
	var asJSON bool
	list := &cobra.Command{Use: "list", RunE: func(command *cobra.Command, _ []string) error {
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		values, err := store.ListCategories(database)
		if err != nil {
			return err
		}
		return output(command, values, asJSON)
	}}
	list.Flags().BoolVar(&asJSON, "json", false, "emit JSON")
	command.AddCommand(add, list)
	return command
}

func itemCommand() *cobra.Command {
	command := &cobra.Command{Use: "item", Short: "Manage roadmap items"}
	var title, description, category, horizon, rationale string
	var epicID int64
	add := &cobra.Command{Use: "add", RunE: func(command *cobra.Command, _ []string) error {
		for _, field := range []struct{ v, n string }{{title, "--title"}, {category, "--category"}, {horizon, "--horizon"}} {
			if err := required(field.v, field.n); err != nil {
				return err
			}
		}
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		if command.Flags().Changed("epic") {
			if epicID < 1 {
				return fmt.Errorf("invalid Epic ID %d", epicID)
			}
			if _, err := store.GetEpic(database, epicID); err != nil {
				return err
			}
		}
		value, err := store.AddItem(database, title, description, category, horizon, rationale)
		if err != nil {
			return err
		}
		if command.Flags().Changed("epic") {
			value, err = store.SetItemEpic(database, value.ID, &epicID)
			if err != nil {
				return err
			}
		}
		return output(command, value, false)
	}}
	add.Flags().StringVar(&title, "title", "", "item title")
	add.Flags().StringVar(&description, "description", "", "item description")
	add.Flags().StringVar(&category, "category", "", "category name")
	add.Flags().StringVar(&horizon, "horizon", "", "Now, Next, or Later")
	add.Flags().StringVar(&rationale, "rationale", "", "item rationale")
	add.Flags().Int64Var(&epicID, "epic", 0, "optional Epic ID")
	var asJSON bool
	list := &cobra.Command{Use: "list", RunE: func(command *cobra.Command, _ []string) error {
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		values, err := store.ListItems(database)
		if err != nil {
			return err
		}
		return output(command, values, asJSON)
	}}
	list.Flags().BoolVar(&asJSON, "json", false, "emit JSON")
	show := showItemCommand()
	update := updateItemCommand()
	ready := itemTransitionCommand("ready", "Ready")
	start := itemTransitionCommand("start", "In Progress")
	complete := itemTransitionCommand("complete", "Done")
	cancel := itemTransitionCommand("cancel", "Won’t Do")
	next := itemNextCommand()
	command.AddCommand(add, list, show, update, itemMetadataCommand(), relationshipCommand(), changeLinkCommand(), artifactCommand(), stateCommand(), lifecycleCommand(), ready, start, complete, cancel, next)
	return command
}

func itemNextCommand() *cobra.Command {
	var asJSON bool
	command := &cobra.Command{Use: "next", Short: "Show the next actionable implementation task", RunE: func(command *cobra.Command, _ []string) error {
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		selection, err := store.SelectNextTask(database)
		if err != nil {
			return err
		}
		if asJSON {
			return output(command, selection, true)
		}
		if selection == nil {
			_, err = fmt.Fprintln(command.OutOrStdout(), "No actionable work found.")
			return err
		}
		_, err = fmt.Fprintf(command.OutOrStdout(), "Item %d %q; task %d %q (%s); next: %s\n", selection.ItemID, selection.ItemTitle, selection.TaskID, selection.TaskTitle, strings.ToLower(selection.TaskStatus), selection.ScopedContextCommand)
		return err
	}}
	command.Flags().BoolVar(&asJSON, "json", false, "emit JSON")
	return command
}
func itemMetadataCommand() *cobra.Command {
	var targetDate, priority, complexity, team, lead, summary, specs, links string
	var progress int
	command := &cobra.Command{Use: "metadata ID", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, args []string) error {
		value, err := id(args[0])
		if err != nil {
			return err
		}
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		item, err := store.UpdateItemMetadata(database, value, targetDate, progress, priority, complexity, team, lead, summary, specs, links)
		if err != nil {
			return err
		}
		return output(command, item, false)
	}}
	command.Flags().StringVar(&targetDate, "target-date", "", "YYYY-MM-DD target date")
	command.Flags().IntVar(&progress, "progress", 0, "completion percent")
	command.Flags().StringVar(&priority, "priority", "Medium", "High, Medium, or Low")
	command.Flags().StringVar(&complexity, "complexity", "Medium", "High, Medium, or Low")
	command.Flags().StringVar(&team, "team", "", "owning team")
	command.Flags().StringVar(&lead, "lead", "", "lead engineer")
	command.Flags().StringVar(&summary, "technical-summary", "", "technical summary")
	command.Flags().StringVar(&specs, "specifications", "[]", "JSON checklist")
	command.Flags().StringVar(&links, "documentation-links", "[]", "JSON documentation links")
	return command
}
func showItemCommand() *cobra.Command {
	var asJSON bool
	command := &cobra.Command{Use: "show ID", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, args []string) error {
		value, err := id(args[0])
		if err != nil {
			return err
		}
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		item, err := store.GetItem(database, value)
		if err != nil {
			return err
		}
		return output(command, item, asJSON)
	}}
	command.Flags().BoolVar(&asJSON, "json", false, "emit JSON")
	return command
}
func updateItemCommand() *cobra.Command {
	var title, description, category, horizon, rationale, featureType string
	var epicID int64
	var clearEpic bool
	command := &cobra.Command{Use: "update ID", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, args []string) error {
		value, err := id(args[0])
		if err != nil {
			return err
		}
		var t, d, c, h, r, ft *string
		if command.Flags().Changed("title") {
			t = &title
		}
		if command.Flags().Changed("description") {
			d = &description
		}
		if command.Flags().Changed("category") {
			c = &category
		}
		if command.Flags().Changed("horizon") {
			h = &horizon
		}
		if command.Flags().Changed("rationale") {
			r = &rationale
		}
		if command.Flags().Changed("feature-type") {
			ft = &featureType
		}
		if command.Flags().Changed("epic") && clearEpic {
			return fmt.Errorf("use either --epic or --clear-epic")
		}
		if t == nil && d == nil && c == nil && h == nil && r == nil && ft == nil && !command.Flags().Changed("epic") && !clearEpic {
			return fmt.Errorf("provide an item field to update")
		}
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		if command.Flags().Changed("epic") && epicID < 1 {
			return fmt.Errorf("invalid Epic ID %d", epicID)
		}
		item, err := store.GetItem(database, value)
		if t != nil || d != nil || c != nil || h != nil || r != nil {
			item, err = store.UpdateItem(database, value, t, d, c, h, r)
			if err != nil {
				return err
			}
		}
		if ft != nil {
			item, err = store.UpdateItemDossier(database, value, ft, nil)
			if err != nil {
				return err
			}
		}
		if command.Flags().Changed("epic") {
			item, err = store.SetItemEpic(database, value, &epicID)
			if err != nil {
				return err
			}
		}
		if clearEpic {
			item, err = store.SetItemEpic(database, value, nil)
			if err != nil {
				return err
			}
		}
		return output(command, item, false)
	}}
	command.Flags().StringVar(&title, "title", "", "item title")
	command.Flags().StringVar(&description, "description", "", "item description")
	command.Flags().StringVar(&category, "category", "", "category name")
	command.Flags().StringVar(&horizon, "horizon", "", "Now, Next, or Later")
	command.Flags().StringVar(&rationale, "rationale", "", "item rationale")
	command.Flags().StringVar(&featureType, "feature-type", "", "Capability, Change, or Gap")
	command.Flags().Int64Var(&epicID, "epic", 0, "assign to Epic ID")
	command.Flags().BoolVar(&clearEpic, "clear-epic", false, "remove the Epic parent")
	return command
}

func epicCommand() *cobra.Command {
	command := &cobra.Command{Use: "epic", Short: "Manage optional Feature groups"}
	var title, description string
	add := &cobra.Command{Use: "add", RunE: func(command *cobra.Command, _ []string) error {
		if err := required(title, "--title"); err != nil {
			return err
		}
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		value, err := store.AddEpic(database, title, description)
		if err != nil {
			return err
		}
		return output(command, value, false)
	}}
	add.Flags().StringVar(&title, "title", "", "Epic title")
	add.Flags().StringVar(&description, "description", "", "Epic description")
	var asJSON bool
	list := &cobra.Command{Use: "list", RunE: func(command *cobra.Command, _ []string) error {
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		values, err := store.ListEpics(database)
		if err != nil {
			return err
		}
		return output(command, values, asJSON)
	}}
	list.Flags().BoolVar(&asJSON, "json", false, "emit JSON")
	show := &cobra.Command{Use: "show ID", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, args []string) error {
		value, err := id(args[0])
		if err != nil {
			return err
		}
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		epic, err := store.GetEpic(database, value)
		if err != nil {
			return err
		}
		if asJSON {
			features, err := store.ListEpicItems(database, value)
			if err != nil {
				return err
			}
			return output(command, struct {
				store.Epic
				Features []store.Item `json:"features"`
			}{Epic: epic, Features: features}, true)
		}
		return output(command, epic, false)
	}}
	show.Flags().BoolVar(&asJSON, "json", false, "emit JSON")
	var updateTitle, updateDescription string
	update := &cobra.Command{Use: "update ID", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, args []string) error {
		value, err := id(args[0])
		if err != nil {
			return err
		}
		var t, d *string
		if command.Flags().Changed("title") {
			t = &updateTitle
		}
		if command.Flags().Changed("description") {
			d = &updateDescription
		}
		if t == nil && d == nil {
			return fmt.Errorf("provide an Epic field to update")
		}
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		epic, err := store.UpdateEpic(database, value, t, d)
		if err != nil {
			return err
		}
		return output(command, epic, false)
	}}
	update.Flags().StringVar(&updateTitle, "title", "", "Epic title")
	update.Flags().StringVar(&updateDescription, "description", "", "Epic description")
	remove := &cobra.Command{Use: "delete ID", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, args []string) error {
		value, err := id(args[0])
		if err != nil {
			return err
		}
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		return store.DeleteEpic(database, value)
	}}
	command.AddCommand(add, list, show, update, remove)
	return command
}

func changeLinkCommand() *cobra.Command {
	command := &cobra.Command{Use: "change-link", Short: "Link delivery changes to enduring capabilities"}
	var changeID, capabilityID int64
	add := &cobra.Command{Use: "add", RunE: func(command *cobra.Command, _ []string) error {
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		value, err := store.AddFeatureChangeLink(database, changeID, capabilityID)
		if err != nil {
			return err
		}
		return output(command, value, false)
	}}
	add.Flags().Int64Var(&changeID, "change", 0, "Change item ID")
	add.Flags().Int64Var(&capabilityID, "capability", 0, "Capability item ID")
	_ = add.MarkFlagRequired("change")
	_ = add.MarkFlagRequired("capability")
	var itemID int64
	var asJSON bool
	list := &cobra.Command{Use: "list", RunE: func(command *cobra.Command, _ []string) error {
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		values, err := store.ListFeatureChangeLinks(database, itemID)
		if err != nil {
			return err
		}
		return output(command, values, asJSON)
	}}
	list.Flags().Int64Var(&itemID, "item", 0, "feature item ID")
	list.Flags().BoolVar(&asJSON, "json", false, "emit JSON")
	_ = list.MarkFlagRequired("item")
	command.AddCommand(add, list)
	return command
}

func artifactCommand() *cobra.Command {
	command := &cobra.Command{Use: "artifact", Short: "Manage attributed dossier findings"}
	var itemID int64
	var kind, authorRole, summary, evidence string
	var supersedesID int64
	add := &cobra.Command{Use: "add", RunE: func(command *cobra.Command, _ []string) error {
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		var supersedes *int64
		if supersedesID > 0 {
			supersedes = &supersedesID
		}
		value, err := store.AddDossierArtifact(database, itemID, kind, authorRole, summary, evidence, supersedes)
		if err != nil {
			return err
		}
		return output(command, value, false)
	}}
	add.Flags().Int64Var(&itemID, "item", 0, "feature item ID")
	add.Flags().StringVar(&kind, "kind", "", "artifact kind")
	add.Flags().StringVar(&authorRole, "author-role", "", "contributing role")
	add.Flags().StringVar(&summary, "summary", "", "compact finding")
	add.Flags().StringVar(&evidence, "evidence", "", "supporting evidence")
	add.Flags().Int64Var(&supersedesID, "supersedes", 0, "prior artifact ID")
	for _, flag := range []string{"item", "kind", "author-role", "summary"} {
		_ = add.MarkFlagRequired(flag)
	}
	var artifactID int64
	var acceptedBy string
	accept := &cobra.Command{Use: "accept", RunE: func(command *cobra.Command, _ []string) error {
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		if err := store.AcceptDossierArtifact(database, artifactID, acceptedBy); err != nil {
			return err
		}
		_, err = fmt.Fprintf(command.OutOrStdout(), "Artifact %d accepted\n", artifactID)
		return err
	}}
	accept.Flags().Int64Var(&artifactID, "artifact", 0, "artifact ID")
	accept.Flags().StringVar(&acceptedBy, "by", "", "accepting orchestrator")
	_ = accept.MarkFlagRequired("artifact")
	_ = accept.MarkFlagRequired("by")
	var asJSON bool
	list := &cobra.Command{Use: "list", RunE: func(command *cobra.Command, _ []string) error {
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		values, err := store.ListDossierArtifactsForItem(database, itemID)
		if err != nil {
			return err
		}
		return output(command, values, asJSON)
	}}
	list.Flags().Int64Var(&itemID, "item", 0, "feature item ID")
	list.Flags().BoolVar(&asJSON, "json", false, "emit JSON")
	_ = list.MarkFlagRequired("item")
	command.AddCommand(add, accept, list)
	return command
}

func stateCommand() *cobra.Command {
	command := &cobra.Command{Use: "state", Short: "Accept capability product state updates"}
	var capabilityID, sourceChangeID int64
	var state, acceptedBy string
	accept := &cobra.Command{Use: "accept", RunE: func(command *cobra.Command, _ []string) error {
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		if err := store.AcceptCapabilityState(database, capabilityID, sourceChangeID, state, acceptedBy); err != nil {
			return err
		}
		_, err = fmt.Fprintf(command.OutOrStdout(), "Capability %d state accepted\n", capabilityID)
		return err
	}}
	accept.Flags().Int64Var(&capabilityID, "capability", 0, "Capability item ID")
	accept.Flags().Int64Var(&sourceChangeID, "source-change", 0, "completed Change item ID")
	accept.Flags().StringVar(&state, "state", "", "accepted current product state")
	accept.Flags().StringVar(&acceptedBy, "by", "", "accepting orchestrator")
	for _, flag := range []string{"capability", "source-change", "state", "by"} {
		_ = accept.MarkFlagRequired(flag)
	}
	var asJSON bool
	list := &cobra.Command{Use: "list", RunE: func(command *cobra.Command, _ []string) error {
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		values, err := store.ListCapabilityStateHistoryForItem(database, capabilityID)
		if err != nil {
			return err
		}
		return output(command, values, asJSON)
	}}
	list.Flags().Int64Var(&capabilityID, "capability", 0, "Capability item ID")
	list.Flags().BoolVar(&asJSON, "json", false, "emit JSON")
	_ = list.MarkFlagRequired("capability")
	command.AddCommand(accept, list)
	return command
}

func relationshipCommand() *cobra.Command {
	command := &cobra.Command{Use: "relationship", Short: "Manage directed feature relationships"}
	var source, target int64
	var relationshipType string
	add := &cobra.Command{Use: "add", RunE: func(command *cobra.Command, _ []string) error {
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		value, err := store.AddFeatureRelationship(database, source, target, relationshipType)
		if err != nil {
			return err
		}
		return output(command, value, false)
	}}
	add.Flags().Int64Var(&source, "from", 0, "source feature ID")
	add.Flags().Int64Var(&target, "to", 0, "target feature ID")
	add.Flags().StringVar(&relationshipType, "type", "", "extends, depends_on, or replaces")
	_ = add.MarkFlagRequired("from")
	_ = add.MarkFlagRequired("to")
	_ = add.MarkFlagRequired("type")
	var itemID int64
	var asJSON bool
	list := &cobra.Command{Use: "list", RunE: func(command *cobra.Command, _ []string) error {
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		values, err := store.ListFeatureRelationships(database, itemID)
		if err != nil {
			return err
		}
		return output(command, values, asJSON)
	}}
	list.Flags().Int64Var(&itemID, "item", 0, "feature ID")
	list.Flags().BoolVar(&asJSON, "json", false, "emit JSON")
	_ = list.MarkFlagRequired("item")
	command.AddCommand(add, list)
	return command
}
func itemTransitionCommand(use, status string) *cobra.Command {
	return &cobra.Command{Use: use + " ID", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, args []string) error {
		value, err := id(args[0])
		if err != nil {
			return err
		}
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		if err := store.TransitionItemStatus(database, value, status); err != nil {
			return err
		}
		_, err = fmt.Fprintf(command.OutOrStdout(), "Roadmap item %d %s\n", value, strings.ToLower(status))
		return err
	}}
}

func planCommand() *cobra.Command {
	command := &cobra.Command{Use: "plan", Short: "Manage revisioned implementation plans"}
	var itemID int64
	var content string
	create := &cobra.Command{Use: "create", RunE: func(command *cobra.Command, _ []string) error {
		if err := required(content, "--content"); err != nil {
			return err
		}
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		value, err := store.CreatePlan(database, itemID, content)
		if err != nil {
			return err
		}
		return output(command, value, false)
	}}
	create.Flags().Int64Var(&itemID, "item", 0, "roadmap item ID")
	create.Flags().StringVar(&content, "content", "", "plan content")
	var revisionContent string
	revise := &cobra.Command{Use: "revise PLAN_ID", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, args []string) error {
		value, err := id(args[0])
		if err != nil {
			return err
		}
		if err := required(revisionContent, "--content"); err != nil {
			return err
		}
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		plan, err := store.RevisePlan(database, value, revisionContent)
		if err != nil {
			return err
		}
		return output(command, plan, false)
	}}
	revise.Flags().StringVar(&revisionContent, "content", "", "revised plan content")
	var asJSON bool
	show := &cobra.Command{Use: "show ID", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, args []string) error {
		value, err := id(args[0])
		if err != nil {
			return err
		}
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		plan, err := store.GetPlan(database, value)
		if err != nil {
			return err
		}
		tasks, err := store.ListTasks(database, value)
		if err != nil {
			return err
		}
		return output(command, struct {
			Plan  store.Plan   `json:"plan"`
			Tasks []store.Task `json:"tasks"`
		}{plan, tasks}, asJSON)
	}}
	show.Flags().BoolVar(&asJSON, "json", false, "emit JSON")
	template := planTemplateCommand()
	validate := planValidateCommand()
	approve := &cobra.Command{Use: "approve ID", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, args []string) error {
		value, err := id(args[0])
		if err != nil {
			return err
		}
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		if err := store.ApprovePlan(database, value); err != nil {
			return err
		}
		_, err = fmt.Fprintf(command.OutOrStdout(), "Plan %d approved\n", value)
		return err
	}}
	command.AddCommand(create, show, revise, template, validate, approve, recordPlanCommand())
	return command
}

func taskCommand() *cobra.Command {
	command := &cobra.Command{Use: "task", Short: "Manage approved-plan tasks"}
	var planID int64
	var title, description string
	var verification []string
	add := &cobra.Command{Use: "add", RunE: func(command *cobra.Command, _ []string) error {
		if err := required(title, "--title"); err != nil {
			return err
		}
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		value, err := store.AddTask(database, planID, title, description, verification)
		if err != nil {
			return err
		}
		return output(command, value, false)
	}}
	add.Flags().Int64Var(&planID, "plan", 0, "plan ID")
	add.Flags().StringVar(&title, "title", "", "task title")
	add.Flags().StringVar(&description, "description", "", "task description")
	add.Flags().StringSliceVar(&verification, "verification", nil, "required verification command or evidence reference (repeatable)")
	var listPlan int64
	var asJSON bool
	list := &cobra.Command{Use: "list", RunE: func(command *cobra.Command, _ []string) error {
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		values, err := store.ListTasks(database, listPlan)
		if err != nil {
			return err
		}
		return output(command, values, asJSON)
	}}
	list.Flags().Int64Var(&listPlan, "plan", 0, "plan ID")
	list.Flags().BoolVar(&asJSON, "json", false, "emit JSON")
	var showJSON bool
	show := &cobra.Command{Use: "show ID", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, args []string) error {
		value, err := id(args[0])
		if err != nil {
			return err
		}
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		task, err := store.GetTask(database, value)
		if err != nil {
			return err
		}
		return output(command, task, showJSON)
	}}
	show.Flags().BoolVar(&showJSON, "json", false, "emit JSON")
	start := taskTransitionCommand("start", "In Progress", false)
	complete := taskTransitionCommand("complete", "Done", true)
	block := taskTransitionCommand("block", "Blocked", true)
	resume := &cobra.Command{Use: "resume ID", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, args []string) error {
		value, err := id(args[0])
		if err != nil {
			return err
		}
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		if err := store.ResumeTask(database, value); err != nil {
			return err
		}
		task, err := store.GetTask(database, value)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(command.OutOrStdout(), "Task %d resumed; next: cassor task complete %d --outcome <result>\n", value, task.ID)
		return err
	}}
	command.AddCommand(add, list, show, start, complete, block, resume)
	return command
}
func taskTransitionCommand(use, status string, needsOutcome bool) *cobra.Command {
	var outcome string
	command := &cobra.Command{Use: use + " ID", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, args []string) error {
		value, err := id(args[0])
		if err != nil {
			return err
		}
		if needsOutcome {
			if err := required(outcome, "--outcome"); err != nil {
				return err
			}
		}
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		if err := store.SetTaskStatus(database, value, status, outcome); err != nil {
			return err
		}
		task, err := store.GetTask(database, value)
		if err != nil {
			return err
		}
		plan, err := store.GetPlan(database, task.PlanID)
		if err != nil {
			return err
		}
		next := fmt.Sprintf("cassor task show %d", value)
		if status == "In Progress" {
			next = fmt.Sprintf("cassor task complete %d --outcome <result>", value)
		} else if status == "Blocked" {
			next = fmt.Sprintf("cassor task resume %d", value)
		} else if status == "Done" {
			next = fmt.Sprintf("cassor context --item %d --role verification --max-bytes 12000", plan.ItemID)
		}
		_, err = fmt.Fprintf(command.OutOrStdout(), "Task %d %s; next: %s\n", value, strings.ToLower(status), next)
		return err
	}}
	if needsOutcome {
		command.Flags().StringVar(&outcome, "outcome", "", "completion or blocking outcome")
	}
	return command
}
