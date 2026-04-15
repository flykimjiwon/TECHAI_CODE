# hanimo → 택가이코드 포팅 가이드

> **작성일**: 2026-04-09
> **목적**: hanimo(오픈소스)에서 택가이코드(폐쇄망 전용)로 이식할 기능 식별 및 구현 전략
> **추가 참고**: PageIndex (VectifyAI) — 계층형 트리 인덱스 기반 Vectorless RAG
> **핵심 원칙**: 외부 인터넷 없이 동작하는 기능만 포팅. 단일 바이너리 유지.
> **다음 구현 시 반드시 이 문서를 참조할 것**

---

## 프로젝트 관계

| 구분 | hanimo | 택가이코드 |
|------|--------|-----------|
| **용도** | 세상에 공개할 오픈소스 | 회사 내부 폐쇄망 전용 |
| **네트워크** | 클라우드 API (OpenAI, Anthropic, Google 등) | 사내 LLM 서버만 (외부 인터넷 없음) |
| **배포** | `go install` / GitHub Release | USB 복사 단일 바이너리 |
| **스택** | Go + Bubble Tea v2 + go-openai | 동일 |
| **프로바이더** | 14+ (OpenAI, Anthropic, Google, Ollama 등) | Ollama + OpenAI-compatible 엔드포인트만 |

**핵심**: hanimo에 이미 구현된 기능 중 **오프라인 호환** 가능한 것만 택가이코드로 이식.

---

## 포팅 대상 분류

### 반드시 포팅 (10개 기능)

폐쇄망에서 즉시 효과가 있는 기능들. 외부 의존성 없음.

| # | 기능 | hanimo 소스 | 줄 수 | 외부 의존 | 바이너리 증가 |
|---|------|------------|-------|----------|-------------|
| 1 | 컨텍스트 압축 | `internal/llm/compaction.go` | 137 | 없음 | ±0 |
| 2 | Git 도구 | `internal/tools/git.go` | 52 | git CLI | ±0 |
| 3 | 해시 앵커 편집 | `internal/tools/hashline.go` | 116 | 없음 | ±0 |
| 4 | 다국어 진단 | `internal/tools/diagnostics.go` | 272 | 린터 CLI | ±0 |
| 5 | 자율 모드 | `internal/agents/auto.go` | 26 | 없음 | ±0 |
| 6 | 모델 능력 레지스트리 | `internal/llm/capabilities.go` | 99 | 없음 | ±0 |
| 7 | 커맨드 팔레트 | `internal/ui/palette.go` | 154 | lipgloss | ±0 |
| 8 | 한/영 전환 | `internal/ui/i18n.go` | 133 | 없음 | ±0 |
| 9 | 메뉴 오버레이 | `internal/ui/menu.go` | 98 | lipgloss | ±0 |
| 10 | 테마 시스템 | `internal/ui/styles.go` | 192 | lipgloss | ±0 |

**소계**: 1,279줄, 바이너리 증가 ±0

---

### 적응 후 포팅 (3개 기능)

구조 변경 또는 폐쇄망 제약에 맞는 수정 필요.

| # | 기능 | hanimo 소스 | 줄 수 | 적응 사항 |
|---|------|------------|-------|----------|
| 11 | SQLite 세션 | `internal/session/` (4파일) | 466 | usage.go의 클라우드 모델 가격표 제거, 사내 모델만 |
| 12 | 프로바이더 레지스트리 | `internal/llm/providers/registry.go` | 110 | Anthropic/Google 제거, Ollama + OpenAI-compat만 |
| 13 | 멀티 모델 설정 | `internal/llm/providers/openai_compat.go` | 202 | 클라우드 URL 제거, 사내 엔드포인트 기본값 |

**소계**: 778줄 → 적응 후 ~500줄 예상

---

### 포팅 제외 (4개 기능)

폐쇄망에서 무의미하거나 불필요한 기능.

| 기능 | 제외 이유 |
|------|----------|
| MCP 서버 연동 | 외부 서비스 접근 불가 |
| 클라우드 프로바이더 (Anthropic, Google, OpenRouter 등) | 인터넷 없음 |
| 비용 추적 (USD 과금) | 사내 서버 = 무료 |
| MCP 서버 테이블 (DB) | MCP 미사용 |

---

## Phase 1: 핵심 기능 포팅 (~14시간)

> 컨텍스트 압축 + Git 도구 + 해시 앵커 편집 + 진단 + 자율 모드 + 모델 능력 레지스트리 + **계층형 지식 인덱스**
> **바이너리 증가**: +~50KB (tree-index.json embed)
> **새 의존성**: 없음

### 1-1. 컨텍스트 압축 (`internal/llm/compaction.go`, 137줄)

**현재 문제**: 택가이코드는 tool loop 20회 하드캡만 있음. 긴 대화에서 컨텍스트 윈도우 초과 시 API 에러.

