package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunReportMarksUnavailableTokenMetrics(t *testing.T) {
	command := newRunReportCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"--mode", "parallel", "--roles", "discovery,verification", "--elapsed", "42s", "--tool-calls", "5", "--verification", "passed"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Tokens: input=unavailable") {
		t.Fatalf("report = %q", output.String())
	}
}

func TestRunReportJSONIncludesObservedTokens(t *testing.T) {
	command := newRunReportCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"--mode", "single-agent", "--elapsed", "1m", "--input-tokens", "120", "--total-tokens", "200", "--json"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "\"input_tokens\": 120") || !strings.Contains(output.String(), "\"total_tokens\": 200") {
		t.Fatalf("json report = %q", output.String())
	}
}
