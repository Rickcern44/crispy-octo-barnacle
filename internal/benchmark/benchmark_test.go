package benchmark

import "testing"

func TestRunProducesFourSuccessfulNonPersistentScenarios(t *testing.T) {
	report, err := Run()
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Scenarios) != 4 || report.ExecutionMode != "single-agent" {
		t.Fatalf("report = %#v", report)
	}
	for _, scenario := range report.Scenarios {
		if scenario.Outcome != "passed" || scenario.Operations == 0 || scenario.ElapsedNS <= 0 || scenario.RepairEffort != 0 || scenario.WorkerComparison != "unavailable" || scenario.TokenTelemetry != "unavailable" || scenario.RawTranscriptsStored {
			t.Fatalf("scenario = %#v", scenario)
		}
	}
}
