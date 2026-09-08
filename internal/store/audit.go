package store

import "database/sql"

// CompletionCandidate is a roadmap item whose approved-plan tasks are all
// complete but whose own lifecycle status still needs an explicit decision.
type CompletionCandidate struct {
	ItemID     int64  `json:"item_id"`
	Title      string `json:"title"`
	ItemStatus string `json:"item_status"`
	TaskCount  int    `json:"task_count"`
}

// CompletionCandidates returns advisory lifecycle-reconciliation candidates.
// It deliberately never changes item status: task completion can be narrower
// than the roadmap outcome represented by its parent item.
func CompletionCandidates(database *sql.DB) ([]CompletionCandidate, error) {
	rows, err := database.Query(`
		SELECT i.id, i.title, i.status, COUNT(t.id)
		FROM roadmap_items i
		JOIN plan_revisions p ON p.roadmap_item_id = i.id AND p.status = 'Approved'
		JOIN tasks t ON t.plan_revision_id = p.id
		WHERE i.status <> 'Done'
		GROUP BY i.id, i.title, i.status
		HAVING COUNT(t.id) > 0
		   AND SUM(CASE WHEN t.status <> 'Done' THEN 1 ELSE 0 END) = 0
		ORDER BY i.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []CompletionCandidate{}
	for rows.Next() {
		var candidate CompletionCandidate
		if err := rows.Scan(&candidate.ItemID, &candidate.Title, &candidate.ItemStatus, &candidate.TaskCount); err != nil {
			return nil, err
		}
		result = append(result, candidate)
	}
	return result, rows.Err()
}
