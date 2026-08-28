package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommandHelp(t *testing.T) {
	t.Parallel()

	command := NewRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"--help"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(output.String(), "Plan, approve, and track repository work") {
		t.Fatalf("help output = %q, want command description", output.String())
	}
}
