package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// Context is the bounded navigation snapshot returned by the default context
// command. It contains references and counts, never full plan artifacts.
type Context struct {
	ActiveTasks       []ContextReference `json:"active_tasks"`
	ProposedItems     []ContextReference `json:"proposed_items"`
	PlansAwaiting     []ContextReference `json:"plans_awaiting_approval"`
	BlockedTasks      []ContextReference `json:"blocked_tasks"`
	RecentlyCompleted []ContextReference `json:"recently_completed_tasks"`
	Capabilities      []ContextReference `json:"capabilities"`
	InFlightChanges   []ContextReference `json:"in_flight_changes"`
	PlannedWork       []ContextReference `json:"planned_work"`
	KnownGaps         []ContextReference `json:"known_gaps"`
	Epics             []ContextReference `json:"epics"`
	Counts            ContextCounts      `json:"counts"`
	Truncated         bool               `json:"truncated"`
}

type ContextCounts struct {
	ActiveTasks       int `json:"active_tasks"`
	ProposedItems     int `json:"proposed_items"`
	PlansAwaiting     int `json:"plans_awaiting_approval"`
	BlockedTasks      int `json:"blocked_tasks"`
	RecentlyCompleted int `json:"recently_completed_tasks"`
	Capabilities      int `json:"capabilities"`
	InFlightChanges   int `json:"in_flight_changes"`
	PlannedWork       int `json:"planned_work"`
	KnownGaps         int `json:"known_gaps"`
	Epics             int `json:"epics"`
}

type ContextReference struct {
	Kind    string `json:"kind"`
	ID      int64  `json:"id"`
	ItemID  int64  `json:"item_id,omitempty"`
	PlanID  int64  `json:"plan_id,omitempty"`
	Title   string `json:"title"`
	Status  string `json:"status"`
	Command string `json:"command,omitempty"`
}

// ScopedContext is the role-aware resumption contract. Required fields remain
// under truncation; omitted detail exposes fetchable IDs.
type ScopedContext struct {
	Version     int                `json:"version"`
	Item        ContextItem        `json:"item"`
	Role        string             `json:"role"`
	Assignment  ContextAssignment  `json:"assignment"`
	Scope       PacketScope        `json:"scope"`
	Decisions   []string           `json:"decisions"`
	Constraints []string           `json:"constraints"`
	Criteria    []ContextCriterion `json:"criteria"`
	NextAction  ContextAction      `json:"next_action"`
	Blockers    []ContextReference `json:"blockers"`
	Evidence    []ContextReference `json:"evidence_references"`
	Related     []ContextReference `json:"related_features"`
	Artifacts   []ContextReference `json:"dossier_artifacts"`
	Epic        *ContextReference  `json:"epic,omitempty"`
	Omitted     []ContextReference `json:"omitted_details,omitempty"`
	Truncated   bool               `json:"truncated"`
	MaxBytes    int                `json:"max_bytes"`
	Bytes       int                `json:"bytes"`
}

type ContextItem struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
	Type   string `json:"feature_type"`
	State  string `json:"current_state,omitempty"`
}

type ContextAssignment struct {
	PlanID    int64              `json:"plan_id"`
	Revision  int                `json:"revision"`
	Goal      string             `json:"goal,omitempty"`
	Reference string             `json:"reference"`
	Tasks     []ContextReference `json:"tasks"`
}

