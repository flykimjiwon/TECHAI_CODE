# 택가이코드 IDE 계획서

> 작성일: 2026-04-23
> 현재 버전: v0.9.6
> 상태: 계획 단계

---

## 개요

택가이코드의 Go 엔진(`internal/*`)을 공유하는 **GUI IDE 제품**을 별도로 제작.
기존 TUI(`cmd/tgc`)는 그대로 유지하며, GUI는 독립적인 프론트엔드로 동작.

```
택가이코드 (현재)          택가이코드 IDE (신규)
┌─────────────┐           ┌──────────────────┐
│  cmd/tgc    │           │  cmd/techai-ide   │
│  (Bubble Tea│           │  (Wails v2        │
│   TUI)      │           │   React Frontend) │
└──────┬──────┘           └────────┬─────────┘
       │                          │
       └──────────┬───────────────┘
                  │
         ┌────────┴────────┐
         │  internal/*     │  ← 공유 엔진
         │  llm/           │     (변경 없음)
         │  tools/         │
         │  config/        │
         │  knowledge/     │
         │  session/       │
         │  hooks/         │
         │  companion/     │
         └─────────────────┘
```

---

## 기술 스택

### Backend: Wails v2 + Go

- **Wails v2**: Go 백엔드 + 시스템 WebView (Electron처럼 Chromium 번들 없음)
- 기존 `internal/*` 패키지 직접 import
- PTY (os/exec): 터미널 프로세스 관리
- 단일 바이너리 출력 (~35MB)

### Frontend: React + TypeScript + Vite

| 라이브러리 | 용도 | 크기 (gzip) |
|-----------|------|------------|
| CodeMirror 6 | 코드 에디터 (Go/JS/Python/YAML/Markdown) | ~80KB |
| xterm.js | 터미널 에뮬레이터 | ~150KB |
| react-resizable-panels | 분할 패널 드래그 | ~10KB |
| @radix-ui/react-tabs | 파일 탭 | ~5KB |
| **총 번들** | | **~250KB** |

### 비교

| | Wails | Electron | Tauri |
|---|---|---|---|
| 바이너리 | ~35MB | ~150MB+ | ~25MB |
| Go 엔진 연동 | 직접 import | HTTP subprocess | subprocess |
| WebView | 시스템 내장 | Chromium 번들 | 시스템 내장 |
| 폐쇄망 배포 | 단일 파일 | 무거움 | 단일 파일 |

---

## 프로젝트 구조

### 방식 C: 단일 레포 + cmd/ 분리 (1단계)

```
택가이코드/
├── cmd/
│   ├── tgc/                ← 기존 TUI (그대로)
│   │   └── main.go
│   └── techai-ide/         ← 신규 GUI (Wails)
│       ├── main.go         ← Wails 앱 진입점
│       ├── app.go          ← Go 메서드 (프론트엔드에서 호출)
│       └── frontend/       ← React 앱
│           ├── src/
│           │   ├── App.tsx
│           │   ├── components/
│           │   │   ├── FileTree.tsx
│           │   │   ├── Editor.tsx
│           │   │   ├── Terminal.tsx
│           │   │   ├── Chat.tsx
│           │   │   ├── ToolLog.tsx
│           │   │   └── StatusBar.tsx
│           │   ├── themes/
│           │   │   └── themes.ts     ← 11개 테마 정의
│           │   └── hooks/
│           │       ├── useStream.ts  ← SSE 스트리밍
│           │       └── useTerminal.ts
│           ├── index.html
│           ├── package.json
│           └── vite.config.ts
├── internal/               ← 공유 엔진 (변경 없음)
├── go.mod                  ← wails 의존성 추가
└── Makefile                ← build-ide 타겟 추가
```

### 방식 B: 엔진 분리 (2단계, 필요 시)

```
택가이코드/
├── core/           ← go.mod 1 (엔진)
│   ├── go.mod
│   └── llm/, tools/, config/, knowledge/ ...
├── tui/            ← go.mod 2 (기존 TUI)
│   ├── go.mod (replace ../core)
│   └── cmd/tgc/main.go
└── gui/            ← go.mod 3 (신규 Wails)
    ├── go.mod (replace ../core)
    ├── cmd/techai-ide/
    └── frontend/
```

---

## 화면 레이아웃

