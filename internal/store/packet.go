package store

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// PlanPacket is the compact, portable plan contract used by agent adapters.
type PlanPacket struct {
	RoadmapItem        PacketRoadmapItem `json:"roadmap_item"`
	Goal               string            `json:"goal"`
	Scope              PacketScope       `json:"scope"`
	Decisions          []string          `json:"decisions"`
	AcceptanceCriteria []PacketCriterion `json:"acceptance_criteria"`
	Constraints        []string          `json:"constraints"`
	Tasks              []PacketTask      `json:"tasks"`
	Risks              []string          `json:"risks"`
	OpenQuestions      []string          `json:"open_questions"`
}
type PacketRoadmapItem struct {
	ID          *int64 `json:"id,omitempty"`
	EpicID      *int64 `json:"epic_id,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Horizon     string `json:"horizon"`
	Rationale   string `json:"rationale"`
}
type PacketScope struct {
	Included []string `json:"included"`
	Excluded []string `json:"excluded"`
}
type PacketCriterion struct {
	ID          string `json:"id,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Required    *bool  `json:"required,omitempty"`
}

// UnmarshalJSON accepts the original compact string form as well as the
// revision-stable object form used by new plan adapters.
func (criterion *PacketCriterion) UnmarshalJSON(data []byte) error {
	var title string
	if err := json.Unmarshal(data, &title); err == nil {
		criterion.Title = title
		return nil
	}
	type plain PacketCriterion
	var value plain
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	*criterion = PacketCriterion(value)
	return nil
}

func (criterion PacketCriterion) key(index int) string {
	if strings.TrimSpace(criterion.ID) != "" {
		return strings.TrimSpace(criterion.ID)
	}
	return fmt.Sprintf("C%d", index+1)
}

func (criterion PacketCriterion) RequiredOrDefault() bool {
	return criterion.Required == nil || *criterion.Required
}

type PacketTask struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Verification []string `json:"verification"`
}
type RecordedPlan struct {
	Plan     Plan                  `json:"plan"`
	Criteria []AcceptanceCriterion `json:"criteria"`
	Tasks    []Task                `json:"tasks"`
}

func (packet PlanPacket) validate() error {
	if strings.TrimSpace(packet.Goal) == "" {
		return fmt.Errorf("plan goal is required")
	}
	if len(packet.Tasks) == 0 {
		return fmt.Errorf("plan requires at least one task")
	}
	if len(packet.AcceptanceCriteria) == 0 {
		return fmt.Errorf("plan requires at least one acceptance criterion")
	}
	keys := map[string]bool{}
	for index, criterion := range packet.AcceptanceCriteria {
		if strings.TrimSpace(criterion.Title) == "" {
			return fmt.Errorf("acceptance criterion %d title is required", index+1)
		}
		key := criterion.key(index)
		if keys[key] {
			return fmt.Errorf("acceptance criterion key %q is duplicated", key)
		}
		keys[key] = true
	}
	for _, task := range packet.Tasks {
		if strings.TrimSpace(task.Title) == "" {
			return fmt.Errorf("plan task title is required")
		}
		if meaningfulValues(task.Verification) == 0 {
			return fmt.Errorf("task %q requires at least one verification requirement", task.Title)
		}
	}
	if packet.RoadmapItem.ID == nil {
		if strings.TrimSpace(packet.RoadmapItem.Title) == "" || strings.TrimSpace(packet.RoadmapItem.Category) == "" || strings.TrimSpace(packet.RoadmapItem.Horizon) == "" {
			return fmt.Errorf("new roadmap items require title, category, and horizon")
		}
	}
	return nil
}

// ValidatePlanPacket checks the structural requirements for an approved plan
// packet without persisting it.
func ValidatePlanPacket(packet PlanPacket) error {
	return packet.validate()
}

// RecordApprovedPlan atomically persists an already user-approved plan packet.
func RecordApprovedPlan(database *sql.DB, packet PlanPacket, approvalNote string) (RecordedPlan, error) {
	return recordApprovedPlan(database, packet, approvalNote, false)
}

