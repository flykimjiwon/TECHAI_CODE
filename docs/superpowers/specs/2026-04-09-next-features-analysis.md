# 택갈이코드 차기 기능 분석 보고서

> **작성일**: 2026-04-09
> **현재 버전**: v0.5.0 (Embedded Knowledge Store 완성)
> **분석 대상**: OpenCode (Go 터미널 코딩 에이전트), gstack (Garry Tan의 Claude Code 스킬셋)
> **목적**: 폐쇄망 환경에서 택갈이코드에 이식할 수 있는 기능 식별 + 브라우저 컴패니언 설계
> **다음 구현 시 반드시 이 문서를 참조할 것**

---

## 단일 바이너리 원칙

택갈이코드의 핵심 가치는 **단일 실행파일 배포**입니다. 모든 추가 기능은 이 원칙을 깨뜨리지 않아야 합니다.

### 빌드 타임 vs 런타임 분리

```
┌─────────────────────────────────────────────────────────────┐
│  techai (단일 바이너리, ~25MB 예상)                          │
│                                                             │
│  빌드 시 내장 (go:embed, 변하지 않음):                       │
│  ├─ knowledge/ (38+ 레퍼런스 문서, ~9MB)                    │
│  ├─ web/ (브라우저 컴패니언 HTML/CSS/JS, ~1.5MB)            │
│  └─ SQLite 엔진 코드 (modernc.org/sqlite, ~5MB)            │
│                                                             │
│  빌드 시 컴파일 (Go 소스):                                   │
│  ├─ internal/knowledge/ (지식 검색/주입)                     │
│  ├─ internal/storage/ (SQLite 세션 관리)                     │
│  ├─ internal/companion/ (브라우저 컴패니언 서버)              │
│  ├─ internal/tools/ (9-11개 도구)                           │
│  └─ internal/app/ (Bubble Tea TUI)                          │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│  ~/.tgc/ (런타임 데이터, 사용자 머신에 자동 생성)             │
│                                                             │
│  ├─ config.yaml          설정 (이미 존재)         ~1KB      │
│  ├─ debug.log            디버그 로그 (이미 존재)   가변      │
│  ├─ sessions.db          SQLite DB [Phase 1 신규]  ~2-50MB  │
│  └─ snapshots/           파일 백업 [Phase 3 신규]  가변      │
│      └─ {session_id}/                                       │
│          └─ {timestamp}/{filepath}                          │
└─────────────────────────────────────────────────────────────┘
```

### 바이너리 크기 변화 예측

| 버전 | 크기 | 변화 | 원인 |
|------|------|------|------|
| v0.4.0 (Phase 1-4) | ~15MB | — | 기본 TUI + LLM 클라이언트 + 7 도구 |
| v0.5.0 (Knowledge Store) | ~16MB | +1MB | knowledge/ embed (38문서) |
| v0.6.0 (Phase 1: SQLite) | ~21MB | +5MB | `modernc.org/sqlite` pure Go 엔진 |
| v0.7.0 (Phase 2: Edit+Git) | ~21.5MB | +0.5MB | `go-diff`, `levenshtein` 라이브러리 |
| v0.8.0 (Phase 3: Snapshot) | ~21.5MB | ±0 | 런타임 파일 복사만, 바이너리 변화 없음 |
| v0.9.0 (Phase 4: Companion) | ~23MB | +1.5MB | web/ embed (Mermaid+highlight.js) |
| v1.0.0 (Phase 5: LSP) | ~24MB | +1MB | JSON-RPC 클라이언트 코드 |
| **최종 예상** | **~24MB** | — | 여전히 단일 파일, USB로 복사 가능 |

### 런타임 데이터 크기 예측

