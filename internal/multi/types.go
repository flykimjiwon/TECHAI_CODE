package multi

import (
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// Strategy defines how two agents collaborate.
type Strategy int

const (
	// StrategyReview: Agent1(Super) generates → Agent2(Dev) reviews.
	StrategyReview Strategy = iota
	// StrategyConsensus: Both agents receive the same prompt, results compared.
	StrategyConsensus
	// StrategyScan: Files split across agents for parallel search.
	StrategyScan
)

func (s Strategy) String() string {
	switch s {
	case StrategyReview:
		return "Review"
	case StrategyConsensus:
		return "Consensus"
	case StrategyScan:
		return "Scan"
	default:
		return "Unknown"
	}
}

// AgentID identifies which agent is acting.
type AgentID int

const (
	Agent1Super AgentID = iota // primary — full tool access (write)
	Agent2Dev                  // secondary — read-only tools
)

func (a AgentID) String() string {
	if a == Agent1Super {
		return "Agent1(Super)"
	}
	return "Agent2(Dev)"
}

// AgentProgress is sent through the progress channel to update the UI.
type AgentProgress struct {
	Agent   AgentID
	Status  string        // "streaming", "tool_call", "waiting", "done", "error", "synthesizing"
	Detail  string        // e.g. tool name, error message
	Tokens  int           // accumulated token count
	Elapsed time.Duration // time since agent started
}

// AgentResult holds the final output from one agent.
type AgentResult struct {
	Agent      AgentID
	Content    string
	ToolCalls  int           // how many tool iterations were used
	Tokens     int           // total tokens consumed
	Elapsed    time.Duration // wall-clock duration
	Err        error
	History    []openai.ChatCompletionMessage // conversation history after completion
}

// MergedResult combines outputs from both agents into a single response.
type MergedResult struct {
	Strategy    Strategy
	Agent1      AgentResult
	Agent2      AgentResult
	FinalOutput string // formatted combined output
	TotalTokens int
	Elapsed     time.Duration
}
