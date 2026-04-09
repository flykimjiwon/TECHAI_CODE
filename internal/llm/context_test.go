package llm

import (
	"fmt"
	"testing"
)

func TestGatherSystemContext(t *testing.T) {
	ctx := GatherSystemContext()
	fmt.Println(ctx)
	if ctx == "" {
		t.Fatal("GatherSystemContext returned empty")
	}
}