| 데이터 | 위치 | 1달 사용 | 1년 사용 | 관리 |
|--------|------|---------|---------|------|
| sessions.db | `~/.tgc/sessions.db` | ~5-20MB | ~50-100MB | `VACUUM`, 아카이브 명령 |
| snapshots/ | `~/.tgc/snapshots/` | ~10-50MB | 자동 정리 | 7일 이상 스냅샷 자동 삭제 |
| debug.log | `~/.tgc/debug.log` | ~1-5MB | 자동 로테이션 | 10MB 초과 시 자동 트림 |
| config.yaml | `~/.tgc/config.yaml` | ~1KB | ~1KB | 수동 관리 |

**결론: 바이너리는 ~24MB로 유지, 런타임 데이터는 `~/.tgc/`에 격리. `techai --reset`으로 초기화 가능.**

---

## 현재 상태 (v0.5.0)

| 항목 | 현재 | 문제점 |
|------|------|--------|
| 도구 | 7개 (file_read, file_write, file_edit, list_files, shell_exec, grep_search, glob_search) | file_edit이 정확일치만 지원 |
| 내장 지식 | 38문서 (BXM/CSS/React/Charts/Vue/Java/Python/Skills) | — |
| 세션 | 인메모리 (`history []Message`) | **재시작하면 전부 소멸** |
| 토큰 카운팅 | `tokenCount++` per SSE chunk | **실제 토큰이 아님** (청크 수) |
| 컨텍스트 관리 | tool loop 20회 하드캡 | **긴 대화에서 오버플로** |
| Git | shell_exec으로 수동 | 구조화된 정보 없음 |
| 편집 안전망 | 없음 | **AI가 파일 망가뜨리면 복구 불가** |
| 브라우저 | 없음 | 터미널에서 플랜/다이어그램 보기 어려움 |

---

## 구현 예정 기능 상세

### Phase 1: SQLite 영구 세션 + 실제 토큰 카운팅

> **목표 버전**: v0.6.0
> **기간**: 3-4주
> **단일 바이너리**: O (`modernc.org/sqlite` = pure Go, CGo 불필요)
> **바이너리 증가**: +5MB
> **런타임 데이터**: `~/.tgc/sessions.db` (자동 생성, ~5-100MB)

#### 기능 1-A: SQLite 영구 세션

**문제**: 현재 대화 내역이 `app.go`의 `history []openai.ChatCompletionMessage`에만 존재. 터미널 닫으면 모든 작업 기록 소멸.

**해결**: SQLite DB에 세션/메시지 저장. 재시작 후에도 이어서 작업 가능.

**사용자 경험**:
```
$ techai                    # 새 세션 자동 생성
> BXM Bean 작성법 알려줘
  (... 작업 ...)
> /sessions                 # 세션 목록 보기
  #1  "BXM Bean 작성"       2026-04-09 16:41  (12 messages)
  #2  "Tailwind 카드 컴포넌트" 2026-04-09 15:20  (8 messages)
> /resume 1                 # 이전 세션 이어가기
> /sessions archive 2       # 오래된 세션 아카이브
```

**구현 상세**:

```
internal/storage/
├── db.go          # SQLite 초기화, 마이그레이션
├── session.go     # sessions 테이블 CRUD
└── message.go     # messages 테이블 CRUD
```

```sql
-- ~/.tgc/sessions.db 스키마
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,              -- UUID
    title TEXT,                       -- 자동 생성 (첫 메시지 요약)
    mode INTEGER DEFAULT 0,           -- 0=super, 1=dev, 2=plan
    model_id TEXT,                    -- 사용 모델
    total_tokens INTEGER DEFAULT 0,   -- 누적 토큰
    message_count INTEGER DEFAULT 0,  -- 메시지 수
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    archived INTEGER DEFAULT 0        -- 아카이브 여부
);

CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    session_id TEXT REFERENCES sessions(id) ON DELETE CASCADE,
    role TEXT NOT NULL,                -- system/user/assistant/tool
    content TEXT NOT NULL,             -- JSON blob (유연한 스키마)
    tool_call_id TEXT,                 -- tool 응답인 경우
    tokens INTEGER DEFAULT 0,         -- 이 메시지의 토큰 수
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_session ON messages(session_id, created_at);
```

