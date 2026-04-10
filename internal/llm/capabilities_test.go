package llm

import "testing"

func TestGetCapability_KnownModel_GptOss120b(t *testing.T) {
	cap := GetCapability("openai/gpt-oss-120b")
	if cap.ContextWindow != 128000 {
		t.Errorf("ContextWindow = %d, want 128000", cap.ContextWindow)
	}
	if cap.CodingTier != CodingStrong {
		t.Errorf("CodingTier = %v, want CodingStrong", cap.CodingTier)
	}
	if !cap.SupportsTools {
		t.Error("SupportsTools = false, want true")
	}
}

func TestGetCapability_KnownModel_Qwen3Coder(t *testing.T) {
	cap := GetCapability("qwen/qwen3-coder-30b")
	// Qwen3-Coder-30B has a 256K context window per models.go
	if cap.ContextWindow != 262144 {
		t.Errorf("ContextWindow = %d, want 262144", cap.ContextWindow)
	}
	if cap.CodingTier != CodingStrong {
		t.Errorf("CodingTier = %v, want CodingStrong", cap.CodingTier)
	}
}

func TestGetCapability_UnknownModel_ReturnsDefault(t *testing.T) {
	cap := GetCapability("unknown/model-9000")
	if cap.ContextWindow == 0 {
		t.Error("default ContextWindow was 0")
	}
	// Default should still be usable
	if !cap.SupportsTools {
		t.Error("default should assume tool support so UX is not degraded")
	}
}

func TestGetCapability_StripsProviderPrefix(t *testing.T) {
	// Even without exact match, stripping provider prefix should find it.
	cap := GetCapability("custom/gpt-oss-120b")
	if cap.ContextWindow != 128000 {
		t.Errorf("ContextWindow = %d, want 128000 (should strip custom/ prefix)", cap.ContextWindow)
	}
}

func TestCodingTierLabel_AllValues(t *testing.T) {
	cases := map[CodingTier]string{
		CodingStrong:   "Strong",
		CodingModerate: "Moderate",
		CodingWeak:     "Weak",
		CodingNone:     "None",
	}
	for tier, want := range cases {
		if got := CodingTierLabel(tier); got != want {
			t.Errorf("CodingTierLabel(%v) = %q, want %q", tier, got, want)
		}
	}
}

func TestRoleLabel_AllValues(t *testing.T) {
	cases := map[RoleType]string{
		RoleAgent:     "Agent",
		RoleAssistant: "Assistant",
		RoleChat:      "Chat",
	}
	for r, want := range cases {
		if got := RoleLabel(r); got != want {
			t.Errorf("RoleLabel(%v) = %q, want %q", r, got, want)
		}
	}
}

func TestAutoAssignRole_GptOss_IsAgent(t *testing.T) {
	// gpt-oss-120b supports tools → should default to Agent role
	if got := AutoAssignRole("openai/gpt-oss-120b"); got != RoleAgent {
		t.Errorf("AutoAssignRole(gpt-oss-120b) = %v, want RoleAgent", got)
	}
}
