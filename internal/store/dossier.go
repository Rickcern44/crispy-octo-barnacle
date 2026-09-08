package store

import (
	"database/sql"
	"fmt"
	"strings"
)

type FeatureChangeLink struct {
	ID               int64  `json:"id"`
	RelationshipType string `json:"relationship_type"`
	ChangeItemID     int64  `json:"change_item_id"`
	ChangeItemTitle  string `json:"change_item_title"`
	CapabilityItemID int64  `json:"capability_item_id"`
	CapabilityTitle  string `json:"capability_title"`
	CreatedAt        string `json:"created_at"`
}

type CapabilityStateRecord struct {
	ID               int64  `json:"id"`
	CapabilityItemID int64  `json:"capability_item_id"`
	SourceChangeID   *int64 `json:"source_change_item_id,omitempty"`
	State            string `json:"state"`
	AcceptedBy       string `json:"accepted_by"`
	AcceptedAt       string `json:"accepted_at"`
	CreatedAt        string `json:"created_at"`
}

type DossierArtifact struct {
	ID           int64   `json:"id"`
	ItemID       int64   `json:"item_id"`
	Kind         string  `json:"kind"`
	AuthorRole   string  `json:"author_role"`
	Status       string  `json:"status"`
	Summary      string  `json:"summary"`
	Evidence     string  `json:"evidence"`
	SupersedesID *int64  `json:"supersedes_id,omitempty"`
	CreatedAt    string  `json:"created_at"`
	AcceptedAt   *string `json:"accepted_at,omitempty"`
	AcceptedBy   string  `json:"accepted_by,omitempty"`
}