**SQLite 설정 (WAL 모드)**:
```go
// OpenCode의 db.ts:90-95 참고
pragma journal_mode = WAL;       // 읽기-쓰기 동시성
pragma synchronous = NORMAL;     // 안전성-성능 균형
pragma busy_timeout = 5000;      // 5초 대기 (락 충돌 시)
pragma foreign_keys = ON;
```

**왜 `modernc.org/sqlite`인가**:
| 비교 | `modernc.org/sqlite` | `mattn/go-sqlite3` |
|------|---------------------|-------------------|
| CGo 필요 | **불필요** (pure Go) | 필요 (C 컴파일러) |
| 단일 바이너리 | **유지** | 깨짐 (동적 링킹) |
| 크로스 컴파일 | **그대로** (`GOOS=windows` OK) | 복잡 (cross-C-compiler 필요) |
| 폐쇄망 빌드 | **`go mod vendor` 후 오프라인 OK** | C 툴체인도 필요 |
| 속도 | 약간 느림 (~10%) | 네이티브 C |
| 바이너리 증가 | ~5MB | ~3MB |

**app.go 변경점** (최소):
```go
// 현재: history []openai.ChatCompletionMessage  (인메모리)
// 변경: storage.Session 사용 (DB 백킹)

// NewModel()에서:
db, _ := storage.Open(filepath.Join(homeDir, ".tgc", "sessions.db"))
session := db.CreateSession(mode)

// sendMessage()에서:
db.AddMessage(session.ID, "user", userInput)
// ... LLM 응답 후 ...
db.AddMessage(session.ID, "assistant", response)
```

**참고 소스**: `opencode/packages/opencode/src/session/index.ts:66-180`

---

#### 기능 1-B: 실제 토큰 카운팅 + 컨텍스트 압축

**문제**: 현재 `tokenCount++`는 SSE 청크 수를 세는 것이지 실제 토큰이 아님. 긴 대화에서 컨텍스트 윈도우 초과 시 API 에러로 세션 크래시.

**해결**: API 응답의 `usage` 필드에서 실제 토큰 추출 + 한도 접근 시 자동 압축.

**사용자 경험**:
```
┌────────────────────────────────────────────────────────────────────┐
│ 슈퍼택가이  GPT-OSS 120B  ■■■■■□□□ 45,231/128,000 tok (35%)      │
└────────────────────────────────────────────────────────────────────┘
              ↑ 상태바에 실제 토큰 표시

# 80% 도달 시 자동 메시지:
[SYSTEM] 컨텍스트 80% 사용. 이전 도구 출력을 압축합니다...
[SYSTEM] 컨텍스트 압축 완료: 102,400 → 61,200 토큰 (40% 절감)
```

**2단계 압축 전략** (OpenCode compaction.ts 참고):

```
Phase 1 (저비용): 오래된 tool output 제거
  messages[5].content = "[tool output truncated — 원본 file_read /src/app.go, 2340 chars]"
  → 보통 30-50% 토큰 절감

Phase 2 (고비용): LLM 요약 요청 (동일 엔드포인트)
  "아래 대화를 요약해줘: Goal, 완료된 작업, 남은 작업, 핵심 파일"
  → 추가 20-30% 절감, API 호출 1회 소요
```

**임계값** (OpenCode 기준):
```go
const (
    PRUNE_TRIGGER = 0.80  // 80% 도달 시 Phase 1
    PRUNE_URGENT  = 0.90  // 90% 도달 시 Phase 2
    PRUNE_TARGET  = 0.60  // 압축 후 목표: 60%
)
```

**참고 소스**: `opencode/packages/opencode/src/session/compaction.ts:10-130`

---

### Phase 2: 퍼지 file_edit + Git 통합 + Diff 뷰