type ContextCriterion struct {
	ID          int64  `json:"id"`
	Key         string `json:"key"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	Required    bool   `json:"required"`
	EvidenceRef string `json:"evidence_reference"`
}

type ContextAction struct {
	Kind         string   `json:"kind"`
	ID           int64    `json:"id,omitempty"`
	Title        string   `json:"title"`
	Description  string   `json:"description,omitempty"`
	Status       string   `json:"status"`
	Command      string   `json:"command"`
	Verification []string `json:"verification,omitempty"`
}

// CompactContext returns a deterministic, small navigation summary.
func CompactContext(database *sql.DB) (Context, error) {
	var context Context
	var err error
	context.ActiveTasks, context.Counts.ActiveTasks, err = navigationTasks(database, "In Progress", 10)
	if err != nil {
		return context, err
	}
	context.ProposedItems, context.Counts.ProposedItems, err = navigationItems(database, "Planned", 10)
	if err != nil {
		return context, err
	}
	context.PlansAwaiting, context.Counts.PlansAwaiting, err = navigationPlans(database, "Draft", 10)
	if err != nil {
		return context, err
	}
	context.BlockedTasks, context.Counts.BlockedTasks, err = navigationTasks(database, "Blocked", 10)
	if err != nil {
		return context, err
	}
	context.RecentlyCompleted, context.Counts.RecentlyCompleted, err = navigationTasks(database, "Done", 10)
	if err != nil {
		return context, err
	}
	context.Capabilities, context.Counts.Capabilities, err = navigationFeatures(database, "Capability", "", 10)
	if err != nil {
		return context, err
	}
	context.InFlightChanges, context.Counts.InFlightChanges, err = navigationFeatures(database, "Change", "In Progress", 10)
	if err != nil {
		return context, err
	}
	context.PlannedWork, context.Counts.PlannedWork, err = navigationFeatures(database, "Change", "Planned,Ready", 10)
	if err != nil {
		return context, err
	}
	context.KnownGaps, context.Counts.KnownGaps, err = navigationFeatures(database, "Gap", "", 10)
	if err != nil {
		return context, err
	}
	context.Epics, context.Counts.Epics, err = navigationEpics(database, 10)
	if err != nil {
		return context, err
	}
	context.Truncated = context.Counts.ActiveTasks > len(context.ActiveTasks) || context.Counts.ProposedItems > len(context.ProposedItems) || context.Counts.PlansAwaiting > len(context.PlansAwaiting) || context.Counts.BlockedTasks > len(context.BlockedTasks) || context.Counts.RecentlyCompleted > len(context.RecentlyCompleted) || context.Counts.Capabilities > len(context.Capabilities) || context.Counts.InFlightChanges > len(context.InFlightChanges) || context.Counts.PlannedWork > len(context.PlannedWork) || context.Counts.KnownGaps > len(context.KnownGaps) || context.Counts.Epics > len(context.Epics)
	return context, nil
}

func navigationEpics(database *sql.DB, limit int) ([]ContextReference, int, error) {
	var count int
	if err := database.QueryRow(`SELECT COUNT(*) FROM epics`).Scan(&count); err != nil {
		return nil, 0, err
	}
	rows, err := database.Query(`SELECT id,title FROM epics ORDER BY updated_at DESC,id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	values := []ContextReference{}
	for rows.Next() {
		var value ContextReference
		if err := rows.Scan(&value.ID, &value.Title); err != nil {
			return nil, 0, err
		}
		value.Kind = "epic"
		value.Status = "group"
		value.Command = fmt.Sprintf("cassor epic show %d", value.ID)
		values = append(values, value)
	}
	return values, count, rows.Err()
}