**hanimo 구현**: 3단계 압축

```
Stage 1 (Snip): 40+ 메시지일 때, 최근 10개 제외한 오래된 tool output을
               200자 이상이면 "[snipped: N lines]"로 대체
               → 보통 30-50% 절감

Stage 2 (Micro): 4000자 초과 메시지를
                head(2000) + "[...truncated...]" + tail(2000)로 절단

Stage 3 (LLM): 여전히 maxTokens 초과 시
              시스템 프롬프트 + 최근 10개 보존, 중간을 LLM에 요약 요청
              → 추가 20-30% 절감
```

**토큰 추정**: `estimateTokens()` = `총문자수 / 4` (외부 토크나이저 불필요)

**포팅 방법**: `compaction.go`를 `택가이코드/internal/llm/`에 복사 후 `openai.ChatCompletionMessage` 타입을 택가이코드의 메시지 타입에 맞춤. `app.go`의 `sendMessage()`에서 전송 전 `Compact()` 호출 추가.

**상태바 통합**: `RenderStatusBar()`에 토큰 퍼센트 바 추가

```
슈퍼택가이  GPT-OSS 120B  ■■■■■□□□ 45,231/128,000 tok (35%)
```

**임계값**:
```go
const (
    PRUNE_TRIGGER = 0.80  // 80% → Stage 1+2
    PRUNE_URGENT  = 0.90  // 90% → Stage 3 (LLM 요약)
    PRUNE_TARGET  = 0.60  // 압축 후 목표
)
```

---

### 1-2. Git 도구 (`internal/tools/git.go`, 52줄)

**현재 문제**: `shell_exec("git status")` 수동 실행만 가능. 구조화된 정보 없음.

**hanimo 구현**: 5개 함수 래퍼 (10초 타임아웃)

```go
GitStatus(path) string  // git status --short
GitDiff(path)   string  // git diff
GitLog(path, n) string  // git log --oneline -n
GitCommit(path, msg)    // git commit -m
GitBranch(path) string  // git rev-parse --abbrev-ref HEAD
```

**포팅 방법**:
1. `git.go` 복사 → `택가이코드/internal/tools/`
2. `registry.go`에 `git_status`, `git_diff`, `git_log` 도구 3개 등록 (읽기 전용, 안전)
3. `GatherFullContext()`에 Git 상태 자동 주입:

```
## Git Context
- Branch: feature/bxm-bean
- Status: 3 modified, 1 untracked
- Last commit: "feat: add bean template" (2h ago)
```

**폐쇄망 호환**: git CLI는 사내 환경에 거의 항상 설치됨. 외부 네트워크 불필요.

---

### 1-3. 해시 앵커 편집 (`internal/tools/hashline.go`, 116줄)

**현재 문제**: `file_edit`이 정확일치(`strings.Replace`)만 지원. 소형 로컬 모델이 공백/들여쓰기를 조금이라도 틀리면 편집 실패 → 재시도 낭비.

**hanimo 구현**: 줄마다 MD5 해시 4자리 앵커 부여

```
읽기 출력:
1#a3f1| function hello() {
2#e4d9|   return 42
3#b2c1| }

편집 요청:
HashlineEdit(path, "2#e4d9", "2#e4d9", "  return 99")

→ 해시 일치 확인 후 해당 줄만 교체
→ 해시 불일치 시 에러 + 최신 해시 안내
```

**왜 중요한가**: 폐쇄망 모델(7B-30B급)은 정확한 문자열 복사 능력이 약함. 해시 앵커는 "줄 번호 + 무결성 검증"을 결합하여 **스테일 에디트(stale edit) 방지**.

**포팅 방법**:
1. `hashline.go` 복사 → `택가이코드/internal/tools/`
2. 새 도구 등록: `hashline_read`, `hashline_edit`
3. 기존 `file_edit`은 유지 (대형 모델용), `hashline_edit`은 소형 모델 자동 선택

---

### 1-4. 다국어 진단 (`internal/tools/diagnostics.go`, 272줄)

**현재 문제**: AI가 코드 수정 후 에러 확인은 `shell_exec("go build")`만. 구조화된 진단 없음.

**hanimo 구현**: 4개 언어 자동 감지 + 린팅

| 언어 | 린터 | 감지 파일 | 타임아웃 |
|------|------|----------|---------|
| Go | `go vet ./...` | `go.mod` | 30초 |
| TypeScript | `npx tsc --noEmit` | `tsconfig.json` | 60초 |
| JavaScript | `npx eslint --format compact` | `package.json` | 30초 |
| Python | `ruff check` | `pyproject.toml` | 30초 |

**진단 구조체**:
```go
type Diagnostic struct {
    File, Message, Source, Severity string
    Line, Column int
}
```