> **목표 버전**: v0.7.0
> **기간**: 2-3주
> **단일 바이너리**: O (pure Go 라이브러리만)
> **바이너리 증가**: +0.5MB
> **런타임 데이터**: 없음

#### 기능 2-A: 퍼지 매칭 file_edit

**문제**: 현재 `file_edit`은 `strings.Replace(content, oldText, newText, 1)` — 정확일치만. LLM이 공백/들여쓰기를 약간 다르게 출력하면 실패 → 재시도 → 폐쇄망의 느린 네트워크에서 시간 낭비.

**해결**: OpenCode처럼 4단계 fallback replacer. 정확매칭 실패 시 점진적으로 유연한 매칭 시도.

**매칭 단계**:
```
1. ExactReplacer         ← 현재와 동일 (정확일치)
   실패 시 ↓
2. LineTrimmedReplacer   ← 각 줄 앞뒤 공백 제거 후 매칭
   실패 시 ↓
3. IndentFlexReplacer    ← 들여쓰기 레벨 무시, 대상 indent 유지
   실패 시 ↓
4. FuzzyReplacer         ← Levenshtein 유사도 85%+ 매칭
   실패 시 → 에러 반환 (기존과 동일)
```

**예시**:
```
LLM이 보낸 old_text:
  "func main() {\n    fmt.Println(\"hello\")\n}"

실제 파일:
  "func main() {\n\tfmt.Println(\"hello\")\n}"
  (탭 vs 스페이스 차이)

→ Step 1 (Exact): 실패
→ Step 2 (LineTrimmed): 성공! 들여쓰기 무시하고 매칭
```

**의존성**: `github.com/agnivade/levenshtein` (pure Go, 파일 1개, ~3KB)

**안전장치**: 퍼지 매칭 시 `[FUZZY MATCH]` 로그 남김. 유사도가 85% 미만이면 거부.

**참고 소스**: `opencode/packages/opencode/src/tool/edit.ts` (9단계 replacer)

---

#### 기능 2-B: Git 구조화 통합

**문제**: 현재 git은 `shell_exec("git status")` 수동 실행만 가능. 시스템 프롬프트에 git 상태가 구조화되어 들어가지 않음.

**해결**: `internal/git/` 패키지로 구조화 + 도구 2개 추가 + 시스템 프롬프트 자동 주입.

**새 도구**:
```
도구 8: git_diff   — 특정 파일이나 전체 변경사항 diff 보기 (읽기 전용, 안전)
도구 9: git_log    — 최근 커밋 히스토리 보기 (읽기 전용, 안전)
```

**시스템 프롬프트 자동 주입** (GatherFullContext() 강화):
```
## Git Context
- Branch: feature/bxm-integration
- Status: 3 modified, 1 untracked
- Last commit: "feat: add BXM bean template" (2h ago)
```

**의존성**: 없음 (`os/exec` + git CLI). 폐쇄망에서 git은 거의 항상 설치되어 있음.

**참고 소스**: `opencode/packages/opencode/src/git/index.ts`

---

#### 기능 2-C: Diff 뷰

**문제**: `file_edit` 실행 시 "수정 완료"만 표시. 뭐가 바뀌었는지 확인 불가.

**해결**: 편집 전후 unified diff를 빨강(삭제)/초록(추가) 컬러로 표시.

```
[file_edit] /src/main.go
  - func main() {
  -     fmt.Println("hello")
  + func main() {
  +     fmt.Println("hello, world!")
      }
```

**의존성**: `github.com/sergi/go-diff` (pure Go, 무의존)
**바이너리 증가**: ~200KB

---

### Phase 3: 파일 스냅샷/Undo + 멀티 모델 설정

> **목표 버전**: v0.8.0
> **기간**: 2-3주
> **단일 바이너리**: O (새 의존성 없음)
> **바이너리 증가**: ±0
> **런타임 데이터**: `~/.tgc/snapshots/` (자동 생성, 7일 후 자동 삭제)

