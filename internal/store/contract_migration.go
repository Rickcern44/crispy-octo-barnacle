package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// backfillDeliveryContract upgrades legacy plan records without manufacturing
// passing evidence. Packet-shaped plans are expanded into revision-scoped
// criteria and task verification requirements. Older free-form plans receive
// an explicit pending legacy criterion so their missing contract is visible.
func backfillDeliveryContract(transaction *sql.Tx) error {
	if _, err := transaction.Exec(`
		UPDATE plan_revisions
		SET active=1
		WHERE status='Approved'
		  AND revision=(SELECT MAX(previous.revision) FROM plan_revisions previous WHERE previous.roadmap_item_id=plan_revisions.roadmap_item_id AND previous.status='Approved')`); err != nil {
		return fmt.Errorf("mark active legacy plans: %w", err)
	}

	rows, err := transaction.Query(`SELECT id,roadmap_item_id,content FROM plan_revisions ORDER BY id`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var planID, itemID int64
		var content string
		if err := rows.Scan(&planID, &itemID, &content); err != nil {
			return err
		}
		var packet PlanPacket
		if err := json.Unmarshal([]byte(content), &packet); err == nil && (len(packet.AcceptanceCriteria) > 0 || len(packet.Tasks) > 0) {
			if err := backfillPacketContract(transaction, planID, itemID, packet); err != nil {
				return err
			}
			continue
		}
		if err := addLegacyCriterion(transaction, planID, itemID); err != nil {
			return err
		}
	}
	return rows.Err()
}

func backfillPacketContract(transaction *sql.Tx, planID, itemID int64, packet PlanPacket) error {
	for index, criterion := range packet.AcceptanceCriteria {
		key := criterion.key(index)
		var count int
		if err := transaction.QueryRow(`SELECT COUNT(*) FROM acceptance_criteria WHERE plan_revision_id=? AND criterion_key=?`, planID, key).Scan(&count); err != nil {
			return err
		}
		if count == 0 {
			if _, err := transaction.Exec(`INSERT INTO acceptance_criteria(roadmap_item_id,plan_revision_id,criterion_key,title,description,required,status,verification_method,evidence,verified_at) VALUES(?,?,?,?,?,?, 'Pending','','',NULL)`, itemID, planID, key, criterion.Title, criterion.Description, boolInt(criterion.RequiredOrDefault())); err != nil {
				return fmt.Errorf("backfill criterion %s for plan %d: %w", key, planID, err)
			}
		}
	}
	for index, task := range packet.Tasks {
		if index == 0 {
			// Keep the query below deterministic while avoiding an UPDATE that
			// could accidentally touch tasks from another plan.
		}
		var taskID int64
		err := transaction.QueryRow(`SELECT id FROM tasks WHERE plan_revision_id=? ORDER BY id LIMIT 1 OFFSET ?`, planID, index).Scan(&taskID)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return err
		}
		verification, err := json.Marshal(task.Verification)
		if err != nil {
			return err
		}
		if _, err := transaction.Exec(`UPDATE tasks SET verification=? WHERE id=? AND verification='[]'`, string(verification), taskID); err != nil {
			return err
		}
	}
	return nil
}

func addLegacyCriterion(transaction *sql.Tx, planID, itemID int64) error {
	var count int
	if err := transaction.QueryRow(`SELECT COUNT(*) FROM acceptance_criteria WHERE plan_revision_id=?`, planID).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := transaction.Exec(`INSERT INTO acceptance_criteria(roadmap_item_id,plan_revision_id,criterion_key,title,description,required,status,verification_method,evidence,verified_at) VALUES(?,?,?,'Legacy acceptance criteria require explicit verification.','This plan predates revision-scoped criteria. Verify or waive it with attributed evidence before relying on completion.',1,'Pending','','',NULL)`, itemID, planID, fmt.Sprintf("legacy-plan-%d", planID))
	return err
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
