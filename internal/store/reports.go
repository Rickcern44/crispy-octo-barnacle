package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type FeatureReport struct {
	ID                int64    `json:"id"`
	ItemID            int64    `json:"item_id"`
	ExecutionMode     string   `json:"execution_mode"`
	Roles             []string `json:"roles"`
	ElapsedNS         int64    `json:"elapsed_ns"`
	ToolCalls         int      `json:"tool_calls"`
	Verification      string   `json:"verification"`
	InputTokens       *int64   `json:"input_tokens"`
	CachedInputTokens *int64   `json:"cached_input_tokens"`
	OutputTokens      *int64   `json:"output_tokens"`
	ReasoningTokens   *int64   `json:"reasoning_tokens"`
	TotalTokens       *int64   `json:"total_tokens"`
	ComparisonKey     *string  `json:"comparison_key"`
	ContextPackets    *int64   `json:"context_packets"`
	ContextBytes      *int64   `json:"context_bytes"`
	Handoffs          *int64   `json:"handoffs"`
	CreatedAt         string   `json:"created_at"`
}

func AddFeatureReport(database *sql.DB, value FeatureReport) (FeatureReport, error) {
	roles, err := json.Marshal(value.Roles)
	if err != nil {
		return value, err
	}
	createdAt := now()
	result, err := database.Exec(`INSERT INTO feature_reports(roadmap_item_id,execution_mode,roles,elapsed_ns,tool_calls,verification,input_tokens,cached_input_tokens,output_tokens,reasoning_tokens,total_tokens,comparison_key,context_packets,context_bytes,handoffs,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, value.ItemID, value.ExecutionMode, string(roles), value.ElapsedNS, value.ToolCalls, value.Verification, value.InputTokens, value.CachedInputTokens, value.OutputTokens, value.ReasoningTokens, value.TotalTokens, value.ComparisonKey, value.ContextPackets, value.ContextBytes, value.Handoffs, createdAt)
	if err != nil {
		return value, fmt.Errorf("add feature report: %w", err)
	}
	value.ID, err = result.LastInsertId()
	if err != nil {
		return value, err
	}
	value.CreatedAt = createdAt
	return value, nil
}

func ListFeatureReports(database *sql.DB) ([]FeatureReport, error) {
	rows, err := database.Query(`SELECT id,roadmap_item_id,execution_mode,roles,elapsed_ns,tool_calls,verification,input_tokens,cached_input_tokens,output_tokens,reasoning_tokens,total_tokens,comparison_key,context_packets,context_bytes,handoffs,created_at FROM feature_reports ORDER BY roadmap_item_id,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []FeatureReport{}
	for rows.Next() {
		var v FeatureReport
		var roles string
		if err := rows.Scan(&v.ID, &v.ItemID, &v.ExecutionMode, &roles, &v.ElapsedNS, &v.ToolCalls, &v.Verification, &v.InputTokens, &v.CachedInputTokens, &v.OutputTokens, &v.ReasoningTokens, &v.TotalTokens, &v.ComparisonKey, &v.ContextPackets, &v.ContextBytes, &v.Handoffs, &v.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(roles), &v.Roles); err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	return values, rows.Err()
}
