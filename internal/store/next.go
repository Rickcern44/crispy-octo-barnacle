package store

import (
	"database/sql"
	"fmt"
)

// NextTaskSelection identifies the task an item-next command should resume.
// A nil result means that there is no actionable work.
type NextTaskSelection struct {
	ItemID               int64  `json:"item_id"`
	ItemTitle            string `json:"item_title"`
	ItemStatus           string `json:"item_status"`
	TaskID               int64  `json:"task_id"`
	TaskTitle            string `json:"task_title"`
	TaskStatus           string `json:"task_status"`
	ScopedContextCommand string `json:"scoped_context_command"`
}

// SelectNextTask returns the deterministic next implementation task.
// In-progress tasks always win. Otherwise, the oldest pending task belonging
// to the oldest Ready or In Progress item with an active approved plan wins.
func SelectNextTask(database *sql.DB) (*NextTaskSelection, error) {
	queries := []string{
		`SELECT i.id,i.title,i.status,t.id,t.title,t.status
		 FROM tasks t
		 JOIN plan_revisions p ON p.id=t.plan_revision_id
		 JOIN roadmap_items i ON i.id=p.roadmap_item_id
		 WHERE t.status='In Progress'
		 ORDER BY t.id
		 LIMIT 1`,
		`SELECT i.id,i.title,i.status,t.id,t.title,t.status
		 FROM tasks t
		 JOIN plan_revisions p ON p.id=t.plan_revision_id AND p.status='Approved' AND p.active=1
		 JOIN roadmap_items i ON i.id=p.roadmap_item_id
		 WHERE t.status='To Do' AND i.status IN ('Ready','In Progress')
		 ORDER BY i.id,t.id
		 LIMIT 1`,
	}
	for _, query := range queries {
		selection := NextTaskSelection{}
		err := database.QueryRow(query).Scan(&selection.ItemID, &selection.ItemTitle, &selection.ItemStatus, &selection.TaskID, &selection.TaskTitle, &selection.TaskStatus)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return nil, err
		}
		selection.ScopedContextCommand = fmt.Sprintf("cassor context --item %d --role implementation --max-bytes 12000", selection.ItemID)
		return &selection, nil
	}
	return nil, nil
}
