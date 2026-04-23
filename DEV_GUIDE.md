# TECHAI CODE — 개발 가이드

> 이 레포는 **두 개의 독립 제품**을 포함합니다.
> 각각 별도로 빌드/배포/버전 관리됩니다.

---

## 레포 구조

```
택가이코드/
├── cmd/tgc/              ← TUI 진입점
├── internal/             ← TUI 공유 엔진 (LLM, 도구, 지식)
│   ├── llm/              ← LLM 스트리밍 클라이언트
│   ├── tools/            ← 15개 도구 (file_read/write, grep, shell 등)
│   ├── config/           ← .tgc/config.yaml 관리
│   ├── knowledge/        ← Knowledge Packs (81개 내장 문서)
│   ├── app/              ← Bubble Tea TUI 앱
│   ├── ui/               ← TUI 렌더링 (chat, menu, palette)
│   ├── companion/        ← 브라우저 컴패니언 (SSE)
│   ├── session/          ← 세션 저장
│   ├── hooks/            ← 라이프사이클 훅
│   └── ...
├── knowledge/            ← 내장 지식 문서 (go:embed)
├── web/                  ← 컴패니언 웹 UI (go:embed)
├── go.mod                ← TUI Go 모듈 (github.com/kimjiwon/tgc)
├── Makefile              ← TUI 빌드 스크립트
│
├── techai-ide/           ← IDE 제품 (별도 Go 모듈)
│   ├── go.mod            ← IDE Go 모듈 (techai-ide)
│   ├── main.go           ← Wails 앱 진입점 + 시스템 메뉴
│   ├── app.go            ← 파일 시스템 API
│   ├── chat.go           ← LLM 채팅 엔진 (자체 구현)
│   ├── git.go            ← Git 연동 (status/diff/log/graph)
│   ├── terminal.go       ← PTY 터미널
│   ├── knowledge.go      ← Knowledge Packs UI
│   ├── toolparse.go      ← text tool_call 파싱
│   ├── settings.go       ← 설정 관리
│   ├── config.go         ← .tgc/config.yaml 로더 (TUI와 공유)
│   ├── frontend/         ← React + TypeScript
│   │   └── src/components/  ← 15개 UI 컴포넌트
│   ├── FEATURES.md       ← 68개 기능 리스트
│   └── TODO.md           ← 남은 구현 계획
│
├── docs/                 ← 문서
│   ├── TECHAI_IDE_PLAN.md
│   └── ...
├── designs-v3.html       ← IDE 디자인 목업 (11 테마)
└── DEV_GUIDE.md          ← 이 파일
```

---

## 버전 관리

| 제품 | 현재 버전 | 태그 형식 | 예시 |
|------|----------|----------|------|
| **TUI** | v0.9.6 | `v{major}.{minor}.{patch}` | v0.9.6, v1.0.0 |
| **IDE** | v0.1.0 | `ide-v{major}.{minor}.{patch}` | ide-v0.1.0, ide-v0.2.0 |

```bash
# TUI 버전 태그
git tag v0.9.7
git push origin v0.9.7

# IDE 버전 태그
git tag ide-v0.1.0
git push origin ide-v0.1.0
```

---

## TUI 개발

### 빌드

```bash
# 개발 빌드 (디버그 모드)
make build

# 프로덕션 릴리스 빌드 (모든 플랫폼)
make build-release

# 온프레미스 빌드 (사내망용, Qwen3-Coder)
make build-onprem

# 테스트
make test
# 또는
go test ./...

# 린트
make lint
```

### 주요 파일

| 파일 | 역할 |
|------|------|
| `cmd/tgc/main.go` | TUI 진입점, 플래그 파싱 |
| `internal/app/app.go` | Bubble Tea 메인 모델 (2000+ 줄) |
| `internal/llm/client.go` | LLM 스트리밍 + tool_call 파싱 |
| `internal/llm/capabilities.go` | 모델 레지스트리 |
| `internal/tools/registry.go` | 15개 도구 실행 |
| `internal/config/config.go` | 설정 로드/저장/마이그레이션 |
| `internal/knowledge/store.go` | 지식 검색/주입 |
| `Makefile` | 빌드 타겟 (build, build-all, build-onprem, build-gemma, build-kimi) |

