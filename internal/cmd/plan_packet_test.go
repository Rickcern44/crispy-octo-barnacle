package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rickcern44/cassor/internal/store"
)

func TestLoadPlanPacketStrictlyDecodesAndValidates(t *testing.T) {
	path := writePlanPacket(t, `{"roadmap_item":{"title":"Skills","category":"Needed","horizon":"Now"},"goal":"Install skills","acceptance_criteria":[{"title":"It works"}],"tasks":[{"title":"Add skill","verification":["go test ./..."]}]}`)

	packet, err := loadPlanPacket(path)
	if err != nil {
		t.Fatal(err)
	}
	if packet.Goal != "Install skills" || len(packet.Tasks) != 1 {
		t.Fatalf("packet = %#v", packet)
	}
}

func TestLoadPlanPacketRejectsUnknownFieldsAndExtraValues(t *testing.T) {
	unknown := writePlanPacket(t, `{"goal":"Goal","unknown":true}`)
	if _, err := loadPlanPacket(unknown); err == nil || !strings.Contains(err.Error(), "decode plan packet") {
		t.Fatalf("unknown field error = %v", err)
	}
	nestedUnknown := writePlanPacket(t, `{"goal":"Goal","acceptance_criteria":[{"title":"It works","unknown":true}],"tasks":[{"title":"Add skill","verification":["go test ./..."]}]}`)
	if _, err := loadPlanPacket(nestedUnknown); err == nil || !strings.Contains(err.Error(), "decode plan packet") {
		t.Fatalf("nested unknown field error = %v", err)
	}

	valid := `{"roadmap_item":{"title":"Skills","category":"Needed","horizon":"Now"},"goal":"Install skills","acceptance_criteria":[{"title":"It works"}],"tasks":[{"title":"Add skill","verification":["go test ./..."]}]}`
	extra := writePlanPacket(t, valid+" "+valid)
	if _, err := loadPlanPacket(extra); err == nil || !strings.Contains(err.Error(), "one JSON value") {
		t.Fatalf("extra value error = %v", err)
	}
}

func TestLoadPlanPacketRejectsInvalidStructure(t *testing.T) {
	contents, err := json.Marshal(store.PlanPacket{Goal: "Goal"})
	if err != nil {
		t.Fatal(err)
	}
	path := writePlanPacket(t, string(contents))
	if _, err := loadPlanPacket(path); err == nil || !strings.Contains(err.Error(), "validate plan packet") {
		t.Fatalf("structural validation error = %v", err)
	}
}

func writePlanPacket(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "plan.json")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
