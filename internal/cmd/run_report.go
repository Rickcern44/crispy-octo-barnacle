package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rickcern44/cassor/internal/store"
	"github.com/spf13/cobra"
)

// RunReport is an ephemeral, caller-supplied summary of one orchestration run.
// It is intentionally not stored in Cassor state.
type RunReport struct {
	ExecutionMode  string        `json:"execution_mode"`
	Roles          []string      `json:"roles"`
	Elapsed        time.Duration `json:"elapsed"`
	ToolCalls      int           `json:"tool_calls"`
	Verification   string        `json:"verification"`
	ComparisonKey  *string       `json:"comparison_key"`
	ContextPackets *int64        `json:"context_packets"`
	ContextBytes   *int64        `json:"context_bytes"`
	Handoffs       *int64        `json:"handoffs"`
	Tokens         TokenUsage    `json:"tokens"`
}

// TokenUsage records only metrics exposed by the runtime. A nil field means
// unavailable rather than zero tokens.
type TokenUsage struct {
	Input     *int64 `json:"input_tokens"`
	Cached    *int64 `json:"cached_input_tokens"`
	Output    *int64 `json:"output_tokens"`
	Reasoning *int64 `json:"reasoning_tokens"`
	Total     *int64 `json:"total_tokens"`
}

func newRunReportCommand() *cobra.Command {
	var mode, roles, elapsed, verification string
	var toolCalls int
	var asJSON bool
	var input, cached, outputTokens, reasoning, total optionalInt64
	var contextPackets, contextBytes, handoffs optionalInt64
	var comparisonKey string
	command := &cobra.Command{
		Use:   "run-report",
		Short: "Render a non-persistent orchestration run report",
		RunE: func(command *cobra.Command, _ []string) error {
			if err := required(mode, "--mode"); err != nil {
				return err
			}
			if err := required(elapsed, "--elapsed"); err != nil {
				return err
			}
			parsedElapsed, err := time.ParseDuration(elapsed)
			if err != nil || parsedElapsed < 0 {
				return fmt.Errorf("--elapsed must be a non-negative duration")
			}
			report := RunReport{
				ExecutionMode:  mode,
				Roles:          splitValues(roles),
				Elapsed:        parsedElapsed,
				ToolCalls:      toolCalls,
				Verification:   verification,
				ComparisonKey:  comparisonKeyPointer(comparisonKey),
				ContextPackets: contextPackets.asPointer(),
				ContextBytes:   contextBytes.asPointer(),
				Handoffs:       handoffs.asPointer(),
				Tokens:         TokenUsage{Input: input.asPointer(), Cached: cached.asPointer(), Output: outputTokens.asPointer(), Reasoning: reasoning.asPointer(), Total: total.asPointer()},
			}
			if asJSON {
				return output(command, report, true)
			}
			return writeRunReport(command, report)
		},
	}
	command.Flags().StringVar(&mode, "mode", "", "single-agent, sequential, or parallel")
	command.Flags().StringVar(&roles, "roles", "", "comma-separated dispatched roles")
	command.Flags().StringVar(&elapsed, "elapsed", "", "elapsed duration, for example 2m15s")
	command.Flags().IntVar(&toolCalls, "tool-calls", 0, "observed tool or command calls")
	command.Flags().StringVar(&verification, "verification", "not-run", "passed, failed, or not-run")
	command.Flags().StringVar(&comparisonKey, "comparison-key", "", "non-empty label for a comparable run set")
	command.Flags().Var(&contextPackets, "context-packets", "observed context payload count")
	command.Flags().Var(&contextBytes, "context-bytes", "observed context payload bytes")
	command.Flags().Var(&handoffs, "handoffs", "observed completed orchestrator-to-worker handoffs")
	command.Flags().Var(&input, "input-tokens", "runtime-reported input tokens")
	command.Flags().Var(&cached, "cached-input-tokens", "runtime-reported cached input tokens")
	command.Flags().Var(&outputTokens, "output-tokens", "runtime-reported output tokens")
	command.Flags().Var(&reasoning, "reasoning-tokens", "runtime-reported reasoning tokens")
	command.Flags().Var(&total, "total-tokens", "runtime-reported total tokens")
	command.Flags().BoolVar(&asJSON, "json", false, "emit JSON")
	command.AddCommand(newRecordRunReportCommand())
	command.AddCommand(newCompareRunReportCommand())
	return command
}

