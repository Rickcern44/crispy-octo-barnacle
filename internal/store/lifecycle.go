package store

import (
	"database/sql"
	"fmt"
	"strings"
)

var lifecyclePhases = map[string]bool{"Intake": true, "Explore": true, "Define": true, "Plan": true, "Implement": true, "Verify": true, "Record": true}

type PhaseRecord struct {
	ID, ItemID         int64
	Phase              string `json:"phase"`
	Revision           int    `json:"revision"`
	Content, CreatedAt string
}
type AcceptanceCriterion struct {
	ID, ItemID, PlanID                                       int64
	Key                                                      string `json:"key"`
	Title, Description, Status, VerificationMethod, Evidence string
	Required                                                 bool    `json:"required"`
	WaivedBy                                                 string  `json:"waived_by,omitempty"`
	WaiverReason                                             string  `json:"waiver_reason,omitempty"`
	VerifiedAt                                               *string `json:"verified_at,omitempty"`
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
	var planID sql.NullInt64
	_ = db.QueryRow(`SELECT id FROM plan_revisions WHERE roadmap_item_id=? AND status='Draft' ORDER BY revision DESC LIMIT 1`, itemID).Scan(&planID)
	key := "legacy-criterion"
	if planID.Valid {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM acceptance_criteria WHERE plan_revision_id=?`, planID.Int64).Scan(&count); err != nil {
			return AcceptanceCriterion{}, err
		}
		key = fmt.Sprintf("C%d", count+1)
	}
	return addAcceptanceCriterion(db, itemID, planID, key, title, description, true)
}

func AddAcceptanceCriterionForPlan(db *sql.DB, planID int64, key, title, description string, required bool) (AcceptanceCriterion, error) {
	var itemID int64
	if err := db.QueryRow(`SELECT roadmap_item_id FROM plan_revisions WHERE id=? AND status='Draft'`, planID).Scan(&itemID); err == sql.ErrNoRows {
		return AcceptanceCriterion{}, fmt.Errorf("draft plan %d not found", planID)
	} else if err != nil {
		return AcceptanceCriterion{}, err
	}
	if strings.TrimSpace(key) == "" {
		return AcceptanceCriterion{}, fmt.Errorf("criterion key is required")
	}
	return addAcceptanceCriterion(db, itemID, sql.NullInt64{Int64: planID, Valid: true}, key, title, description, required)
}

func addAcceptanceCriterion(db *sql.DB, itemID int64, planID sql.NullInt64, key, title, description string, required bool) (AcceptanceCriterion, error) {
	if strings.TrimSpace(title) == "" {
		return AcceptanceCriterion{}, fmt.Errorf("criterion title is required")
	}
	v := AcceptanceCriterion{ItemID: itemID, PlanID: planID.Int64, Key: key, Title: title, Description: description, Required: required, Status: "Pending"}
	r, err := db.Exec(`INSERT INTO acceptance_criteria(roadmap_item_id,plan_revision_id,criterion_key,title,description,required,status,verification_method,evidence,waived_by,waiver_reason) VALUES(?,?,?,?,?,?, 'Pending','','','','')`, itemID, nullableInt64(planID), key, title, description, boolInt(required))
	if err != nil {
		return v, err
	}
	v.ID, _ = r.LastInsertId()
	return v, nil
}
func ListAcceptanceCriteria(db *sql.DB) ([]AcceptanceCriterion, error) {
	rows, err := db.Query(`SELECT id,roadmap_item_id,COALESCE(plan_revision_id,0),criterion_key,title,description,required,status,verification_method,evidence,waived_by,waiver_reason,verified_at FROM acceptance_criteria ORDER BY roadmap_item_id,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	v := []AcceptanceCriterion{}
	for rows.Next() {
		var x AcceptanceCriterion
		if err := rows.Scan(&x.ID, &x.ItemID, &x.PlanID, &x.Key, &x.Title, &x.Description, &x.Required, &x.Status, &x.VerificationMethod, &x.Evidence, &x.WaivedBy, &x.WaiverReason, &x.VerifiedAt); err != nil {
			return nil, err
		}
		v = append(v, x)
	}
	return v, rows.Err()
}
func VerifyAcceptanceCriterion(db *sql.DB, id int64, status, method, evidence string, details ...string) error {
	if status != "Passed" && status != "Failed" && status != "Waived" {
		return fmt.Errorf("unsupported criterion status %q", status)
	}
	if strings.TrimSpace(evidence) == "" {
		return fmt.Errorf("verification evidence is required")
	}
	by, reason := "", ""
	if len(details) > 0 {
		by = strings.TrimSpace(details[0])
	}
	if len(details) > 1 {
		reason = strings.TrimSpace(details[1])
	}
	if status == "Waived" && (by == "" || reason == "") {
		return fmt.Errorf("waivers require attribution and a reason")
	}
	transaction, err := db.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	var exists int
	if err := transaction.QueryRow(`SELECT COUNT(*) FROM acceptance_criteria WHERE id=?`, id).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("acceptance criterion %d not found", id)
	}
	timestamp := now()
	if _, err := transaction.Exec(`INSERT INTO criterion_evidence(acceptance_criterion_id,status,verification_method,evidence,recorded_by,waiver_reason,recorded_at) VALUES(?,?,?,?,?,?,?)`, id, status, method, evidence, by, reason, timestamp); err != nil {
		return err
	}
	if _, err := transaction.Exec(`UPDATE acceptance_criteria SET status=?,verification_method=?,evidence=?,verified_at=?,waived_by=?,waiver_reason=? WHERE id=?`, status, method, evidence, timestamp, by, reason, id); err != nil {
		return err
	}
	return transaction.Commit()
}
func UnresolvedAcceptanceCriteria(db *sql.DB, itemID int64) (int, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM acceptance_criteria c JOIN plan_revisions p ON p.id=c.plan_revision_id WHERE p.roadmap_item_id=? AND p.active=1 AND c.required=1 AND c.status NOT IN ('Passed','Waived')`, itemID).Scan(&n)
	return n, err
}

func OpenTasksForActivePlan(db *sql.DB, itemID int64) (int, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM tasks t JOIN plan_revisions p ON p.id=t.plan_revision_id WHERE p.roadmap_item_id=? AND p.active=1 AND t.status<>'Done'`, itemID).Scan(&n)
	return n, err
}

func nullableInt64(value sql.NullInt64) any {
	if value.Valid {
		return value.Int64
	}
	return nil
}
