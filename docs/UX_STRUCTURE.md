# TECHAI CODE — UX 구조 문서

> 사용자 관점에서 본 전체 화면·입력·출력·모드·상태 흐름 가이드
> 작성일: 2026-04-16
> 대상: 신규 개발자, 기여자, UX 개선 작업자

---

## 목차

1. [전체 화면 구성](#1-전체-화면-구성-screen-layout)
2. [입력 흐름](#2-입력-흐름-input-flow)
3. [출력 흐름](#3-출력-흐름-output-flow)
4. [모드별 동작](#4-모드별-동작-mode-behavior)
5. [키보드 조작](#5-키보드-조작-keyboard-controls)
6. [상태바 (HUD)](#6-상태바-hud)
7. [오버레이](#7-오버레이-overlays)
8. [도구 실행 사이클](#8-도구-실행-사이클-tool-execution-cycle)
9. [데이터 흐름](#9-데이터-흐름-data-flow)
10. [에러 및 복구](#10-에러-및-복구-error--recovery)

---

## 1. 전체 화면 구성 (Screen Layout)

TECHAI CODE는 Bubble Tea v2 기반 전체화면 TUI(Terminal User Interface)로 동작한다.
터미널을 전체 점유하며 세 개의 고정 영역으로 나뉜다.

### 1.1 레이아웃 다이어그램

```
┌─────────────────────────────────────────────────────────────────────────┐  ← line 0
│  ████████╗███████╗ ██████╗██╗  ██╗ █████╗ ██╗                          │
│  ╚══██╔══╝██╔════╝██╔════╝██║  ██║██╔══██╗██║                          │
│     ██║   █████╗  ██║     ███████║███████║██║                          │
│     ██║   ██╔══╝  ██║     ██╔══██║██╔══██║██║                          │
│     ██║   ███████╗╚██████╗██║  ██║██║  ██║██║                          │  LOGO AREA
│     ╚═╝   ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝                          │  (super.go)
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━                         │
│     ██████╗  ██████╗  ██████╗  ███████╗                                 │
│    ██╔════╝ ██╔═══██╗ ██╔══██╗ ██╔════╝                                 │
│    ██║      ██║   ██║ ██║  ██║ █████╗                                   │
│    ██║      ██║   ██║ ██║  ██║ ██╔══╝                                   │
│    ╚██████╗ ╚██████╔╝ ██████╔╝ ███████╗                                 │
│     ╚═════╝  ╚═════╝  ╚═════╝  ╚══════╝                                 │
│   v0.3.x                                                                │
├─────────────────────────────────────────────────────────────────────────┤
│ ╭─────────────────────────────────────────────╮                         │  MODE INFO BOX
│ │ Super — gpt-oss-120b                         │                         │  (super.go)
│ │ All-purpose. Code, analysis, conversation   │                         │
│ ╰─────────────────────────────────────────────╯                         │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ▌ 사용자 메시지 (파란 배경 블록)                                         │  CHAT AREA
│                                                                         │  (viewport)
│  ▎ AI 응답 (마크다운 렌더링, 왼쪽 녹색 바)                                │
│                                                                         │
│  >> reading file.go                    ← tool preview (tool 호출 시)    │
│  << file_read: 결과 요약                ← tool result                   │
│                                                                         │
│  시스템 메시지 (녹색 텍스트)                                              │
│                                                                         │
├─────────────────────────────────────────────────────────────────────────┤
│ [paste hint — 6줄 이상 붙여넣기 시 표시]                                  │  PASTE HINT
├─────────────────────────────────────────────────────────────────────────┤
│ ❯ 사용자 입력창 (textarea)                                                │  INPUT BOX
│   (1~3줄 가변 높이, Shift+Enter로 줄바꿈)                                 │
├─────────────────────────────────────────────────────────────────────────┤
│  Super  gpt-oss-120b  ./프로젝트  ⎇ main  Tool:ON(15)  2048tok  ctx:3%  │  STATUS BAR (HUD)
│                                            Ctrl+K Palette  Esc Menu     │  (chat.go)
└─────────────────────────────────────────────────────────────────────────┘  ← line H
```

### 1.2 영역별 설명

| 영역 | 소스 파일 | 설명 |
|------|----------|------|
| **Logo Area** | `internal/ui/super.go` | TECHAI CODE ASCII 로고 + 버전 번호. 앱 시작 시 대화 위에 표시. 메시지가 쌓이면 viewport 위로 밀려남 |
| **Mode Info Box** | `internal/ui/super.go` | 현재 모드(Super/Deep Agent/Plan) + 모델명을 Rounded Border 박스로 표시. 색상이 모드마다 다름 |
| **Chat Area** | `internal/ui/chat.go` | 스크롤 가능한 대화 영역. Bubble Tea `viewport` 컴포넌트. 글자 너비 기준 자동 줄바꿈 |
| **Paste Hint** | `internal/app/app.go` | 6줄 이상 붙여넣기 시 표시되는 한 줄 힌트. 다음 Enter 시 사라짐 |
| **Input Box** | `charm.land/bubbles/v2/textarea` | 사용자 입력 영역. 1~3줄 가변 높이. CharLimit 4096 |
| **Status Bar (HUD)** | `internal/ui/chat.go` | 하단 고정 한 줄. 모드·모델·경로·git·도구·토큰·비용·컨텍스트 표시 |

### 1.3 초기 화면 vs 대화 중 화면

**초기 화면** (메시지 없음):
```
[Logo 전체 표시]
[Mode Info Box]
[빈 chat area]
[입력창]
[상태바]
```

**대화 진행 중** (메시지 쌓임):
```
[Logo는 viewport 위로 스크롤됨 — 보이지 않음]
[Mode Info Box도 위로 밀림]
[chat area: 메시지들]
[스트리밍 중이면: 파란색 "Thinking..." or 스트리밍 텍스트]
[입력창]
[상태바]
```

---

## 2. 입력 흐름 (Input Flow)

### 2.1 기본 메시지 전송 흐름

```
사용자 키 입력
      │
      ▼
[textarea 업데이트]
      │
   Enter 키?
   ├─ No  → 계속 입력 대기
   └─ Yes
        │
    스트리밍 중?
    ├─ Yes  → pendingQueue에 추가 (큐잉)
    │          → 화면: "메시지가 대기 중..." 표시
    └─ No
         │
      슬래시 명령어? (첫 글자 '/')
      ├─ Yes → [슬래시 명령어 흐름] (§2.3)
      └─ No
           │
        입력 비어있음?
        ├─ Yes → 무시
        └─ No
             │
          [입력을 msgs에 추가 — RoleUser]
          [history에 추가 — openai.ChatCompletionMessage]
          [textarea 초기화]
          [pasteHint 초기화]
          [historyIdx = -1 리셋]
          [현재 입력을 inputHistory[0]에 추가]
                │
                ▼
         [AI 스트리밍 시작] (§3)
```

### 2.2 Queue 시스템 (스트리밍 중 입력)

스트리밍이 진행되는 동안 사용자가 Enter를 누르면 메시지가 즉시 전송되지 않고 `pendingQueue`에 저장된다.

```
스트리밍 완료
      │
      ▼
pendingQueue 비어있음?
├─ Yes → 대기 상태로 전환
└─ No
      │
  queue[0] 꺼내기
      │
  AI 스트리밍 즉시 시작 (연쇄 처리)
      │
  (반복)
```

**동작 원칙**: 사용자는 AI가 응답하는 중에도 자유롭게 다음 질문을 입력·전송할 수 있다. 순서가 보장된다.

### 2.3 슬래시 명령어 흐름

```
입력이 '/'로 시작
      │
      ▼
명령어 파싱 (공백 기준 분리)
      │
내장 명령어 테이블 조회
      │
  매칭됨?
  ├─ Yes → 해당 핸들러 실행 (동기 or goroutine)
  │         결과는 slashResultMsg로 UI에 반환
  └─ No
        │
    커스텀 명령어 조회 (.tgc/commands/*.md, ~/.tgc/commands/*.md)
        │
    매칭됨?
    ├─ Yes → 파일 내용을 $ARGUMENTS 치환 후 AI에 전송
    └─ No  → "unknown command" 시스템 메시지 표시
```

**주요 내장 명령어 카테고리**:

| 카테고리 | 명령어 | 처리 방식 |
|---------|--------|----------|
| 세션 | `/new`, `/sessions`, `/session <id>`, `/compact`, `/clear` | 동기 처리 |
| 파일 | `/undo`, `/undo <N>`, `/undo list`, `/diff` | `/diff`는 goroutine (non-blocking) |
| 복사/내보내기 | `/copy`, `/copy <N>`, `/export` | 동기 처리 |
| AI 모드 | `/auto`, `/multi on/off/review/consensus/scan/auto` | 상태 변경 |
| 메모리 | `/remember`, `/forget` | `.tgc/memories.json` 읽기/쓰기 |
| 진단 | `/diagnostics`, `/git`, `/mcp`, `/companion` | goroutine |
| 시스템 | `/init`, `/init deep`, `/setup`, `/version`, `/help`, `/exit` | 혼합 |

### 2.4 붙여넣기 동작 (Paste Behavior)

**5줄 이하 붙여넣기** (Ctrl+V / Cmd+V):
```
클립보드 내용
      │
      ▼
줄 수 계산
      │
   ≤ 5줄?
   └─ Yes → textarea에 직접 삽입 (그대로 표시)
```

**6줄 이상 붙여넣기**:
```
클립보드 내용
      │
      ▼
줄 수 계산 → 6줄 이상
      │
pasteHint 설정:
"[Pasted X lines — Press Enter to send, or Ctrl+U to clear]"
      │
      ▼
입력창 위 힌트 영역에 표시
      │
Enter → 전송 + pasteHint 초기화
Ctrl+U → 입력창 초기화 + pasteHint 초기화
```

**근거**: 대량 붙여넣기 시 사용자가 의도치 않게 전송하는 것을 방지한다.

### 2.5 입력 히스토리 탐색

```
↑ 화살표
      │
   historyIdx == -1?
   └─ Yes → historyDraft에 현재 입력 저장
            historyIdx = 0
   └─ No  → historyIdx++
      │
   inputHistory[historyIdx] 로드 → textarea에 표시

↓ 화살표
      │
   historyIdx == 0?
   └─ Yes → historyDraft 복원
            historyIdx = -1
   └─ No  → historyIdx--
      │
   inputHistory[historyIdx] 로드 → textarea에 표시
```

최대 100개 이전 입력 저장. Enter 전송 시 historyIdx는 -1로 리셋된다.

---

## 3. 출력 흐름 (Output Flow)

### 3.1 AI 응답 스트리밍 흐름

```
AI 스트리밍 시작
      │
      ▼
goroutine: llm.StreamChat()
      │
      ├─ chunk 도착 → streamChunkMsg → Update()
      │    │
      │    ▼
      │  streamBuf += chunk.content
      │  viewport 업데이트 (스트리밍 텍스트 표시)
      │  상태바: "Thinking..." → 경과 시간 표시 (lastElapsed)
      │
      ├─ tool_calls 도착 → toolResultMsg 준비
      │    │
      │    └─ [도구 실행 흐름] (§8 참조)
      │
      └─ done == true
           │
           ▼
        streamBuf → msgs에 추가 (RoleAssistant)
        history에 추가
        streaming = false
        viewport.GotoBottom() → 자동 스크롤
        lastElapsed 기록 (상태바에 표시)
        pendingQueue 확인 → 다음 메시지 처리 (§2.2)
```

### 3.2 스트리밍 상태 표시

| 상태 | 화면 표시 | 색상 |
|------|----------|------|
| 스트리밍 시작 | `Connecting...` | 파란색 (ColorPrimary) |
| 토큰 수신 중 | 실시간 텍스트 + 경과 시간 | 파란색 bold |
| tool 호출 중 | `>> tool_name: arg...` | Accent 색상 |
| 완료 | 전체 응답 렌더링 + 경과시간 | 정상 |

### 3.3 메시지 렌더링 방식

**사용자 메시지 (RoleUser)**:
```
  ▌ 메시지 내용 (파란 배경 블록)
  ▌ 두 번째 줄...
```
- 좌측 `▌` 바 + 파란 배경(`#0D1520`) 블록
- `wrapText()`로 터미널 너비에 맞게 줄바꿈
- 배경 컬러가 끝까지 채워짐 (공백 패딩)

**AI 응답 (RoleAssistant)**:
```
  ▎ # 제목 (glamour 렌더링)
  ▎ 본문 텍스트...
  ▎ ```go
  ▎ // 코드 블록
  ▎ ```
```
- 좌측 `▎` 바 (녹색, 배경 없음)
- `glamour` 라이브러리로 Markdown 렌더링 (커스텀 다크 테마)
- 20줄 초과 시 상단에 `[N lines]` 카운트 표시
- `[ASK_USER]` 태그는 자동으로 제거되어 표시되지 않음

**시스템 메시지 (RoleSystem)**:
```
  시스템 알림 (녹색 텍스트)
```
- `SystemMsg` 스타일 (녹색, bold 없음)
- 줄바꿈 후 빈 줄 추가

**도구 메시지 (RoleTool)**:
```
  >> file_read: main.go (preview)
  << file_read: 245 lines read
```
- Accent 색상 (청록)
- `>>` = tool preview (호출 전), `<<` = tool result (실행 후)

### 3.4 마크다운 렌더링 커스텀 테마

`customDarkStyle`(chat.go)가 glamour 기본 dark 테마를 오버라이드한다:

| 요소 | 표시 |
|------|------|
| H1 | 노란색 bold, 청보라 배경 블록 |
| H2 | `▌ ` 접두사 |
| H3 | `▌▌ ` 접두사 |
| 코드 블록 | 어두운 배경 + 구문 강조 |
| 인용 | `│ ` 들여쓰기 |
| 목록 | 2-space 들여쓰기 |

---

## 4. 모드별 동작 (Mode Behavior)

Tab 키로 순환 전환: Super → Deep Agent → Plan → (다시 Super)

### 4.1 Super 모드 (기본값)

```
모드 색상: 파란색
모델: openai/gpt-oss-120b (128K context)
도구: 전체 14개 + MCP
지식 예산: 8,000 tokens (6 섹션)
컴팩션 트리거: 80% / 90%
보존 메시지: 15개
```

**동작 원칙**:
- 질문 의도를 자동 감지 (코드 작성 / 분석 / 대화 / 검색)
- 모호한 요청 → `clarifyFirstDirective`에 따라 먼저 질문
- 도구를 자유롭게 선택·연쇄 호출
- 최대 20 tool iteration per response (toolIter 카운터)

**사용자 관점**: 무엇이든 물어볼 수 있는 만능 모드. 기본으로 두고 쓰면 됨.

### 4.2 Deep Agent 모드

```
모드 색상: 보라색
모델: openai/gpt-oss-120b (동일)
도구: 전체 14개 + MCP
최대 자율 반복: 100 iterations
종료 마커: [TASK_COMPLETE]
```

**동작 원칙**:
- 사용자 입력 없이 AI가 자율적으로 작업 반복
- 매 iteration: AI 응답 → tool 실행 → 결과 → AI 계속
- `ASK_USER` 호출은 최소화 (프롬프트로 억제)
- `[TASK_COMPLETE]` 또는 100 iteration 도달 시 자동 종료

```
사용자: "이 레포지토리 전체를 리팩토링해줘"
      │
      ▼
Deep Agent 시작
      │
  iteration 1: 코드 탐색 (list_files, file_read)
  iteration 2: 계획 수립
  iteration 3~N: 파일 수정 (file_edit, file_write)
  ...
  iteration N: [TASK_COMPLETE] 마커 출력
      │
      ▼
자율 모드 종료 → 사용자에게 제어권 반환
```

**주의**: 대규모 작업에 적합. 중단하려면 Esc → 메뉴 → Cancel 또는 Ctrl+C.

### 4.3 Plan 모드

```
모드 색상: 노란색/주황색
모델: openai/gpt-oss-120b (동일)
도구: 전체 14개 포함 (⚠️ 핫픽스 필요 — 현재 write 도구 포함)
```

**설계 의도 동작** (hanimo 기준):
```
사용자 요청
      │
      ▼
AI: 단계별 실행 계획 작성
      │
      ▼
사용자 검토 + 승인 ("yes" / "approve")
      │
      ▼
AI: 계획에 따라 실행
```

**실제 현재 동작**: 계획 수립과 실행을 모두 허용 (read-only 제한 미적용).
`⚠️ Plan 모드 write 도구 차단은 미구현 상태 — MASTER-OVERVIEW 기준 IW-1/T1-2 핫픽스 필요`

**사용자 관점**: 복잡한 아키텍처 결정이 필요한 작업 전, 먼저 계획을 짜고 싶을 때 사용.

### 4.4 모드 전환 시 동작

```
Tab 키
  │
  ▼
activeTab = (activeTab + 1) % 3
  │
  ▼
Mode Info Box 업데이트 (색상 + 텍스트 변경)
  │
  ▼
history/msgs는 유지됨 (모드 전환해도 대화 이어짐)
  │
  ▼
다음 AI 요청부터 새 모드의 시스템 프롬프트 적용
```

**핵심**: 대화 내용(history)은 공유된다. 모드는 다음 요청의 시스템 프롬프트만 바꾼다.

---

## 5. 키보드 조작 (Keyboard Controls)

### 5.1 전체 키바인딩 표

| 키 | 동작 | 컨텍스트 |
|----|------|----------|
| `Enter` | 메시지 전송 | 입력 중 |
| `Shift+Enter` | 줄바꿈 (멀티라인 입력) | 입력 중 |
| `Ctrl+J` | 줄바꿈 (Windows 대체) | 입력 중 (Windows CMD/PowerShell) |
| `Tab` | 모드 전환 (Super→Deep Agent→Plan) | 항상 |
| `↑` / `↓` | 입력 히스토리 탐색 (최근 100개) | 입력 중 (히스토리 모드) |
| `Ctrl+K` | 커맨드 팔레트 열기 | 항상 |
| `Esc` | 메뉴 열기 / 스트리밍 취소 / 오버레이 닫기 | 컨텍스트별 |
| `Ctrl+U` | 입력창 전체 지우기 | 입력 중 |
| `Ctrl+B` | 마우스 모드 토글 (스크롤 ↔ 텍스트 선택) | 항상 |
| `Ctrl+L` | 대화 전체 지우기 (Ctrl+C + 재시작 없이) | 항상 |
| `Ctrl+C` | 앱 종료 | 항상 |
| `Ctrl+V` / `Cmd+V` | 붙여넣기 | 입력 중 |
| `Alt+↑` / `Alt+↓` | 3줄 스크롤 | 항상 |
| `PgUp` / `PgDown` | 페이지 스크롤 | 항상 |

### 5.2 오버레이별 추가 키

**커맨드 팔레트 (Ctrl+K 활성화 시)**:

| 키 | 동작 |
|----|------|
| `↑` / `↓` | 목록 탐색 |
| `Enter` | 선택된 명령 실행 |
| `Esc` | 팔레트 닫기 |
| 일반 문자 입력 | Fuzzy 검색 필터링 |

**메뉴 (Esc 활성화 시)**:

| 키 | 동작 |
|----|------|
| `↑` / `↓` | 항목 탐색 |
| `Enter` | 선택된 항목 실행 |
| `Esc` | 메뉴 닫기 |

**세션 피커 (메뉴 → Sessions 또는 /sessions)**:

| 키 | 동작 |
|----|------|
| `↑` / `↓` | 세션 탐색 (최근 20개) |
| `Enter` | 선택된 세션 복원 |
| `Esc` | 세션 피커 닫기 |

### 5.3 Esc 키의 컨텍스트별 동작

```
Esc 키
  │
  ├─ 팔레트 열려있음? → 팔레트 닫기
  ├─ 세션 피커 열려있음? → 세션 피커 닫기
  ├─ 메뉴 열려있음? → 메뉴 닫기
  ├─ 스트리밍 중? → 스트리밍 취소 (streamCancel())
  └─ 그 외 → 메뉴 열기
```

### 5.4 OS별 주의사항

**Windows (CMD / PowerShell)**:
- `Shift+Enter`가 동작하지 않을 수 있음 → `Ctrl+J` 사용
- Windows Terminal 사용 권장 (색상·마크다운·마우스 스크롤 지원)
- Windows Terminal에서 Shift+Enter 활성화:
  ```json
  { "keys": "shift+enter", "command": { "action": "sendInput", "input": "\u000A" } }
  ```

**마우스 모드 (기본값: OFF)**:
- `Ctrl+B`로 토글
- **ON 상태**: 마우스 휠 스크롤 가능, 텍스트 드래그 선택 불가
- **OFF 상태**: 텍스트 드래그 + 복사 가능, 스크롤은 키보드(`Alt+↑/↓`, `PgUp/PgDn`)만
- 사용 팁: `Ctrl+B` → 텍스트 선택·복사 → `Ctrl+B`로 스크롤 모드 복귀

---

## 6. 상태바 (HUD)

상태바는 화면 최하단 1줄을 차지하는 고정 영역이다 (`chat.go: RenderStatusBar()`).
배경색 `#0F172A` (딥 네이비).

### 6.1 레이아웃

```
┌────────────────────────────────────────────────────────────────────────────────┐
│  Super  gpt-oss-120b  ./myproject  ⎇ main*  [DEBUG]  Tool:ON(15)  Multi:ON    │
│  2048tok  $0.001  ctx:12%  2.3s           Ctrl+K Palette  Esc Menu  Tab Switch │
└────────────────────────────────────────────────────────────────────────────────┘
  ←─────────────────── LEFT ──────────────────────────────────────────────────────→
                                                       ←── RIGHT (힌트) ─────────→
```

### 6.2 각 요소 설명

| 요소 | 예시 | 색상 | 의미 |
|------|------|------|------|
| **Mode** | `Super` | 모드별 (파랑/보라/노랑) | 현재 활성 모드 |
| **Model** | `gpt-oss-120b` | 회색 (Subtle) | provider prefix 제거 후 표시 (`openai/` 등 제거) |
| **CWD** | `./myproject` | 회색 | 현재 작업 디렉토리 (basename만) |
| **Git Branch** | `⎇ main` | 초록색 | 현재 git 브랜치 |
| **Git Dirty** | `⎇ main*` | 노란색 + bold | 워킹 트리에 변경사항 있음 |
| **Debug** | `[DEBUG]` | 빨간색 | `--debug` 플래그 활성 시 표시 |
| **Tool Count** | `Tool:ON(15)` | 초록색 | 활성 도구 수 (14 built-in + MCP) |
| **Tool OFF** | `Tool:OFF` | 빨간색 | 도구 비활성 (현재 미사용 상태) |
| **Multi** | `Multi:ON` | 보라색 | 멀티 에이전트 활성 시 |
| **Tokens** | `2048tok` | 회색 | 현재 세션 누적 토큰 |
| **Cost** | `$0.001` | 회색 | 추정 비용 (gpt-oss-120b 기준 $0.30/1M) |
| **Context %** | `ctx:12%` | 회색→노랑→빨강 | 컨텍스트 윈도우 사용률 |
| **Elapsed** | `2.3s` | 회색 | 마지막 AI 응답 소요 시간 |
| **Right Hints** | `Ctrl+K Palette  Esc Menu  Tab Switch  Ctrl+C` | 회색 | 키 힌트 (항상 오른쪽 정렬) |

### 6.3 Context % 색상 임계값

```
ctx < 70%  → Subtle (회색)      — 정상
ctx 70~89% → #FBBF24 (노란색)  — 경고: /compact 권장
ctx ≥ 90%  → #F87171 (빨간색)  — 위험: 자동 컴팩션 트리거
```

### 6.4 토큰이 0일 때

토큰·비용·ctx%는 0이면 표시 안 함. 첫 AI 응답 후 점진적으로 나타난다.

---

## 7. 오버레이 (Overlays)

세 가지 오버레이가 존재한다. 동시에 하나만 표시된다.

### 7.1 커맨드 팔레트 (Ctrl+K)

```
┌─────────────────────────────────────────────┐
│ > /comp                     ← 사용자 입력    │
├─────────────────────────────────────────────┤
│ ▶ /compact  — Compress conversation history │
│   /companion — Open browser dashboard       │
│   /commands  — List loaded custom commands  │
└─────────────────────────────────────────────┘
```

**동작 방식**:
1. `Ctrl+K` → `showPalette = true`
2. 문자 입력 → `paletteQuery` 업데이트 → fuzzy 필터링
3. Fuzzy 매칭: 입력 문자가 명령어에 순서대로 포함되면 매칭 (예: `cmp` → `/compact`)
4. `↑/↓` → `paletteSelected` 이동
5. `Enter` → 선택된 명령어를 textarea에 삽입 후 팔레트 닫기
6. `Esc` → 팔레트 닫기

**포함 항목**: 모든 내장 슬래시 명령어 + 로드된 커스텀 명령어 (`/commands`로 확인 가능)

소스: `internal/ui/palette.go`

### 7.2 메뉴 (Esc)

스트리밍 중이 아닐 때 Esc를 누르면 활성화되는 빠른 실행 메뉴.

```
┌─────────────────────────┐
│  TECHAI CODE             │
├─────────────────────────┤
│ ▶ New Session           │
│   Browse Sessions       │
│   Compact History       │
│   Toggle Multi-Agent    │
│   Cancel Streaming      │
│   Help                  │
│   Quit                  │
└─────────────────────────┘
```

**동작 방식**:
1. `Esc` → `showMenu = true`
2. `↑/↓` → `menuSelected` 이동
3. `Enter` → 선택된 항목 실행 + 메뉴 닫기
4. `Esc` → 메뉴 닫기

소스: `internal/ui/menu.go`

### 7.3 세션 피커 (메뉴 → Browse Sessions 또는 `/sessions`)

```
┌──────────────────────────────────────────────┐
│  Sessions                                     │
├──────────────────────────────────────────────┤
│ ▶ [#42] 2026-04-16 14:23 — 리팩토링 작업      │
│   [#41] 2026-04-15 10:01 — API 연동 디버깅    │
│   [#40] 2026-04-14 22:17 — DB 스키마 설계     │
│   ...                                         │
└──────────────────────────────────────────────┘
```

**동작 방식**:
1. 최근 20개 세션 목록 표시 (SQLite `~/.tgc/sessions.db`)
2. `↑/↓` → 세션 탐색
3. `Enter` → 선택된 세션 복원 (전체 대화 history + 모드 + 설정)
4. `Esc` → 피커 닫기

세션 첫 번째 user 메시지가 타이틀이 된다.

소스: `internal/session/store.go`, `internal/app/app.go`

---

## 8. 도구 실행 사이클 (Tool Execution Cycle)

### 8.1 단일 에이전트 도구 흐름

```
사용자 질문
      │
      ▼
AI 스트리밍 시작
      │
AI: tool_calls 결정
      │
      ▼
┌─────────────────────────────────────────────────┐
│  Tool Preview 표시                               │
│  >> tool_name: 인자 미리보기                      │
│  (Accent 색상)                                   │
└─────────────────────────────────────────────────┘
      │
      ▼
goroutine: tools.Execute(toolCall)
      │
      ▼
┌─────────────────────────────────────────────────┐
│  Tool Result 표시                                │
│  << tool_name: 결과 요약                          │
│  (Accent 색상)                                   │
└─────────────────────────────────────────────────┘
      │
      ▼
결과를 history에 추가 (role=tool)
      │
      ▼
AI에 결과 전달 → 다음 스트리밍 (AI가 계속 판단)
      │
      ▼
toolIter++
      │
   toolIter >= 20?
   ├─ Yes → "Tool loop limit reached" 시스템 메시지 + 종료
   └─ No  → 계속 (AI가 더 tool 호출할 수 있음)
```

### 8.2 도구별 safety 동작

| 도구 | Preview 표시 | 실행 전 확인 | 특이사항 |
|------|------------|------------|---------|
| `file_read` | `>> reading file.go` | 없음 | 최대 50KB |
| `file_write` | `>> writing file.go` | 없음 | 자동 snapshot 생성 |
| `file_edit` | `>> editing file.go` | 없음 | 4단계 fuzzy 매칭 + 자동 snapshot + diff 미리보기 |
| `shell_exec` | `>> shell: git status` | 위험 명령 경고 | `rm -rf /`, `sudo` 등 차단 |
| `hashline_edit` | `>> hashline: file.go` | 없음 | 해시 불일치 시 실패 |
| `glob_search` | `>> glob: **/*.go` | 없음 | .gitignore 존중, max 2000 |
| `grep_search` | `>> grep: pattern` | 없음 | ripgrep 우선, max 300 |
| `knowledge_search` | `>> knowledge: query` | 없음 | 3단계 파이프라인 |

### 8.3 멀티 에이전트 흐름

`/multi on` 또는 `/multi auto` 활성화 시:

```
사용자 입력
      │
      ▼
┌─────────────────────────────────────────┐
│  Auto-Detect (multiAuto == true)        │
│  1. 키워드 매칭 (0ms)                   │
│     "리뷰", "검토" → Review             │
│     "비교", "확인" → Consensus          │
│     "스캔", "검색" → Scan              │
│  2. 불명확 → LLM Gate (~5s)            │
│     LLM이 전략 결정                      │
│  입력 >300자 paste → 자동 skip          │
└─────────────────────────────────────────┘
      │
      ▼
전략별 분기:

[Review 전략]                [Consensus 전략]         [Scan 전략]
Agent1 (GPT-OSS-120B)       Agent1 + Agent2          파일 분할
  전체 14개 도구              동일 프롬프트 병렬          Agent1: dir/a/
  코드 생성                   Agent1 ──┐               Agent2: dir/b/
      │                      Agent2 ──┤               병렬 탐색
      ▼                      결과 비교  │                    │
Agent2 (Qwen3-30B)                   │                    │
  read-only 도구                      │                    │
  코드 리뷰                            │                    │
      │                              │                    │
      └──────────────────────────────┘────────────────────┘
                          │
                          ▼
                   LLM Synthesis
                   (Agent2 = "no issues"이면 skip)
                   (Agent2 에러이면 skip)
                          │
                          ▼
                    최종 응답 표시
```

**멀티 에이전트 모델 차이**:

| 항목 | Agent1 (Super) | Agent2 (Dev) |
|------|--------------|--------------|
| 모델 | GPT-OSS-120B | Qwen3-Coder-30B |
| Context | 128K tokens | 32K tokens |
| 도구 | 전체 14개 | read-only (검색·읽기만) |
| 지식 예산 | 8,000 tokens | 2,000 tokens |
| 컴팩션 | 80%/90% | 60%/75% |

소스: `internal/multi/orchestrator.go`, `internal/multi/strategy.go`

### 8.4 도구 루프 제한 (20 iterations)

```
toolIter 카운터 (매 tool 실행마다 +1)
      │
   toolIter >= 20?
   └─ Yes
        │
        ▼
시스템 메시지: "Tool loop limit reached (20). Stopping to prevent infinite loops."
        │
스트리밍 강제 종료
        │
사용자에게 제어권 반환
```

**목적**: AI가 같은 도구를 무한 반복 호출하는 doom-loop 방지.

---

## 9. 데이터 흐름 (Data Flow)

### 9.1 시스템 프롬프트 구성

매 AI 요청마다 아래 레이어가 합산되어 system prompt를 구성한다:

```
┌─────────────────────────────────────────────────────────────────┐
│  Layer 1: Base Prompt (internal/llm/prompt.go)                  │
│  모드별 기본 지시사항:                                            │
│  - Super: 만능 에이전트, 도구 활용, clarify-first                │
│  - Deep Agent: 자율 반복, ASK_USER 최소화, [TASK_COMPLETE]       │
│  - Plan: 계획 먼저, 승인 후 실행                                  │
├─────────────────────────────────────────────────────────────────┤
│  Layer 2: Project Context (.techai.md)                          │
│  /init 또는 /init deep으로 생성된 프로젝트 가이드                 │
│  앱 시작 시 읽어서 projectCtx에 저장                             │
├─────────────────────────────────────────────────────────────────┤
│  Layer 3: Environment Context (llm/environment.go)              │
│  40+ 도구 감지 결과: node, python, go, git, docker 등            │
│  프로젝트 타입 자동 감지 (Go, Node.js, Python, Java 등)          │
├─────────────────────────────────────────────────────────────────┤
│  Layer 4: Memory (tools/memory.go)                              │
│  /remember로 저장한 사실들                                        │
│  프로젝트 로컬: .tgc/memories.json                               │
│  글로벌: ~/.tgc/memories.json                                    │
├─────────────────────────────────────────────────────────────────┤
│  Layer 5: Knowledge Injection (knowledge/injector.go)           │
│  사용자 질문 → 3단계 검색 → 관련 문서 주입                         │
│  매 질문마다 동적으로 교체됨 (최대 8,000 tokens)                   │
└─────────────────────────────────────────────────────────────────┘
```

### 9.2 지식 주입 파이프라인 (3단계)

```
사용자 질문
      │
      ▼
Level 1: 키워드 추출 + 유의어 확장 (~0ms)
  "스프링 인증" → spring-core.md + spring-security.md
  충분한 결과? ──Yes──→ 주입
      │ No
      ▼
Level 2: BM25 본문 검색 (<1ms, 메모리 내)
  전체 81개 내장 문서 대상 TF-IDF 스코어링
  결과 있음? ──Yes──→ 주입
      │ No
      ▼
Level 3: LLM 판단 매칭 (~200ms, API 1회)
  "아래 문서 중 관련된 것을 골라주세요" → LLM 응답
      │
      ▼
선택된 문서 전문 → system prompt에 주입
```

**내장 문서**: 81개 (바이너리에 컴파일됨 — `knowledge/docs/`)
**사용자 문서**: `.tgc/knowledge/*.md` (다음 실행 시 자동 인덱싱)

### 9.3 세션 영속화

```
앱 시작
  │
  ▼
~/.tgc/sessions.db (SQLite) 열기
  │
세션 ID 생성 또는 복원
  │
대화 중: 매 메시지마다 SQLite에 저장
  │
앱 재시작 후: /sessions → 피커 → 세션 복원
```

**저장 내용**: 메시지 전체 (role, content, timestamp) + 모드 + 모델 + 메타데이터

**세션 타이틀**: 첫 번째 user 메시지 (자동), 이후 변경 없음 (`titleSet = true`)

### 9.4 메모리 시스템 흐름

```
/remember "Redis는 캐시 레이어에 사용"
      │
      ▼
.tgc/memories.json에 저장 {id, text, hitCount, createdAt}
      │
다음 질문 시
      │
      ▼
모든 memories → system prompt 상단에 주입
"## Project Memories\n- Redis는 캐시 레이어에 사용"
      │
      ▼
AI가 메모리 컨텍스트를 인지하고 답변
```

**글로벌 메모리**: `/remember -g "텍스트"` → `~/.tgc/memories.json` (모든 프로젝트에서 공유)
**우선순위**: 프로젝트 로컬 > 글로벌

### 9.5 파일 스냅샷 시스템

```
AI: file_write 또는 file_edit 호출
      │
      ▼
tools/snapshot.go: 수정 전 파일을 자동 백업
~/.tgc/snapshots/{timestamp}_{filename}
      │
      ▼
파일 수정 실행
      │
오류 시 또는 사용자 /undo 호출 시
      │
      ▼
가장 최근 snapshot 복원
```

---

## 10. 에러 및 복구 (Error & Recovery)

### 10.1 네트워크 타임아웃

```
AI 스트리밍 중 네트워크 끊김 / 타임아웃
      │
      ▼
streamChunkMsg{err: error} 수신
      │
      ▼
시스템 메시지 표시:
"[Error] Connection timeout: <error message>"
      │
streaming = false → 사용자에게 제어권 반환
      │
      ▼
사용자: 다시 메시지를 보내면 재시작
(history에 이전 partial response 없음 — streamBuf 버려짐)
```

**주의**: 스트리밍 중 에러가 나면 partial AI 응답은 저장되지 않는다.

### 10.2 도구 실패

```
tools.Execute() → error
      │
      ▼
RoleTool 메시지로 에러 표시:
"<< tool_name: ERROR — <error message>"
      │
      ▼
에러를 history에 추가 (tool role)
      │
      ▼
AI에 에러 결과 전달 → AI가 에러를 인식하고 다른 방법 시도
      │
toolIter 카운트는 증가 (loop limit 진행)
```

**file_edit fuzzy 폴백**: 정확 매칭 실패 시 4단계 순서로 시도:
1. ExactMatch (완전 일치)
2. LineTrimmed (앞뒤 공백 제거 후 비교)
3. IndentFlex (들여쓰기 무시)
4. Levenshtein 85% (유사도 85% 이상이면 매칭)

모두 실패 시: `"file_edit: no matching block found"` 에러 → AI가 인식 후 재시도 또는 다른 전략 사용

### 10.3 컨텍스트 오버플로 → 자동 컴팩션

```
매 응답 후: ContextPercent(tokens, contextWindow) 계산
      │
   ctx >= 90%?
   └─ Yes
        │
        ▼
자동 컴팩션 실행 (3단계):
  1. Snip: 오래된 tool 출력 결과 제거 (가장 큰 절약)
  2. Truncate: 긴 메시지 앞부분 자름 (중간 크기 절약)
  3. LLM Summary: 전체를 요약문 하나로 교체 (최후 수단)
        │
        ▼
시스템 메시지: "Conversation compacted (X → Y tokens)"
        │
        ▼
계속 대화 가능
```

**수동 실행**: `/compact` 명령 또는 메뉴에서 실행 가능.
**컴팩션 타겟**: 컨텍스트 윈도우의 50% 이하로 줄이기.

### 10.4 파일 수정 실패 → /undo 복구

```
AI: file_edit 실패 또는 잘못된 수정
      │
사용자: /undo 입력
      │
      ▼
~/.tgc/snapshots/ 에서 가장 최근 snapshot 찾기
      │
      ▼
원본 파일 복원
      │
      ▼
시스템 메시지: "Restored: <filename> (from snapshot <timestamp>)"
```

**확장 사용**:
- `/undo 3` → 최근 3개 수정 되돌리기
- `/undo list` → 최근 20개 snapshot 타임스탬프 목록 표시

### 10.5 Plan 모드 보안 이슈 (현재 미해결)

```
현재 상태:
  Plan 모드 Tools 목록에 file_write, file_edit 포함
  → AI가 계획 단계에서도 실제 파일 수정 가능

설계 의도:
  Plan 모드 = read-only
  (file_read, list_files, grep_search, glob_search, git_status 만 허용)
  계획 수립 후 사용자 승인 → Super 모드로 전환 후 실행

수정 방향 (hanimo MASTER-OVERVIEW IW-1/T1-2):
  internal/llm/models.go에서 Plan Tools를 ReadOnlyTools()로 교체
```

### 10.6 Companion 대시보드 연결 실패

```
/companion 또는 앱 시작 시 companion 서버 시작
      │
localhost:8787 바인딩 실패
      │
      ▼
시스템 메시지: "Companion server failed to start: <error>"
      │
앱은 정상 동작 계속 (companion은 선택 기능)
```

---

## 부록 A — 파일 구조 매핑

| 파일 | UX 역할 |
|------|---------|
| `internal/app/app.go` | Model struct, Update(), View() — 전체 TUI 상태 머신 |
| `internal/ui/chat.go` | RenderMessages(), RenderStatusBar() — 메시지·HUD 렌더링 |
| `internal/ui/super.go` | RenderLogo(), ModeWelcome(), ModeInfoBox() — 로고·모드 박스 |
| `internal/ui/palette.go` | 커맨드 팔레트 UI |
| `internal/ui/menu.go` | 메뉴 오버레이 UI |
| `internal/ui/styles.go` | 색상·스타일 상수 (ColorPrimary, ColorAccent 등) |
| `internal/ui/tabbar.go` | 탭 정의 (Tabs[0/1/2]) |
| `internal/llm/prompt.go` | 모드별 시스템 프롬프트 |
| `internal/llm/compaction.go` | 3단계 컴팩션 로직 |
| `internal/llm/environment.go` | 환경 프로브 (40+ 도구 감지) |
| `internal/tools/file.go` | file_read/write/edit (fuzzy 4단계) |
| `internal/tools/snapshot.go` | 자동 스냅샷 + /undo |
| `internal/tools/shell.go` | shell_exec + 위험 명령 차단 |
| `internal/knowledge/injector.go` | 3단계 지식 검색 파이프라인 |
| `internal/multi/orchestrator.go` | 멀티 에이전트 실행기 |
| `internal/session/store.go` | SQLite 세션 영속화 |
| `internal/agents/auto.go` | Deep Agent 자율 반복 로직 |

---

## 부록 B — 용어 사전

| 용어 | 설명 |
|------|------|
| **HUD** | Heads-Up Display. 하단 상태바 전체를 지칭 |
| **streamBuf** | 스트리밍 중 실시간으로 쌓이는 텍스트 버퍼. 완료 시 msgs에 추가 |
| **pendingQueue** | 스트리밍 중 사용자가 보낸 메시지 대기열 |
| **toolIter** | 단일 AI 응답 내 tool 호출 횟수 카운터. 20 초과 시 강제 종료 |
| **projectCtx** | `.techai.md` + 환경 정보 + 사용자 문서 TOC를 합산한 문자열 |
| **clarifyFirstDirective** | 모호한 요청에 먼저 질문하도록 유도하는 프롬프트 지시사항 |
| **ASK_USER** | AI가 사용자 확인이 필요할 때 출력하는 구조화된 태그. UI에서 제거되어 표시 안 됨 |
| **TASK_COMPLETE** | Deep Agent 모드에서 AI가 작업 완료를 알리는 마커 |
| **Compaction** | 컨텍스트 초과 시 대화 히스토리를 압축하는 3단계 과정 |
| **Fuzzy Edit** | file_edit에서 정확한 블록을 못 찾을 때 유사도 기반으로 매칭하는 4단계 전략 |
| **Knowledge RAG** | 81개 내장 문서 + 사용자 문서를 3단계 검색으로 system prompt에 주입하는 시스템 |
| **Snapshot** | file_write/file_edit 전 자동 생성되는 백업 파일. /undo로 복원 |
| **MCP** | Model Context Protocol. 외부 도구(Jira, Wiki 등)를 AI에 연결하는 표준 프로토콜 |
| **Onprem** | 사내망 전용 빌드. 엔드포인트·모델 고정, 설정 분리 (`~/.tgc-onprem/`) |
| **Companion** | `localhost:8787` 브라우저 대시보드. SSE로 실시간 AI 활동 표시 |

---

_Last updated: 2026-04-16 · TECHAI CODE 기준_
