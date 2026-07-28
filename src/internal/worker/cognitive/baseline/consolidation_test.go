package baseline

import (
	"strings"
	"testing"

	"plasmod/src/internal/schemas"
	"plasmod/src/internal/storage"
)

func TestMemoryConsolidationKeepsExistingSummaryStable(t *testing.T) {
	store := storage.NewMemoryRuntimeStorage().Objects()
	worker := CreateInMemoryMemoryConsolidationWorker("consolidate-test", store, nil)

	store.PutMemory(schemas.Memory{
		MemoryID:   "mem_1",
		AgentID:    "agent1",
		SessionID:  "sess1",
		Content:    "first memory",
		Level:      0,
		IsActive:   true,
		Version:    1,
	})
	if err := worker.Consolidate("agent1", "sess1"); err != nil {
		t.Fatalf("initial Consolidate: %v", err)
	}
	summaryID := schemas.IDPrefixSummary + "agent1_sess1"
	first, ok := store.GetMemory(summaryID)
	if !ok {
		t.Fatal("summary missing after initial consolidation")
	}

	store.PutMemory(schemas.Memory{
		MemoryID:   "mem_2",
		AgentID:    "agent1",
		SessionID:  "sess1",
		Content:    "second memory should not force a summary rebuild",
		Level:      0,
		IsActive:   true,
		Version:    1,
	})
	if err := worker.Consolidate("agent1", "sess1"); err != nil {
		t.Fatalf("second Consolidate: %v", err)
	}
	second, ok := store.GetMemory(summaryID)
	if !ok {
		t.Fatal("summary missing after second consolidation")
	}
	if second.Content != first.Content {
		t.Fatalf("existing summary was rebuilt: got %q, want %q", second.Content, first.Content)
	}
	if strings.Contains(second.Content, "second memory") {
		t.Fatal("existing summary pulled in newly added level-0 memory")
	}
}