```
┌──────────────────────────────────────────────────────────────────┐
│  택가이코드 IDE          ⎇ main*  ●3 modified     🌙 Dark  v0.9.6│
├──────────┬───────────────────────────────┬───────────────────────┤
│ 📁 FILES │  app.go  ×  │ client.go  ×   │  💬 TECHAI Chat       │
│          ├──────────────────────────────┤                       │
│ ▾ cmd/   │  1│ package main             │  ▌ user message       │
│   tgc/   │  2│                          │                       │
│ ▾ internal│  3│ import (                │  ▎ AI response        │
│   llm/   │  4│   "fmt"                  │    with code blocks   │
│   tools/ │  5│   "os"                   │                       │
│   config/│  6│ )                        │  >> tool_call preview │
│          │  7│                          │  << tool result       │
│          │  8│ func main() {            │                       │
│          │  9│   ...                    │  ❯ input box          │
│          ├──────────────────────────────┼───────────────────────┤
│          │ $ TERMINAL                   │  🔧 Tool Log          │
│          │ ~/project $ go test ./...    │  file_read app.go ✓   │
│          │ ok  internal/llm  0.4s       │  grep "TODO" → 3건    │
│          │ ~/project $ _                │  shell go vet ✓       │
└──────────┴──────────────────────────────┴───────────────────────┘
```

### 패널 구성

| 패널 | 위치 | 단축키 | 설명 |
|------|------|--------|------|
| File Tree | 좌 | Ctrl+1 | 프로젝트 파일 탐색 |
| Editor | 중앙 상 | Ctrl+2 | 코드 편집 (CodeMirror 6) |
| Terminal | 중앙 하 | Ctrl+` | 실제 쉘 세션 (xterm.js) |
| Chat | 우 상 | Ctrl+3 | TECHAI AI 채팅 |
| Tool Log | 우 하 | - | 도구 실행 이력 |
| Status Bar | 하단 | - | Git/모델/토큰/경로 |

---

## 기능 목록

### MVP (v1.0)

| # | 기능 | 구현 방식 | 우선순위 |
|---|------|----------|---------|
| 1 | 파일 트리 | Go: os.ReadDir / React: 자체 트리 컴포넌트 | P0 |
| 2 | 코드 에디터 | CodeMirror 6 (Go/JS/Python/YAML/Markdown) | P0 |
| 3 | 파일 탭 | 다중 파일 열기/닫기 | P0 |
| 4 | Chat UI | SSE 스트리밍 + 도구 실행 표시 | P0 |
| 5 | 터미널 | xterm.js + Go PTY | P0 |
| 6 | Git 상태 | 파일트리 M/A 뱃지 + 상태바 branch | P0 |
| 7 | 테마 선택 | CSS 변수 기반 11개 테마 (다크/라이트) | P0 |
| 8 | 마크다운 렌더링 | Chat 내 코드블록 구문강조 | P0 |
| 9 | 전체 검색 | Ctrl+Shift+F → grep_search 연동 | P1 |
| 10 | 파일 자동 갱신 | AI file_write 시 에디터 새로고침 | P1 |
| 11 | 분할 패널 리사이즈 | react-resizable-panels 드래그 | P1 |
| 12 | Git commit/diff | Chat에서 /git 명령 | P1 |

### 향후 (v1.1+)

| 기능 | 설명 |
|------|------|
| LSP 연동 | gopls, typescript-language-server → 에러/경고 마커 |
| 미니맵 | CodeMirror 미니맵 확장 |
| 검색/치환 | 에디터 내 Ctrl+H |
| Diff 뷰 | 변경사항 side-by-side 비교 |
| 멀티 터미널 | 탭으로 여러 터미널 세션 |
| 키보드 단축키 커스텀 | 사용자 정의 단축키 |

---

## Go Backend API (Wails 바인딩)

```go
// cmd/techai-ide/app.go

type App struct {
    cfg     config.Config
    client  *llm.Client
    knStore *knowledge.Store
}

// ── 파일 시스템 ──
func (a *App) ListFiles(path string) ([]FileEntry, error)
func (a *App) ReadFile(path string) (string, error)
func (a *App) WriteFile(path, content string) error
func (a *App) SearchFiles(pattern, path string) ([]SearchResult, error)

// ── AI 채팅 ──
func (a *App) SendMessage(prompt string) // → SSE 이벤트로 스트리밍
func (a *App) GetHistory() []Message
func (a *App) ClearHistory()

// ── 터미널 ──
func (a *App) CreateTerminal() (termID string, error)
func (a *App) WriteTerminal(termID, input string) error
func (a *App) CloseTerminal(termID string) error

// ── Git ──
func (a *App) GitStatus() (gitinfo.Info, error)
func (a *App) GitDiff() (string, error)
func (a *App) GitCommit(message string) error

