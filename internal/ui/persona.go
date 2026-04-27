package ui

import (
	"encoding/json"
	"strings"
)

// Persona controls the "flavour text" of the TUI — thinking verbs,
// cold-start messages, stall warnings.
type Persona struct {
	Name          string
	Emoji         string
	ThinkingVerbs []string
	ColdStart     string
	Stall5s       string
	Stall15s      string
}

// TechaiDefaultPersona is 택가이코드's default flavour text.
var TechaiDefaultPersona = Persona{
	Name:  "techai-default",
	Emoji: "",
	ThinkingVerbs: []string{
		"생각하는 중",
		"분석 중",
		"작업 중",
		"정리 중",
		"검토 중",
	},
	ColdStart: "연결 중",
	Stall5s:   "응답 준비 중... 조금 오래 걸리네요",
	Stall15s:  "응답 지연 — 네트워크 또는 모델 부하일 수 있습니다",
}

// ActivePersona is the persona currently mounted on this binary.
var ActivePersona = TechaiDefaultPersona

// ThinkingVerbFor returns the rotating thinking verb.
func ThinkingVerbFor(elapsedSeconds float64) string {
	verbs := ActivePersona.ThinkingVerbs
	if len(verbs) == 0 {
		return ActivePersona.ColdStart
	}
	idx := int(elapsedSeconds/2) % len(verbs)
	if idx < 0 {
		idx = 0
	}
	return verbs[idx]
}

// FormatToolDiff produces a visual diff block for file_edit / file_write calls.
func FormatToolDiff(name, argsJSON, output string) string {
	if name != "file_edit" && name != "file_write" {
		return ""
	}
	var args struct {
		Path      string `json:"path"`
		OldString string `json:"old_string"`
		NewString string `json:"new_string"`
		Content   string `json:"content"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return ""
	}

	header := "<< " + name
	if args.Path != "" {
		header += " " + args.Path
	}

	if name == "file_write" {
		lineCount := 1 + strings.Count(args.Content, "\n")
		return header + "\n  | +" + itoa(lineCount) + "줄 새로 작성"
	}

	const maxSide = 8
	oldLines := splitLinesKeep(args.OldString)
	newLines := splitLinesKeep(args.NewString)
	trimmedOld := false
	trimmedNew := false
	if len(oldLines) > maxSide {
		oldLines = oldLines[:maxSide]
		trimmedOld = true
	}
	if len(newLines) > maxSide {
		newLines = newLines[:maxSide]
		trimmedNew = true
	}

	var b []string
	b = append(b, header)
	b = append(b, "  | -"+itoa(len(splitLinesKeep(args.OldString)))+"줄  +"+itoa(len(splitLinesKeep(args.NewString)))+"줄")
	for _, l := range oldLines {
		b = append(b, "  | - "+trimRunes(l, 150))
	}
	if trimmedOld {
		b = append(b, "  | -  ... (더 있음)")
	}
	for _, l := range newLines {
		b = append(b, "  | + "+trimRunes(l, 150))
	}
	if trimmedNew {
		b = append(b, "  | +  ... (더 있음)")
	}
	return joinNL(b)
}

// FormatToolResult renders a tool result with auto-collapsing.
func FormatToolResult(name, output string) string {
	const (
		maxLines  = 15
		headLines = 6
		tailLines = 4
	)
	raw := splitLinesKeep(output)
	total := len(raw)
	if total == 0 {
		return "<< " + name + ": (빈 결과)"
	}

	var b []string
	if total <= maxLines {
		b = append(b, "<< "+name)
		for _, l := range raw {
			b = append(b, "  | "+trimRunes(l, 160))
		}
		return joinNL(b)
	}

	b = append(b, "<< "+name+" ("+formatLineCount(total)+")")
	for _, l := range raw[:headLines] {
		b = append(b, "  | "+trimRunes(l, 160))
	}
	omitted := total - headLines - tailLines
	b = append(b, "  | ... +"+formatLineCount(omitted)+" 생략")
	for _, l := range raw[total-tailLines:] {
		b = append(b, "  | "+trimRunes(l, 160))
	}
	return joinNL(b)
}

func splitLinesKeep(s string) []string {
	lines := []string{}
	cur := ""
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, cur)
			cur = ""
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}

func joinNL(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += "\n"
		}
		out += p
	}
	return out
}

func trimRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "..."
}

func formatLineCount(n int) string {
	return itoa(n) + "줄"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
