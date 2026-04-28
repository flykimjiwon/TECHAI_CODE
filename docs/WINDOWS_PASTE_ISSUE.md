# Windows 붙여넣기 (Ctrl+V) 프리징 이슈 — 조사 & 해결 기록

> 작성일: 2026-04-28
> 작성자: 김지원
> 상태: **해결 (자동 감지 + 수동 폴백)**

---

## 증상

Windows 환경에서 특정 텍스트를 Ctrl+V로 붙여넣으면 TUI가 프리징됨.
macOS에서는 동일 텍스트가 100% 정상 동작.

---

## 근본 원인

### 맥 vs Windows 붙여넣기 경로 차이

| | macOS | Windows |
|---|---|---|
| 붙여넣기 방식 | Bracket Paste (ESC[200~) | 개별 문자 전달 |
| 앱이 받는 것 | `tea.PasteMsg` (전체 텍스트 한번에) | `tea.KeyPressMsg` × N번 (글자마다) |
| InsertString 호출 | **1번** | **N번** (100자면 100번) |
| wrap() 계산 | **1번** | **N번** |
| 결과 | 빠름 | **프리징** (특정 텍스트) |

### textarea wrap() 함수의 CJK 혼합 텍스트 버그

- 한글 = 2칸 (full-width), 영문 = 1칸 (half-width)
- `wrap()` 함수가 mixed-width 문자 조합에서 특정 폭 경계에 걸리면 계산이 비정상적으로 느려짐
- 참고: [charmbracelet/bubbletea#831](https://github.com/charmbracelet/bubbletea/issues/831)

### Windows Terminal의 Ctrl+V 동작

- Windows Terminal이 Ctrl+V를 가로채서 앱에 전달하지 않음
- 대신 클립보드 내용을 개별 문자 키 이벤트로 변환하여 전송
- 참고: [charmbracelet/bubbletea#1301](https://github.com/charmbracelet/bubbletea/issues/1301)
- 우클릭 붙여넣기도 동일한 동작

---

## 해결 방법 (v0.9.7)

### 1. 자동 프리징 방지 (기본 동작)

사용자가 Ctrl+V를 누르면:

```
Windows Terminal이 글자를 하나씩 전달
    ↓
1~19번째 글자: textarea에 정상 전달
    ↓
20번째 글자 (5ms 이내 도착): "붙여넣기 감지!"
    ↓
textarea 초기화 → clipboard.ReadAll() → InsertString (한번에!)
    ↓
21번째~ 글자들: 전부 무시 (이미 클립보드에서 읽었으므로)
    ↓
사용자에게는 정상 붙여넣기처럼 보임
```

**감지 기준**: 5ms 내 20자 이상 연속 입력 = 사람이 아닌 붙여넣기
(사람이 5ms에 1글자 = 초당 200자 = 물리적으로 불가능)

### 2. F5 단축키 (수동 폴백)

F5를 누르면 클립보드에서 직접 읽어서 InsertString.
자동 감지가 안 될 경우 수동으로 사용.

### 3. /paste 또는 /v 명령어 (최종 폴백)

`/paste` 또는 `/v` 입력 후 Enter.
가장 확실한 방법이지만 타이핑이 필요.

### 4. MaxWidth 확장 (보조)

textarea의 MaxWidth를 10000으로 설정하여 soft-wrap 발생 자체를 억제.
macOS에서의 wrap 관련 잠재 이슈도 예방.

---

## 시도했다가 실패한 방법들

| 방법 | 실패 이유 |
|------|----------|
| rapid-input 감지 → Enter를 줄바꿈으로 전환 | 일반 타이핑도 오탐, 임계값 조정 불가 |
| Layout 재계산 debounce | textarea.Update() 자체가 프리징 원인 |
| 키 입력 버퍼링 | msg.String()으로 한글/특수키 추출 불완전 |
| Ctrl+V 가로채기 (case "ctrl+v") | Windows Terminal이 먹어서 앱에 안 도달 |
| IME guard (Zero-Width Space) | textarea 내부 동작 충돌로 crash |
| Bubble Tea v2.0.6 업그레이드 | 의존성 꼬여서 crash (ultraviolet, bubbles 불일치) |
| tea.ReadClipboard() (OSC52) | Msg 타입이라 Cmd로 반환 불가, 구조적 불일치 |

---

## 구현 위치

| 기능 | 파일 | 위치 |
|------|------|------|
| 자동 감지 | `internal/app/app.go` | default 키 핸들러 (rapidCount >= 20) |
| F5 단축키 | `internal/app/app.go` | `case "f5"` |
| /paste, /v | `internal/app/app.go` | `handleSlashCommand` |
| MaxWidth | `internal/app/app.go` | `ta.MaxWidth = 10000` |
| pasteHint 표시 | `internal/app/app.go` | View() 함수 |

---

## 참고 자료

- [charmbracelet/bubbletea#831 — Textarea slow on paste](https://github.com/charmbracelet/bubbletea/issues/831)
- [charmbracelet/bubbletea#1301 — Cannot detect Ctrl+V on Windows](https://github.com/charmbracelet/bubbletea/issues/1301)
- [charmbracelet/bubbletea#1453 — KeyPressMsg text empty on Windows](https://github.com/charmbracelet/bubbletea/issues/1453)
- [OpenCode TUI — paste handling with PowerShell clipboard API](https://github.com/anomalyco/opencode)
