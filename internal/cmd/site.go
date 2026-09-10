package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/rickcern44/cassor/internal/project"
	"github.com/rickcern44/cassor/internal/site"
	"github.com/rickcern44/cassor/internal/store"
)

func newContextCommand() *cobra.Command {
	var asJSON bool
	var itemID int64
	var role string
	var maxBytes int
	command := &cobra.Command{Use: "context", Short: "Show compact active Cassor context", RunE: func(command *cobra.Command, _ []string) error {
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		if itemID > 0 {
			context, err := store.BuildScopedContext(database, itemID, role, maxBytes)
			if err != nil {
				return err
			}
			if asJSON {
				encoded, err := json.Marshal(context)
				if err != nil {
					return err
				}
				_, err = command.OutOrStdout().Write(encoded)
				return err
			}
			_, err = fmt.Fprintf(command.OutOrStdout(), "Item %d · %s\nPlan %d revision %d · %s\nNext: %s\nContext bytes: %d/%d · truncated: %t\n", context.Item.ID, context.Item.Title, context.Assignment.PlanID, context.Assignment.Revision, context.Role, context.NextAction.Command, context.Bytes, context.MaxBytes, context.Truncated)
			return err
		}
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
	command.Flags().Int64Var(&itemID, "item", 0, "roadmap item ID for scoped context")
	command.Flags().StringVar(&role, "role", "implementation", "scoped role: implementation, verification, or planning")
	command.Flags().IntVar(&maxBytes, "max-bytes", 12000, "maximum encoded bytes for scoped context")
	return command
}

func newSiteCommand() *cobra.Command {
	command := &cobra.Command{Use: "site", Short: "Build and preview the static developer documentation"}
	command.AddCommand(&cobra.Command{Use: "build", Short: "Generate the static multi-page documentation site", RunE: func(command *cobra.Command, _ []string) error {
		root, state, err := currentProject()
		if err != nil {
			return err
		}
		output, err := site.Generate(root, state)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(command.OutOrStdout(), "Generated %s\n", output)
		return err
	}})
	var address string
	var watch bool
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
		if !watch {
			return http.ListenAndServe(address, http.FileServer(http.Dir(directory)))
		}
		var clients sync.Map
		go watchRoadmap(root, filepath.Join(root, ".cassor", "cassor.db"), func() {
			clients.Range(func(key, _ any) bool {
				select {
				case key.(chan struct{}) <- struct{}{}:
				default:
				}
				return true
			})
		})
		mux := http.NewServeMux()
		mux.Handle("/", http.FileServer(http.Dir(directory)))
		mux.HandleFunc("/__cassor_reload", func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "text/event-stream")
			writer.Header().Set("Cache-Control", "no-cache")
			writer.Header().Set("Connection", "keep-alive")
			flusher, ok := writer.(http.Flusher)
			if !ok {
				http.Error(writer, "streaming unsupported", http.StatusInternalServerError)
				return
			}
			channel := make(chan struct{}, 1)
			clients.Store(channel, struct{}{})
			defer clients.Delete(channel)
			fmt.Fprint(writer, "data: connected\n\n")
			flusher.Flush()
			select {
			case <-request.Context().Done():
			case <-channel:
				fmt.Fprint(writer, "data: reload\n\n")
				flusher.Flush()
			}
		})
		return http.ListenAndServe(address, mux)
	}}
	serve.Flags().StringVar(&address, "addr", "127.0.0.1:8080", "listen address")
	serve.Flags().BoolVar(&watch, "watch", false, "rebuild and reload local browsers when Cassor state changes")
	command.AddCommand(serve)
	return command
}

func watchRoadmap(root, database string, reload func()) {
	var previous time.Time
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		info, err := os.Stat(database)
		if err != nil || !info.ModTime().After(previous) {
			continue
		}
		previous = info.ModTime()
		if state, err := project.FindStateDirectory(root); err == nil {
			if _, err := site.Generate(root, state); err == nil {
				reload()
			}
		}
	}
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