// ── 설정 ──
func (a *App) GetConfig() config.Config
func (a *App) SetTheme(theme string) error
```

---

## 테마 시스템 (11개)

### 다크 테마 (8개)
1. **Slate** (기본) — 블루-그레이
2. **Cursor** — 오렌지 액센트
3. **Linear** — 퍼플
4. **GitHub Dark** — GitHub 스타일
5. **Dracula** — 클래식 다크
6. **Nord** — 극지 팔레트
7. **Solarized Dark** — 따뜻한 다크
8. **One Dark** — Atom/JetBrains
9. **Monokai Pro** — 노란 액센트

### 라이트 테마 (2개)
10. **Claude** — 테라코타 (크림 배경)
11. **Vercel** — 모노크롬 화이트

### 구현
- CSS 변수 40개 (`--bg-deep`, `--accent`, `--keyword` 등)
- `body.theme-{name}` 클래스 토글
- CodeMirror 테마 연동 (CSS 변수 참조)
- `localStorage`에 선택 저장

---

## 디자인 참고 파일

| 파일 | 내용 |
|------|------|
| `designs.html` | v1 디자인 목업 (6 테마, 에디터+채팅) |
| `designs-v2.html` | **v2 디자인 목업 (11 테마, 에디터+터미널+채팅+도구)** |

브라우저에서 열어 테마별 미리보기 가능.

---

## 개발 일정

| 단계 | 내용 | 예상 기간 |
|------|------|----------|
| 1. 뼈대 | Wails 프로젝트 생성 + 4분할 레이아웃 + 테마 | 반나절 |
| 2. 파일트리 | Go: ListFiles / React: 트리 컴포넌트 | 반나절 |
| 3. 에디터 | CodeMirror 6 + 파일 탭 + 구문강조 | 1일 |
| 4. Chat | Go: StreamChat 바인딩 / React: 메시지 UI | 1일 |
| 5. 터미널 | Go: PTY / React: xterm.js | 1일 |
| 6. Git | Go: gitinfo / React: 상태바+뱃지 | 반나절 |
| 7. 통합/QA | 전체 연동 + 테마 + 단축키 | 1일 |
| **MVP 합계** | | **~5일** |

---

## 빌드 & 배포

### Makefile 타겟

```makefile
build-ide:
    cd cmd/techai-ide/frontend && npm run build
    wails build -o dist/techai-ide

build-ide-all:
    # macOS (ARM/Intel)
    wails build -platform darwin/arm64 -o dist/techai-ide-darwin-arm64
    wails build -platform darwin/amd64 -o dist/techai-ide-darwin-amd64
    # Windows
    wails build -platform windows/amd64 -o dist/techai-ide-windows-amd64.exe
    # Linux
    wails build -platform linux/amd64 -o dist/techai-ide-linux-amd64
```

### 바이너리 크기 예측

| 구성 | 크기 |
|------|------|
| Go 엔진 (internal/*) | ~20MB |
| Wails 런타임 | ~5MB |
| Knowledge docs (embedded) | ~5MB |
| React 번들 (embedded) | ~0.3MB |
| **합계** | **~30-35MB** |

### 폐쇄망 배포
- USB로 `techai-ide.exe` 단일 파일 복사
- `.tgc/config.yaml` 자동 생성 (API URL/Key)
- 기존 TUI와 설정 공유 (같은 ConfigDir)

---

## VS Code Extension 대안 경로

Wails IDE와 별개로, **VS Code Extension도 병행 가능**:

### 장점
- VS Code 익스텐션 생태계 전체 활용
- `.vsix` 파일로 폐쇄망 설치 가능
- 개발 비용 낮음 (WebView 패널 1개)

### 구조
```
techai-vscode/
├── src/
│   ├── extension.ts       ← VS Code API
│   ├── chatPanel.ts       ← WebView 패널 (Chat UI)
│   └── techai-client.ts   ← TECHAI TUI 바이너리와 통신
├── webview/               ← Chat UI (HTML/CSS/JS)
├── package.json           ← VS Code 확장 매니페스트
└── 빌드 → techai-code.vsix (~3MB)
```

### 통신 방식
1. VS Code Extension → `techai exec "prompt"` subprocess 호출
2. 또는 TECHAI를 HTTP 서버 모드로 실행 → Extension이 SSE로 연결

### 우선순위
- **1순위**: Wails IDE (독립 제품, 폐쇄망 단일 바이너리)
- **2순위**: VS Code Extension (기존 VS Code 사용자용)

---

## 참고 자료

- [Wails v2 공식 문서](https://wails.io/docs/)
- [CodeMirror 6](https://codemirror.net/)
- [xterm.js](https://xtermjs.org/)
- designs-v2.html (11테마 디자인 목업)
- docs/superpowers/specs/2026-04-10-unified-roadmap-design.md (기존 로드맵)
- docs/UX_STRUCTURE.md (현재 TUI 구조)
