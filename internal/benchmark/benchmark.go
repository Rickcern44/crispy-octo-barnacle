// Package benchmark runs reproducible, non-persistent workflow scenarios.
package benchmark

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/rickcern44/cassor/internal/store"
)

type ScenarioResult struct {
	Name                 string `json:"name"`
	Kind                 string `json:"kind"`
	Outcome              string `json:"outcome"`
	Operations           int    `json:"observed_operations"`
	ElapsedNS            int64  `json:"elapsed_ns"`
	OutputBytes          int    `json:"output_bytes"`
	RepairEffort         int    `json:"repair_effort"`
	WorkerComparison     string `json:"worker_comparison"`
	TokenTelemetry       string `json:"token_telemetry"`
	RawTranscriptsStored bool   `json:"raw_transcripts_stored"`
}

type Report struct {
	Version         int              `json:"version"`
	ExecutionMode   string           `json:"execution_mode"`
	QualityGate     string           `json:"quality_gate"`
	TelemetryPolicy string           `json:"telemetry_policy"`
	Scenarios       []ScenarioResult `json:"scenarios"`
}

func Run() (Report, error) {
	scenarios := []struct {
		name string
		kind string
		run  func(*sql.DB) (int, int, error)
	}{
		{"small-fix", "small local change", runSmallFix},
		{"feature", "new feature", runFeature},
		{"amendment", "approved plan amendment", runAmendment},
		{"resumed-task", "fresh-session resumption", runResumedTask},
	}
	report := Report{Version: 1, ExecutionMode: "single-agent", QualityGate: "scenario completed without error", TelemetryPolicy: "worker comparison and token telemetry unavailable in this runtime; no estimates are emitted", Scenarios: []ScenarioResult{}}
	for _, scenario := range scenarios {
		database, cleanup, err := temporaryDatabase()
		if err != nil {
			return report, err
		}
		started := time.Now()
		operations, outputBytes, runErr := scenario.run(database)
		elapsed := time.Since(started)
		cleanup()
		result := ScenarioResult{Name: scenario.name, Kind: scenario.kind, Outcome: "passed", Operations: operations, ElapsedNS: elapsed.Nanoseconds(), OutputBytes: outputBytes, WorkerComparison: "unavailable", TokenTelemetry: "unavailable", RawTranscriptsStored: false}
		if runErr != nil {
			result.Outcome = runErr.Error()
			return report, fmt.Errorf("benchmark scenario %s: %w", scenario.name, runErr)
		}
		report.Scenarios = append(report.Scenarios, result)
	}
	return report, nil
}

func temporaryDatabase() (*sql.DB, func(), error) {
	file, err := os.CreateTemp("", "cassor-benchmark-*.db")
	if err != nil {
		return nil, nil, err
	}
	path := file.Name()
	if err := file.Close(); err != nil {
		os.Remove(path)
		return nil, nil, err
	}
	if err := store.Initialize(path); err != nil {
		os.Remove(path)
		return nil, nil, err
	}
	database, err := store.Open(path)
	if err != nil {
		os.Remove(path)
		return nil, nil, err
	}
	return database, func() {
		database.Close()
		os.Remove(path)
	}, nil
}

func runSmallFix(database *sql.DB) (int, int, error) {
	operations := 0
	item, err := store.AddItem(database, "Benchmark small fix", "", "Needed", "Now", "")
	operations++
	if err != nil {
		return operations, 0, err
	}
	if err := store.TransitionItemStatus(database, item.ID, "Ready"); err != nil {
		return operations, 0, err
	}
	operations++
	plan, err := store.CreatePlan(database, item.ID, "small fix plan")
	operations++
	if err != nil {
		return operations, 0, err
	}
	criterion, err := store.AddAcceptanceCriterionForPlan(database, plan.ID, "fixed", "The small fix works", "", true)
	operations++
	if err != nil {
		return operations, 0, err
	}
	task, err := store.AddTask(database, plan.ID, "Implement small fix", "", []string{"go test ./..."})
	operations++
	if err != nil {
		return operations, 0, err
	}
	if err := store.ApprovePlan(database, plan.ID); err != nil {
		return operations, 0, err
	}
	operations++
	if err := store.SetTaskStatus(database, task.ID, "In Progress", ""); err != nil {
		return operations, 0, err
	}
	operations++
	if err := store.SetTaskStatus(database, task.ID, "Done", "verified"); err != nil {
		return operations, 0, err
	}
	operations++
	if err := store.VerifyAcceptanceCriterion(database, criterion.ID, "Passed", "automated", "benchmark verification", "benchmark"); err != nil {
		return operations, 0, err
	}
	operations++
	if err := store.TransitionItemStatus(database, item.ID, "In Progress"); err != nil {
		return operations, 0, err
	}
	operations++
	if err := store.TransitionItemStatus(database, item.ID, "Done"); err != nil {
		return operations, 0, err
	}
	operations++
	encoded, _ := json.Marshal(struct {
		Status string `json:"status"`
	}{"passed"})
	return operations, len(encoded), nil
}