func navigationItems(database *sql.DB, status string, limit int) ([]ContextReference, int, error) {
	var count int
	if err := database.QueryRow(`SELECT COUNT(*) FROM roadmap_items WHERE status=?`, status).Scan(&count); err != nil {
		return nil, 0, err
	}
	rows, err := database.Query(`SELECT id,title,status FROM roadmap_items WHERE status=? ORDER BY updated_at DESC,id DESC LIMIT ?`, status, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	values := []ContextReference{}
	for rows.Next() {
		var value ContextReference
		if err := rows.Scan(&value.ID, &value.Title, &value.Status); err != nil {
			return nil, 0, err
		}
		value.Kind = "item"
		value.Command = fmt.Sprintf("cassor item show %d", value.ID)
		values = append(values, value)
	}
	return values, count, rows.Err()
}

func navigationFeatures(database *sql.DB, featureType, statuses string, limit int) ([]ContextReference, int, error) {
	args := []any{featureType}
	where := "feature_type=?"
	if statuses != "" {
		parts := strings.Split(statuses, ",")
		placeholders := make([]string, len(parts))
		for index, part := range parts {
			placeholders[index] = "?"
			args = append(args, strings.TrimSpace(part))
		}
		where += " AND status IN (" + strings.Join(placeholders, ",") + ")"
	}
	var count int
	if err := database.QueryRow(`SELECT COUNT(*) FROM roadmap_items WHERE `+where, args...).Scan(&count); err != nil {
		return nil, 0, err
	}
	rows, err := database.Query(`SELECT id,title,status FROM roadmap_items WHERE `+where+` ORDER BY updated_at DESC,id DESC LIMIT ?`, append(args, limit)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	values := []ContextReference{}
	for rows.Next() {
		var value ContextReference
		if err := rows.Scan(&value.ID, &value.Title, &value.Status); err != nil {
			return nil, 0, err
		}
		value.Kind = strings.ToLower(featureType)
		value.Command = fmt.Sprintf("cassor item show %d", value.ID)
		values = append(values, value)
	}
	return values, count, rows.Err()
}

func navigationPlans(database *sql.DB, status string, limit int) ([]ContextReference, int, error) {
	var count int
	if err := database.QueryRow(`SELECT COUNT(*) FROM plan_revisions WHERE status=?`, status).Scan(&count); err != nil {
		return nil, 0, err
	}
	rows, err := database.Query(`SELECT id,roadmap_item_id,revision,status FROM plan_revisions WHERE status=? ORDER BY created_at,id LIMIT ?`, status, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	values := []ContextReference{}
	for rows.Next() {
		var value ContextReference
		var revision int
		if err := rows.Scan(&value.ID, &value.ItemID, &revision, &value.Status); err != nil {
			return nil, 0, err
		}
		value.Kind = "plan"
		value.Title = fmt.Sprintf("plan revision %d", revision)
		value.Command = fmt.Sprintf("cassor plan show %d", value.ID)
		values = append(values, value)
	}
	return values, count, rows.Err()
}

func navigationTasks(database *sql.DB, status string, limit int) ([]ContextReference, int, error) {
	var count int
	if err := database.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status=?`, status).Scan(&count); err != nil {
		return nil, 0, err
	}
	rows, err := database.Query(`SELECT id,plan_revision_id,title,status FROM tasks WHERE status=? ORDER BY COALESCE(completed_at,blocked_at,started_at,created_at) DESC,id DESC LIMIT ?`, status, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	values := []ContextReference{}
	for rows.Next() {
		var value ContextReference
		if err := rows.Scan(&value.ID, &value.PlanID, &value.Title, &value.Status); err != nil {
			return nil, 0, err
		}
		value.Kind = "task"
		value.Command = fmt.Sprintf("cassor task show %d", value.ID)
		values = append(values, value)
	}
	return values, count, rows.Err()
}

var scopedRoles = map[string]bool{"implementation": true, "verification": true, "planning": true}

func BuildScopedContext(database *sql.DB, itemID int64, role string, maxBytes int) (ScopedContext, error) {
	role = strings.ToLower(strings.TrimSpace(role))
	if !scopedRoles[role] {
		return ScopedContext{}, fmt.Errorf("unsupported context role %q; use implementation, verification, or planning", role)
	}
	if maxBytes < 256 {
		return ScopedContext{}, fmt.Errorf("--max-bytes must be at least 256 to retain required context")
	}
	item, err := GetItem(database, itemID)
	if err != nil {
		return ScopedContext{}, err
	}
	context := ScopedContext{Version: 1, Item: ContextItem{ID: item.ID, Title: item.Title, Status: item.Status, Type: item.FeatureType, State: item.CurrentState}, Role: role, Scope: PacketScope{Included: []string{}, Excluded: []string{}}, Decisions: []string{}, MaxBytes: maxBytes, Constraints: []string{}, Criteria: []ContextCriterion{}, Blockers: []ContextReference{}, Evidence: []ContextReference{}, Related: []ContextReference{}, Artifacts: []ContextReference{}, Omitted: []ContextReference{}}
	if item.EpicID != nil {
		epic, err := GetEpic(database, *item.EpicID)
		if err != nil {
			return ScopedContext{}, err
		}
		context.Epic = &ContextReference{Kind: "epic", ID: epic.ID, Title: epic.Title, Status: "parent", Command: fmt.Sprintf("cassor epic show %d", epic.ID)}
	}
	links, err := ListFeatureChangeLinks(database, itemID)
	if err != nil {
		return ScopedContext{}, err
	}
	for _, link := range links {
		if link.ChangeItemID == itemID {
			context.Related = append(context.Related, ContextReference{Kind: "capability", ID: link.CapabilityItemID, ItemID: link.CapabilityItemID, Title: link.CapabilityTitle, Status: "linked", Command: fmt.Sprintf("cassor item show %d", link.CapabilityItemID)})
		} else {
			context.Related = append(context.Related, ContextReference{Kind: "change", ID: link.ChangeItemID, ItemID: link.ChangeItemID, Title: link.ChangeItemTitle, Status: "linked", Command: fmt.Sprintf("cassor item show %d", link.ChangeItemID)})
		}
	}
	artifacts, err := ListDossierArtifactsForItem(database, itemID)
	if err != nil {
		return ScopedContext{}, err
	}
	for _, artifact := range artifacts {
		context.Artifacts = append(context.Artifacts, ContextReference{Kind: "artifact", ID: artifact.ID, ItemID: itemID, Title: artifact.Summary, Status: artifact.Status, Command: fmt.Sprintf("cassor item artifact list --item %d", itemID)})
	}
	plan, err := scanPlan(database.QueryRow(`SELECT id,roadmap_item_id,revision,content,status,active,created_at,approved_at,approval_note FROM plan_revisions WHERE roadmap_item_id=? AND active=1`, itemID))
	if err == sql.ErrNoRows {
		return ScopedContext{}, fmt.Errorf("roadmap item %d has no active approved plan", itemID)
	} else if err != nil {
		return ScopedContext{}, err
	}
	context.Assignment = ContextAssignment{PlanID: plan.ID, Revision: plan.Revision, Reference: fmt.Sprintf("cassor plan show %d", plan.ID), Tasks: []ContextReference{}}
	var packet PlanPacket
	if err := json.Unmarshal([]byte(plan.Content), &packet); err == nil {
		context.Assignment.Goal = packet.Goal
		context.Scope = packet.Scope
		context.Decisions = packet.Decisions
		context.Constraints = append(context.Constraints, packet.Constraints...)
	}
	if len(context.Constraints) == 0 {
		context.Constraints = []string{"Full plan constraints: " + context.Assignment.Reference}
	}
	criteria, err := criteriaForPlan(database, plan.ID)
	if err != nil {
		return ScopedContext{}, err
	}
	for _, criterion := range criteria {
		context.Criteria = append(context.Criteria, ContextCriterion{ID: criterion.ID, Key: criterion.Key, Title: criterion.Title, Status: criterion.Status, Required: criterion.Required, EvidenceRef: fmt.Sprintf("cassor lifecycle criterion show %d", criterion.ID)})
		if criterion.Status != "Pending" {
			context.Evidence = append(context.Evidence, ContextReference{Kind: "criterion-evidence", ID: criterion.ID, PlanID: plan.ID, ItemID: itemID, Title: criterion.Title, Status: criterion.Status, Command: fmt.Sprintf("cassor lifecycle criterion show %d", criterion.ID)})
		}
	}
	tasks, err := ListTasks(database, plan.ID)
	if err != nil {
		return ScopedContext{}, err
	}
	for _, task := range tasks {
		ref := ContextReference{Kind: "task", ID: task.ID, ItemID: itemID, PlanID: plan.ID, Title: task.Title, Status: task.Status, Command: fmt.Sprintf("cassor task show %d", task.ID)}
		context.Assignment.Tasks = append(context.Assignment.Tasks, ref)
		if task.Status == "Blocked" {
			context.Blockers = append(context.Blockers, ref)
		}
		if context.NextAction.Kind == "" && (task.Status == "In Progress" || task.Status == "To Do" || task.Status == "Blocked") {
			command := fmt.Sprintf("cassor task start %d", task.ID)
			if task.Status == "Blocked" {
				command = fmt.Sprintf("cassor task resume %d", task.ID)
			}
			context.NextAction = ContextAction{Kind: "task", ID: task.ID, Title: task.Title, Description: task.Description, Status: task.Status, Command: command, Verification: task.Verification}
		}
	}
	if context.NextAction.Kind == "" {
		context.NextAction = ContextAction{Kind: "complete", Title: "Verify applicable criteria and complete the item", Status: item.Status, Command: fmt.Sprintf("cassor item complete %d", itemID)}
	}
	return boundScopedContext(context)
}

func criteriaForPlan(database *sql.DB, planID int64) ([]AcceptanceCriterion, error) {
	all, err := ListAcceptanceCriteria(database)
	if err != nil {
		return nil, err
	}
	values := []AcceptanceCriterion{}
	for _, criterion := range all {
		if criterion.PlanID == planID {
			values = append(values, criterion)
		}
	}
	return values, nil
}

func boundScopedContext(context ScopedContext) (ScopedContext, error) {
	encoded, err := stableScopedJSON(&context)
	if err != nil {
		return context, err
	}
	if len(encoded) <= context.MaxBytes {
		context.Bytes = len(encoded)
		return context, nil
	}
	context.Truncated = true
	for index := range context.Criteria {
		context.Omitted = appendScopedOmission(context.Omitted, ContextReference{Kind: "criterion-detail", ID: context.Criteria[index].ID, PlanID: context.Assignment.PlanID, ItemID: context.Item.ID, Status: "omitted", Command: fmt.Sprintf("cassor lifecycle criterion show %d", context.Criteria[index].ID)})
	}
	for index := range context.Assignment.Tasks {
		if context.Assignment.Tasks[index].ID == context.NextAction.ID {
			continue
		}
		context.Omitted = appendScopedOmission(context.Omitted, ContextReference{Kind: "task-detail", ID: context.Assignment.Tasks[index].ID, PlanID: context.Assignment.PlanID, ItemID: context.Item.ID, Status: "omitted", Command: context.Assignment.Tasks[index].Command})
	}
	context.Assignment.Tasks = keepTask(context.Assignment.Tasks, context.NextAction.ID)
	context.Evidence = nil
	context.Assignment.Goal = ""
	encoded, err = stableScopedJSON(&context)
	if err != nil {
		return context, err
	}
	if len(encoded) > context.MaxBytes {
		// Plan scope and decisions are useful enrichment, but the plan reference
		// is a stable, compact way to recover them when the packet is tight.
		if len(context.Decisions) > 0 {
			context.Omitted = appendScopedOmission(context.Omitted, ContextReference{Kind: "decision-detail", ID: context.Assignment.PlanID, PlanID: context.Assignment.PlanID, ItemID: context.Item.ID, Status: "omitted", Command: context.Assignment.Reference})
			context.Decisions = nil
		}
		if len(context.Scope.Included) > 0 || len(context.Scope.Excluded) > 0 {
			context.Omitted = appendScopedOmission(context.Omitted, ContextReference{Kind: "scope-detail", ID: context.Assignment.PlanID, PlanID: context.Assignment.PlanID, ItemID: context.Item.ID, Status: "omitted", Command: context.Assignment.Reference})
			context.Scope = PacketScope{}
		}
		encoded, err = stableScopedJSON(&context)
		if err != nil {
			return context, err
		}
	}
	if len(encoded) > context.MaxBytes && context.NextAction.Description != "" {
		context.Omitted = appendScopedOmission(context.Omitted, ContextReference{Kind: "next-action-description", ID: context.NextAction.ID, PlanID: context.Assignment.PlanID, ItemID: context.Item.ID, Status: "omitted", Command: fmt.Sprintf("cassor task show %d", context.NextAction.ID)})
		context.NextAction.Description = ""
		encoded, err = stableScopedJSON(&context)
		if err != nil {
			return context, err
		}
	}
	if len(encoded) > context.MaxBytes {
		return context, fmt.Errorf("--max-bytes %d is too small for required context; minimum is %d", context.MaxBytes, len(encoded))
	}
	return context, nil
}

func appendScopedOmission(values []ContextReference, value ContextReference) []ContextReference {
	for _, existing := range values {
		if existing.Kind == value.Kind && existing.ID == value.ID && existing.Command == value.Command {
			return values
		}
	}
	return append(values, value)
}

func stableScopedJSON(context *ScopedContext) ([]byte, error) {
	for attempt := 0; attempt < 4; attempt++ {
		encoded, err := json.Marshal(context)
		if err != nil {
			return nil, err
		}
		if context.Bytes == len(encoded) {
			return encoded, nil
		}
		context.Bytes = len(encoded)
	}
	return json.Marshal(context)
}

func keepTask(values []ContextReference, id int64) []ContextReference {
	kept := []ContextReference{}
	for _, value := range values {
		if value.ID == id {
			kept = append(kept, value)
		}
	}
	return kept
}
