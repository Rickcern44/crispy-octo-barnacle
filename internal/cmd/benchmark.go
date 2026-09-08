package cmd

import (
	"encoding/json"

	"github.com/rickcern44/cassor/internal/benchmark"
	"github.com/spf13/cobra"
)

func newBenchmarkCommand() *cobra.Command {
	var asJSON bool
	command := &cobra.Command{Use: "benchmark", Short: "Run non-persistent workflow efficiency scenarios", RunE: func(command *cobra.Command, _ []string) error {
		report, err := benchmark.Run()
		if err != nil {
			return err
		}
		if asJSON {
			encoder := json.NewEncoder(command.OutOrStdout())
			encoder.SetIndent("", "  ")
			return encoder.Encode(report)
		}
		return output(command, report, false)
	}}
	command.Flags().BoolVar(&asJSON, "json", false, "emit JSON")
	return command
}