#### 기능 3-A: 파일 스냅샷 / Undo

**문제**: AI가 `file_edit`/`file_write`로 파일을 망가뜨리면 복구 방법 없음. git commit 하지 않았다면 끝.

**해결**: 모든 파일 수정 전에 원본을 자동 백업. `/undo` 명령으로 복원.

**사용자 경험**:
```
> React 컴포넌트 리팩토링해줘

  [file_edit] /src/App.tsx — 수정 완료 (원본 스냅샷 저장)
  [file_edit] /src/utils.ts — 수정 완료 (원본 스냅샷 저장)

> /undo                      # 마지막 편집 되돌리기
  ✓ /src/utils.ts 복원됨

> /undo all                  # 이 세션의 모든 편집 되돌리기
  ✓ /src/App.tsx 복원됨
  ✓ /src/utils.ts 복원됨

> /diff                      # 현재 세션에서 변경된 파일 보기
  Modified: /src/App.tsx (+15 -8)
  Modified: /src/utils.ts (+3 -1)
```

**저장 구조**:
```
~/.tgc/snapshots/
└── {session_id}/
    └── 2026-04-09T16:41:17/
        └── src/App.tsx       ← 수정 전 원본 복사본
```

**자동 정리**: 7일 이상된 스냅샷 자동 삭제 (앱 시작 시 체크)
**최대 크기**: 스냅샷 총 500MB 초과 시 오래된 것부터 삭제

**참고 소스**: `opencode/packages/opencode/src/snapshot/index.ts`

---

#### 기능 3-B: 멀티 모델 / 엔드포인트 설정

**문제**: 현재 2개 모델이 `models.go`에 하드코딩. 폐쇄망에서는 환경마다 다른 모델/엔드포인트 사용.

**해결**: `config.yaml`에서 모드별 모델/엔드포인트/컨텍스트 윈도우 설정.

```yaml
# ~/.tgc/config.yaml
models:
  super:
    id: "qwen3-235b"
    display_name: "Qwen3-235B (슈퍼)"
    endpoint: "http://gpu-server-1:8080/v1"
    api_key: "sk-internal-..."
    context_window: 128000
  dev:
    id: "deepseek-coder-v2"
    display_name: "DeepSeek-Coder-V2"
    endpoint: "http://gpu-server-2:8080/v1"
    api_key: "sk-internal-..."
    context_window: 32000
  plan:
    id: "qwen3-30b"
    display_name: "Qwen3-30B (플랜)"
    endpoint: "http://gpu-server-1:8080/v1"
    context_window: 64000
```

**핵심**: 각 모드가 다른 물리 서버/모델을 사용 가능. 컨텍스트 압축(Phase 1)이 `context_window` 값을 참조.

---

### Phase 4: 브라우저 컴패니언

> **목표 버전**: v0.9.0
> **기간**: 1-2주
> **단일 바이너리**: O (`//go:embed web/*`로 HTML/CSS/JS 내장)
> **바이너리 증가**: +1.5MB (Mermaid ~800KB + highlight.js ~50KB + diff2html ~40KB)
> **런타임 데이터**: 없음 (메모리 내 이벤트 버퍼만)
> **외부 의존성**: 없음 (`net/http` 표준 라이브러리)

**문제**: 터미널에서 플랜, 아키텍처 다이어그램, 코드 diff를 보기 어려움. 복잡한 작업 진행 상황 파악이 힘듦.

**해결**: `/companion` 명령으로 로컬 HTTP 서버 시작 → 브라우저에서 실시간 대시보드 표시.

