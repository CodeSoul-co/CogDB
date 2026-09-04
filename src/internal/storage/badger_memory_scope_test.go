package storage

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"plasmod/src/internal/schemas"
)

func TestBadgerObjectStoreListMemoriesFilteredByScopeDoesNotDecodeUnrelatedMemories(t *testing.T) {
	db, err := openBadgerInMemory()
	if err != nil {
		t.Fatalf("openBadgerInMemory: %v", err)
	}
	defer db.Close()

	store := newBadgerObjectStore(db)
	store.PutMemory(schemas.Memory{
		MemoryID:  "target-1",
		AgentID:   "agent-target",
		SessionID: "session-target",
		Content:   "target memory",
		IsActive:  true,
	})
	store.PutMemory(schemas.Memory{
		MemoryID:  "target-2",
		AgentID:   "agent-target",
		SessionID: "session-target",
		Content:   "second target memory",
		IsActive:  true,
	})

	largeContent := strings.Repeat("unrelated payload ", 256)
	for i := 0; i < 20000; i++ {
		store.PutMemory(schemas.Memory{
			MemoryID:  fmt.Sprintf("noise-%05d", i),
			AgentID:   "agent-noise",
			SessionID: "session-noise",
			Content:   largeContent,
			IsActive:  true,
		})
	}

	started := time.Now()
	got := store.ListMemories("agent-target", "session-target")
	elapsed := time.Since(started)

	if len(got) != 2 {
		t.Fatalf("ListMemories returned %d memories, want 2", len(got))
	}
	ids := map[string]bool{}
	for _, memory := range got {
		ids[memory.MemoryID] = true
	}
	if !ids["target-1"] || !ids["target-2"] {
		t.Fatalf("ListMemories returned ids %v, want target-1 and target-2", ids)
	}
	if elapsed > 25*time.Millisecond {
		t.Fatalf("filtered ListMemories took %s with unrelated records, want <= 25ms", elapsed)
	}
}
