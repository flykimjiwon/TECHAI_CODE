package agents

import (
	"regexp"
	"strings"
)

// Marker strings that bracket an interactive ASK_USER request in an LLM response.
const (
	AskUserStartMarker = "[ASK_USER]"
	AskUserEndMarker   = "[/ASK_USER]"
)

// AskQuestionType differentiates the three interactive question styles.
type AskQuestionType int

const (
	AskTypeText AskQuestionType = iota
	AskTypeChoice
	AskTypeConfirm
)

// AskQuestion describes a single ASK_USER request parsed from model output.
type AskQuestion struct {
	Type     AskQuestionType
	Question string
	Options  []string // populated for choice + confirm
}

var askUserRegex = regexp.MustCompile(`(?s)\[ASK_USER\](.*?)\[/ASK_USER\]`)

// ParseAskUser extracts the first ASK_USER block from a response and returns
// the structured question. Returns nil if no marker is present.
func ParseAskUser(response string) *AskQuestion {
	matches := askUserRegex.FindStringSubmatch(response)
	if len(matches) < 2 {
		return nil
	}

	body := strings.TrimSpace(matches[1])
	lines := strings.Split(body, "\n")

	q := &AskQuestion{Type: AskTypeText}
	inOptions := false
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		switch {
		case strings.HasPrefix(strings.ToLower(line), "question:"):
			q.Question = strings.TrimSpace(line[len("question:"):])
			inOptions = false
		case strings.HasPrefix(strings.ToLower(line), "type:"):
			t := strings.ToLower(strings.TrimSpace(line[len("type:"):]))
			switch t {
			case "choice":
				q.Type = AskTypeChoice
			case "confirm":
				q.Type = AskTypeConfirm
			default:
				q.Type = AskTypeText
			}
			inOptions = false
		case strings.HasPrefix(strings.ToLower(line), "options:"):
			inOptions = true
		case strings.HasPrefix(line, "- "):
			opt := strings.TrimSpace(line[2:])
			if opt != "" {
				q.Options = append(q.Options, opt)
			}
		default:
			if inOptions {
				q.Options = append(q.Options, line)
			} else if q.Question == "" {
				q.Question = line
			}
		}
	}

	if q.Type == AskTypeConfirm && len(q.Options) == 0 {
		q.Options = []string{"예", "아니오"}
	}
	if q.Type == AskTypeChoice && len(q.Options) == 0 {
		q.Type = AskTypeText
	}
	return q
}

// StripAskUser removes ASK_USER blocks from the response.
func StripAskUser(response string) string {
	return strings.TrimSpace(askUserRegex.ReplaceAllString(response, ""))
}

// FormatAnswer wraps the user's response for feeding back into the LLM.
func FormatAnswer(q *AskQuestion, answer string) string {
	if q == nil {
		return answer
	}
	switch q.Type {
	case AskTypeChoice:
		return "사용자 선택: " + answer
	case AskTypeConfirm:
		return "사용자 응답: " + answer
	default:
		return "사용자 응답: " + answer
	}
}

// AskUserPromptSuffix is appended to system prompts to teach the LLM how and
// when to use ASK_USER.
const AskUserPromptSuffix = `

## 대화형 질문 (ASK_USER)

실행을 일시 중지하고 사용자에게 질문할 수 있습니다. 요구사항이 모호하거나,
여러 접근법이 유효하거나, 되돌리기 어려운 중요한 결정을 내릴 때만 사용하세요.

형식:

[ASK_USER]
question: 어떤 데이터베이스를 사용할까요?
type: choice
options:
- PostgreSQL
- MySQL
- SQLite
[/ASK_USER]

지원 유형:
- choice   — 옵션 목록, 사용자가 하나 선택
- text     — 자유 텍스트 입력
- confirm  — 예/아니�� 확인

규칙:
- 응답당 ASK_USER 블록은 최대 하나만.
- 직접 답할 수 있는 사소한 질문은 하지 마세요.
- 사용자가 답하면, 해당 답변을 사용하여 작업을 계속하세요.`