// RecordApprovedPlanAllowOpenQuestions atomically persists an explicitly
// approved plan packet that still contains open questions. The questions are
// retained in the immutable plan content; callers should make the override
// conspicuous to the approving user.
func RecordApprovedPlanAllowOpenQuestions(database *sql.DB, packet PlanPacket, approvalNote string) (RecordedPlan, error) {
	return recordApprovedPlan(database, packet, approvalNote, true)
}

func recordApprovedPlan(database *sql.DB, packet PlanPacket, approvalNote string, allowOpenQuestions bool) (RecordedPlan, error) {
	if err := packet.validate(); err != nil {
		return RecordedPlan{}, err
	}
	if len(packet.OpenQuestions) > 0 && !allowOpenQuestions {
		return RecordedPlan{}, fmt.Errorf("plan has unresolved open questions; record with the explicit approved override")
	}
	content, err := json.Marshal(packet)
	if err != nil {
		return RecordedPlan{}, err
	}
	transaction, err := database.Begin()
	if err != nil {
		return RecordedPlan{}, err
	}
	defer transaction.Rollback()
	if packet.RoadmapItem.EpicID != nil {
		var exists bool
		if err := transaction.QueryRow(`SELECT EXISTS(SELECT 1 FROM epics WHERE id=?)`, *packet.RoadmapItem.EpicID).Scan(&exists); err != nil {
			return RecordedPlan{}, err
		}
		if !exists {
			return RecordedPlan{}, fmt.Errorf("epic %d not found", *packet.RoadmapItem.EpicID)
		}
	}
	timestamp := now()
	var itemID int64
	if packet.RoadmapItem.ID != nil {
		itemID = *packet.RoadmapItem.ID
		var status string
		if err := transaction.QueryRow(`SELECT status FROM roadmap_items WHERE id=?`, itemID).Scan(&status); err == sql.ErrNoRows {
			return RecordedPlan{}, fmt.Errorf("roadmap item %d not found", itemID)
		} else if err != nil {
			return RecordedPlan{}, err
		} else if status != "Ready" && status != "In Progress" {
			return RecordedPlan{}, fmt.Errorf("roadmap item %d is not ready or in progress", itemID)
		}
	} else {
		result, err := transaction.Exec(`INSERT INTO roadmap_items(title,description,category_id,horizon,status,rationale,created_at,updated_at,epic_id) SELECT ?,?,id,?,'Ready',?,?,?,? FROM categories WHERE name=?`, packet.RoadmapItem.Title, packet.RoadmapItem.Description, packet.RoadmapItem.Horizon, packet.RoadmapItem.Rationale, timestamp, timestamp, packet.RoadmapItem.EpicID, packet.RoadmapItem.Category)
		if err != nil {
			return RecordedPlan{}, err
		}
		itemID, err = result.LastInsertId()
		if err != nil || itemID == 0 {
			return RecordedPlan{}, fmt.Errorf("category %q does not exist", packet.RoadmapItem.Category)
		}
	}
	var revision int
	if err := transaction.QueryRow(`SELECT COALESCE(MAX(revision),0)+1 FROM plan_revisions WHERE roadmap_item_id=?`, itemID).Scan(&revision); err != nil {
		return RecordedPlan{}, err
	}
	result, err := transaction.Exec(`INSERT INTO plan_revisions(roadmap_item_id,revision,content,status,active,created_at,approved_at,approval_note) VALUES(?,?,?,'Approved',1,?,?,?)`, itemID, revision, string(content), timestamp, timestamp, approvalNote)
	if err != nil {
		return RecordedPlan{}, err
	}
	planID, err := result.LastInsertId()
	if err != nil {
		return RecordedPlan{}, err
	}
	if _, err := transaction.Exec(`UPDATE plan_revisions SET active=0 WHERE roadmap_item_id=? AND id<>?`, itemID, planID); err != nil {
		return RecordedPlan{}, err
	}
	criteria := make([]AcceptanceCriterion, 0, len(packet.AcceptanceCriteria))
	for index, packetCriterion := range packet.AcceptanceCriteria {
		key := packetCriterion.key(index)
		status, method, evidence, verifiedAt, waivedBy, waiverReason := carriedCriterion(transaction, itemID, key, packetCriterion)
		result, err := transaction.Exec(`INSERT INTO acceptance_criteria(roadmap_item_id,plan_revision_id,criterion_key,title,description,required,status,verification_method,evidence,verified_at,waived_by,waiver_reason) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, itemID, planID, key, packetCriterion.Title, packetCriterion.Description, boolInt(packetCriterion.RequiredOrDefault()), status, method, evidence, verifiedAt, waivedBy, waiverReason)
		if err != nil {
			return RecordedPlan{}, err
		}
		criterionID, err := result.LastInsertId()
		if err != nil {
			return RecordedPlan{}, err
		}
		criteria = append(criteria, AcceptanceCriterion{ID: criterionID, ItemID: itemID, PlanID: planID, Key: key, Title: packetCriterion.Title, Description: packetCriterion.Description, Required: packetCriterion.RequiredOrDefault(), Status: pointerValue(status), VerificationMethod: pointerValue(method), Evidence: pointerValue(evidence), VerifiedAt: verifiedAt, WaivedBy: pointerValue(waivedBy), WaiverReason: pointerValue(waiverReason)})
	}
	tasks := make([]Task, 0, len(packet.Tasks))
	for _, packetTask := range packet.Tasks {
		verification, err := json.Marshal(packetTask.Verification)
		if err != nil {
			return RecordedPlan{}, err
		}
		result, err := transaction.Exec(`INSERT INTO tasks(plan_revision_id,title,description,verification,created_at) VALUES(?,?,?,?,?)`, planID, packetTask.Title, packetTask.Description, string(verification), timestamp)
		if err != nil {
			return RecordedPlan{}, err
		}
		taskID, err := result.LastInsertId()
		if err != nil {
			return RecordedPlan{}, err
		}
		tasks = append(tasks, Task{ID: taskID, PlanID: planID, Title: packetTask.Title, Description: packetTask.Description, Verification: packetTask.Verification, Status: "To Do", CreatedAt: timestamp})
	}
	if err := transaction.Commit(); err != nil {
		return RecordedPlan{}, err
	}
	approvedAt := timestamp
	return RecordedPlan{Plan: Plan{ID: planID, ItemID: itemID, Revision: revision, Content: string(content), Status: "Approved", Active: true, CreatedAt: timestamp, ApprovedAt: &approvedAt, ApprovalNote: approvalNote}, Criteria: criteria, Tasks: tasks}, nil
}

func meaningfulValues(values []string) int {
	count := 0
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			count++
		}
	}
	return count
}

func carriedCriterion(transaction *sql.Tx, itemID int64, key string, criterion PacketCriterion) (status, method, evidence, verifiedAt, waivedBy, waiverReason *string) {
	var previousTitle, previousDescription string
	var previousRequired, previousStatus string
	var previousMethod, previousEvidence, previousVerifiedAt, previousWaivedBy, previousWaiverReason sql.NullString
	err := transaction.QueryRow(`SELECT title,description,required,status,verification_method,evidence,verified_at,waived_by,waiver_reason FROM acceptance_criteria WHERE plan_revision_id=(SELECT id FROM plan_revisions WHERE roadmap_item_id=? AND active=1 LIMIT 1) AND criterion_key=?`, itemID, key).Scan(&previousTitle, &previousDescription, &previousRequired, &previousStatus, &previousMethod, &previousEvidence, &previousVerifiedAt, &previousWaivedBy, &previousWaiverReason)
	if err != nil || previousTitle != criterion.Title || previousDescription != criterion.Description || (previousRequired == "1") != criterion.RequiredOrDefault() {
		return stringPointer("Pending"), stringPointer(""), stringPointer(""), nil, stringPointer(""), stringPointer("")
	}
	return stringPointer(previousStatus), nullStringPointer(previousMethod), nullStringPointer(previousEvidence), nullStringPointer(previousVerifiedAt), nullStringPointer(previousWaivedBy), nullStringPointer(previousWaiverReason)
}

func stringPointer(value string) *string { return &value }
func nullStringPointer(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func pointerValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