func runFeature(database *sql.DB) (int, int, error) {
	recorded, err := store.RecordApprovedPlan(database, store.PlanPacket{
		RoadmapItem:        store.PacketRoadmapItem{Title: "Benchmark feature", Category: "Needed", Horizon: "Now"},
		Goal:               "Deliver a feature",
		AcceptanceCriteria: []store.PacketCriterion{{ID: "feature", Title: "The feature is verified"}},
		Tasks:              []store.PacketTask{{Title: "Implement feature", Verification: []string{"go test ./..."}}},
	}, "benchmark")
	if err != nil {
		return 1, 0, err
	}
	if err := store.SetTaskStatus(database, recorded.Tasks[0].ID, "In Progress", ""); err != nil {
		return 2, 0, err
	}
	if err := store.SetTaskStatus(database, recorded.Tasks[0].ID, "Done", "verified"); err != nil {
		return 3, 0, err
	}
	if err := store.VerifyAcceptanceCriterion(database, recorded.Criteria[0].ID, "Passed", "automated", "benchmark verification", "benchmark"); err != nil {
		return 4, 0, err
	}
	if err := store.TransitionItemStatus(database, recorded.Plan.ItemID, "In Progress"); err != nil {
		return 5, 0, err
	}
	if err := store.TransitionItemStatus(database, recorded.Plan.ItemID, "Done"); err != nil {
		return 6, 0, err
	}
	encoded, _ := json.Marshal(recorded)
	return 7, len(encoded), nil
}

func runAmendment(database *sql.DB) (int, int, error) {
	recorded, err := store.RecordApprovedPlan(database, store.PlanPacket{
		RoadmapItem:        store.PacketRoadmapItem{Title: "Benchmark amendment", Category: "Needed", Horizon: "Now"},
		Goal:               "Amend an approved plan",
		AcceptanceCriteria: []store.PacketCriterion{{ID: "amendment", Title: "The amendment is approved"}},
		Tasks:              []store.PacketTask{{Title: "Original work", Verification: []string{"go test ./..."}}},
	}, "benchmark")
	if err != nil {
		return 1, 0, err
	}
	plan, err := store.RevisePlan(database, recorded.Plan.ID, recorded.Plan.Content)
	if err != nil {
		return 2, 0, err
	}
	if err := store.ApprovePlan(database, plan.ID); err != nil {
		return 3, 0, err
	}
	encoded, _ := json.Marshal(plan)
	return 4, len(encoded), nil
}

func runResumedTask(database *sql.DB) (int, int, error) {
	recorded, err := store.RecordApprovedPlan(database, store.PlanPacket{
		RoadmapItem:        store.PacketRoadmapItem{Title: "Benchmark resumed task", Category: "Needed", Horizon: "Now"},
		Goal:               "Resume from bounded context",
		Constraints:        []string{"Keep the output bounded"},
		AcceptanceCriteria: []store.PacketCriterion{{ID: "resume", Title: "The task can resume"}},
		Tasks:              []store.PacketTask{{Title: "Resume task", Verification: []string{"go test ./..."}}},
	}, "benchmark")
	if err != nil {
		return 1, 0, err
	}
	context, err := store.BuildScopedContext(database, recorded.Plan.ItemID, "implementation", 12000)
	if err != nil {
		return 2, 0, err
	}
	encoded, err := json.Marshal(context)
	if err != nil {
		return 2, 0, err
	}
	return 2, len(encoded), nil
}