func ListAllFeatureChangeLinks(database *sql.DB) ([]FeatureChangeLink, error) {
	rows, err := database.Query(`SELECT l.id,l.change_item_id,c.title,l.capability_item_id,p.title,l.created_at
		FROM feature_change_links l
		JOIN roadmap_items c ON c.id=l.change_item_id
		JOIN roadmap_items p ON p.id=l.capability_item_id
		ORDER BY l.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []FeatureChangeLink{}
	for rows.Next() {
		var value FeatureChangeLink
		if err := rows.Scan(&value.ID, &value.ChangeItemID, &value.ChangeItemTitle, &value.CapabilityItemID, &value.CapabilityTitle, &value.CreatedAt); err != nil {
			return nil, err
		}
		value.RelationshipType = "change_of"
		values = append(values, value)
	}
	return values, rows.Err()
}

func AddFeatureChangeLink(database *sql.DB, changeItemID, capabilityItemID int64) (FeatureChangeLink, error) {
	change, err := GetItem(database, changeItemID)
	if err != nil {
		return FeatureChangeLink{}, err
	}
	capability, err := GetItem(database, capabilityItemID)
	if err != nil {
		return FeatureChangeLink{}, err
	}
	if change.FeatureType != "Change" {
		return FeatureChangeLink{}, fmt.Errorf("item %d must have feature type Change", changeItemID)
	}
	if capability.FeatureType != "Capability" {
		return FeatureChangeLink{}, fmt.Errorf("item %d must have feature type Capability", capabilityItemID)
	}
	if changeItemID == capabilityItemID {
		return FeatureChangeLink{}, fmt.Errorf("a change cannot link to itself")
	}
	result, err := database.Exec(`INSERT INTO feature_change_links(change_item_id,capability_item_id,created_at) VALUES(?,?,?)`, changeItemID, capabilityItemID, now())
	if err != nil {
		return FeatureChangeLink{}, fmt.Errorf("link change to capability: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return FeatureChangeLink{}, err
	}
	values, err := featureChangeLinks(database, "WHERE l.id=?", id)
	if err != nil {
		return FeatureChangeLink{}, err
	}
	return values[0], nil
}

func ListFeatureChangeLinks(database *sql.DB, itemID int64) ([]FeatureChangeLink, error) {
	if _, err := GetItem(database, itemID); err != nil {
		return nil, err
	}
	return featureChangeLinks(database, "WHERE l.change_item_id=? OR l.capability_item_id=?", itemID, itemID)
}

func featureChangeLinks(database *sql.DB, where string, args ...any) ([]FeatureChangeLink, error) {
	rows, err := database.Query(`SELECT l.id,l.change_item_id,c.title,l.capability_item_id,p.title,l.created_at
		FROM feature_change_links l
		JOIN roadmap_items c ON c.id=l.change_item_id
		JOIN roadmap_items p ON p.id=l.capability_item_id `+where+` ORDER BY l.id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []FeatureChangeLink{}
	for rows.Next() {
		var value FeatureChangeLink
		if err := rows.Scan(&value.ID, &value.ChangeItemID, &value.ChangeItemTitle, &value.CapabilityItemID, &value.CapabilityTitle, &value.CreatedAt); err != nil {
			return nil, err
		}
		value.RelationshipType = "change_of"
		values = append(values, value)
	}
	return values, rows.Err()
}

func AddDossierArtifact(database *sql.DB, itemID int64, kind, authorRole, summary, evidence string, supersedesID *int64) (DossierArtifact, error) {
	if _, err := GetItem(database, itemID); err != nil {
		return DossierArtifact{}, err
	}
	if strings.TrimSpace(kind) == "" || strings.TrimSpace(authorRole) == "" || strings.TrimSpace(summary) == "" {
		return DossierArtifact{}, fmt.Errorf("artifact kind, author role, and summary are required")
	}
	if supersedesID != nil {
		var previousItemID int64
		if err := database.QueryRow(`SELECT roadmap_item_id FROM dossier_artifacts WHERE id=?`, *supersedesID).Scan(&previousItemID); err != nil {
			return DossierArtifact{}, fmt.Errorf("superseded artifact %d not found", *supersedesID)
		}
		if previousItemID != itemID {
			return DossierArtifact{}, fmt.Errorf("superseded artifact must belong to item %d", itemID)
		}
	}
	result, err := database.Exec(`INSERT INTO dossier_artifacts(roadmap_item_id,kind,author_role,status,summary,evidence,supersedes_id,created_at) VALUES(?,?,?,'Draft',?,?,?,?)`, itemID, kind, authorRole, summary, evidence, supersedesID, now())
	if err != nil {
		return DossierArtifact{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return DossierArtifact{}, err
	}
	values, err := dossierArtifacts(database, "WHERE a.id=?", id)
	if err != nil {
		return DossierArtifact{}, err
	}
	return values[0], nil
}

func AcceptDossierArtifact(database *sql.DB, artifactID int64, acceptedBy string) error {
	if strings.TrimSpace(acceptedBy) == "" {
		return fmt.Errorf("artifact acceptance requires attribution")
	}
	result, err := database.Exec(`UPDATE dossier_artifacts SET status='Accepted',accepted_at=?,accepted_by=? WHERE id=? AND status='Draft' AND trim(evidence)<>''`, now(), acceptedBy, artifactID)
	if err != nil {
		return err
	}
	changed, _ := result.RowsAffected()
	if changed == 0 {
		return fmt.Errorf("artifact %d must be a draft with evidence before acceptance", artifactID)
	}
	return nil
}

func AcceptCapabilityState(database *sql.DB, capabilityItemID, sourceChangeID int64, state, acceptedBy string) error {
	if strings.TrimSpace(state) == "" || strings.TrimSpace(acceptedBy) == "" {
		return fmt.Errorf("accepted capability state and attribution are required")
	}
	capability, err := GetItem(database, capabilityItemID)
	if err != nil {
		return err
	}
	if capability.FeatureType != "Capability" {
		return fmt.Errorf("item %d must have feature type Capability", capabilityItemID)
	}
	change, err := GetItem(database, sourceChangeID)
	if err != nil {
		return err
	}
	if change.FeatureType != "Change" || change.Status != "Done" {
		return fmt.Errorf("source change %d must be a completed Change", sourceChangeID)
	}
	var linked bool
	if err := database.QueryRow(`SELECT EXISTS(SELECT 1 FROM feature_change_links WHERE change_item_id=? AND capability_item_id=?)`, sourceChangeID, capabilityItemID).Scan(&linked); err != nil {
		return err
	}
	if !linked {
		return fmt.Errorf("source change %d is not linked to capability %d", sourceChangeID, capabilityItemID)
	}
	timestamp := now()
	transaction, err := database.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	if _, err := transaction.Exec(`INSERT INTO capability_state_history(capability_item_id,source_change_item_id,state,accepted_by,accepted_at,created_at) VALUES(?,?,?,?,?,?)`, capabilityItemID, sourceChangeID, state, acceptedBy, timestamp, timestamp); err != nil {
		return err
	}
	if _, err := transaction.Exec(`UPDATE roadmap_items SET current_state=?,updated_at=? WHERE id=?`, state, timestamp, capabilityItemID); err != nil {
		return err
	}
	return transaction.Commit()
}

func ListCapabilityStateHistory(database *sql.DB) ([]CapabilityStateRecord, error) {
	rows, err := database.Query(`SELECT id,capability_item_id,source_change_item_id,state,accepted_by,accepted_at,created_at FROM capability_state_history ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []CapabilityStateRecord{}
	for rows.Next() {
		var value CapabilityStateRecord
		var source sql.NullInt64
		if err := rows.Scan(&value.ID, &value.CapabilityItemID, &source, &value.State, &value.AcceptedBy, &value.AcceptedAt, &value.CreatedAt); err != nil {
			return nil, err
		}
		if source.Valid {
			value.SourceChangeID = &source.Int64
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

func ListCapabilityStateHistoryForItem(database *sql.DB, itemID int64) ([]CapabilityStateRecord, error) {
	if _, err := GetItem(database, itemID); err != nil {
		return nil, err
	}
	rows, err := database.Query(`SELECT id,capability_item_id,source_change_item_id,state,accepted_by,accepted_at,created_at FROM capability_state_history WHERE capability_item_id=? ORDER BY id`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCapabilityStateHistory(rows)
}

func ListDossierArtifacts(database *sql.DB) ([]DossierArtifact, error) {
	return dossierArtifacts(database, "")
}

func ListDossierArtifactsForItem(database *sql.DB, itemID int64) ([]DossierArtifact, error) {
	if _, err := GetItem(database, itemID); err != nil {
		return nil, err
	}
	return dossierArtifacts(database, "WHERE a.roadmap_item_id=?", itemID)
}

func dossierArtifacts(database *sql.DB, where string, args ...any) ([]DossierArtifact, error) {
	rows, err := database.Query(`SELECT id,roadmap_item_id,kind,author_role,status,summary,evidence,supersedes_id,created_at,accepted_at,COALESCE(accepted_by,'') FROM dossier_artifacts a `+where+` ORDER BY id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDossierArtifacts(rows)
}

func scanCapabilityStateHistory(rows *sql.Rows) ([]CapabilityStateRecord, error) {
	values := []CapabilityStateRecord{}
	for rows.Next() {
		var value CapabilityStateRecord
		var source sql.NullInt64
		if err := rows.Scan(&value.ID, &value.CapabilityItemID, &source, &value.State, &value.AcceptedBy, &value.AcceptedAt, &value.CreatedAt); err != nil {
			return nil, err
		}
		if source.Valid {
			value.SourceChangeID = &source.Int64
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

func scanDossierArtifacts(rows *sql.Rows) ([]DossierArtifact, error) {
	values := []DossierArtifact{}
	for rows.Next() {
		var value DossierArtifact
		var supersedes sql.NullInt64
		var acceptedAt sql.NullString
		if err := rows.Scan(&value.ID, &value.ItemID, &value.Kind, &value.AuthorRole, &value.Status, &value.Summary, &value.Evidence, &supersedes, &value.CreatedAt, &acceptedAt, &value.AcceptedBy); err != nil {
			return nil, err
		}
		if supersedes.Valid {
			value.SupersedesID = &supersedes.Int64
		}
		if acceptedAt.Valid {
			value.AcceptedAt = &acceptedAt.String
		}
		values = append(values, value)
	}
	return values, rows.Err()
}
