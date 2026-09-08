package store

import (
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
	AcceptanceCriteria []string          `json:"acceptance_criteria"`
	Constraints        []string          `json:"constraints"`
	Tasks              []PacketTask      `json:"tasks"`
	Risks              []string          `json:"risks"`
	OpenQuestions      []string          `json:"open_questions"`
}
type PacketRoadmapItem struct {
	ID          *int64 `json:"id,omitempty"`
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
type PacketTask struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Verification []string `json:"verification"`
}
type RecordedPlan struct {
	Plan  Plan   `json:"plan"`
	Tasks []Task `json:"tasks"`
}

func (packet PlanPacket) validate() error {
	if strings.TrimSpace(packet.Goal) == "" {
		return fmt.Errorf("plan goal is required")
	}
	if len(packet.OpenQuestions) > 0 {
		return fmt.Errorf("plan has unresolved open questions")
	}
	if len(packet.Tasks) == 0 {
		return fmt.Errorf("plan requires at least one task")
	}
	for _, task := range packet.Tasks {
		if strings.TrimSpace(task.Title) == "" {
			return fmt.Errorf("plan task title is required")
		}
	}
	if packet.RoadmapItem.ID == nil {
		if strings.TrimSpace(packet.RoadmapItem.Title) == "" || strings.TrimSpace(packet.RoadmapItem.Category) == "" || strings.TrimSpace(packet.RoadmapItem.Horizon) == "" {
			return fmt.Errorf("new roadmap items require title, category, and horizon")
		}
	}
	return nil
}

// RecordApprovedPlan atomically persists an already user-approved plan packet.
func RecordApprovedPlan(database *sql.DB, packet PlanPacket, approvalNote string) (RecordedPlan, error) {
	if err := packet.validate(); err != nil {
		return RecordedPlan{}, err
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
		result, err := transaction.Exec(`INSERT INTO roadmap_items(title,description,category_id,horizon,status,rationale,created_at,updated_at) SELECT ?,?,id,?,'Ready',?,?,? FROM categories WHERE name=?`, packet.RoadmapItem.Title, packet.RoadmapItem.Description, packet.RoadmapItem.Horizon, packet.RoadmapItem.Rationale, timestamp, timestamp, packet.RoadmapItem.Category)
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
	result, err := transaction.Exec(`INSERT INTO plan_revisions(roadmap_item_id,revision,content,status,created_at,approved_at,approval_note) VALUES(?,?,?,'Approved',?,?,?)`, itemID, revision, string(content), timestamp, timestamp, approvalNote)
	if err != nil {
		return RecordedPlan{}, err
	}
	planID, err := result.LastInsertId()
	if err != nil {
		return RecordedPlan{}, err
	}
	tasks := make([]Task, 0, len(packet.Tasks))
	for _, packetTask := range packet.Tasks {
		result, err := transaction.Exec(`INSERT INTO tasks(plan_revision_id,title,description,created_at) VALUES(?,?,?,?)`, planID, packetTask.Title, packetTask.Description, timestamp)
		if err != nil {
			return RecordedPlan{}, err
		}
		taskID, err := result.LastInsertId()
		if err != nil {
			return RecordedPlan{}, err
		}
		tasks = append(tasks, Task{ID: taskID, PlanID: planID, Title: packetTask.Title, Description: packetTask.Description, Status: "To Do", CreatedAt: timestamp})
	}
	if err := transaction.Commit(); err != nil {
		return RecordedPlan{}, err
	}
	approvedAt := timestamp
	return RecordedPlan{Plan: Plan{ID: planID, ItemID: itemID, Revision: revision, Content: string(content), Status: "Approved", CreatedAt: timestamp, ApprovedAt: &approvedAt, ApprovalNote: approvalNote}, Tasks: tasks}, nil
}