**포팅 방법**:
1. `diagnostics.go` 복사 → `택가이코드/internal/tools/`
2. 새 도구: `diagnostics` (프로젝트 루트 자동 감지)
3. 폐쇄망 적응: `npx` 대신 전역 설치된 린터 경로 사용 가능하도록 config 옵션

**폐쇄망 고려**: Go/Python 린터는 보통 설치됨. JS/TS는 사내 개발 서버에 따라 다름. 린터 미설치 시 graceful skip.

---

### 1-5. 자율 모드 (`internal/agents/auto.go`, 26줄)

**현재 문제**: 매 도구 실행마다 사용자 입력 대기. 반복 작업(파일 10개 수정 등)에서 비효율.

**hanimo 구현**: 시스템 프롬프트에 자율 모드 마커 주입

```go
const (
    MaxAutoIterations  = 20
    AutoCompleteMarker = "[AUTO_COMPLETE]"
    AutoPauseMarker    = "[AUTO_PAUSE]"
)
```

LLM이 `[AUTO_COMPLETE]` 출력하면 자동 종료, `[AUTO_PAUSE]` 출력하면 사용자 입력 요청.

**포팅 방법**:
1. `auto.go` 복사 → `택가이코드/internal/agents/`
2. `/auto` 슬래시 명령 추가 (handleSlashCommand)
3. `sendMessage()` 루프에 마커 감지 로직 삽입
4. 상태바에 `[AUTO]` 표시

---

### 1-6. 모델 능력 레지스트리 (`internal/llm/capabilities.go`, 99줄)

**현재 문제**: 모든 모델을 동일하게 취급. 소형 모델에게 복잡한 도구 체이닝을 시키면 실패.

**hanimo 구현**: 모델별 코딩 능력 + 역할 자동 배정

```go
type CodingTier int  // Strong, Moderate, Weak, None
type RoleType int    // Agent(전도구), Assistant(읽기만), Chat(도구없음)

// 17개 모델 매핑
var knownModels = map[string]ModelCapability{
    "qwen3-235b": {128000, CodingStrong, RoleAgent, true},
    "llama3.1:8b": {128000, CodingWeak, RoleAssistant, true},
    ...
}
```

**포팅 적응**: 클라우드 모델(GPT-4o, Claude 등) 제거, 사내 모델만:
```go
var knownModels = map[string]ModelCapability{
    "qwen3-235b":      {128000, CodingStrong, RoleAgent, true},
    "deepseek-coder":  {128000, CodingStrong, RoleAgent, true},
    "llama3.1:70b":    {128000, CodingModerate, RoleAgent, true},
    "codellama:13b":   {16000, CodingModerate, RoleAssistant, true},
    "gemma-4-31b-it":  {128000, CodingModerate, RoleAgent, true},
}
```

**연동**: `sendMessage()`에서 모델 능력에 따라 도구 목록 필터링. Weak 모델에는 `file_edit` 대신 `hashline_edit`만 제공.

---

### 1-7. 계층형 지식 인덱스 (PageIndex 개념 차용, ~200줄 신규)

