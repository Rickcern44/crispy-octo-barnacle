package store

import (
	"database/sql"
	"fmt"
)

var lifecyclePhases = map[string]bool{"Intake": true, "Explore": true, "Define": true, "Plan": true, "Implement": true, "Verify": true, "Record": true}

type PhaseRecord struct {
	ID, ItemID         int64
	Phase              string `json:"phase"`
	Revision           int    `json:"revision"`
	Content, CreatedAt string
}
type AcceptanceCriterion struct {
	ID, ItemID                                               int64
	Title, Description, Status, VerificationMethod, Evidence string
	VerifiedAt                                               *string
}

func AddPhaseRecord(db *sql.DB, itemID int64, phase, content string) (PhaseRecord, error) {
	if !lifecyclePhases[phase] {
		return PhaseRecord{}, fmt.Errorf("unsupported lifecycle phase %q", phase)
	}
	var revision int
	if err := db.QueryRow(`SELECT COALESCE(MAX(revision),0)+1 FROM feature_phase_records WHERE roadmap_item_id=? AND phase=?`, itemID, phase).Scan(&revision); err != nil {
		return PhaseRecord{}, err
	}
	v := PhaseRecord{ItemID: itemID, Phase: phase, Revision: revision, Content: content, CreatedAt: now()}
	r, err := db.Exec(`INSERT INTO feature_phase_records(roadmap_item_id,phase,revision,content,created_at) VALUES(?,?,?,?,?)`, v.ItemID, v.Phase, v.Revision, v.Content, v.CreatedAt)
	if err != nil {
		return v, err
	}
	v.ID, _ = r.LastInsertId()
	return v, nil
}
func ListPhaseRecords(db *sql.DB) ([]PhaseRecord, error) {
	rows, err := db.Query(`SELECT id,roadmap_item_id,phase,revision,content,created_at FROM feature_phase_records ORDER BY roadmap_item_id,phase,revision`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	v := []PhaseRecord{}
	for rows.Next() {
		var x PhaseRecord
		if err := rows.Scan(&x.ID, &x.ItemID, &x.Phase, &x.Revision, &x.Content, &x.CreatedAt); err != nil {
			return nil, err
		}
		v = append(v, x)
	}
	return v, rows.Err()
}
func AddAcceptanceCriterion(db *sql.DB, itemID int64, title, description string) (AcceptanceCriterion, error) {
	v := AcceptanceCriterion{ItemID: itemID, Title: title, Description: description, Status: "Pending"}
	r, err := db.Exec(`INSERT INTO acceptance_criteria(roadmap_item_id,title,description,status,verification_method,evidence) VALUES(?,?,?,'Pending','','')`, itemID, title, description)
	if err != nil {
		return v, err
	}
	v.ID, _ = r.LastInsertId()
	return v, nil
}
func ListAcceptanceCriteria(db *sql.DB) ([]AcceptanceCriterion, error) {
	rows, err := db.Query(`SELECT id,roadmap_item_id,title,description,status,verification_method,evidence,verified_at FROM acceptance_criteria ORDER BY roadmap_item_id,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	v := []AcceptanceCriterion{}
	for rows.Next() {
		var x AcceptanceCriterion
		if err := rows.Scan(&x.ID, &x.ItemID, &x.Title, &x.Description, &x.Status, &x.VerificationMethod, &x.Evidence, &x.VerifiedAt); err != nil {
			return nil, err
		}
		v = append(v, x)
	}
	return v, rows.Err()
}
func VerifyAcceptanceCriterion(db *sql.DB, id int64, status, method, evidence string) error {
	if status != "Passed" && status != "Failed" && status != "Waived" {
		return fmt.Errorf("unsupported criterion status %q", status)
	}
	if evidence == "" {
		return fmt.Errorf("verification evidence is required")
	}
	_, err := db.Exec(`UPDATE acceptance_criteria SET status=?,verification_method=?,evidence=?,verified_at=? WHERE id=?`, status, method, evidence, now(), id)
	return err
}
func UnresolvedAcceptanceCriteria(db *sql.DB, itemID int64) (int, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM acceptance_criteria WHERE roadmap_item_id=? AND status NOT IN ('Passed','Waived')`, itemID).Scan(&n)
	return n, err
}