func newRecordRunReportCommand() *cobra.Command {
	var itemID int64
	var mode, roles, elapsed, verification string
	var toolCalls int
	var input, cached, outputTokens, reasoning, total optionalInt64
	var contextPackets, contextBytes, handoffs optionalInt64
	var comparisonKey string
	command := &cobra.Command{Use: "record", Short: "Persist an observed report against a roadmap feature", RunE: func(command *cobra.Command, _ []string) error {
		if itemID < 1 {
			return fmt.Errorf("--item is required")
		}
		if err := required(mode, "--mode"); err != nil {
			return err
		}
		parsed, err := time.ParseDuration(elapsed)
		if err != nil || parsed < 0 {
			return fmt.Errorf("--elapsed must be a non-negative duration")
		}
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		value, err := store.AddFeatureReport(database, store.FeatureReport{ItemID: itemID, ExecutionMode: mode, Roles: splitValues(roles), ElapsedNS: parsed.Nanoseconds(), ToolCalls: toolCalls, Verification: verification, InputTokens: input.asPointer(), CachedInputTokens: cached.asPointer(), OutputTokens: outputTokens.asPointer(), ReasoningTokens: reasoning.asPointer(), TotalTokens: total.asPointer(), ComparisonKey: comparisonKeyPointer(comparisonKey), ContextPackets: contextPackets.asPointer(), ContextBytes: contextBytes.asPointer(), Handoffs: handoffs.asPointer()})
		if err != nil {
			return err
		}
		return output(command, value, false)
	}}
	command.Flags().Int64Var(&itemID, "item", 0, "roadmap feature ID")
	command.Flags().StringVar(&mode, "mode", "", "single-agent, sequential, or parallel")
	command.Flags().StringVar(&roles, "roles", "", "comma-separated dispatched roles")
	command.Flags().StringVar(&elapsed, "elapsed", "", "elapsed duration")
	command.Flags().IntVar(&toolCalls, "tool-calls", 0, "observed tool or command calls")
	command.Flags().StringVar(&verification, "verification", "not-run", "passed, failed, or not-run")
	command.Flags().StringVar(&comparisonKey, "comparison-key", "", "non-empty label for a comparable run set")
	command.Flags().Var(&contextPackets, "context-packets", "observed context payload count")
	command.Flags().Var(&contextBytes, "context-bytes", "observed context payload bytes")
	command.Flags().Var(&handoffs, "handoffs", "observed completed orchestrator-to-worker handoffs")
	command.Flags().Var(&input, "input-tokens", "runtime-reported input tokens")
	command.Flags().Var(&cached, "cached-input-tokens", "runtime-reported cached input tokens")
	command.Flags().Var(&outputTokens, "output-tokens", "runtime-reported output tokens")
	command.Flags().Var(&reasoning, "reasoning-tokens", "runtime-reported reasoning tokens")
	command.Flags().Var(&total, "total-tokens", "runtime-reported total tokens")
	return command
}

type comparisonReportRow struct {
	ReportID          int64         `json:"report_id"`
	ItemID            int64         `json:"item_id"`
	ItemTitle         string        `json:"item_title"`
	ExecutionMode     string        `json:"execution_mode"`
	Roles             []string      `json:"roles"`
	ContextPackets    *int64        `json:"context_packets"`
	ContextBytes      *int64        `json:"context_bytes"`
	Handoffs          *int64        `json:"handoffs"`
	Elapsed           time.Duration `json:"elapsed"`
	Verification      string        `json:"verification"`
	InputTokens       *int64        `json:"input_tokens"`
	CachedInputTokens *int64        `json:"cached_input_tokens"`
	OutputTokens      *int64        `json:"output_tokens"`
	ReasoningTokens   *int64        `json:"reasoning_tokens"`
	TotalTokens       *int64        `json:"total_tokens"`
}

