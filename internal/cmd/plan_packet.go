package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/rickcern44/cassor/internal/store"
)

// loadPlanPacket reads and validates one strict plan-packet JSON document.
// Keeping this separate from record allows other plan commands to share the
// same decoding and structural validation contract.
func loadPlanPacket(path string) (store.PlanPacket, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return store.PlanPacket{}, fmt.Errorf("read plan packet: %w", err)
	}

	var packet store.PlanPacket
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&packet); err != nil {
		return store.PlanPacket{}, fmt.Errorf("decode plan packet: %w", err)
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		return store.PlanPacket{}, fmt.Errorf("plan packet must contain one JSON value")
	}
	if err := store.ValidatePlanPacket(packet); err != nil {
		return store.PlanPacket{}, fmt.Errorf("validate plan packet: %w", err)
	}
	return packet, nil
}