**사용자 경험**:
```
> /companion
  Companion: http://localhost:8787 (브라우저 자동 열림)

> BXM Service 계층 설계해줘

  [터미널]                          [브라우저 — localhost:8787]
  ┌──────────────────────┐          ┌─────────────────────────────┐
  │ 슈퍼택가이 GPT-OSS   │          │ 📋 Plan                     │
  │                      │          │ ┌─────────────────────────┐ │
  │ BXM Service 계층은   │    SSE   │ │ [Mermaid 다이어그램]     │ │
  │ IO → DBIO → Bean →  │ ──────►  │ │  IO ──► DBIO ──► Bean   │ │
  │ Service 순서로...    │          │ │    ──► Service           │ │
  │                      │          │ └─────────────────────────┘ │
  │ [file_edit] Bean.java│          │                             │
  │  - old code          │          │ 📝 Diff View                │
  │  + new code          │          │ - old code (red)            │
  │                      │          │ + new code (green)          │
  └──────────────────────┘          └─────────────────────────────┘
```

**아키텍처**:
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

**왜 SSE (WebSocket 아님)?**:
- TUI→브라우저 단방향 흐름 → SSE로 충분
- `net/http` 표준 라이브러리만 사용 (외부 의존성 0)
- gstack이 이 패턴으로 실제 프로덕션 검증 완료
- 브라우저의 `EventSource` API가 자동 재연결 지원

**파일 구조**:
```
internal/companion/
├── hub.go          # Ring buffer(500개) + pub/sub (~80줄)
├── server.go       # net/http 서버 + SSE endpoint (~120줄)
├── events.go       # 이벤트 타입 정의 (~40줄)
├── browser.go      # OS별 브라우저 열기 (~25줄)
└── embed.go        # //go:embed web/* (~5줄)

web/                # 빌드 스텝 없음, embed로 바이너리 내장
├── index.html      # SPA 대시보드
├── style.css       # 다크 테마 (TUI 팔레트와 동일)
├── app.js          # SSE 클라이언트 + DOM 업데이트
└── vendor/         # 오프라인용 번들
    ├── mermaid.min.js      # 다이어그램 (~800KB)
    ├── highlight.min.js    # 코드 하이라이팅 (~50KB)
    └── diff2html.min.js    # Diff 렌더링 (~40KB)
```

**핵심 제약**: `DEBUG_TRANSPORT_FREEZE` 준수 — 컴패니언 서버는 반드시 별도 고루틴에서 `net/http` 직접 사용. LLM HTTP 클라이언트와 절대 공유하지 않음.

**참고 소스**: `gstack/browse/src/activity.ts:1-210`, `gstack/browse/src/server.ts:1614-1676`

---

### Phase 5: LSP 통합 (Language Server Protocol)

> **목표 버전**: v1.0.0
> **기간**: 4-6주
> **단일 바이너리**: O (JSON-RPC 코드만 추가)
> **바이너리 증가**: +1MB
> **런타임 의존**: 언어 서버 바이너리 (`gopls`, `typescript-language-server` 등)
> **외부 의존**: 언어 서버가 설치되어 있어야 함

**문제**: AI가 코드를 수정할 때 타입 에러, 미사용 변수, import 누락 등을 감지하지 못함. 현재는 `shell_exec("go build")` 후에야 에러 확인 가능.

**해결**: LSP 클라이언트를 내장하여 실시간 코드 진단 + 자동완성 정보를 AI에 제공.

**새 도구**:
```
도구 10: lsp_diagnostics  — 현재 파일의 에러/경고 목록
도구 11: lsp_hover         — 심볼 위에 타입/문서 정보
```

**단계적 구현**:
```
Phase 5-1 (2주): gopls only
  └─ textDocument/diagnostics + textDocument/hover

Phase 5-2 (2주): definition + references
  └─ textDocument/definition + textDocument/references

Phase 5-3 (2주): 멀티 서버 (config.yaml에서 설정)
  └─ typescript-language-server, pyright 등
```

