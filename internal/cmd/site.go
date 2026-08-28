package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/rickcern44/cassor/internal/project"
	"github.com/rickcern44/cassor/internal/site"
	"github.com/rickcern44/cassor/internal/store"
)

func newContextCommand() *cobra.Command {
	var asJSON bool
	command := &cobra.Command{Use: "context", Short: "Show compact active Cassor context", RunE: func(command *cobra.Command, _ []string) error {
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		context, err := store.CompactContext(database)
		if err != nil {
			return err
		}
		if asJSON {
			encoder := json.NewEncoder(command.OutOrStdout())
			encoder.SetIndent("", "  ")
			return encoder.Encode(context)
		}
		_, err = fmt.Fprintf(command.OutOrStdout(), "Active tasks: %d\nProposed items: %d\nPlans awaiting approval: %d\nBlocked tasks: %d\nRecently completed tasks: %d\n", len(context.ActiveTasks), len(context.ProposedItems), len(context.PlansAwaiting), len(context.BlockedTasks), len(context.RecentlyCompleted))
		return err
	}}
	command.Flags().BoolVar(&asJSON, "json", false, "emit JSON")
	return command
}

func newSiteCommand() *cobra.Command {
	command := &cobra.Command{Use: "site", Short: "Build and preview the static roadmap"}
	command.AddCommand(&cobra.Command{Use: "build", Short: "Generate docs/roadmap/index.html", RunE: func(command *cobra.Command, _ []string) error {
		root, state, err := currentProject()
		if err != nil {
			return err
		}
		output, err := site.Build(root, state)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(command.OutOrStdout(), "Generated %s\n", output)
		return err
	}})
	var address string
	serve := &cobra.Command{Use: "serve", Short: "Serve the generated roadmap locally", RunE: func(command *cobra.Command, _ []string) error {
		root, _, err := currentProject()
		if err != nil {
			return err
		}
		directory := filepath.Join(root, "docs", "roadmap")
		if _, err := os.Stat(filepath.Join(directory, "index.html")); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("roadmap has not been generated; run cassor site build first")
			}
			return err
		}
		_, err = fmt.Fprintf(command.OutOrStdout(), "Serving roadmap at http://%s\n", address)
		if err != nil {
			return err
		}
		return http.ListenAndServe(address, http.FileServer(http.Dir(directory)))
	}}
	serve.Flags().StringVar(&address, "addr", "127.0.0.1:8080", "listen address")
	command.AddCommand(serve)
	return command
}

func currentProject() (string, string, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	state, err := project.FindStateDirectory(workingDirectory)
	if err != nil {
		return "", "", err
	}
	return filepath.Dir(state), state, nil
}
