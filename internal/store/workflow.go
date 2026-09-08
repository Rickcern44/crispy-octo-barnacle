package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type Category struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}
type Item struct {
	ID                 int64  `json:"id"`
	Title              string `json:"title"`
	Description        string `json:"description"`
	Category           string `json:"category"`
	Horizon            string `json:"horizon"`
	Status             string `json:"status"`
	Rationale          string `json:"rationale"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
	TargetDate         string `json:"target_date"`
	Progress           int    `json:"progress"`
	Priority           string `json:"priority"`
	Complexity         string `json:"complexity"`
	Team               string `json:"team"`
	LeadEngineer       string `json:"lead_engineer"`
	TechnicalSummary   string `json:"technical_summary"`
	Specifications     string `json:"specifications"`
	DocumentationLinks string `json:"documentation_links"`
	FeatureType        string `json:"feature_type"`
	CurrentState       string `json:"current_state"`
}
type FeatureRelationship struct {
	ID               int64  `json:"id"`
	SourceItemID     int64  `json:"source_item_id"`
	SourceItemTitle  string `json:"source_item_title"`
	TargetItemID     int64  `json:"target_item_id"`
	TargetItemTitle  string `json:"target_item_title"`
	RelationshipType string `json:"relationship_type"`
	CreatedAt        string `json:"created_at"`
}
type Plan struct {
	ID           int64   `json:"id"`
	ItemID       int64   `json:"item_id"`
	Revision     int     `json:"revision"`
	Content      string  `json:"content"`
	Status       string  `json:"status"`
	Active       bool    `json:"active"`
	CreatedAt    string  `json:"created_at"`
	ApprovedAt   *string `json:"approved_at,omitempty"`
	ApprovalNote string  `json:"approval_note,omitempty"`
}
type Task struct {
	ID           int64    `json:"id"`
	PlanID       int64    `json:"plan_id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Verification []string `json:"verification"`
	Status       string   `json:"status"`
	Outcome      string   `json:"outcome"`
	CreatedAt    string   `json:"created_at"`
	StartedAt    *string  `json:"started_at,omitempty"`
	CompletedAt  *string  `json:"completed_at,omitempty"`
	BlockedAt    *string  `json:"blocked_at,omitempty"`
}

func Open(path string) (*sql.DB, error) {
	database, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := database.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		database.Close()
		return nil, err
	}
	return database, nil
}
func now() string { return time.Now().UTC().Format(time.RFC3339Nano) }