**폐쇄망 제약**: 언어 서버 자동 다운로드 불가 → 사전 설치 필요. `config.yaml`에서 경로 지정:
```yaml
lsp:
  go:
    command: "/usr/local/bin/gopls"
    args: ["serve"]
  typescript:
    command: "/usr/local/bin/typescript-language-server"
    args: ["--stdio"]
```

**참고 소스**: `opencode/packages/opencode/src/lsp/index.ts`, `opencode/packages/opencode/src/lsp/server.ts`

---

### Phase 6: Vim 키바인딩

> **목표 버전**: v1.1.0
> **기간**: 1-2주
> **단일 바이너리**: O (코드만 추가)
> **바이너리 증가**: ±0
> **런타임 데이터**: 없음

**문제**: 텍스트 입력 영역이 기본 Emacs 바인딩만 지원 (Ctrl+A/E/K). Vim 사용자에게 불편.

**해결**: Bubble Tea textarea에 Vim 모드 추가 (Normal/Insert/Visual).

```
[NORMAL] hjkl 이동, dd 줄삭제, yy 복사, p 붙여넣기, w/b 단어이동
[INSERT] i/a/o로 진입, ESC로 Normal 복귀
[VISUAL] v로 진입, 범위 선택 후 d/y
```

**설정**:
```yaml
# ~/.tgc/config.yaml
editor:
  mode: "vim"   # "default" | "vim"
```

**참고 소스**: `opencode/packages/opencode/src/cli/cmd/tui/component/textarea-keybindings.ts`

---

## 구현 로드맵 전체

```
v0.5.0 (현재)
  └─ Embedded Knowledge Store (38문서, BXM Tier 0)

v0.6.0 — Phase 1 (3-4주)
  ├─ SQLite 영구 세션 (/sessions, /resume)
  └─ 실제 토큰 카운팅 + 컨텍스트 자동 압축

v0.7.0 — Phase 2 (2-3주)
  ├─ 퍼지 file_edit (4단계 fallback)
  ├─ Git 구조화 통합 (git_diff, git_log 도구)
  └─ Diff 뷰 (빨강/초록 컬러)

v0.8.0 — Phase 3 (2-3주)
  ├─ 파일 스냅샷 / Undo (/undo, /diff)
  └─ 멀티 모델 / 엔드포인트 설정

v0.9.0 — Phase 4 (1-2주)
  └─ 브라우저 컴패니언 (/companion, SSE, Mermaid)

v1.0.0 — Phase 5 (4-6주)
  └─ LSP 통합 (gopls → 멀티 서버)

v1.1.0 — Phase 6 (1-2주)
  └─ Vim 키바인딩
```

**총 예상 기간**: 14-22주 (전체), 첫 3개 Phase가 핵심 (~8주)

---

## 단일 바이너리 호환성 총정리

| Phase | 기능 | 새 Go 의존성 | CGo 필요 | 바이너리 증가 | 런타임 파일 | 단일 빌드 |
|-------|------|-------------|---------|-------------|------------|---------|
| 1 | SQLite 세션 | `modernc.org/sqlite` | **아니오** | +5MB | `~/.tgc/sessions.db` | **O** |
| 1 | 토큰 카운팅 | 없음 | 아니오 | ±0 | 없음 | **O** |
| 2 | 퍼지 edit | `agnivade/levenshtein` | **아니오** | +3KB | 없음 | **O** |
| 2 | Git 통합 | 없음 | 아니오 | ±0 | 없음 (git CLI) | **O** |
| 2 | Diff 뷰 | `sergi/go-diff` | **아니오** | +200KB | 없음 | **O** |
| 3 | 스냅샷/Undo | 없음 | 아니오 | ±0 | `~/.tgc/snapshots/` | **O** |
| 3 | 멀티 모델 | 없음 | 아니오 | ±0 | 없음 | **O** |
| 4 | 브라우저 | 없음 (`net/http` 표준) | **아니오** | +1.5MB | 없음 | **O** |
| 5 | LSP | `sourcegraph/jsonrpc2` | **아니오** | +1MB | 없음 (서버 별도) | **O** |
| 6 | Vim 키 | 없음 | 아니오 | ±0 | 없음 | **O** |

