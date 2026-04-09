# 택갈이코드 차기 기능 분석 보고서

> **작성일**: 2026-04-09
> **분석 대상**: OpenCode (Go 터미널 코딩 에이전트), gstack (Garry Tan의 Claude Code 스킬셋)
> **목적**: 폐쇄망 환경에서 택갈이코드에 이식할 수 있는 기능 식별 + 브라우저 컴패니언 설계

---

## 현재 상태 (v0.5.0)

| 항목 | 현재 |
|------|------|
| 도구 | 7개 (file_read, file_write, file_edit, list_files, shell_exec, grep_search, glob_search) |
| 내장 지식 | 38문서 (BXM Tier 0 + CSS/React/Charts Tier 1 + Vue/Java Tier 2 + Python Tier 3 + Skills) |
| 세션 | 인메모리 (재시작 시 소멸) |
| 토큰 카운팅 | `tokenCount++` per SSE chunk (실제 토큰 아님) |
| file_edit | 정확일치 `strings.Replace` only |
| Git | shell_exec으로 수동 실행 |
| 컨텍스트 관리 | 없음 (tool loop 20회 하드캡만) |
| 스냅샷/Undo | 없음 |
| 브라우저 연동 | 없음 |

---

## Part 1: OpenCode에서 가져올 기능 (가치×난이도 순)

### Rank 1: SQLite 영구 세션 — Score 9/10

**가치**: 극히 높음 (세션이 재시작 후에도 유지, 작업 이어가기)
**난이도**: 중 (2-3주)
**의존성**: `modernc.org/sqlite` (pure Go, CGo 불필요, 단일 바이너리 유지)

**OpenCode 구현 참고**:
- `opencode/packages/opencode/src/storage/db.ts:90-95` — SQLite WAL 모드 + pragma 설정
- `opencode/packages/opencode/src/session/session.sql.ts` — 스키마: sessions, messages, parts, todos
- `opencode/packages/opencode/src/session/index.ts:66-180` — CRUD: create, fork, archive, search, pagination

**택갈이코드 구현 계획**:
```
internal/storage/
├── db.go          # modernc.org/sqlite 초기화, WAL 모드, busy_timeout=5000
├── session.go     # sessions 테이블 CRUD
└── message.go     # messages 테이블 CRUD (JSON blob)
```

