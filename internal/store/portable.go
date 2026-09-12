package store

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
)

const (
	ExportFormat  = "cassor-state"
	ExportVersion = 1
)

type CriterionEvidence struct {
	ID                 int64  `json:"id"`
	CriterionID        int64  `json:"criterion_id"`
	Status             string `json:"status"`
	VerificationMethod string `json:"verification_method"`
	Evidence           string `json:"evidence"`
	RecordedBy         string `json:"recorded_by"`
	WaiverReason       string `json:"waiver_reason"`
	RecordedAt         string `json:"recorded_at"`
}

// PortableState contains supported logical state in deterministic primary-key
// order. It intentionally excludes schema metadata and generated site files.
type PortableState struct {
	Format            string                  `json:"format"`
	Version           int                     `json:"version"`
	Categories        []Category              `json:"categories"`
	Epics             []Epic                  `json:"epics,omitempty"`
	Items             []Item                  `json:"items"`
	Plans             []Plan                  `json:"plans"`
	Tasks             []Task                  `json:"tasks"`
	PhaseRecords      []PhaseRecord           `json:"phase_records"`
	Criteria          []AcceptanceCriterion   `json:"criteria"`
	CriterionEvidence []CriterionEvidence     `json:"criterion_evidence"`
	TaskEvents        []TaskEvent             `json:"task_events"`
	Relationships     []FeatureRelationship   `json:"relationships"`
	ChangeLinks       []FeatureChangeLink     `json:"change_links"`
	CapabilityStates  []CapabilityStateRecord `json:"capability_state_history"`
	Artifacts         []DossierArtifact       `json:"dossier_artifacts"`
	Reports           []FeatureReport         `json:"reports"`
}

func ExportState(database *sql.DB) ([]byte, error) {
	state := PortableState{Format: ExportFormat, Version: ExportVersion}
	var err error
	if state.Categories, err = exportCategories(database); err != nil {
		return nil, err
	}
	if state.Items, err = ListItems(database); err != nil {
		return nil, err
	}
	if state.Epics, err = ListEpics(database); err != nil {
		return nil, err
	}
	if state.Plans, err = exportPlans(database); err != nil {
		return nil, err
	}
	if state.Tasks, err = exportTasks(database); err != nil {
		return nil, err
	}
	if state.PhaseRecords, err = ListPhaseRecords(database); err != nil {
		return nil, err
	}
	if state.Criteria, err = ListAcceptanceCriteria(database); err != nil {
		return nil, err
	}
	if state.CriterionEvidence, err = exportCriterionEvidence(database); err != nil {
		return nil, err
	}
	if state.TaskEvents, err = exportTaskEvents(database); err != nil {
		return nil, err
	}
	if state.Relationships, err = ListAllFeatureRelationships(database); err != nil {
		return nil, err
	}
	if state.ChangeLinks, err = ListAllFeatureChangeLinks(database); err != nil {
		return nil, err
	}
	if state.CapabilityStates, err = ListCapabilityStateHistory(database); err != nil {
		return nil, err
	}
	if state.Artifacts, err = ListDossierArtifacts(database); err != nil {
		return nil, err
	}
	if state.Reports, err = ListFeatureReports(database); err != nil {
		return nil, err
	}
	encoded, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(encoded, '\n'), nil
}

