package store

import (
	"database/sql"
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
}
type Plan struct {
	ID           int64   `json:"id"`
	ItemID       int64   `json:"item_id"`
	Revision     int     `json:"revision"`
	Content      string  `json:"content"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"created_at"`
	ApprovedAt   *string `json:"approved_at,omitempty"`
	ApprovalNote string  `json:"approval_note,omitempty"`
}
type Task struct {
	ID          int64   `json:"id"`
	PlanID      int64   `json:"plan_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Outcome     string  `json:"outcome"`
	CreatedAt   string  `json:"created_at"`
	StartedAt   *string `json:"started_at,omitempty"`
	CompletedAt *string `json:"completed_at,omitempty"`
	BlockedAt   *string `json:"blocked_at,omitempty"`
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
	err := scanner.Scan(&value.ID, &value.Title, &value.Description, &value.Category, &value.Horizon, &value.Status, &value.Rationale, &value.CreatedAt, &value.UpdatedAt, &value.TargetDate, &value.Progress, &value.Priority, &value.Complexity, &value.Team, &value.LeadEngineer, &value.TechnicalSummary, &value.Specifications, &value.DocumentationLinks)
	return value, err
}

const itemColumns = `i.id, i.title, i.description, c.name, i.horizon, i.status, i.rationale, i.created_at, i.updated_at, i.target_date, i.progress, i.priority, i.complexity, i.team, i.lead_engineer, i.technical_summary, i.specifications, i.documentation_links FROM roadmap_items i JOIN categories c ON c.id=i.category_id`

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
func SetItemStatus(database *sql.DB, id int64, status string) error {
	result, err := database.Exec(`UPDATE roadmap_items SET status=?, updated_at=? WHERE id=? AND status='Planned'`, status, now(), id)
	if err != nil {
		return err
	}
	changed, _ := result.RowsAffected()
	if changed == 0 {
		return fmt.Errorf("only proposed roadmap items may be %s", status)
	}
	return nil
}

func scanPlan(scanner interface{ Scan(...any) error }) (Plan, error) {
	var value Plan
	err := scanner.Scan(&value.ID, &value.ItemID, &value.Revision, &value.Content, &value.Status, &value.CreatedAt, &value.ApprovedAt, &value.ApprovalNote)
	return value, err
}
func GetPlan(database *sql.DB, id int64) (Plan, error) {
	value, err := scanPlan(database.QueryRow(`SELECT id,roadmap_item_id,revision,content,status,created_at,approved_at,approval_note FROM plan_revisions WHERE id=?`, id))
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
	if status != "Ready" {
		return Plan{}, fmt.Errorf("only ready roadmap items may receive plans")
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
	return CreatePlan(database, previous.ItemID, content)
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
	if itemStatus != "Ready" {
		transaction.Rollback()
		return fmt.Errorf("only plans for ready roadmap items may be approved")
	}
	if _, err := transaction.Exec(`UPDATE plan_revisions SET status='Approved', approved_at=? WHERE id=?`, now(), id); err != nil {
		transaction.Rollback()
		return err
	}
	return transaction.Commit()
}
func ListPlans(database *sql.DB, itemID int64) ([]Plan, error) {
	rows, err := database.Query(`SELECT id,roadmap_item_id,revision,content,status,created_at,approved_at,approval_note FROM plan_revisions WHERE roadmap_item_id=? ORDER BY revision`, itemID)
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
	err := scanner.Scan(&value.ID, &value.PlanID, &value.Title, &value.Description, &value.Status, &value.Outcome, &value.CreatedAt, &value.StartedAt, &value.CompletedAt, &value.BlockedAt)
	return value, err
}
func GetTask(database *sql.DB, id int64) (Task, error) {
	value, err := scanTask(database.QueryRow(`SELECT id,plan_revision_id,title,description,status,outcome,created_at,started_at,completed_at,blocked_at FROM tasks WHERE id=?`, id))
	if err == sql.ErrNoRows {
		return value, fmt.Errorf("task %d not found", id)
	}
	return value, err
}
func AddTask(database *sql.DB, planID int64, title, description string) (Task, error) {
	plan, err := GetPlan(database, planID)
	if err != nil {
		return Task{}, err
	}
	if plan.Status != "Draft" {
		return Task{}, fmt.Errorf("tasks can only be added to a draft plan")
	}
	result, err := database.Exec(`INSERT INTO tasks(plan_revision_id,title,description,created_at) VALUES(?,?,?,?)`, planID, title, description, now())
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
	rows, err := database.Query(`SELECT id,plan_revision_id,title,description,status,outcome,created_at,started_at,completed_at,blocked_at FROM tasks WHERE plan_revision_id=? ORDER BY id`, planID)
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
	task, err := GetTask(database, id)
	if err != nil {
		return err
	}
	if status == "In Progress" {
		if task.Status != "To Do" {
			return fmt.Errorf("only pending tasks may start")
		}
		var planStatus string
		if err := database.QueryRow(`SELECT status FROM plan_revisions WHERE id=?`, task.PlanID).Scan(&planStatus); err != nil {
			return err
		}
		if planStatus != "Approved" {
			return fmt.Errorf("tasks cannot start without an approved plan revision")
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
	_, err = database.Exec(query, args...)
	return err
}