### 배포

```bash
# 전체 릴리스 (16개 바이너리)
make build-release

# GitHub 릴리스
gh release create v0.9.7 dist/* --title "v0.9.7 — ..." --notes "..."
```

### 설정 공유

TUI와 IDE는 같은 설정 파일을 사용:
```
~/.tgc/config.yaml        ← 기본 (Novita.ai)
~/.tgc-onprem/config.yaml ← 온프레미스 (사내망)
```

---

## IDE 개발

### 사전 요구사항

```bash
# Wails CLI (최초 1회)
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Node.js (v18+)
node --version

# frontend 의존성 (최초 1회 또는 package.json 변경 시)
cd techai-ide/frontend && npm install
```

### 개발 모드

```bash
cd techai-ide

# 개발 서버 (Hot Reload — Go + React 둘 다 자동 반영)
wails dev

# → 앱 창이 자동으로 뜸
# → frontend 수정 → 즉시 반영 (Vite HMR)
# → Go 수정 → 자동 리빌드 + 앱 재시작
```

### 프로덕션 빌드

```bash
cd techai-ide

# macOS (현재 아키텍처)
wails build

# macOS Intel
wails build -platform darwin/amd64

# Windows
wails build -platform windows/amd64

# 결과물
# build/bin/techai-ide.app  (macOS, 12MB)
# build/bin/techai-ide.exe  (Windows, 12MB)
```

### Wails 바인딩 재생성

Go에 새 public 메서드를 추가하면 프론트엔드 타입 재생성 필요:

```bash
cd techai-ide
wails generate module

# 결과: frontend/wailsjs/go/main/App.d.ts 자동 업데이트
# → TypeScript에서 import { NewMethod } from '../../wailsjs/go/main/App' 사용 가능
```

### 주요 파일

| 파일 | 역할 |
|------|------|
| **Go Backend** | |
| `main.go` | Wails 앱 설정 + 시스템 메뉴 정의 |
| `app.go` | 파일 시스템 API (ListFiles, ReadFile, WriteFile, OpenFolder 등) |
| `chat.go` | LLM 채팅 엔진 (SendMessage, ClearChat, 도구 실행, 스트리밍) |
| `git.go` | Git 연동 (GetGitInfo, GitDiff, GitGraph, GitStage, GitCommit) |
| `terminal.go` | PTY 터미널 (StartTerminal, WriteTerminal, SetShell) |
| `knowledge.go` | Knowledge Packs (74개 문서 스캔, 토글, 시스템 프롬프트 주입) |
| `toolparse.go` | text tool_call 파싱 (3포맷: tool_call, function, parameter) |
| `settings.go` | 설정 (GetSettings, SaveSettings) |
| `config.go` | .tgc/config.yaml 로더 |
| **React Frontend** | |
| `src/App.tsx` | 메인 레이아웃 + 단축키 + 패널 상태 관리 |
| `src/components/Editor.tsx` | 에디터 (탭, 저장, 브레드크럼, 검색, 이미지 미리보기) |
| `src/components/CodeEditor.tsx` | CodeMirror 6 래퍼 (15개 언어, 테마, 확장) |
| `src/components/ChatPanel.tsx` | AI 채팅 (스트리밍, 마크다운, 슬래시 명령, Knowledge) |
| `src/components/FileTree.tsx` | 파일 트리 (60종 아이콘, 컨텍스트 메뉴, Live Server) |
| `src/components/Terminal.tsx` | 터미널 (다중 탭, 쉘 선택, 명령어 히스토리) |
| `src/components/GitPanel.tsx` | Git 패널 (Stage, Commit, Inline Diff) |
| `src/components/GitGraph.tsx` | Git Graph (커밋 히스토리, 브랜치 시각화) |
| `src/components/SearchPanel.tsx` | 검색 (실시간, Include/Exclude 필터) |
| `src/components/ThemePicker.tsx` | 테마 선택 (11개) |
| `src/components/SettingsPanel.tsx` | 설정 (API URL/Key/Model) |
| `src/components/QuickOpen.tsx` | Cmd+P 파일 빠른 열기 |
| `src/components/CommandPalette.tsx` | Cmd+Shift+P 커맨드 팔레트 |
| `src/components/Toast.tsx` | 알림 토스트 |
| `src/components/AboutDialog.tsx` | About 다이얼로그 |
| `src/components/ResizeHandle.tsx` | 패널 리사이즈 드래그 |
| `src/style.css` | 글로벌 CSS 변수 + 11개 테마 |