스키마:
```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    title TEXT,
    mode INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    archived INTEGER DEFAULT 0
);

CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    session_id TEXT REFERENCES sessions(id),
    role TEXT NOT NULL,
    content TEXT NOT NULL,  -- JSON blob
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**변경점**: `app.go`의 `history []openai.ChatCompletionMessage`를 DB 백킹으로 교체 + `/sessions` 슬래시 명령 추가.
**바이너리 증가**: ~5MB (pure-Go SQLite)

---

### Rank 2: 컨텍스트 압축 + 실제 토큰 카운팅 — Score 8.5/10

**가치**: 극히 높음 (컨텍스트 오버플로 방지, 긴 세션 가능)
**난이도**: 중 (1-2주)
**의존성**: 없음 (go-openai 응답의 `usage` 필드 활용)

**OpenCode 구현 참고**:
- `opencode/packages/opencode/src/session/compaction.ts:10-11` — `PRUNE_MINIMUM = 20000`, `PRUNE_PROTECT = 40000`
- 2단계 pruning: (1) 이전 tool output을 "[truncated]"로 교체, (2) LLM 요약

**택갈이코드 구현 계획**:
1. `openai.ChatCompletionResponse.Usage.TotalTokens`에서 실제 토큰 추출
2. 누적 토큰 추적 (세션별)
3. 모델 한도 80% 도달 시:
   - Phase 1: 오래된 tool output → `"[output truncated]"` 교체
   - Phase 2: 동일 LLM 엔드포인트에 대화 요약 요청
4. 상태바에 토큰 사용량 표시

**현재 문제**: `tokenCount++`는 SSE 청크 수이지 토큰이 아님 (app.go ~350줄)

---

### Rank 3: 퍼지 매칭 file_edit — Score 8/10

**가치**: 높음 (tool 재시도 감소 → 폐쇄망의 느린 네트워크에서 귀중)
**난이도**: 쉬움~중 (1주)
**의존성**: `github.com/agnivade/levenshtein` (pure Go, 소형)

**OpenCode 구현 참고**:
- `opencode/packages/opencode/src/tool/edit.ts` — 9단계 fallback replacer

**택갈이코드에 이식할 4단계 (우선순위)**:
1. `LineTrimmedReplacer` — `strings.TrimSpace` per line 후 매칭
2. `WhitespaceNormalizedReplacer` — `\s+`를 단일 공백으로 정규화
3. `IndentationFlexibleReplacer` — 들여쓰기 무시 매칭, 대상 indent 재적용
4. `BlockAnchorReplacer` — Levenshtein 유사도로 퍼지 매칭 (정확매칭 실패 시에만)

**현재 코드**: `internal/tools/registry.go` — `strings.Replace(content, oldText, newText, 1)` 단일 정확매칭

---

### Rank 4: Git 구조화 통합 — Score 7.5/10

**가치**: 높음 (폐쇄망에서 git은 핵심 도구)
**난이도**: 쉬움 (3-5일)
**의존성**: 없음 (`os/exec` + git CLI)

**OpenCode 구현 참고**:
- `opencode/packages/opencode/src/git/index.ts` — `branch`, `status`, `diff`, `stats`, `mergeBase`, `show`
- `--no-optional-locks` 플래그로 충돌 방지

**택갈이코드 구현 계획**:
```
internal/git/
└── git.go    # Branch(), Status(), Diff(), DiffStat(), Log(n)
```

- `git_diff`와 `git_log`를 새 도구로 `registry.go`에 추가 (읽기 전용, 안전)
- `GatherFullContext()`에 구조화된 git 정보 주입 강화

---

### Rank 5: 파일 스냅샷 / Undo — Score 7/10

**가치**: 높음 (AI 생성 편집의 안전망)
**난이도**: 중 (2-3주)

**OpenCode 구현 참고**:
- `opencode/packages/opencode/src/snapshot/index.ts` — Shadow git repo per project
- `track()` → `restore()` → `revert()` → `diff()`

**택갈이코드 간소화 구현**:
1. `file_write`/`file_edit` 전에 원본을 `~/.tgc/snapshots/{session_id}/{timestamp}/{filepath}`에 복사
2. 편집 목록을 세션별 manifest JSON으로 추적
3. `/undo` 슬래시 명령으로 마지막 스냅샷 복원
4. `/diff` 명령으로 변경사항 표시 (`github.com/sergi/go-diff`)

**트레이드오프**: Shadow git (OpenCode 방식)은 강건하지만 복잡. 단순 파일 복사가 가치의 80%를 20% 노력으로 달성.

---

### Rank 6: Diff 뷰 — Score 6.5/10

**가치**: 중상 (AI 변경 시각적 확인)
**난이도**: 쉬움 (3-5일)
**의존성**: `github.com/sergi/go-diff` (pure Go, 무의존)

**구현**: file_edit/file_write 후 unified diff 생성 → Lip Gloss red/green 컬러링으로 Bubble Tea viewport에 렌더링

---

### Rank 7: 멀티 모델 설정 — Score 6/10

**가치**: 중 (폐쇄망에서 1-3개 모델 운영 시)
**난이도**: 쉬움 (2-3일)

**구현**: `config.yaml`에서 모드별 endpoint/model/context_window 지정:
```yaml
models:
  super:
    id: "my-large-model"
    context_window: 128000
    endpoint: "http://internal-llm:8080/v1"
  dev:
    id: "my-code-model"
    context_window: 32000
    endpoint: "http://internal-llm:8080/v1"