func newCompareRunReportCommand() *cobra.Command {
	var comparisonKey string
	var asJSON bool
	command := &cobra.Command{Use: "compare", Aliases: []string{"comparison"}, Short: "Compare persisted reports sharing a comparison key", RunE: func(command *cobra.Command, _ []string) error {
		key := strings.TrimSpace(comparisonKey)
		if key == "" {
			return fmt.Errorf("--comparison-key is required and must be non-empty")
		}
		database, err := databaseForCommand()
		if err != nil {
			return err
		}
		defer database.Close()
		reports, err := store.ListFeatureReports(database)
		if err != nil {
			return err
		}
		rows := make([]comparisonReportRow, 0)
		for _, report := range reports {
			if report.ComparisonKey == nil || strings.TrimSpace(*report.ComparisonKey) != key {
				continue
			}
			item, err := store.GetItem(database, report.ItemID)
			if err != nil {
				return err
			}
			rows = append(rows, comparisonReportRow{ReportID: report.ID, ItemID: report.ItemID, ItemTitle: item.Title, ExecutionMode: report.ExecutionMode, Roles: report.Roles, ContextPackets: report.ContextPackets, ContextBytes: report.ContextBytes, Handoffs: report.Handoffs, Elapsed: time.Duration(report.ElapsedNS), Verification: report.Verification, InputTokens: report.InputTokens, CachedInputTokens: report.CachedInputTokens, OutputTokens: report.OutputTokens, ReasoningTokens: report.ReasoningTokens, TotalTokens: report.TotalTokens})
		}
		if asJSON {
			return output(command, rows, true)
		}
		return writeComparisonReports(command, rows)
	}}
	command.Flags().StringVar(&comparisonKey, "comparison-key", "", "required non-empty label to compare")
	command.Flags().BoolVar(&asJSON, "json", false, "emit JSON")
	return command
}

func writeComparisonReports(command *cobra.Command, rows []comparisonReportRow) error {
	if _, err := fmt.Fprintln(command.OutOrStdout(), "Item\tMode\tRoles\tContext packets\tContext bytes\tHandoffs\tElapsed\tVerification\tTokens (input/cached/output/reasoning/total)"); err != nil {
		return err
	}
	for _, row := range rows {
		if _, err := fmt.Fprintf(command.OutOrStdout(), "RM-%d %q\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s/%s/%s/%s/%s\n", row.ItemID, row.ItemTitle, row.ExecutionMode, valueList(row.Roles), optionalValue(row.ContextPackets), optionalValue(row.ContextBytes), optionalValue(row.Handoffs), row.Elapsed, row.Verification, tokenValue(row.InputTokens), tokenValue(row.CachedInputTokens), tokenValue(row.OutputTokens), tokenValue(row.ReasoningTokens), tokenValue(row.TotalTokens)); err != nil {
			return err
		}
	}
	return nil
}

func writeRunReport(command *cobra.Command, report RunReport) error {
	verification := report.Verification
	if verification == "" {
		verification = "not-run"
	}
	comparisonKey := "unavailable"
	if report.ComparisonKey != nil {
		comparisonKey = *report.ComparisonKey
	}
	_, err := fmt.Fprintf(command.OutOrStdout(), "Execution mode: %s\nRoles: %s\nElapsed: %s\nTool calls: %d\nVerification: %s\nComparison key: %s\nContext packets: %s\nContext bytes: %s\nHandoffs: %s\nTokens: input=%s cached=%s output=%s reasoning=%s total=%s\n", report.ExecutionMode, valueList(report.Roles), report.Elapsed, report.ToolCalls, verification, comparisonKey, optionalValue(report.ContextPackets), optionalValue(report.ContextBytes), optionalValue(report.Handoffs), tokenValue(report.Tokens.Input), tokenValue(report.Tokens.Cached), tokenValue(report.Tokens.Output), tokenValue(report.Tokens.Reasoning), tokenValue(report.Tokens.Total))
	return err
}

type optionalInt64 struct {
	set    bool
	number int64
}

func (value *optionalInt64) String() string { return strconv.FormatInt(value.number, 10) }
func (value *optionalInt64) Set(raw string) error {
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || parsed < 0 {
		return fmt.Errorf("must be a non-negative integer")
	}
	value.number, value.set = parsed, true
	return nil
}
func (value *optionalInt64) Type() string { return "integer" }
func (value *optionalInt64) asPointer() *int64 {
	if !value.set {
		return nil
	}
	result := value.number
	return &result
}

func splitValues(raw string) []string {
	values := []string{}
	for _, value := range strings.Split(raw, ",") {
		if value = strings.TrimSpace(value); value != "" {
			values = append(values, value)
		}
	}
	return values
}

func valueList(values []string) string {
	if len(values) == 0 {
		return "none"
	}
	return strings.Join(values, ", ")
}

func tokenValue(value *int64) string {
	if value == nil {
		return "unavailable"
	}
	return strconv.FormatInt(*value, 10)
}

func optionalValue(value *int64) string { return tokenValue(value) }

func comparisonKeyPointer(raw string) *string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil
	}
	return &value
}
