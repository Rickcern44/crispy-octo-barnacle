// Package check validates persisted Cassor state without mutating it.
package check

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rickcern44/cassor/internal/config"
	"github.com/rickcern44/cassor/internal/migrations"
	"github.com/rickcern44/cassor/internal/site"
	"github.com/rickcern44/cassor/internal/store"
)

// Run performs configuration, migration, relational, workflow, and render
// checks. It returns all discoverable validation failures together.
func Run(stateDir string) error {
	problems := []string{}
	if _, err := config.Read(config.Path(stateDir)); err != nil {
		problems = append(problems, "configuration: "+err.Error())
	}
	database, err := store.Open(filepath.Join(stateDir, store.DatabaseFileName))
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer database.Close()
	if err := migrationCheck(database); err != nil {
		problems = append(problems, err.Error())
	}
	if err := foreignKeyCheck(database); err != nil {
		problems = append(problems, err.Error())
	}
	if err := workflowCheck(database); err != nil {
		problems = append(problems, err.Error())
	}
	if err := site.Validate(stateDir); err != nil {
		problems = append(problems, "site render: "+err.Error())
	}
	if len(problems) > 0 {
		return fmt.Errorf("Cassor check failed:\n- %s", strings.Join(problems, "\n- "))
	}
	return nil
}

func migrationCheck(database *sql.DB) error {
	names, err := migrations.Names()
	if err != nil {
		return fmt.Errorf("migrations: %w", err)
	}
	expected := make(map[string]bool, len(names))
	for _, name := range names {
		expected[strings.TrimSuffix(filepath.Base(name), ".sql")] = true
	}
	rows, err := database.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("migration metadata: %w", err)
	}
	defer rows.Close()
	actual := map[string]bool{}
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return err
		}
		actual[version] = true
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for version := range expected {
		if !actual[version] {
			return fmt.Errorf("migration %s has not been applied", version)
		}
	}
	for version := range actual {
		if !expected[version] {
			return fmt.Errorf("unknown applied migration %s", version)
		}
	}
	return nil
}

func foreignKeyCheck(database *sql.DB) error {
	rows, err := database.Query(`PRAGMA foreign_key_check`)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		var table string
		var rowID sql.NullInt64
		var parent string
		var foreignKey int
		if err := rows.Scan(&table, &rowID, &parent, &foreignKey); err != nil {
			return err
		}
		return fmt.Errorf("foreign key violation in %s row %v referencing %s", table, rowID, parent)
	}
	return rows.Err()
}

func workflowCheck(database *sql.DB) error {
	checks := []struct{ query, message string }{
		{`SELECT COUNT(*) FROM plan_revisions p JOIN roadmap_items i ON i.id=p.roadmap_item_id WHERE p.status='Approved' AND (p.approved_at IS NULL OR i.status NOT IN ('Ready','In Progress','Done'))`, "approved plans must belong to active or completed items and have approval timestamps"},
		{`SELECT COUNT(*) FROM plan_revisions WHERE active=1 AND status<>'Approved'`, "only approved plan revisions may be active"},
		{`SELECT COUNT(*) FROM roadmap_items i JOIN plan_revisions p ON p.roadmap_item_id=i.id AND p.active=1 WHERE i.status IN ('Ready','In Progress') AND (SELECT COUNT(*) FROM acceptance_criteria c WHERE c.plan_revision_id=p.id AND trim(c.title)<>'')=0`, "active executable plans need meaningful acceptance criteria"},
		{`SELECT COUNT(*) FROM roadmap_items i JOIN plan_revisions p ON p.roadmap_item_id=i.id AND p.active=1 JOIN tasks t ON t.plan_revision_id=p.id WHERE i.status IN ('Ready','In Progress') AND (trim(t.verification)='' OR trim(t.verification)='[]')`, "active executable tasks need verification requirements"},
		{`SELECT COUNT(*) FROM plan_revisions WHERE status='Draft' AND approved_at IS NOT NULL`, "draft plans cannot have approval timestamps"},
		{`SELECT COUNT(*) FROM tasks t JOIN plan_revisions p ON p.id=t.plan_revision_id WHERE t.status IN ('In Progress','Done','Blocked') AND p.status!='Approved'`, "started, done, and blocked tasks must belong to approved plans"},
		{`SELECT COUNT(*) FROM tasks WHERE status='To Do' AND (started_at IS NOT NULL OR completed_at IS NOT NULL OR blocked_at IS NOT NULL)`, "to-do tasks cannot have lifecycle timestamps"},
		{`SELECT COUNT(*) FROM tasks WHERE status='In Progress' AND started_at IS NULL`, "in-progress tasks need a start timestamp"},
		{`SELECT COUNT(*) FROM tasks WHERE status='Done' AND (started_at IS NULL OR completed_at IS NULL OR outcome='')`, "done tasks need start, completion, and outcome data"},
		{`SELECT COUNT(*) FROM tasks WHERE status='Blocked' AND (started_at IS NULL OR blocked_at IS NULL OR outcome='')`, "blocked tasks need start, block, and outcome data"},
		{`SELECT COUNT(*) FROM tasks t LEFT JOIN task_events e ON e.task_id=t.id AND e.status='Blocked' WHERE t.status='Blocked' AND e.id IS NULL`, "blocked tasks need a blocking history event"},
		{`SELECT COUNT(*) FROM acceptance_criteria WHERE status IN ('Passed','Failed','Waived') AND (verified_at IS NULL OR trim(evidence)='')`, "verified criteria need timestamps and evidence"},
		{`SELECT COUNT(*) FROM acceptance_criteria WHERE status='Waived' AND (trim(waived_by)='' OR trim(waiver_reason)='')`, "waived criteria need attribution and a reason"},
		{`SELECT COUNT(*) FROM feature_change_links l JOIN roadmap_items c ON c.id=l.change_item_id JOIN roadmap_items p ON p.id=l.capability_item_id WHERE c.feature_type<>'Change' OR p.feature_type<>'Capability'`, "change links must connect Changes to Capabilities"},
		{`SELECT COUNT(*) FROM capability_state_history h JOIN roadmap_items c ON c.id=h.capability_item_id LEFT JOIN roadmap_items s ON s.id=h.source_change_item_id WHERE c.feature_type<>'Capability' OR (h.source_change_item_id IS NOT NULL AND s.feature_type<>'Change') OR trim(h.state)='' OR trim(h.accepted_by)='' OR trim(h.accepted_at)=''`, "capability state history must be attributed and type-safe"},
		{`SELECT COUNT(*) FROM dossier_artifacts WHERE status='Accepted' AND (trim(evidence)='' OR accepted_at IS NULL OR trim(COALESCE(accepted_by,''))='')`, "accepted dossier artifacts need evidence and attribution"},
		{`SELECT COUNT(*) FROM roadmap_items i LEFT JOIN epics e ON e.id=i.epic_id WHERE i.epic_id IS NOT NULL AND e.id IS NULL`, "features with an Epic parent must reference an existing Epic"},
	}
	for _, check := range checks {
		var count int
		if err := database.QueryRow(check.query).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("workflow: %s", check.message)
		}
	}
	return nil
}
