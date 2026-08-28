package store

import "database/sql"

// Context is the intentionally small workflow snapshot consumed by agents.
type Context struct {
	ActiveTasks       []Task `json:"active_tasks"`
	ProposedItems     []Item `json:"proposed_items"`
	PlansAwaiting     []Plan `json:"plans_awaiting_approval"`
	BlockedTasks      []Task `json:"blocked_tasks"`
	RecentlyCompleted []Task `json:"recently_completed_tasks"`
}

// CompactContext omits historical detail and returns only work relevant to the
// next planning or implementation decision.
func CompactContext(database *sql.DB) (Context, error) {
	var context Context
	var err error
	if context.ActiveTasks, err = tasksByStatus(database, "In Progress", 0); err != nil {
		return context, err
	}
	if context.ProposedItems, err = itemsByStatus(database, "Planned"); err != nil {
		return context, err
	}
	if context.PlansAwaiting, err = plansByStatus(database, "Draft"); err != nil {
		return context, err
	}
	if context.BlockedTasks, err = tasksByStatus(database, "Blocked", 0); err != nil {
		return context, err
	}
	if context.RecentlyCompleted, err = tasksByStatus(database, "Done", 10); err != nil {
		return context, err
	}
	return context, nil
}

func itemsByStatus(database *sql.DB, status string) ([]Item, error) {
	rows, err := database.Query(`SELECT `+itemColumns+` WHERE i.status=? ORDER BY i.updated_at DESC, i.id DESC`, status)
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

func plansByStatus(database *sql.DB, status string) ([]Plan, error) {
	rows, err := database.Query(`SELECT id,roadmap_item_id,revision,content,status,created_at,approved_at,approval_note FROM plan_revisions WHERE status=? ORDER BY created_at, id`, status)
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

func tasksByStatus(database *sql.DB, status string, limit int) ([]Task, error) {
	query := `SELECT id,plan_revision_id,title,description,status,outcome,created_at,started_at,completed_at,blocked_at FROM tasks WHERE status=? ORDER BY COALESCE(completed_at, blocked_at, started_at, created_at) DESC, id DESC`
	arguments := []any{status}
	if limit > 0 {
		query += ` LIMIT ?`
		arguments = append(arguments, limit)
	}
	rows, err := database.Query(query, arguments...)
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
