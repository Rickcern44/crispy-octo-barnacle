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
	root.AddCommand(categoryCommand(), itemCommand(), planCommand(), taskCommand())
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
		value, err := store.AddItem(database, title, description, category, horizon, rationale)
		if err != nil {
			return err
		}
		return output(command, value, false)
	}}
	add.Flags().StringVar(&title, "title", "", "item title")
	add.Flags().StringVar(&description, "description", "", "item description")
	add.Flags().StringVar(&category, "category", "", "category name")
	add.Flags().StringVar(&horizon, "horizon", "", "Now, Next, or Later")
	add.Flags().StringVar(&rationale, "rationale", "", "item rationale")
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
	approve := itemTransitionCommand("approve", "Approved")
	decline := itemTransitionCommand("decline", "Declined")
	command.AddCommand(add, list, show, update, approve, decline)
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
	var title, description, category, horizon, rationale string
	command := &cobra.Command{Use: "update ID", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, args []string) error {
		value, err := id(args[0])
		if err != nil {
			return err
		}
		var t, d, c, h, r *string
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
		if t == nil && d == nil && c == nil && h == nil && r == nil {
			return fmt.Errorf("provide an item field to update")
		}
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		item, err := store.UpdateItem(database, value, t, d, c, h, r)
		if err != nil {
			return err
		}
		return output(command, item, false)
	}}
	command.Flags().StringVar(&title, "title", "", "item title")
	command.Flags().StringVar(&description, "description", "", "item description")
	command.Flags().StringVar(&category, "category", "", "category name")
	command.Flags().StringVar(&horizon, "horizon", "", "Now, Next, or Later")
	command.Flags().StringVar(&rationale, "rationale", "", "item rationale")
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
		if err := store.SetItemStatus(database, value, status); err != nil {
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
	command.AddCommand(create, show, revise, approve, recordPlanCommand())
	return command
}

func taskCommand() *cobra.Command {
	command := &cobra.Command{Use: "task", Short: "Manage approved-plan tasks"}
	var planID int64
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
		value, err := store.AddTask(database, planID, title, description)
		if err != nil {
			return err
		}
		return output(command, value, false)
	}}
	add.Flags().Int64Var(&planID, "plan", 0, "plan ID")
	add.Flags().StringVar(&title, "title", "", "task title")
	add.Flags().StringVar(&description, "description", "", "task description")
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
	start := taskTransitionCommand("start", "Active", false)
	complete := taskTransitionCommand("complete", "Completed", true)
	block := taskTransitionCommand("block", "Blocked", true)
	command.AddCommand(add, list, show, start, complete, block)
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
		_, err = fmt.Fprintf(command.OutOrStdout(), "Task %d %s\n", value, strings.ToLower(status))
		return err
	}}
	if needsOutcome {
		command.Flags().StringVar(&outcome, "outcome", "", "completion or blocking outcome")
	}
	return command
}