```

---

### Rank 8: LSP 통합 — Score 5.5/10 (가치 최고, 난이도도 최고)

**가치**: 매우 높음 (클라우드 없이 코드 인텔리전스)
**난이도**: 어려움 (4-6주)

**Phase 1** (2주): `gopls` only — diagnostics + hover
**Phase 2** (2주): definition + references 추가
**Phase 3** (2주): 멀티 서버 (설정 기반)

**권장**: Rank 1-5 완료 후 진행

---

### Rank 9: Vim 키바인딩 — Score 4/10

**가치**: 낮음 (QoL)
**난이도**: 중 (1-2주)

**권장**: 모든 기능 구현 후 마지막

---

## Part 2: 브라우저 컴패니언 설계 (gstack 기반)

### gstack 핵심 아키텍처

gstack은 **headless Chromium 데몬** + **로컬 HTTP 서버** + **SSE 실시간 스트리밍** 패턴 사용.

핵심 참고 파일:
- `gstack/browse/src/activity.ts:1-210` — Ring buffer + pub/sub + SSE 스트리밍
- `gstack/browse/src/server.ts:1614-1676` — SSE endpoint (`text/event-stream`, heartbeat, abort cleanup)
- `gstack/extension/sidepanel.js:1-86` — SSE 클라이언트 (EventSource + 자동 재연결)
- `gstack/ARCHITECTURE.md` — 데몬 모델 선택 이유 (상태 유지, 서브초 지연시간)

### 택갈이코드 브라우저 컴패니언 아키텍처

```
┌─────────────────────────────────┐      SSE       ┌────────────────────────────────┐
│  Terminal (Bubble Tea TUI)      │ ─────────────► │  Browser (localhost:8787)      │
│                                 │                │                                │
│  app.Model                      │                │  index.html (Go embed)         │
│   └─ companion.Hub              │                │   ├─ 채팅 실시간 뷰            │
│       .Emit(event) ─────────────┤                │   ├─ 플랜/태스크 뷰어          │
│                                 │                │   │   └─ Mermaid.js 다이어그램  │
│  goroutine: companion.Server    │                │   ├─ 코드 diff (highlight.js)  │
│   ├─ GET /           (HTML)     │                │   ├─ 도구 활동 피드            │
│   ├─ GET /events     (SSE)      │                │   └─ 세션 통계 (토큰, 시간)    │
│   ├─ GET /api/state  (JSON)     │                │                                │
│   └─ GET /assets/*   (static)   │                │  EventSource('/events')        │
└─────────────────────────────────┘                └────────────────────────────────┘
```

### 이벤트 흐름

```
app.Model.Update()
  ├─ streamChunkMsg  ──► hub.Emit(EventChunk{content, tokens})
  ├─ toolResultMsg   ──► hub.Emit(EventToolResult{name, output})
  ├─ sendMessage()   ──► hub.Emit(EventUserMessage{input})
  ├─ tab switch      ──► hub.Emit(EventModeChange{mode})
  └─ stream complete ──► hub.Emit(EventComplete{elapsed, tokens})
```

### 파일 구조

```
internal/companion/
├── hub.go          # Ring buffer + pub/sub (gstack activity.ts 번역)
├── server.go       # net/http 서버, SSE endpoint, REST state endpoint
├── events.go       # 이벤트 타입 정의
├── browser.go      # 기본 브라우저 열기 (darwin/linux/windows)
└── embed.go        # //go:embed web/* 선언

web/                # 정적 에셋 (빌드 스텝 없음, embed로 바이너리 내장)
├── index.html      # 단일 페이지 대시보드
├── style.css       # 다크 테마 (TUI 팔레트 #0F172A, #60A5FA 등)
├── app.js          # SSE 클라이언트 + DOM 업데이트
└── vendor/
    ├── mermaid.min.js      # 다이어그램 렌더링 (~800KB)
    ├── highlight.min.js    # 코드 하이라이팅 (~50KB)
    └── diff2html.min.js    # Diff 렌더링 (~40KB)
```

### 핵심 구현: Hub (gstack activity.ts → Go)

```go
// internal/companion/hub.go
type Event struct {
    ID        int64       `json:"id"`
    Timestamp int64       `json:"timestamp"`
    Type      string      `json:"type"`
    Data      interface{} `json:"data"`
}

type Hub struct {
    mu          sync.RWMutex
    buffer      []Event        // ring buffer (capacity: 500)
    nextID      int64
    subscribers map[chan Event]struct{}
}

func (h *Hub) Emit(evt Event)                    // 논블로킹, 모든 구독자 알림
func (h *Hub) Subscribe() (chan Event, func())    // 채널 + 구독 해제 함수
func (h *Hub) After(id int64) ([]Event, bool)     // 백로그 + gap 감지
```

### 핵심 구현: SSE Endpoint

```go
// internal/companion/server.go
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")

    ch, unsub := s.hub.Subscribe()
    defer unsub()

    for {
        select {
        case evt := <-ch:
            fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Type, jsonMarshal(evt))
            w.(http.Flusher).Flush()
        case <-r.Context().Done():
            return
        case <-time.After(15 * time.Second):
            fmt.Fprintf(w, ": heartbeat\n\n")  // 연결 유지
            w.(http.Flusher).Flush()
        }
    }
}
```

### app.go 통합 (최소 변경)

```go
// Model에 필드 추가
type Model struct {
    // ... 기존 필드 ...
    companion *companion.Hub  // nil이면 비활성
}

// /companion 슬래시 명령
case "/companion":
    if m.companionServer == nil {
        m.companionServer = companion.NewServer(m.companion, 8787)
        go m.companionServer.Start()
    }
    companion.OpenBrowser("http://localhost:8787")
```

### 설계 결정 근거

| 결정 | 이유 |
|------|------|
| SSE (WebSocket 대신) | 단방향 TUI→브라우저 충분. gstack이 검증. `net/http` 표준 라이브러리만 사용 |
| `//go:embed` 에셋 | 단일 바이너리 유지, 폐쇄망 OK, npm 빌드 불필요 |
| 별도 고루틴 | DEBUG_TRANSPORT_FREEZE 준수: LLM HTTP 클라이언트와 완전 분리 |
| `/companion` 명령 | 항상 서버 시작하지 않음, 필요할 때만 on-demand |
| Ring buffer (500개) | gstack 패턴. 브라우저 접속 전 이벤트도 백로그로 복구 |

### 구현 규모

| 컴포넌트 | 줄 수 | 소요 시간 |
|----------|-------|----------|
| hub.go | ~80줄 | 1시간 |
| server.go | ~120줄 | 2시간 |
| events.go + browser.go | ~65줄 | 45분 |
| web/ (HTML+CSS+JS) | ~350줄 | 3시간 |
| 벤더 번들링 | 다운로드 | 30분 |
| app.go 수정 | ~30줄 변경 | 1시간 |
| **합계** | **~650줄** | **~10시간** |

**바이너리 증가**: ~1.5MB (Mermaid ~800KB + highlight.js ~50KB + diff2html ~40KB + 커스텀)

---

## Part 3: 구현 로드맵

| Phase | 기능 | 기간 | 의존성 |
|-------|------|------|--------|
| **Phase 1** | SQLite 세션 (Rank 1) + 토큰 카운팅 (Rank 2) | 3-4주 | `modernc.org/sqlite` |
| **Phase 2** | 퍼지 file_edit (Rank 3) + Git 통합 (Rank 4) + Diff 뷰 (Rank 6) | 2-3주 | `go-diff`, `levenshtein` |
| **Phase 3** | 파일 스냅샷 (Rank 5) + 멀티 모델 (Rank 7) | 2-3주 | Phase 1 (세션 추적) |
| **Phase 4** | **브라우저 컴패니언** | 1-2주 | Phase 1-2 (이벤트 구조) |
| **Phase 5** | LSP 통합 (Rank 8) | 4-6주 | Phase 2 (Git 프로젝트 감지) |
| **Phase 6** | Vim 키바인딩 (Rank 9) | 1-2주 | 없음 |

---

## 핵심 제약 사항 (반드시 준수)

1. **DEBUG_TRANSPORT_FREEZE**: HTTP transport 래핑 금지. 컴패니언 서버는 별도 고루틴에서 `net/http` 직접 사용
2. **구 app.go 기반 유지**: `app.go` 변경은 최소 단위로 추가 (필드 추가 + Emit 호출만)
3. **단일 바이너리**: 모든 에셋은 `//go:embed`로 내장. 외부 파일 의존 금지
4. **폐쇄망 호환**: 네트워크 의존 기능 금지. 모든 것이 바이너리 안에 포함

---

## 참고 소스 파일

### OpenCode (`/Users/kimjiwon/Desktop/kimjiwon/opencode/`)
- `packages/opencode/src/session/index.ts:66-180` — 세션 CRUD
- `packages/opencode/src/session/compaction.ts:10-130` — 컨텍스트 압축
- `packages/opencode/src/session/session.sql.ts` — SQLite 스키마
- `packages/opencode/src/storage/db.ts:90-95` — SQLite WAL pragma
- `packages/opencode/src/tool/edit.ts` — 9단계 fallback replacer
- `packages/opencode/src/git/index.ts` — 구조화 git 연산
- `packages/opencode/src/snapshot/index.ts` — Shadow git 스냅샷
- `packages/opencode/src/lsp/index.ts` — LSP 클라이언트
- `packages/opencode/src/provider/provider.ts` — 20+ LLM 프로바이더

### gstack (`/Users/kimjiwon/Desktop/kimjiwon/gstack/`)
- `browse/src/activity.ts:1-210` — Ring buffer + pub/sub + SSE 핵심 패턴
- `browse/src/server.ts:1614-1676` — SSE endpoint 구현
- `extension/sidepanel.js:1-86` — SSE 클라이언트 (EventSource + 자동 재연결)
- `ARCHITECTURE.md` — 데몬 모델 + CLI 설계 원칙
- `BROWSER.md:25-48` — CLI↔데몬 HTTP 아키텍처 다이어그램

### 택갈이코드 (`/Users/kimjiwon/Desktop/kimjiwon/택갈이코드/`)
- `internal/app/app.go:42-76` — Model 구조체 (컴패니언 hub 필드 추가점)
- `internal/app/app.go:313-407` — 스트림/도구 핸들러 (Emit 호출점)
- `internal/app/app.go:485-517` — handleSlashCommand (/companion 추가점)
- `internal/app/app.go:665-696` — sendMessage (사용자 메시지 Emit점)
- `internal/tools/registry.go:232-372` — 도구 실행 (tool_start/result Emit점)
- `internal/ui/styles.go:9-22` — 컬러 팔레트 (웹 UI 동기화)
- `internal/config/config.go` — 설정 구조
- `docs/DEBUG_TRANSPORT_FREEZE.md` — HTTP 래핑 금지 제약
