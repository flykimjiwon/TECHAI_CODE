package llm

import "strings"

// CodingTier represents a model's coding ability level.
type CodingTier int

const (
	CodingStrong   CodingTier = iota // Production-grade code generation
	CodingModerate                   // Good for standard tasks
	CodingWeak                       // Basic completions only
	CodingNone                       // No coding ability
)

// RoleType determines what tools are available by default.
type RoleType int

const (
	RoleAgent     RoleType = iota // Full tools (read + write + exec)
	RoleAssistant                 // Read-only tools
	RoleChat                      // No tools
)

// ModelCapability describes a model's known capabilities.
type ModelCapability struct {
	ContextWindow int
	CodingTier    CodingTier
	DefaultRole   RoleType
	SupportsTools bool
}

// knownModels maps model IDs to their capabilities. Keyed by the suffix
// after the provider prefix so lookups work for both "openai/gpt-oss-120b"
// and bare "gpt-oss-120b" ids.
var knownModels = map[string]ModelCapability{
	// Novita.ai production models
	"qwen3-coder-30b-a3b-instruct": {ContextWindow: 262144, CodingTier: CodingStrong, DefaultRole: RoleAgent, SupportsTools: true},
	"gpt-oss-120b":                 {ContextWindow: 128000, CodingTier: CodingStrong, DefaultRole: RoleAgent, SupportsTools: true},
	"qwen3-coder-30b":              {ContextWindow: 262144, CodingTier: CodingStrong, DefaultRole: RoleAgent, SupportsTools: true},

	// Common reference models, kept so the registry can grow without churn.
	"gpt-4o":            {ContextWindow: 128000, CodingTier: CodingStrong, DefaultRole: RoleAgent, SupportsTools: true},
	"gpt-4o-mini":       {ContextWindow: 128000, CodingTier: CodingModerate, DefaultRole: RoleAgent, SupportsTools: true},
	"claude-sonnet-4":   {ContextWindow: 200000, CodingTier: CodingStrong, DefaultRole: RoleAgent, SupportsTools: true},
	"claude-haiku-4":    {ContextWindow: 200000, CodingTier: CodingModerate, DefaultRole: RoleAgent, SupportsTools: true},
	"deepseek-chat":     {ContextWindow: 128000, CodingTier: CodingStrong, DefaultRole: RoleAgent, SupportsTools: true},
	"deepseek-reasoner": {ContextWindow: 128000, CodingTier: CodingStrong, DefaultRole: RoleAgent, SupportsTools: true},
	"llama3.1:8b":       {ContextWindow: 128000, CodingTier: CodingWeak, DefaultRole: RoleChat, SupportsTools: false},
	"llama3.1:70b":      {ContextWindow: 128000, CodingTier: CodingModerate, DefaultRole: RoleAgent, SupportsTools: true},
}

// defaultCapability is returned for unknown models. It assumes tool support
// so the UX does not silently degrade when the user switches to a new model
// that just hasn't been added to the registry yet.
var defaultCapability = ModelCapability{
	ContextWindow: 32768,
	CodingTier:    CodingModerate,
	DefaultRole:   RoleAssistant,
	SupportsTools: true,
}

// GetCapability returns the capability for a model, falling back to the
// default when the id is unknown. It understands the "provider/model" form
// ("openai/gpt-oss-120b") and looks up the suffix.
func GetCapability(model string) ModelCapability {
	if cap, ok := knownModels[model]; ok {
		return cap
	}
	if idx := strings.LastIndex(model, "/"); idx >= 0 {
		if cap, ok := knownModels[model[idx+1:]]; ok {
			return cap
		}
	}
	return defaultCapability
}

// AutoAssignRole returns the recommended role for a given model, based on
// its registered capability.
func AutoAssignRole(model string) RoleType {
	return GetCapability(model).DefaultRole
}

// RoleLabel returns a human-readable label for a role type.
func RoleLabel(r RoleType) string {
	switch r {
	case RoleAgent:
		return "Agent"
	case RoleAssistant:
		return "Assistant"
	case RoleChat:
		return "Chat"
	default:
		return "Unknown"
	}
}

// CodingTierLabel returns a human-readable label for a coding tier.
func CodingTierLabel(t CodingTier) string {
	switch t {
	case CodingStrong:
		return "Strong"
	case CodingModerate:
		return "Moderate"
	case CodingWeak:
		return "Weak"
	case CodingNone:
		return "None"
	default:
		return "Unknown"
	}
}
