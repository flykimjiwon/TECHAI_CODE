package agents

import (
	"strings"
	"testing"
)

func TestMaxAutoIterations_Is20(t *testing.T) {
	// RED: the loop limit for /auto is 20 per the roadmap.
	if MaxAutoIterations != 20 {
		t.Errorf("MaxAutoIterations = %d, want 20", MaxAutoIterations)
	}
}

func TestCheckAutoMarkers_Complete(t *testing.T) {
	complete, pause := CheckAutoMarkers("All done! [AUTO_COMPLETE]")
	if !complete {
		t.Error("complete = false, want true")
	}
	if pause {
		t.Error("pause = true, want false")
	}
}

func TestCheckAutoMarkers_Pause(t *testing.T) {
	complete, pause := CheckAutoMarkers("Need credentials [AUTO_PAUSE]")
	if complete {
		t.Error("complete = true, want false")
	}
	if !pause {
		t.Error("pause = false, want true")
	}
}

func TestCheckAutoMarkers_Neither(t *testing.T) {
	complete, pause := CheckAutoMarkers("Still working on it...")
	if complete || pause {
		t.Errorf("complete=%v pause=%v, want both false", complete, pause)
	}
}

func TestAutoPromptSuffix_ContainsKeyInstructions(t *testing.T) {
	// The suffix must tell the model about the two markers and the mode.
	required := []string{
		"AUTONOMOUS",
		"[AUTO_COMPLETE]",
		"[AUTO_PAUSE]",
	}
	for _, want := range required {
		if !strings.Contains(AutoPromptSuffix, want) {
			t.Errorf("AutoPromptSuffix missing %q", want)
		}
	}
}