func ListCategories(database *sql.DB) ([]Category, error) {
	rows, err := database.Query(`SELECT id, name, created_at FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Category{}
	for rows.Next() {
		var value Category
		if err := rows.Scan(&value.ID, &value.Name, &value.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}
func AddCategory(database *sql.DB, name string) (Category, error) {
	var value Category
	value.CreatedAt = now()
	result, err := database.Exec(`INSERT INTO categories(name, created_at) VALUES(?, ?)`, name, value.CreatedAt)
	if err != nil {
		return value, fmt.Errorf("add category: %w", err)
	}
	value.ID, err = result.LastInsertId()
	value.Name = name
	return value, err
}

func AddItem(database *sql.DB, title, description, category, horizon, rationale string) (Item, error) {
	var value Item
	timestamp := now()
	result, err := database.Exec(`INSERT INTO roadmap_items(title, description, category_id, horizon, rationale, created_at, updated_at) SELECT ?, ?, id, ?, ?, ?, ? FROM categories WHERE name = ?`, title, description, horizon, rationale, timestamp, timestamp, category)
	if err != nil {
		return value, fmt.Errorf("add item: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil || id == 0 {
		return value, fmt.Errorf("add item: category %q does not exist", category)
	}
	return GetItem(database, id)
}
func scanItem(scanner interface{ Scan(...any) error }) (Item, error) {
	var value Item
	err := scanner.Scan(&value.ID, &value.Title, &value.Description, &value.Category, &value.Horizon, &value.Status, &value.Rationale, &value.CreatedAt, &value.UpdatedAt, &value.TargetDate, &value.Progress, &value.Priority, &value.Complexity, &value.Team, &value.LeadEngineer, &value.TechnicalSummary, &value.Specifications, &value.DocumentationLinks, &value.FeatureType, &value.CurrentState)
	return value, err
}

const itemColumns = `i.id, i.title, i.description, c.name, i.horizon, i.status, i.rationale, i.created_at, i.updated_at, i.target_date, i.progress, i.priority, i.complexity, i.team, i.lead_engineer, i.technical_summary, i.specifications, i.documentation_links, i.feature_type, i.current_state FROM roadmap_items i JOIN categories c ON c.id=i.category_id`

func UpdateItemMetadata(database *sql.DB, id int64, targetDate string, progress int, priority, complexity, team, lead, summary, specs, links string) (Item, error) {
	_, err := database.Exec(`UPDATE roadmap_items SET target_date=?,progress=?,priority=?,complexity=?,team=?,lead_engineer=?,technical_summary=?,specifications=?,documentation_links=?,updated_at=? WHERE id=?`, targetDate, progress, priority, complexity, team, lead, summary, specs, links, now(), id)
	if err != nil {
		return Item{}, err
	}
	return GetItem(database, id)
}

func GetItem(database *sql.DB, id int64) (Item, error) {
	value, err := scanItem(database.QueryRow(`SELECT `+itemColumns+` WHERE i.id=?`, id))
	if err == sql.ErrNoRows {
		return value, fmt.Errorf("roadmap item %d not found", id)
	}
	return value, err
}
func ListItems(database *sql.DB) ([]Item, error) {
	rows, err := database.Query(`SELECT ` + itemColumns + ` ORDER BY i.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []Item{}
	for rows.Next() {
		value, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}
func UpdateItem(database *sql.DB, id int64, title, description, category, horizon, rationale *string) (Item, error) {
	item, err := GetItem(database, id)
	if err != nil {
		return item, err
	}
	if title != nil {
		item.Title = *title
	}
	if description != nil {
		item.Description = *description
	}
	if category != nil {
		item.Category = *category
	}
	if horizon != nil {
		item.Horizon = *horizon
	}
	if rationale != nil {
		item.Rationale = *rationale
	}
	result, err := database.Exec(`UPDATE roadmap_items SET title=?, description=?, category_id=(SELECT id FROM categories WHERE name=?), horizon=?, rationale=?, updated_at=? WHERE id=?`, item.Title, item.Description, item.Category, item.Horizon, item.Rationale, now(), id)
	if err != nil {
		return item, err
	}
	changed, _ := result.RowsAffected()
	if changed == 0 {
		return item, fmt.Errorf("category %q does not exist", item.Category)
	}
	return GetItem(database, id)
}

func UpdateItemDossier(database *sql.DB, id int64, featureType, currentState *string) (Item, error) {
	item, err := GetItem(database, id)
	if err != nil {
		return item, err
	}
	if currentState != nil {
		return item, fmt.Errorf("current state must be accepted with a completed linked change")
	}
	if featureType != nil {
		if !validFeatureType(*featureType) {
			return item, fmt.Errorf("unsupported feature type %q", *featureType)
		}
		item.FeatureType = *featureType
	}
	if _, err := database.Exec(`UPDATE roadmap_items SET feature_type=?,current_state=?,updated_at=? WHERE id=?`, item.FeatureType, item.CurrentState, now(), id); err != nil {
		return item, err
	}
	return GetItem(database, id)
}

func AddFeatureRelationship(database *sql.DB, sourceItemID, targetItemID int64, relationshipType string) (FeatureRelationship, error) {
	if !validRelationshipType(relationshipType) {
		return FeatureRelationship{}, fmt.Errorf("unsupported relationship type %q", relationshipType)
	}
	if sourceItemID == targetItemID {
		return FeatureRelationship{}, fmt.Errorf("a feature cannot relate to itself")
	}
	for _, itemID := range []int64{sourceItemID, targetItemID} {
		if _, err := GetItem(database, itemID); err != nil {
			return FeatureRelationship{}, err
		}
	}
	var exists bool
	if err := database.QueryRow(`SELECT EXISTS(SELECT 1 FROM feature_relationships WHERE source_item_id=? AND target_item_id=? AND relationship_type=?)`, sourceItemID, targetItemID, relationshipType).Scan(&exists); err != nil {
		return FeatureRelationship{}, err
	}
	if exists {
		return FeatureRelationship{}, fmt.Errorf("feature relationship already exists")
	}
	result, err := database.Exec(`INSERT INTO feature_relationships(source_item_id,target_item_id,relationship_type,created_at) VALUES(?,?,?,?)`, sourceItemID, targetItemID, relationshipType, now())
	if err != nil {
		return FeatureRelationship{}, fmt.Errorf("add feature relationship: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return FeatureRelationship{}, err
	}
	return getFeatureRelationship(database, id)
}

func ListFeatureRelationships(database *sql.DB, itemID int64) ([]FeatureRelationship, error) {
	if _, err := GetItem(database, itemID); err != nil {
		return nil, err
	}
	return featureRelationships(database, `WHERE r.source_item_id=? OR r.target_item_id=?`, itemID, itemID)
}

func ListAllFeatureRelationships(database *sql.DB) ([]FeatureRelationship, error) {
	return featureRelationships(database, ``)
}

func getFeatureRelationship(database *sql.DB, id int64) (FeatureRelationship, error) {
	values, err := featureRelationships(database, `WHERE r.id=?`, id)
	if err != nil {
		return FeatureRelationship{}, err
	}
	if len(values) == 0 {
		return FeatureRelationship{}, fmt.Errorf("feature relationship %d not found", id)
	}
	return values[0], nil
}

func featureRelationships(database *sql.DB, where string, arguments ...any) ([]FeatureRelationship, error) {
	query := `SELECT r.id,r.source_item_id,s.title,r.target_item_id,t.title,r.relationship_type,r.created_at FROM feature_relationships r JOIN roadmap_items s ON s.id=r.source_item_id JOIN roadmap_items t ON t.id=r.target_item_id ` + where + ` ORDER BY r.id`
	rows, err := database.Query(query, arguments...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []FeatureRelationship{}
	for rows.Next() {
		var value FeatureRelationship
		if err := rows.Scan(&value.ID, &value.SourceItemID, &value.SourceItemTitle, &value.TargetItemID, &value.TargetItemTitle, &value.RelationshipType, &value.CreatedAt); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

func validFeatureType(value string) bool {
	return value == "Capability" || value == "Change" || value == "Gap"
}

func validRelationshipType(value string) bool {
	return value == "extends" || value == "depends_on" || value == "replaces"
}
func TransitionItemStatus(database *sql.DB, id int64, status string) error {
	previous, ok := map[string]string{
		"Ready":       "Planned",
		"In Progress": "Ready",
		"Done":        "In Progress",
		"Won’t Do":    "Planned",
	}[status]
	if !ok {
		return fmt.Errorf("unsupported roadmap item status %q", status)
	}
	if status == "Done" {
		unresolved, err := UnresolvedAcceptanceCriteria(database, id)
		if err != nil {
			return err
		}
		if unresolved > 0 {
			return fmt.Errorf("roadmap item has %d acceptance criteria that have not passed or been waived", unresolved)
		}
		openTasks, err := OpenTasksForActivePlan(database, id)
		if err != nil {
			return err
		}
		if openTasks > 0 {
			return fmt.Errorf("roadmap item has %d active-plan tasks that are not done", openTasks)
		}
	}
	result, err := database.Exec(`UPDATE roadmap_items SET status=?, updated_at=? WHERE id=? AND status=?`, status, now(), id, previous)
	if err != nil {
		return err
	}
	changed, _ := result.RowsAffected()
	if changed == 0 {
		return fmt.Errorf("roadmap item must be %s before it can be %s", previous, status)
	}
	return nil
}

// SetItemStatus is retained for callers that advance proposed items to Ready or Won’t Do.
func SetItemStatus(database *sql.DB, id int64, status string) error {
	return TransitionItemStatus(database, id, status)
}

func scanPlan(scanner interface{ Scan(...any) error }) (Plan, error) {
	var value Plan
	err := scanner.Scan(&value.ID, &value.ItemID, &value.Revision, &value.Content, &value.Status, &value.Active, &value.CreatedAt, &value.ApprovedAt, &value.ApprovalNote)
	return value, err
}
func GetPlan(database *sql.DB, id int64) (Plan, error) {
	value, err := scanPlan(database.QueryRow(`SELECT id,roadmap_item_id,revision,content,status,active,created_at,approved_at,approval_note FROM plan_revisions WHERE id=?`, id))
	if err == sql.ErrNoRows {
		return value, fmt.Errorf("plan %d not found", id)
	}
	return value, err
}
func CreatePlan(database *sql.DB, itemID int64, content string) (Plan, error) {
	var status string
	if err := database.QueryRow(`SELECT status FROM roadmap_items WHERE id=?`, itemID).Scan(&status); err == sql.ErrNoRows {
		return Plan{}, fmt.Errorf("roadmap item %d not found", itemID)
	} else if err != nil {
		return Plan{}, err
	}
	if status != "Ready" && status != "In Progress" {
		return Plan{}, fmt.Errorf("only ready or in-progress roadmap items may receive plans")
	}
	var revision int
	if err := database.QueryRow(`SELECT COALESCE(MAX(revision),0)+1 FROM plan_revisions WHERE roadmap_item_id=?`, itemID).Scan(&revision); err != nil {
		return Plan{}, err
	}
	result, err := database.Exec(`INSERT INTO plan_revisions(roadmap_item_id,revision,content,created_at) VALUES(?,?,?,?)`, itemID, revision, content, now())
	if err != nil {
		return Plan{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Plan{}, err
	}
	return GetPlan(database, id)
}
func RevisePlan(database *sql.DB, id int64, content string) (Plan, error) {
	previous, err := GetPlan(database, id)
	if err != nil {
		return Plan{}, err
	}
	if previous.Status != "Approved" {
		return Plan{}, fmt.Errorf("only approved plans may be revised")
	}
	plan, err := CreatePlan(database, previous.ItemID, content)
	if err != nil {
		return Plan{}, err
	}
	if err := carryForwardCriteria(database, previous.ID, plan.ID); err != nil {
		return Plan{}, err
	}
	if err := carryForwardTasks(database, previous.ID, plan.ID); err != nil {
		return Plan{}, err
	}
	return GetPlan(database, plan.ID)
}

func carryForwardCriteria(database *sql.DB, previousPlanID, planID int64) error {
	var itemID int64
	if err := database.QueryRow(`SELECT roadmap_item_id FROM plan_revisions WHERE id=?`, planID).Scan(&itemID); err != nil {
		return err
	}
	rows, err := database.Query(`SELECT criterion_key,title,description,required,status,verification_method,evidence,verified_at,waived_by,waiver_reason FROM acceptance_criteria WHERE plan_revision_id=? ORDER BY id`, previousPlanID)
	if err != nil {
		return err
	}
	values := []struct {
		key, title, description, status, method, evidence, waivedBy, waiverReason string
		required                                                                  bool
		verifiedAt                                                                *string
	}{}
	for rows.Next() {
		var key, title, description, status, method, evidence, waivedBy, waiverReason string
		var required bool
		var verifiedAt *string
		if err := rows.Scan(&key, &title, &description, &required, &status, &method, &evidence, &verifiedAt, &waivedBy, &waiverReason); err != nil {
			rows.Close()
			return err
		}
		values = append(values, struct {
			key, title, description, status, method, evidence, waivedBy, waiverReason string
			required                                                                  bool
			verifiedAt                                                                *string
		}{key, title, description, status, method, evidence, waivedBy, waiverReason, required, verifiedAt})
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, value := range values {
		if _, err := database.Exec(`INSERT INTO acceptance_criteria(roadmap_item_id,plan_revision_id,criterion_key,title,description,required,status,verification_method,evidence,verified_at,waived_by,waiver_reason) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, itemID, planID, value.key, value.title, value.description, boolInt(value.required), value.status, value.method, value.evidence, value.verifiedAt, value.waivedBy, value.waiverReason); err != nil {
			return err
		}
	}
	return nil
}

func validatePlanContract(transaction *sql.Tx, planID int64) error {
	var criteria, tasks, incompleteTasks int
	if err := transaction.QueryRow(`SELECT COUNT(*) FROM acceptance_criteria WHERE plan_revision_id=? AND trim(title)<>''`, planID).Scan(&criteria); err != nil {
		return err
	}
	if criteria == 0 {
		return fmt.Errorf("plan requires at least one meaningful acceptance criterion")
	}
	if err := transaction.QueryRow(`SELECT COUNT(*) FROM tasks WHERE plan_revision_id=?`, planID).Scan(&tasks); err != nil {
		return err
	}
	if tasks == 0 {
		return fmt.Errorf("plan requires at least one task")
	}
	if err := transaction.QueryRow(`SELECT COUNT(*) FROM tasks WHERE plan_revision_id=? AND verification='[]'`, planID).Scan(&incompleteTasks); err != nil {
		return err
	}
	if incompleteTasks > 0 {
		return fmt.Errorf("plan tasks require verification requirements")
	}
	return nil
}

func carryForwardTasks(database *sql.DB, previousPlanID, planID int64) error {
	rows, err := database.Query(`SELECT title,description,verification FROM tasks WHERE plan_revision_id=? AND status<>'Done' ORDER BY id`, previousPlanID)
	if err != nil {
		return err
	}
	values := [][3]string{}
	for rows.Next() {
		var value [3]string
		if err := rows.Scan(&value[0], &value[1], &value[2]); err != nil {
			rows.Close()
			return err
		}
		values = append(values, value)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, value := range values {
		if _, err := database.Exec(`INSERT INTO tasks(plan_revision_id,title,description,verification,created_at) VALUES(?,?,?,?,?)`, planID, value[0], value[1], value[2], now()); err != nil {
			return err
		}
	}
	return nil
}
func ApprovePlan(database *sql.DB, id int64) error {
	transaction, err := database.Begin()
	if err != nil {
		return err
	}
	var itemID int64
	var status string
	if err := transaction.QueryRow(`SELECT roadmap_item_id,status FROM plan_revisions WHERE id=?`, id).Scan(&itemID, &status); err != nil {
		transaction.Rollback()
		if err == sql.ErrNoRows {
			return fmt.Errorf("plan %d not found", id)
		}
		return err
	}
	if status != "Draft" {
		transaction.Rollback()
		return fmt.Errorf("only draft plans may be approved")
	}
	var itemStatus string
	if err := transaction.QueryRow(`SELECT status FROM roadmap_items WHERE id=?`, itemID).Scan(&itemStatus); err != nil {
		transaction.Rollback()
		return err
	}
	if itemStatus != "Ready" && itemStatus != "In Progress" {
		transaction.Rollback()
		return fmt.Errorf("only plans for ready or in-progress roadmap items may be approved")
	}
	if err := validatePlanContract(transaction, id); err != nil {
		transaction.Rollback()
		return err
	}
	if _, err := transaction.Exec(`UPDATE plan_revisions SET active=0 WHERE roadmap_item_id=?`, itemID); err != nil {
		transaction.Rollback()
		return err
	}
	if _, err := transaction.Exec(`UPDATE plan_revisions SET status='Approved', active=1, approved_at=? WHERE id=?`, now(), id); err != nil {
		transaction.Rollback()
		return err
	}
	return transaction.Commit()
}
func ListPlans(database *sql.DB, itemID int64) ([]Plan, error) {
	rows, err := database.Query(`SELECT id,roadmap_item_id,revision,content,status,active,created_at,approved_at,approval_note FROM plan_revisions WHERE roadmap_item_id=? ORDER BY revision`, itemID)
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

func scanTask(scanner interface{ Scan(...any) error }) (Task, error) {
	var value Task
	var verification string
	err := scanner.Scan(&value.ID, &value.PlanID, &value.Title, &value.Description, &verification, &value.Status, &value.Outcome, &value.CreatedAt, &value.StartedAt, &value.CompletedAt, &value.BlockedAt)
	if err == nil {
		if decodeErr := json.Unmarshal([]byte(verification), &value.Verification); decodeErr != nil {
			return value, fmt.Errorf("decode task verification: %w", decodeErr)
		}
	}
	return value, err
}
func GetTask(database *sql.DB, id int64) (Task, error) {
	value, err := scanTask(database.QueryRow(`SELECT id,plan_revision_id,title,description,verification,status,outcome,created_at,started_at,completed_at,blocked_at FROM tasks WHERE id=?`, id))
	if err == sql.ErrNoRows {
		return value, fmt.Errorf("task %d not found", id)
	}
	return value, err
}
func AddTask(database *sql.DB, planID int64, title, description string, verificationValues ...[]string) (Task, error) {
	plan, err := GetPlan(database, planID)
	if err != nil {
		return Task{}, err
	}
	if plan.Status != "Draft" {
		return Task{}, fmt.Errorf("tasks can only be added to a draft plan")
	}
	verification := []string{}
	if len(verificationValues) > 0 {
		verification = verificationValues[0]
	}
	encoded, err := json.Marshal(verification)
	if err != nil {
		return Task{}, err
	}
	result, err := database.Exec(`INSERT INTO tasks(plan_revision_id,title,description,verification,created_at) VALUES(?,?,?,?,?)`, planID, title, description, string(encoded), now())
	if err != nil {
		return Task{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Task{}, err
	}
	return GetTask(database, id)
}
func ListTasks(database *sql.DB, planID int64) ([]Task, error) {
	rows, err := database.Query(`SELECT id,plan_revision_id,title,description,verification,status,outcome,created_at,started_at,completed_at,blocked_at FROM tasks WHERE plan_revision_id=? ORDER BY id`, planID)
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
func SetTaskStatus(database *sql.DB, id int64, status, outcome string) error {
	if status != "In Progress" && status != "Done" && status != "Blocked" {
		return fmt.Errorf("unsupported task status %q", status)
	}
	task, err := GetTask(database, id)
	if err != nil {
		return err
	}
	if status == "In Progress" {
		if task.Status != "To Do" {
			return fmt.Errorf("only pending tasks may start")
		}
		var planStatus string
		var active bool
		if err := database.QueryRow(`SELECT status,active FROM plan_revisions WHERE id=?`, task.PlanID).Scan(&planStatus, &active); err != nil {
			return err
		}
		if planStatus != "Approved" || !active {
			return fmt.Errorf("tasks cannot start without the active approved plan revision")
		}
	} else if task.Status != "In Progress" {
		return fmt.Errorf("only active tasks may be %s", status)
	}
	timestamp := now()
	query := `UPDATE tasks SET status=?, outcome=?`
	args := []any{status, outcome}
	if status == "In Progress" {
		query += `, started_at=?`
		args = append(args, timestamp)
	} else if status == "Done" {
		query += `, completed_at=?`
		args = append(args, timestamp)
	} else {
		query += `, blocked_at=?`
		args = append(args, timestamp)
	}
	query += ` WHERE id=?`
	args = append(args, id)
	if _, err = database.Exec(query, args...); err != nil {
		return err
	}
	_, err = database.Exec(`INSERT INTO task_events(task_id,status,outcome,recorded_at) VALUES(?,?,?,?)`, id, status, outcome, timestamp)
	return err
}

func ResumeTask(database *sql.DB, id int64) error {
	task, err := GetTask(database, id)
	if err != nil {
		return err
	}
	if task.Status != "Blocked" {
		return fmt.Errorf("only blocked tasks may resume")
	}
	var active bool
	var planStatus string
	if err := database.QueryRow(`SELECT status,active FROM plan_revisions WHERE id=?`, task.PlanID).Scan(&planStatus, &active); err != nil {
		return err
	}
	if planStatus != "Approved" || !active {
		return fmt.Errorf("tasks cannot resume without the active approved plan revision")
	}
	timestamp := now()
	if _, err := database.Exec(`UPDATE tasks SET status='In Progress',started_at=? WHERE id=?`, timestamp, id); err != nil {
		return err
	}
	_, err = database.Exec(`INSERT INTO task_events(task_id,status,outcome,recorded_at) VALUES(?,?,?,?)`, id, "In Progress", "resumed", timestamp)
	return err
}