func ImportState(database *sql.DB, data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var state PortableState
	if err := decoder.Decode(&state); err != nil {
		return fmt.Errorf("decode Cassor state: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return fmt.Errorf("state export must contain one JSON value")
	}
	if state.Format != ExportFormat || state.Version != ExportVersion {
		return fmt.Errorf("unsupported Cassor state format %q version %d", state.Format, state.Version)
	}
	if err := validatePortableState(state); err != nil {
		return err
	}
	transaction, err := database.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	if err := ensureEmptyDestination(transaction, state); err != nil {
		return err
	}
	if err := importCategories(transaction, state.Categories); err != nil {
		return err
	}
	if err := importEpics(transaction, state.Epics); err != nil {
		return err
	}
	if err := importItems(transaction, state.Items); err != nil {
		return err
	}
	if err := importPlans(transaction, state.Plans); err != nil {
		return err
	}
	if err := importTasks(transaction, state.Tasks); err != nil {
		return err
	}
	if err := importPhaseRecords(transaction, state.PhaseRecords); err != nil {
		return err
	}
	if err := importCriteria(transaction, state.Criteria); err != nil {
		return err
	}
	if err := importCriterionEvidence(transaction, state.CriterionEvidence); err != nil {
		return err
	}
	if err := importTaskEvents(transaction, state.TaskEvents); err != nil {
		return err
	}
	if err := importRelationships(transaction, state.Relationships); err != nil {
		return err
	}
	if err := importChangeLinks(transaction, state.ChangeLinks); err != nil {
		return err
	}
	if err := importCapabilityStates(transaction, state.CapabilityStates); err != nil {
		return err
	}
	if err := importArtifacts(transaction, state.Artifacts); err != nil {
		return err
	}
	if err := importReports(transaction, state.Reports); err != nil {
		return err
	}
	return transaction.Commit()
}

func exportCategories(database *sql.DB) ([]Category, error) { return ListCategories(database) }
func exportPlans(database *sql.DB) ([]Plan, error) {
	rows, err := database.Query(`SELECT id,roadmap_item_id,revision,content,status,active,created_at,approved_at,approval_note FROM plan_revisions ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []Plan{}
	for rows.Next() {
		value, err := scanPlan(rows)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}
func exportTasks(database *sql.DB) ([]Task, error) {
	rows, err := database.Query(`SELECT id,plan_revision_id,title,description,verification,status,outcome,created_at,started_at,completed_at,blocked_at FROM tasks ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []Task{}
	for rows.Next() {
		value, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}
func exportCriterionEvidence(database *sql.DB) ([]CriterionEvidence, error) {
	rows, err := database.Query(`SELECT id,acceptance_criterion_id,status,verification_method,evidence,recorded_by,waiver_reason,recorded_at FROM criterion_evidence ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []CriterionEvidence{}
	for rows.Next() {
		var value CriterionEvidence
		if err := rows.Scan(&value.ID, &value.CriterionID, &value.Status, &value.VerificationMethod, &value.Evidence, &value.RecordedBy, &value.WaiverReason, &value.RecordedAt); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}
func exportTaskEvents(database *sql.DB) ([]TaskEvent, error) {
	rows, err := database.Query(`SELECT id,task_id,status,outcome,recorded_at FROM task_events ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []TaskEvent{}
	for rows.Next() {
		var value TaskEvent
		if err := rows.Scan(&value.ID, &value.TaskID, &value.Status, &value.Outcome, &value.RecordedAt); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

func validatePortableState(state PortableState) error {
	items := map[int64]bool{}
	plans := map[int64]bool{}
	tasks := map[int64]bool{}
	criteria := map[int64]bool{}
	for _, item := range state.Items {
		if items[item.ID] {
			return fmt.Errorf("duplicate item ID %d", item.ID)
		}
		items[item.ID] = true
	}
	epics := map[int64]bool{}
	for _, epic := range state.Epics {
		if epics[epic.ID] {
			return fmt.Errorf("duplicate epic ID %d", epic.ID)
		}
		epics[epic.ID] = true
	}
	for _, item := range state.Items {
		if item.EpicID != nil && !epics[*item.EpicID] {
			return fmt.Errorf("item %d references missing epic %d", item.ID, *item.EpicID)
		}
	}
	for _, plan := range state.Plans {
		if plans[plan.ID] {
			return fmt.Errorf("duplicate plan ID %d", plan.ID)
		}
		if !items[plan.ItemID] {
			return fmt.Errorf("plan %d references missing item %d", plan.ID, plan.ItemID)
		}
		plans[plan.ID] = true
	}
	for _, task := range state.Tasks {
		if tasks[task.ID] {
			return fmt.Errorf("duplicate task ID %d", task.ID)
		}
		if !plans[task.PlanID] {
			return fmt.Errorf("task %d references missing plan %d", task.ID, task.PlanID)
		}
		tasks[task.ID] = true
	}
	for _, criterion := range state.Criteria {
		if criteria[criterion.ID] {
			return fmt.Errorf("duplicate criterion ID %d", criterion.ID)
		}
		if !items[criterion.ItemID] {
			return fmt.Errorf("criterion %d references missing item %d", criterion.ID, criterion.ItemID)
		}
		if criterion.PlanID != 0 && !plans[criterion.PlanID] {
			return fmt.Errorf("criterion %d references missing plan %d", criterion.ID, criterion.PlanID)
		}
		criteria[criterion.ID] = true
	}
	for _, evidence := range state.CriterionEvidence {
		if !criteria[evidence.CriterionID] {
			return fmt.Errorf("evidence %d references missing criterion %d", evidence.ID, evidence.CriterionID)
		}
	}
	for _, event := range state.TaskEvents {
		if !tasks[event.TaskID] {
			return fmt.Errorf("task event %d references missing task %d", event.ID, event.TaskID)
		}
	}
	for _, relationship := range state.Relationships {
		if !items[relationship.SourceItemID] || !items[relationship.TargetItemID] {
			return fmt.Errorf("relationship %d references missing item", relationship.ID)
		}
	}
	itemValues := map[int64]Item{}
	for _, item := range state.Items {
		itemValues[item.ID] = item
	}
	for _, link := range state.ChangeLinks {
		if !items[link.ChangeItemID] || !items[link.CapabilityItemID] {
			return fmt.Errorf("change link %d references missing item", link.ID)
		}
		if itemValues[link.ChangeItemID].FeatureType != "Change" || itemValues[link.CapabilityItemID].FeatureType != "Capability" {
			return fmt.Errorf("change link %d must connect a Change to a Capability", link.ID)
		}
	}
	for _, stateRecord := range state.CapabilityStates {
		if !items[stateRecord.CapabilityItemID] {
			return fmt.Errorf("capability state %d references missing capability %d", stateRecord.ID, stateRecord.CapabilityItemID)
		}
		if stateRecord.SourceChangeID != nil && !items[*stateRecord.SourceChangeID] {
			return fmt.Errorf("capability state %d references missing change %d", stateRecord.ID, *stateRecord.SourceChangeID)
		}
	}
	artifacts := map[int64]bool{}
	for _, artifact := range state.Artifacts {
		if artifacts[artifact.ID] {
			return fmt.Errorf("duplicate artifact ID %d", artifact.ID)
		}
		if !items[artifact.ItemID] {
			return fmt.Errorf("artifact %d references missing item %d", artifact.ID, artifact.ItemID)
		}
		if artifact.SupersedesID != nil && !artifacts[*artifact.SupersedesID] {
			return fmt.Errorf("artifact %d references missing prior artifact %d", artifact.ID, *artifact.SupersedesID)
		}
		artifacts[artifact.ID] = true
	}
	for _, report := range state.Reports {
		if !items[report.ItemID] {
			return fmt.Errorf("report %d references missing item %d", report.ID, report.ItemID)
		}
	}
	return nil
}

func ensureEmptyDestination(transaction *sql.Tx, state PortableState) error {
	for _, table := range []string{"epics", "roadmap_items", "plan_revisions", "tasks", "feature_phase_records", "acceptance_criteria", "criterion_evidence", "task_events", "feature_relationships", "feature_change_links", "capability_state_history", "dossier_artifacts", "feature_reports"} {
		var count int
		if err := transaction.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("cannot import into non-empty %s (%d records)", table, count)
		}
	}
	for _, category := range state.Categories {
		var name string
		err := transaction.QueryRow(`SELECT name FROM categories WHERE id=?`, category.ID).Scan(&name)
		if err == nil && name != category.Name {
			return fmt.Errorf("category ID %d conflicts with %q", category.ID, name)
		}
		if err != nil && err != sql.ErrNoRows {
			return err
		}
	}
	return nil
}

func importEpics(transaction *sql.Tx, values []Epic) error {
	for _, value := range values {
		if _, err := transaction.Exec(`INSERT INTO epics(id,title,description,created_at,updated_at) VALUES(?,?,?,?,?)`, value.ID, value.Title, value.Description, value.CreatedAt, value.UpdatedAt); err != nil {
			return err
		}
	}
	return nil
}

func importCategories(transaction *sql.Tx, values []Category) error {
	for _, value := range values {
		var exists string
		err := transaction.QueryRow(`SELECT name FROM categories WHERE id=?`, value.ID).Scan(&exists)
		if err == nil {
			if _, err := transaction.Exec(`UPDATE categories SET name=?,created_at=? WHERE id=?`, value.Name, value.CreatedAt, value.ID); err != nil {
				return err
			}
			continue
		}
		if err != sql.ErrNoRows {
			return err
		}
		if _, err := transaction.Exec(`INSERT INTO categories(id,name,created_at) VALUES(?,?,?)`, value.ID, value.Name, value.CreatedAt); err != nil {
			return err
		}
	}
	return nil
}
func importItems(transaction *sql.Tx, values []Item) error {
	for _, value := range values {
		if _, err := transaction.Exec(`INSERT INTO roadmap_items(id,title,description,category_id,horizon,status,rationale,created_at,updated_at,target_date,progress,priority,complexity,team,lead_engineer,technical_summary,specifications,documentation_links,feature_type,current_state,epic_id) VALUES(?,?,?,(SELECT id FROM categories WHERE name=?),?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, value.ID, value.Title, value.Description, value.Category, value.Horizon, value.Status, value.Rationale, value.CreatedAt, value.UpdatedAt, value.TargetDate, value.Progress, value.Priority, value.Complexity, value.Team, value.LeadEngineer, value.TechnicalSummary, value.Specifications, value.DocumentationLinks, value.FeatureType, value.CurrentState, value.EpicID); err != nil {
			return err
		}
	}
	return nil
}
func importPlans(transaction *sql.Tx, values []Plan) error {
	for _, value := range values {
		if _, err := transaction.Exec(`INSERT INTO plan_revisions(id,roadmap_item_id,revision,content,status,active,created_at,approved_at,approval_note) VALUES(?,?,?,?,?,?,?,?,?)`, value.ID, value.ItemID, value.Revision, value.Content, value.Status, boolInt(value.Active), value.CreatedAt, value.ApprovedAt, value.ApprovalNote); err != nil {
			return err
		}
	}
	return nil
}
func importTasks(transaction *sql.Tx, values []Task) error {
	for _, value := range values {
		verification, err := json.Marshal(value.Verification)
		if err != nil {
			return err
		}
		if _, err := transaction.Exec(`INSERT INTO tasks(id,plan_revision_id,title,description,verification,status,outcome,created_at,started_at,completed_at,blocked_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, value.ID, value.PlanID, value.Title, value.Description, string(verification), value.Status, value.Outcome, value.CreatedAt, value.StartedAt, value.CompletedAt, value.BlockedAt); err != nil {
			return err
		}
	}
	return nil
}
func importPhaseRecords(transaction *sql.Tx, values []PhaseRecord) error {
	for _, value := range values {
		if _, err := transaction.Exec(`INSERT INTO feature_phase_records(id,roadmap_item_id,phase,revision,content,created_at) VALUES(?,?,?,?,?,?)`, value.ID, value.ItemID, value.Phase, value.Revision, value.Content, value.CreatedAt); err != nil {
			return err
		}
	}
	return nil
}
func importCriteria(transaction *sql.Tx, values []AcceptanceCriterion) error {
	for _, value := range values {
		var plan any
		if value.PlanID != 0 {
			plan = value.PlanID
		}
		if _, err := transaction.Exec(`INSERT INTO acceptance_criteria(id,roadmap_item_id,plan_revision_id,criterion_key,title,description,required,status,verification_method,evidence,verified_at,waived_by,waiver_reason) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`, value.ID, value.ItemID, plan, value.Key, value.Title, value.Description, boolInt(value.Required), value.Status, value.VerificationMethod, value.Evidence, value.VerifiedAt, value.WaivedBy, value.WaiverReason); err != nil {
			return err
		}
	}
	return nil
}
func importCriterionEvidence(transaction *sql.Tx, values []CriterionEvidence) error {
	for _, value := range values {
		if _, err := transaction.Exec(`INSERT INTO criterion_evidence(id,acceptance_criterion_id,status,verification_method,evidence,recorded_by,waiver_reason,recorded_at) VALUES(?,?,?,?,?,?,?,?)`, value.ID, value.CriterionID, value.Status, value.VerificationMethod, value.Evidence, value.RecordedBy, value.WaiverReason, value.RecordedAt); err != nil {
			return err
		}
	}
	return nil
}
func importTaskEvents(transaction *sql.Tx, values []TaskEvent) error {
	for _, value := range values {
		if _, err := transaction.Exec(`INSERT INTO task_events(id,task_id,status,outcome,recorded_at) VALUES(?,?,?,?,?)`, value.ID, value.TaskID, value.Status, value.Outcome, value.RecordedAt); err != nil {
			return err
		}
	}
	return nil
}
func importRelationships(transaction *sql.Tx, values []FeatureRelationship) error {
	for _, value := range values {
		if _, err := transaction.Exec(`INSERT INTO feature_relationships(id,source_item_id,target_item_id,relationship_type,created_at) VALUES(?,?,?,?,?)`, value.ID, value.SourceItemID, value.TargetItemID, value.RelationshipType, value.CreatedAt); err != nil {
			return err
		}
	}
	return nil
}
func importChangeLinks(transaction *sql.Tx, values []FeatureChangeLink) error {
	for _, value := range values {
		if _, err := transaction.Exec(`INSERT INTO feature_change_links(id,change_item_id,capability_item_id,created_at) VALUES(?,?,?,?)`, value.ID, value.ChangeItemID, value.CapabilityItemID, value.CreatedAt); err != nil {
			return err
		}
	}
	return nil
}
func importCapabilityStates(transaction *sql.Tx, values []CapabilityStateRecord) error {
	for _, value := range values {
		var source any
		if value.SourceChangeID != nil {
			source = *value.SourceChangeID
		}
		if _, err := transaction.Exec(`INSERT INTO capability_state_history(id,capability_item_id,source_change_item_id,state,accepted_by,accepted_at,created_at) VALUES(?,?,?,?,?,?,?)`, value.ID, value.CapabilityItemID, source, value.State, value.AcceptedBy, value.AcceptedAt, value.CreatedAt); err != nil {
			return err
		}
	}
	return nil
}
func importArtifacts(transaction *sql.Tx, values []DossierArtifact) error {
	for _, value := range values {
		if _, err := transaction.Exec(`INSERT INTO dossier_artifacts(id,roadmap_item_id,kind,author_role,status,summary,evidence,supersedes_id,created_at,accepted_at,accepted_by) VALUES(?,?,?,?,?,?,?,NULL,?,?,?)`, value.ID, value.ItemID, value.Kind, value.AuthorRole, value.Status, value.Summary, value.Evidence, value.CreatedAt, value.AcceptedAt, value.AcceptedBy); err != nil {
			return err
		}
	}
	for _, value := range values {
		if value.SupersedesID != nil {
			if _, err := transaction.Exec(`UPDATE dossier_artifacts SET supersedes_id=? WHERE id=?`, *value.SupersedesID, value.ID); err != nil {
				return err
			}
		}
	}
	return nil
}
func importReports(transaction *sql.Tx, values []FeatureReport) error {
	for _, value := range values {
		roles, err := json.Marshal(value.Roles)
		if err != nil {
			return err
		}
		if _, err := transaction.Exec(`INSERT INTO feature_reports(id,roadmap_item_id,execution_mode,roles,elapsed_ns,tool_calls,verification,input_tokens,cached_input_tokens,output_tokens,reasoning_tokens,total_tokens,comparison_key,context_packets,context_bytes,handoffs,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, value.ID, value.ItemID, value.ExecutionMode, string(roles), value.ElapsedNS, value.ToolCalls, value.Verification, value.InputTokens, value.CachedInputTokens, value.OutputTokens, value.ReasoningTokens, value.TotalTokens, value.ComparisonKey, value.ContextPackets, value.ContextBytes, value.Handoffs, value.CreatedAt); err != nil {
			return err
		}
	}
	return nil
}
