package store

import "database/sql"

// TaskEvent preserves every blocking and subsequent lifecycle transition.
type TaskEvent struct {
	ID         int64  `json:"id"`
	TaskID     int64  `json:"task_id"`
	Status     string `json:"status"`
	Outcome    string `json:"outcome"`
	RecordedAt string `json:"recorded_at"`
}

func ListTaskEvents(database *sql.DB, taskID int64) ([]TaskEvent, error) {
	rows, err := database.Query(`SELECT id,task_id,status,outcome,recorded_at FROM task_events WHERE task_id=? ORDER BY id`, taskID)
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