> **영감**: [PageIndex](https://github.com/VectifyAI/PageIndex) — 벡터 DB 없이 LLM 추론으로 문서 검색
> **핵심 아이디어**: 현재 키워드 매칭을 **트리 인덱스 + 섹션 단위 검색**으로 업그레이드

**현재 문제**: v0.5.0 Knowledge Store는 키워드 매칭으로 **문서 전체**를 주입. "BXM Bean 작성법"을 물으면 BXM 전체 문서(~40KB)가 컨텍스트에 들어감. 정작 필요한 건 Bean 섹션 2KB뿐.

**PageIndex가 해결하는 방식**:
1. 문서의 `#` 헤딩을 파싱하여 목차(TOC) 트리 생성
2. 각 노드에 요약(summary) 부여
3. 검색 시 LLM이 트리를 탐색하여 **관련 섹션만** 반환

**택가이코드 적응**: PageIndex는 Python + 런타임 LLM 호출이지만, 택가이코드는 **빌드 타임에 트리 + 요약을 미리 생성**하여 JSON으로 임베딩.

**트리 노드 구조**:
```go
type KnowledgeNode struct {
    Title    string           `json:"title"`     // 섹션 제목
    NodeID   string           `json:"node_id"`   // "0001", "0002", ...
    LineNum  int              `json:"line_num"`   // 원본 MD 시작 줄
    Summary  string           `json:"summary"`   // 미리 생성된 1-2문장 요약
    Keywords []string         `json:"keywords"`  // 기존 키워드 (하위호환)
    Tier     int              `json:"tier"`       // 기존 Tier (0=BXM, 1=daily, ...)
    Children []KnowledgeNode  `json:"children,omitempty"`
}
```

**빌드 파이프라인** (빌드 스크립트로 1회 실행):
```
knowledge/docs/**/*.md
       ↓  (Go 스크립트: # 헤딩 파싱 → 트리 구축)
knowledge/tree-index.json   ← //go:embed로 바이너리 내장
       ↓
런타임: 트리 로드 → 키워드 매칭 + 트리 탐색 병행
```

**`tree-index.json` 예시**:
```json
{
  "docs": [
    {
      "doc_name": "bxm-overview.md",
      "doc_description": "BXM 프레임워크 전체 개요 — 아키텍처, 레이어 구조, 주요 개념",
      "tier": 0,
      "structure": [
        {
          "title": "BXM 개요",
          "node_id": "0001",
          "line_num": 1,
          "summary": "BXM 프레임워크의 목적과 전체 구조 설명",
          "keywords": ["bxm", "프레임워크", "아키텍처"],
          "children": [
            {
              "title": "Bean 구조",
              "node_id": "0002",
              "line_num": 45,
              "summary": "IO/DBIO/Bean 3계층과 Bean 작성 규칙",
              "keywords": ["bean", "io", "dbio"]
            },
            {
              "title": "Service 계층",
              "node_id": "0003",
              "line_num": 120,
              "summary": "비즈니스 로직 서비스 레이어 구현 패턴",
              "keywords": ["service", "비즈니스로직"]
            }
          ]
        }
      ]
    }
  ]
}
```

**검색 전략 (2단계 하이브리드)**:

```
쿼리: "BXM Bean 작성법"

1단계 (기존 키워드 매칭 — 빠름, 0ms):
   키워드 추출: ["bxm", "bean"]
   → 매칭 문서: bxm-overview.md, bxm-bean.md

2단계 (트리 탐색 — 정밀):
   매칭된 문서의 트리에서 키워드가 포함된 노드만 선택
   → bxm-overview.md → "Bean 구조" 섹션 (line 45-119)
   → bxm-bean.md → 전체 (단일 주제 문서)

결과: 전체 문서 40KB 대신 관련 섹션 ~5KB만 주입
```

**구현 파일**:
```
internal/knowledge/
├── store.go      (기존) + TreeIndex 필드 추가
├── tree.go       (신규, ~120줄) — 트리 로딩, 노드 검색
├── extractor.go  (기존) 그대로
└── injector.go   (기존) + 섹션 단위 주입 로직
```

**효과**:
- 컨텍스트 토큰 사용 **60-80% 절감** (문서 전체 → 섹션만)
- 컨텍스트 압축(1-1)과 시너지: 더 많은 문서를 동시 참조 가능
- 빌드 타임에 트리 생성하므로 **런타임 LLM 호출 불필요** (폐쇄망 친화)

**Obsidian 영감 — Frontmatter 메타데이터**:
```markdown
---
title: BXM Bean 구조
tier: 0
tags: [bxm, bean, io, dbio, java]
related: [[bxm-service]], [[bxm-select]]
---
# BXM Bean 구조
...
```

빌드 스크립트가 YAML frontmatter를 파싱하여 `tree-index.json`에 메타데이터 포함. `related` 필드로 문서 간 연결 그래프 구축.

---

## Phase 2: UI 강화 포팅 (~14시간)

> 커맨드 팔레트 + 한/영 전환 + 메뉴 오버레이 + 테마 시스템 + **Obsidian 스타일 지식 탐색**
> **바이너리 증가**: ±0
> **새 의존성**: 없음 (lipgloss 이미 사용 중)

### 2-1. 커맨드 팔레트 (`internal/ui/palette.go`, 154줄)

**현재 문제**: `/help`로만 명령 목록 확인. 명령어를 외워야 함.

**hanimo 구현**: `Ctrl+K`로 팔레트 열기 → 퍼지 검색 → Enter 실행

```
┌─────────────────────────────────┐
│ > 검색...                        │
├─────────────────────────────────┤
│ ▶ 세션 저장      /save          │
│   세션 불러오기   /load          │
│   모델 전환      /model         │
│   진단 실행      /diagnostics   │
│   테마 변경      /theme         │
│   언어 전환      /lang          │
│   도움말         /help          │
├─────────────────────────────────┤
│ ↑↓ 이동  Enter 선택  Esc 닫기   │
└─────────────────────────────────┘
```

**13개 기본 명령**: save, load, search, model, provider, usage, diagnostics, remember, memories, lang, config, theme, clear

**퍼지 필터**: 라벨/설명/액션 substring 매칭

**포팅 방법**: `palette.go` 복사 + `app.go`에 `Ctrl+K` 키바인딩 추가. 팔레트 아이템을 택가이코드의 슬래시 명령에 매핑.

---

### 2-2. 한/영 전환 (`internal/ui/i18n.go`, 133줄)

**현재 문제**: UI 문자열이 한국어 하드코딩. 외국인 개발자 사용 불가.

**hanimo 구현**: 27개 UI 문자열의 한/영 번역 + `/lang` 토글

```go
var KO = Strings{
    SendMessage: "메시지 입력...",
    ModeSuper:   "슈퍼택가이",
    ToolOn:      "도구 활성",
    ...
}
var EN = Strings{
    SendMessage: "Type a message...",
    ModeSuper:   "Super",
    ToolOn:      "Tools ON",
    ...
}

func T() *Strings { /* 현재 언어 반환 */ }
```

**포팅 방법**: `i18n.go` 복사. UI 렌더링 코드에서 하드코딩 문자열을 `ui.T().XXX`로 교체.

---

### 2-3. 메뉴 오버레이 (`internal/ui/menu.go`, 98줄)

**현재 문제**: Tab으로 모드 전환만 가능. 모델/프로바이더 변경은 슬래시 명령으로만.

**hanimo 구현**: `Esc` 또는 `Ctrl+M`으로 플로팅 메뉴

```
┌─ hanimo ────────────────────────┐
│ ▶ 모델 전환                      │
│   프로바이더 전환                  │
│   사용량 통계                     │
│   진단 실행                      │
│   설정 보기                      │
│   커맨드 팔레트 (Ctrl+K)         │
│   도움말                         │
├─────────────────────────────────┤
│ ↑↓ 이동  Enter 선택  Esc 닫기   │
└─────────────────────────────────┘
```

**서브메뉴**: "모델 전환" 선택 시 → 사용 가능 모델 목록 표시

**포팅 방법**: `menu.go` 복사 + `app.go`에 `Esc` 키 핸들링 추가. 서브메뉴 아이템을 택가이코드의 모드/모델에 매핑.

---

### 2-4. 테마 시스템 (`internal/ui/styles.go`, 192줄)

**현재 문제**: 블루 톤 단일 테마 하드코딩.

**hanimo 구현**: 5개 프리셋 테마 + `/theme` 전환

| 테마 | Primary | Accent | Background | 느낌 |
|------|---------|--------|------------|------|
| **honey** (기본) | `#F9E2AF` | `#CBA6F7` | `#1E1E2E` | 따뜻한 꿀색 |
| **ocean** | `#89B4FA` | `#74C7EC` | `#1E1E2E` | 시원한 파란색 |
| **dracula** | `#BD93F9` | `#FF79C6` | `#282A36` | 드라큘라 |
| **nord** | `#88C0D0` | `#81A1C1` | `#2E3440` | 노르딕 |
| **forest** | `#A6E3A1` | `#94E2D5` | `#1E1E2E` | 초록 숲 |

```go
func ApplyTheme(name string) bool  // 전역 컬러 변수 업데이트
```

**포팅 방법**: 택가이코드의 `styles.go` 컬러 상수를 변수로 변환 + `ApplyTheme()` 함수 추가. `config.yaml`에 `theme: "ocean"` 저장.

---

### 2-5. Obsidian 스타일 지식 탐색 (PageIndex + Obsidian 개념, ~150줄 신규)

> **영감**: Obsidian의 `[[wiki-link]]` + 그래프 뷰 + PageIndex의 트리 탐색
> **핵심**: 내장 지식을 **대화형으로 탐색**할 수 있는 UI

**현재 문제**: 지식 주입은 자동이지만, 사용자가 "어떤 지식이 들어있는지" 확인하거나 탐색할 방법이 없음.

**Obsidian에서 차용할 개념**:
1. **`[[wiki-link]]`**: 문서 간 상호 참조 (`[[bxm-service]]` → bxm-service.md로 연결)
2. **태그 기반 필터링**: `#bxm`, `#tailwind`, `#react` 태그로 관련 문서 묶기
3. **백링크**: "이 문서를 참조하는 다른 문서" 자동 표시
4. **그래프 뷰**: 문서 연결 관계를 시각적으로 (터미널에서는 텍스트 기반)

**`/knowledge` 명령 — 대화형 지식 탐색기**:

```
> /knowledge

┌─ 내장 지식 (38 문서) ────────────────────────────┐
│                                                   │
│ [Tier 0 — BXM]  13 문서                           │
│  ├─ bxm-overview     BXM 프레임워크 전체 개요      │
│  │   ├─ Bean 구조                                 │
│  │   ├─ Service 계층                              │
│  │   └─ Select/Paging                             │
│  ├─ bxm-bean         Bean 작성 가이드              │
│  ├─ bxm-service      Service 구현 패턴            │
│  └─ ...                                           │
│                                                   │
│ [Tier 1 — Daily]  8 문서                           │
│  ├─ tailwind-guide   Tailwind CSS 핵심 유틸리티    │
│  ├─ react-patterns   React 컴포넌트 패턴          │
│  └─ ...                                           │
│                                                   │
│ ↑↓ 이동  Enter 펼치기  / 검색  Esc 닫기           │
└───────────────────────────────────────────────────┘
```

**`/knowledge <query>` — 트리 검색 미리보기**:

```
> /knowledge bean 작성

  검색 결과 (3 섹션 매칭):

  [Tier 0] bxm-overview → Bean 구조 (line 45-119)
    "IO/DBIO/Bean 3계층과 Bean 작성 규칙"

  [Tier 0] bxm-bean → 전체 문서
    "BXM Bean 정의, 필드 매핑, 유효성 검사"

  [Tier 0] bxm-service → Bean 활용 (line 88-102)
    "Service에서 Bean을 사용한 데이터 처리 흐름"

  ← 다음 질문에 이 3개 섹션이 자동 주입됩니다
```

**문서 간 링크 그래프 (빌드 타임 생성)**:

```go
// knowledge/link-graph.json — //go:embed
{
  "bxm-overview": {
    "links_to": ["bxm-bean", "bxm-service", "bxm-select"],
    "tags": ["bxm", "framework", "java"]
  },
  "bxm-bean": {
    "links_to": ["bxm-overview", "bxm-dbio"],
    "linked_from": ["bxm-overview", "bxm-service"],
    "tags": ["bxm", "bean", "io"]
  }
}
```

**백링크 활용**: "bxm-bean" 섹션이 주입될 때, `linked_from`에 있는 `bxm-service`도 관련 문서로 LLM에 힌트 제공:
```
[KNOWLEDGE] bxm-bean → Bean 구조 (주입)
[KNOWLEDGE] 관련: bxm-service (백링크), bxm-overview (백링크)
```

**구현 파일**:
```
internal/knowledge/
├── tree.go       (Phase 1에서 생성) + 대화형 탐색 UI 연동
├── links.go      (신규, ~80줄) — [[wiki-link]] 파싱, 링크 그래프 로딩
└── browser.go    (신규, ~70줄) — /knowledge 명령 TUI 렌더링

knowledge/
├── tree-index.json   (Phase 1)
└── link-graph.json   (신규, 빌드 타임 생성)
```

**빌드 스크립트 확장** (`cmd/build-index/`):
1. MD 파일에서 `[[...]]` 패턴 파싱
2. frontmatter의 `related:` 필드 파싱
3. 양방향 링크 그래프 생성 → `link-graph.json`

---

## Phase 3: SQLite 세션 + 프로바이더 + LLM 트리 검색 (~12시간)

> SQLite 영구 세션 + 프로바이더 레지스트리 (폐쇄망 적응) + **LLM 추론 기반 지식 검색**
> **바이너리 증가**: +5MB (`modernc.org/sqlite`)
> **새 의존성**: `modernc.org/sqlite`, `github.com/google/uuid`

### 3-1. SQLite 세션 (`internal/session/`, 466줄 → ~400줄)

hanimo의 세션 레이어를 그대로 가져오되 다음을 제거/수정:

| 파일 | 줄 수 | 적응 사항 |
|------|-------|----------|
| `db.go` | 119 | `mcp_servers` 테이블 제거 |
| `store.go` | 177 | 그대로 복사 |
| `memory.go` | 87 | 그대로 복사 |
| `usage.go` | 83 | 클라우드 가격표 제거, 토큰 카운트만 유지 |

**스키마 (적응 후)**:
```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    name TEXT,
    project_dir TEXT,
    provider TEXT,
    model TEXT,
    mode INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT REFERENCES sessions(id),
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    tool_calls TEXT,
    tool_result TEXT,
    input_tokens INTEGER DEFAULT 0,
    output_tokens INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE memories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_dir TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    source TEXT DEFAULT 'user',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_dir, key)
);
```

**슬래시 명령**:
- `/sessions` — 세션 목록
- `/resume <id>` — 이전 세션 이어가기
- `/save [name]` — 세션 이름 지정
- `/search <query>` — 세션 내용 검색
- `/fork` — 현재 세션 분기
- `/remember <key> <value>` — 프로젝트 메모리 저장
- `/memories` — 저장된 메모리 목록

---

### 3-2. 프로바이더 레지스트리 (적응)

hanimo의 프로바이더 레지스트리에서 **Ollama + OpenAI-compatible**만 유지:

```go
// 택가이코드용 프로바이더 (폐쇄망)
func init() {
    Register("ollama", NewOllama)        // 로컬 Ollama 서버
    Register("openai-compat", NewOpenAICompat)  // 사내 vLLM/TGI 서버
}
```

**config.yaml 확장**:
```yaml
providers:
  default: "openai-compat"
  openai-compat:
    base_url: "http://gpu-server-1:8080/v1"
    api_key: "sk-internal-..."
  ollama:
    base_url: "http://localhost:11434"
```

---

### 3-3. LLM 추론 기반 트리 검색 (PageIndex 핵심 개념, ~100줄 신규)

> **영감**: PageIndex의 Reasoning-based Retrieval — LLM이 트리를 탐색하여 관련 노드 선택
> **Phase 1의 키워드 매칭과 병행**: 키워드로 빠르게 후보 선택 → LLM으로 정밀 필터링

**현재 (Phase 1 키워드 매칭)의 한계**:
- "BXM에서 페이징 처리할 때 성능 이슈 해결법" → 키워드 "bxm", "페이징" 매칭
- 하지만 정작 답은 `bxm-select.md`의 "대용량 조회 최적화" 섹션에 있음 → 키워드에 "성능"이 없으면 놓침

**PageIndex 방식 적용**: 트리 인덱스(제목+요약만, 텍스트 제외)를 LLM에 보내고 관련 노드를 추론하게 함.

**검색 프롬프트** (PageIndex `tree-search` 튜토리얼 참고):
```
당신은 내장 기술 문서의 검색 도우미입니다.
아래 트리 구조에서 사용자 질문에 답할 수 있는 노드를 찾으세요.
각 노드에는 제목(title)과 요약(summary)이 있습니다.

질문: {query}

문서 트리:
{tree_index_json_without_text}

다음 JSON으로 답하세요:
{
  "thinking": "어떤 노드가 관련있는지 추론 과정",
  "nodes": ["node_id_1", "node_id_2"]
}
```

**2단계 하이브리드 검색 (최종)**:

```
쿼리 입력
    ↓
[Stage A: 키워드 매칭 — 0ms, LLM 호출 없음]
  키워드 추출 → 매칭 문서 + 섹션 후보 선별
  → 5개 이하 섹션 매칭: Stage A 결과만 사용 (충분)
  → 5개 초과 or 키워드 매칭 0건: Stage B 진행
    ↓
[Stage B: LLM 트리 추론 — 1회 API 호출]
  트리 인덱스(제목+요약만, ~2K 토큰) → LLM에 전송
  → LLM이 관련 node_id 목록 반환
  → 해당 노드의 텍스트만 추출하여 주입
```

**왜 Phase 3인가**: LLM 트리 검색은 **추가 API 호출 1회**가 필요. Phase 1-2는 LLM 호출 없이 동작하는 기능에 집중. Phase 3에서 SQLite + 프로바이더가 안정화된 후, LLM 의존 기능을 추가하는 것이 안전.

**토큰 효율**:
```
기존 (v0.5.0):  전체 문서 주입    ~32K chars → ~8K tokens
Phase 1 트리:   섹션 단위 주입    ~5K chars  → ~1.3K tokens (84% 절감)
Phase 3 LLM:    추론 기반 정밀    ~3K chars  → ~750 tokens (91% 절감)
                + 검색 호출 1회    ~2K tokens (트리 인덱스 전송)
```

**구현 파일**:
```
internal/knowledge/
├── tree.go       (Phase 1) + LLM 검색 함수 추가
└── search.go     (신규, ~100줄) — LLM 프롬프트 구성, JSON 파싱, 노드 매핑
```

**폐쇄망 고려**: 검색 프롬프트는 사내 LLM 서버로 전송. 트리 인덱스는 바이너리에 내장되어 있으므로 외부 네트워크 불필요. LLM 서버 미연결 시 Phase 1의 키워드 매칭으로 자동 fallback.

---

## 포팅 총정리

### 줄 수 및 작업량

| Phase | 기능 수 | hanimo 줄 수 | 신규 줄 수 | 예상 적응 줄 수 | 작업 시간 |
|-------|--------|-------------|----------|---------------|----------|
| Phase 1 | 7 (6 hanimo + 1 PageIndex) | 702 | ~200 | ~850 | ~14시간 |
| Phase 2 | 5 (4 hanimo + 1 Obsidian) | 577 | ~150 | ~700 | ~14시간 |
| Phase 3 | 3 (2 hanimo + 1 PageIndex) | 778 | ~100 | ~600 | ~12시간 |
| **합계** | **15** | **2,057** | **~450** | **~2,150** | **~40시간** |

### 바이너리 영향

| 항목 | 변화 |
|------|------|
| Phase 1 (Go 코드 + tree-index.json embed) | +~50KB |
| Phase 2 (Go 코드 + link-graph.json embed) | +~10KB |
| Phase 3 SQLite (`modernc.org/sqlite`) | +5MB |
| Phase 3 UUID (`google/uuid`) | +50KB |
| **총 바이너리 증가** | **~5.1MB** (16MB → 21.1MB) |

### 새 도구 목록 (현재 7개 → 12개)

| # | 도구 | Phase | 용도 |
|---|------|-------|------|
| 8 | `git_status` | 1 | Git 상태 조회 |
| 9 | `git_diff` | 1 | Git diff 조회 |
| 10 | `git_log` | 1 | Git 커밋 히스토리 |
| 11 | `hashline_read` | 1 | 해시 앵커 파일 읽기 |
| 12 | `hashline_edit` | 1 | 해시 앵커 파일 편집 |

**+ diagnostics는 도구가 아닌 `/diagnostics` 슬래시 명령으로 제공**

### 새 슬래시 명령 (현재 /clear만 → 17개)

| 명령 | Phase | 설명 |
|------|-------|------|
| `/auto` | 1 | 자율 모드 토글 |
| `/diagnostics` | 1 | 프로젝트 진단 실행 |
| `/lang` | 2 | 한/영 전환 |
| `/theme <name>` | 2 | 테마 변경 |
| `/knowledge` | 2 | 내장 지식 트리 탐색 (Obsidian 스타일) |
| `/knowledge <q>` | 2 | 지식 검색 미리보기 |
| `/sessions` | 3 | 세션 목록 |
| `/resume <id>` | 3 | 세션 이어가기 |
| `/save [name]` | 3 | 세션 이름 지정 |
| `/search <query>` | 3 | 세션 검색 |
| `/fork` | 3 | 세션 분기 |
| `/remember <k> <v>` | 3 | 프로젝트 메모리 저장 |
| `/memories` | 3 | 메모리 목록 |
| `/model <name>` | 3 | 모델 전환 |
| `/provider <name>` | 3 | 프로바이더 전환 |
| `/usage` | 3 | 토큰 사용량 통계 |

### 키 바인딩 추가

| 키 | Phase | 기능 |
|----|-------|------|
| `Ctrl+K` | 2 | 커맨드 팔레트 열기 |
| `Esc` | 2 | 메뉴 오버레이 토글 |

### 지식 시스템 진화 경로

```
v0.5.0 (현재)
  └─ 키워드 매칭 → 문서 전체 주입 (~8K tokens)

v0.6.0 (Phase 1: 계층형 인덱스)
  └─ 키워드 매칭 → 트리 탐색 → 섹션 단위 주입 (~1.3K tokens, 84% 절감)
  └─ Frontmatter 메타데이터 (tags, related)
  └─ 빌드 타임 트리 생성 (LLM 호출 불필요)

v0.7.0 (Phase 2: Obsidian 스타일)
  └─ [[wiki-link]] 기반 문서 간 연결 그래프
  └─ /knowledge 대화형 탐색기
  └─ 백링크 → 관련 문서 자동 서피싱

v0.8.0 (Phase 3: LLM 추론 검색)
  └─ 키워드 fallback + LLM 트리 추론 하이브리드 (~750 tokens, 91% 절감)
  └─ PageIndex 스타일 "제목+요약 → LLM이 노드 선택"
  └─ LLM 미연결 시 키워드 매칭 자동 fallback
```

---

## 핵심 제약 (기존 유지)

1. **DEBUG_TRANSPORT_FREEZE**: HTTP transport 래핑 금지. 새 기능도 LLM HTTP와 분리.
2. **구 app.go 기반 유지**: `app.go` 변경은 필드 추가 + 메서드 호출만. 대규모 리팩토링 금지.
3. **단일 바이너리**: `modernc.org/sqlite` (pure Go). CGo 금지.
4. **폐쇄망 호환**: 런타임 외부 네트워크 의존 금지 (LLM 엔드포인트 제외).

---

## 참고 프로젝트

| 프로젝트 | 경로 | 차용 개념 |
|---------|------|----------|
| **hanimo** | `/Users/kimjiwon/Desktop/kimjiwon/hanimo/` | 12개 기능 직접 포팅 (Phase 1-3 핵심) |
| **PageIndex** | `/Users/kimjiwon/Desktop/kimjiwon/PageIndex/` | 계층형 트리 인덱스, LLM 추론 기반 검색 |
| **Obsidian** | (개념 차용) | `[[wiki-link]]`, frontmatter, 태그, 백링크, 그래프 |
| OpenCode | `/Users/kimjiwon/Desktop/kimjiwon/opencode/` | 설계 참고 (next-features-analysis.md) |
| gstack | `/Users/kimjiwon/Desktop/kimjiwon/gstack/` | 브라우저 컴패니언 설계 참고 |

---

## 기존 문서와의 관계

| 문서 | 역할 |
|------|------|
| `2026-04-09-next-features-analysis.md` | OpenCode + gstack 기반 **설계 스펙** (Phase 1-6 상세) |
| **이 문서** (hanimo-porting-guide.md) | hanimo + PageIndex + Obsidian 기반 **포팅 가이드** (실제 소스 매핑) |

**핵심 차이**: `next-features-analysis.md`는 "무엇을 만들 것인가" (설계), 이 문서는 "어디서 가져와서 어떻게 적응할 것인가" (포팅). hanimo에 이미 구현된 기능은 처음부터 만들 필요 없이 **복사 + 적응**으로 빠르게 이식 가능. PageIndex의 트리 인덱스 + Obsidian의 링크 그래프 개념을 추가하여 지식 시스템을 전면 업그레이드.

**예상 절감**: OpenCode 참고 시 14-22주 → hanimo 포팅 + PageIndex/Obsidian 차용으로 **~5주 (40시간)**로 단축.