**결론: 모든 기능이 CGo 없이, 단일 `go build`로, 크로스 컴파일 가능하게 구현됩니다.**

빌드 명령은 기존과 동일:
```bash
make build              # 현재 OS 바이너리
make build-all          # 5개 플랫폼 크로스 컴파일
```

---

## 핵심 제약 사항 (반드시 준수)

1. **DEBUG_TRANSPORT_FREEZE**: HTTP transport 래핑 금지. 컴패니언 서버/LSP 클라이언트는 별도 고루틴에서 `net/http` 직접 사용. LLM 스트리밍과 절대 공유 불가.
2. **구 app.go 기반 유지**: `app.go` 변경은 최소 단위로만 (필드 추가 + 메서드 호출). 대규모 리팩토링 금지.
3. **단일 바이너리**: 모든 에셋은 `//go:embed`로 내장. 외부 파일 의존 금지.
4. **폐쇄망 호환**: 빌드 타임에 네트워크 사용 (`go mod download`), 런타임에 네트워크 의존 금지 (LLM 엔드포인트 제외).
5. **pure Go only**: CGo 필요 라이브러리 사용 금지. 크로스 컴파일 깨짐.

---

## 참고 소스 파일

### OpenCode (`/Users/kimjiwon/Desktop/kimjiwon/opencode/`)
- `packages/opencode/src/session/index.ts:66-180` — 세션 CRUD
- `packages/opencode/src/session/compaction.ts:10-130` — 컨텍스트 압축 (PRUNE_MINIMUM, 2단계)
- `packages/opencode/src/session/session.sql.ts` — SQLite 스키마
- `packages/opencode/src/storage/db.ts:90-95` — SQLite WAL pragma
- `packages/opencode/src/tool/edit.ts` — 9단계 fallback replacer
- `packages/opencode/src/git/index.ts` — 구조화 git 연산 (`--no-optional-locks`)
- `packages/opencode/src/snapshot/index.ts` — Shadow git 스냅샷
- `packages/opencode/src/lsp/index.ts` — LSP 클라이언트 (25+ 서버)
- `packages/opencode/src/provider/provider.ts` — 20+ LLM 프로바이더

### gstack (`/Users/kimjiwon/Desktop/kimjiwon/gstack/`)
- `browse/src/activity.ts:1-210` — Ring buffer + pub/sub + SSE 핵심 패턴
- `browse/src/server.ts:1614-1676` — SSE endpoint 구현
- `extension/sidepanel.js:1-86` — SSE 클라이언트 (EventSource + 자동 재연결)
- `ARCHITECTURE.md` — 데몬 모델 + CLI 설계 원칙
- `BROWSER.md:25-48` — CLI↔데몬 HTTP 아키텍처 다이어그램

### 택갈이코드 (`/Users/kimjiwon/Desktop/kimjiwon/택갈이코드/`)
- `internal/app/app.go:42-76` — Model 구조체 (통합 포인트)
- `internal/app/app.go:313-407` — 스트림/도구 핸들러 (이벤트 Emit 포인트)
- `internal/app/app.go:485-517` — handleSlashCommand (슬래시 명령 추가점)
- `internal/app/app.go:665-696` — sendMessage (사용자 메시지 처리)
- `internal/tools/registry.go` — 도구 실행 (퍼지 edit 수정점)
- `internal/ui/styles.go:9-22` — 컬러 팔레트 (브라우저 UI 동기화)
- `internal/config/config.go` — 설정 구조 (멀티 모델 확장점)
- `internal/knowledge/` — 지식 스토어 (v0.5.0 완성)
- `docs/DEBUG_TRANSPORT_FREEZE.md` — HTTP 래핑 금지 제약