### 새 기능 추가 패턴

**Go API 추가:**
```go
// app.go (또는 새 파일)
func (a *App) MyNewFeature(arg string) (string, error) {
    // ... 로직
    return result, nil
}
```

```bash
# 바인딩 재생성
cd techai-ide && wails generate module
```

**React에서 호출:**
```tsx
import { MyNewFeature } from '../../wailsjs/go/main/App'

const result = await MyNewFeature("arg")
```

**Wails 이벤트 (Go → React):**
```go
// Go에서 emit
runtime.EventsEmit(a.ctx, "my:event", data)
```
```tsx
// React에서 수신
EventsOn('my:event', (data) => { ... })
```

### 테마 추가 방법

`src/style.css`에 CSS 변수 블록 추가:
```css
body.t-mytheme {
  --bg-base: #...; --bg-activity: #...;
  --accent: #...; --status-bg: #...;
  /* ... 전체 변수 */
}
```

`src/components/ThemePicker.tsx`의 `themes` 배열에 추가:
```tsx
{ id: 't-mytheme', name: 'My Theme', dot: '#...', dark: true },
```

---

## 양쪽 동시 작업 시 주의사항

1. **go.mod 충돌 없음** — TUI(`./go.mod`)와 IDE(`techai-ide/go.mod`)는 별도 모듈
2. **설정 공유** — 둘 다 `~/.tgc/config.yaml` 사용. 한쪽에서 설정 변경하면 다른 쪽도 반영
3. **커밋** — 한 레포이므로 커밋 시 TUI/IDE 변경 혼재 가능. 커밋 메시지에 `[TUI]` 또는 `[IDE]` 접두사 권장
4. **태그** — TUI: `v0.x.x`, IDE: `ide-v0.x.x` 형식으로 구분
5. **internal/ 변경** — TUI 엔진 변경 시 IDE에 직접 영향 없음 (IDE는 자체 LLM 클라이언트 보유). 향후 엔진 통합 시 영향.

---

## 키보드 단축키 (IDE)

| 단축키 | 기능 |
|--------|------|
| `Cmd+P` | 파일 빠른 열기 |
| `Cmd+Shift+P` | 커맨드 팔레트 |
| `Cmd+S` | 저장 |
| `Cmd+W` | 탭 닫기 |
| `Cmd+F` | 파일 내 검색 |
| `Cmd+Shift+F` | 프로젝트 검색 |
| `Cmd+B` | 사이드바 토글 |
| `Cmd+J` / `Cmd+\`` | 터미널 토글 |
| `Cmd+\\` | 에디터 스플릿 |
| `Cmd+1` | Explorer |
| `Cmd+2` | Search |
| `Cmd+3` | Git |
| `Cmd+,` | 테마 선택 |
| `Cmd+O` | 폴더 열기 |

---

## 참고 문서

| 문서 | 내용 |
|------|------|
| `docs/TECHAI_IDE_PLAN.md` | IDE 전체 계획서 (기술 스택, 레이아웃, API 설계) |
| `techai-ide/FEATURES.md` | 68개 기능 전체 리스트 |
| `techai-ide/TODO.md` | 남은 구현 + 아키텍처 노트 |
| `designs-v3.html` | 디자인 목업 (11 테마, 브라우저에서 열기) |
| `docs/UX_STRUCTURE.md` | TUI UX 구조 |
| `docs/superpowers/specs/2026-04-10-unified-roadmap-design.md` | TUI 통합 로드맵 |
