package agents

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Plan represents a multi-step task execution plan.
type Plan struct {
	Task      string     `json:"task"`
	Goal      string     `json:"goal"`
	Steps     []PlanStep `json:"steps"`
	Current   int        `json:"current"` // index of the currently executing step
	Status    string     `json:"status"`  // draft, approved, executing, completed, failed, paused
	CreatedAt string     `json:"created_at"`
}

// PlanStep represents a single step in a Plan.
type PlanStep struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"` // pending, in_progress, completed, failed, skipped
	Result      string `json:"result,omitempty"`
}

// Plan mode markers and prompts.
const (
	StepCompleteMarker = "[STEP_COMPLETE]"
	StepFailedMarker   = "[STEP_FAILED]"
	PlanReviseMarker   = "[PLAN_REVISE]"
)

// PlanPromptPrefix is prepended to the system when asking the LLM to create a plan.
const PlanPromptPrefix = `You are in PLANNING MODE. Create a step-by-step execution plan for the following task.

Output ONLY valid JSON matching this schema:
{
  "goal": "One-line summary of what we'll accomplish",
  "steps": [
    {"id": 1, "title": "Short title", "description": "What to do in this step"},
    {"id": 2, "title": "...", "description": "..."}
  ]
}

Guidelines:
- Break the task into 3-15 concrete, actionable steps
- Each step should be independently testable
- Order steps so earlier ones enable later ones
- Include testing/verification steps
- Do NOT execute anything — just plan

Task:
`

// ExecutePromptPrefix is prepended when asking the LLM to execute a single plan step.
// Format args: plan render, current step (1-indexed), total steps, step title, step description.
const ExecutePromptPrefix = `You are in PLAN EXECUTION MODE. Execute the current step of an approved plan.

Current Plan:
%s

Current Step: %d of %d
Step Title: %s
Step Description: %s

Instructions:
- Use tools to accomplish this step
- Focus ONLY on this step, don't skip ahead
- When done, output [STEP_COMPLETE] on its own line
- If blocked, output [STEP_FAILED] followed by reason
- If you need to revise the plan, output [PLAN_REVISE] followed by new steps
`

// ParsePlan parses an LLM response as a Plan. The response may be raw JSON or
// wrapped in Markdown code fences.
func ParsePlan(response string) (*Plan, error) {
	jsonStr := extractJSON(response)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found in response")
	}

	var parsed struct {
		Goal  string     `json:"goal"`
		Steps []PlanStep `json:"steps"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil, fmt.Errorf("parse plan: %w", err)
	}
	if len(parsed.Steps) == 0 {
		return nil, fmt.Errorf("plan has no steps")
	}

	for i := range parsed.Steps {
		if parsed.Steps[i].ID == 0 {
			parsed.Steps[i].ID = i + 1
		}
		parsed.Steps[i].Status = "pending"
	}

	return &Plan{
		Goal:   parsed.Goal,
		Steps:  parsed.Steps,
		Status: "draft",
	}, nil
}

func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
		if idx := strings.LastIndex(s, "```"); idx >= 0 {
			s = s[:idx]
		}
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		if idx := strings.LastIndex(s, "```"); idx >= 0 {
			s = s[:idx]
		}
	}
	s = strings.TrimSpace(s)
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start < 0 || end < 0 || end <= start {
		return ""
	}
	return s[start : end+1]
}

// Render returns a human-readable plan display.
func (p *Plan) Render() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("  계획: %s\n\n", p.Goal))
	for i, step := range p.Steps {
		icon := "[ ]"
		switch step.Status {
		case "completed":
			icon = "[x]"
		case "in_progress":
			icon = "[>]"
		case "failed":
			icon = "[!]"
		case "skipped":
			icon = "[-]"
		}
		marker := "  "
		if i == p.Current && p.Status == "executing" {
			marker = "->"
		}
		b.WriteString(fmt.Sprintf("  %s %s %d. %s\n", marker, icon, step.ID, step.Title))
		if step.Description != "" {
			b.WriteString(fmt.Sprintf("         %s\n", step.Description))
		}
	}
	return b.String()
}

// Progress returns a progress string like "3/10" counting completed steps.
func (p *Plan) Progress() string {
	if p == nil || len(p.Steps) == 0 {
		return ""
	}
	completed := 0
	for _, s := range p.Steps {
		if s.Status == "completed" {
			completed++
		}
	}
	return fmt.Sprintf("%d/%d", completed, len(p.Steps))
}

// CheckPlanMarkers inspects assistant output for plan-execution control markers.
func CheckPlanMarkers(content string) (stepDone bool, stepFailed bool, revise bool) {
	stepDone = strings.Contains(content, StepCompleteMarker)
	stepFailed = strings.Contains(content, StepFailedMarker)
	revise = strings.Contains(content, PlanReviseMarker)
	return
}
