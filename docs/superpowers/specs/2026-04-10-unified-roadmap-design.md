# 택가이코드 통합 로드맵 설계서

> **작성일**: 2026-04-10
> **현재 버전**: v0.5.0 (Embedded Knowledge Store 완성)
> **대상 모델**: GPT-OSS-120B (슈퍼택가이/플랜), qwen3-coder-30b (개발)
> **서버 구성**: 단일 온프레미스 GPU 서버, 2개 모델 동시 서빙
> **목적**: 기존 두 스펙 문서를 통합한 마스터 로드맵 + 지식 시스템 + 개인화 설계
> **선행 문서**:
>   - [`next-features-analysis.md`](2026-04-09-next-features-analysis.md) — "무엇을 만들 것인가" (설계 스펙)
>   - [`hanimo-porting-guide.md`](2026-04-09-hanimo-porting-guide.md) — "어디서 가져올 것인가" (포팅 가이드)
> **다음 구현 시 이 문서를 마스터로 참조할 것**

---

## 설계 원칙

### 1. 단일 바이너리 원칙
- 모든 기능은 `go build` 한 번으로 생성되는 단일 실행파일 안에 포함
- 외부 런타임 의존 없음 (LLM 엔드포인트 제외)
- USB 하나에 복사 → 폐쇄망 PC에서 바로 실행
- 최종 예상 크기: **~21.5MB** (Phase 1-4 완료 기준)

### 2. 2-모델 최적화
- 큰 모델(120B)과 작은 모델(30B)은 능력이 다름
- 시스템이 모델 능력에 맞게 **도구, 지식 주입량, 압축 전략을 자동 조절**
- 사용자는 Tab 키로 모드만 전환, 나머지는 자동

### 3. 점진적 지능
- 사용할수록 개인화됨 (메모리 축적)
- 지식은 빌드 내장 + 사용자 추가 두 트랙
- 지식 검색은 키워드(0ms) → 트리(정밀) → LLM 추론(최정밀) 3단계 진화

### 4. 기존 코드 최소 침범
- `app.go` 변경은 필드 추가 + 메서드 호출만
- 새 기능은 새 패키지 (`internal/xxx/`)로 격리
- DEBUG_TRANSPORT_FREEZE 제약 항상 준수

---

## 현재 상태 (v0.5.0)

| 항목 | 현재 | 문제점 |
|------|------|--------|
| 도구 | 7개 (file_read/write/edit, list_files, shell_exec, grep/glob_search) | file_edit 정확일치만 |
| 내장 지식 | 38문서 (BXM/CSS/React/Charts/Vue/Java/Python/Skills) | 문서 전체 주입 (토큰 낭비) |
| 세션 | 인메모리 (`history []Message`) | 재시작 시 소멸 |
| 토큰 카운팅 | `tokenCount++` per SSE chunk | 실제 토큰 아님 |
| 컨텍스트 관리 | tool loop 20회 하드캡 | 긴 대화에서 오버플로 |
| 모델 관리 | 하드코딩 2개 | 모드별 차등 설정 없음 |
| Git | shell_exec 수동 | 구조화된 정보 없음 |
| 편집 안전망 | 없음 | 복구 불가 |
| 개인화 | 없음 | 프로젝트별 학습 없음 |

---

## 통합 Phase 정의

기존 두 문서의 Phase가 충돌했으므로 **폐쇄망 사용자 체감 가치** 순서로 재정의.

| Phase | 버전 | 테마 | 바이너리 | 기간 |
|-------|------|------|---------|------|
| **1: 생존** | v0.6.0 | 크래시 방지 + 기본 운용 | 16MB (±0) | ~1주 |
| **2: 신뢰성** | v0.7.0 | 편집 성공률 + 지식 정밀화 | ~16.5MB (+0.5MB) | ~2주 |
| **3: 영속성** | v0.8.0 | 세션 저장 + 개인화 | ~21.5MB (+5MB) | ~2주 |
| **4: UX 완성** | v0.9.0 | 편의 기능 | 21.5MB (±0) | ~1주 |
| **추후** | v1.0+ | 사용자 지식 + 자동학습 + 브라우저 등 | TBD | TBD |

---

## Phase 1: 생존 (v0.6.0)

> **테마**: "긴 대화에서 안 죽게"
> **의존성 추가**: 없음
> **바이너리**: 16MB 유지
> **예상 기간**: ~1주

### 1-1. 컨텍스트 압축

**소스**: hanimo `internal/llm/compaction.go` (137줄)

**현재 문제**: tool loop 20회 하드캡만. 긴 대화에서 컨텍스트 윈도우 초과 → API 에러 → 세션 사망.

**3단계 압축**:

```
Stage 1 (Snip): 40+ 메시지일 때, 최근 N개 제외한 오래된 tool output을
               200자 이상이면 "[snipped: N lines]"로 대체
               → 30-50% 절감

Stage 2 (Micro): 4000자 초과 메시지를
                head(2000) + "[...truncated...]" + tail(2000)로 절단

Stage 3 (LLM): 여전히 maxTokens 초과 시
              시스템 프롬프트 + 최근 N개 보존, 중간을 LLM에 요약 요청
              → 추가 20-30% 절감
```

**토큰 추정**: `estimateTokens() = len(text) / 4` (외부 토크나이저 불필요)

**모델별 차등 임계값**:

| 설정 | GPT-OSS-120B (128K) | qwen3-coder-30b (32K) |
|------|---------------------|----------------------|
| Stage 1+2 트리거 | 80% (102,400 tok) | 60% (19,200 tok) |
| Stage 3 트리거 | 90% (115,200 tok) | 75% (24,000 tok) |
| 압축 후 목표 | 60% | 50% |
| 보존 메시지 수 | 최근 15개 | 최근 8개 |

**포팅 방법**: `compaction.go` 복사 → `택가이코드/internal/llm/` 배치. `sendMessage()`에서 전송 전 `Compact()` 호출 추가.

**구현 파일**:
```
internal/llm/
├── compaction.go    (신규, hanimo 복사+적응, ~150줄)
└── client.go        (기존, Compact() 호출 추가)
```

---

### 1-2. 실제 토큰 카운팅

**현재 문제**: `tokenCount++`는 SSE 청크 수. 실제 토큰과 무관.

**해결**: SSE 스트리밍 완료 후 API 응답의 `usage` 필드에서 실제 토큰 추출.

```go
// 스트리밍 완료 시 (app.go의 streamDone 핸들러)
if resp.Usage != nil {
    m.inputTokens += resp.Usage.PromptTokens
    m.outputTokens += resp.Usage.CompletionTokens
    m.totalTokens = m.inputTokens + m.outputTokens
}
```

**상태바 표시**:
```
슈퍼택가이  GPT-OSS 120B  ■■■■■□□□ 45,231/128,000 tok (35%)
```

**컨텍스트 압축 연동**: `totalTokens / contextWindow` 비율이 임계값 초과 시 자동 압축 트리거.

**구현 변경**:
```
internal/app/app.go    — Model에 inputTokens, outputTokens, totalTokens 필드 추가
                         streamDone 핸들러에서 usage 파싱
internal/ui/super.go   — RenderStatusBar()에 토큰 바 추가
```

---

### 1-3. 모델 능력 레지스트리

**소스**: hanimo `internal/llm/capabilities.go` (99줄)

**현재 문제**: 두 모델을 동일하게 취급. 30B 모델에게 복잡한 도구 체이닝 → 실패.

**모델 능력 구조체**:
```go
type CodingTier int
const (
    CodingStrong   CodingTier = iota  // 120B급: 복잡한 도구 체이닝 가능
    CodingModerate                     // 30B급: 단순 도구, 해시 편집 선호
    CodingWeak                         // 7B급: 읽기 전용 추천
)

type ModelCapability struct {
    ContextWindow   int
    CodingTier      CodingTier
    SupportsTools   bool

    // 지식 주입 설정
    KnowledgeBudget int      // 토큰
    MaxSections     int      // 최대 주입 섹션 수
    SearchTiers     []int    // 검색 대상 Tier (0=BXM, 1=Daily, ...)

    // 컨텍스트 압축 설정
    PruneTrigger    float64  // Stage 1+2 트리거
    PruneUrgent     float64  // Stage 3 트리거
    PreserveRecent  int      // 보존할 최근 메시지 수

    // 편집 전략
    EditTool        string   // "file_edit" | "hashline_edit"
}
```

**2-모델 프로필**:
```go
var knownModels = map[string]ModelCapability{
    "gpt-oss-120b": {
        ContextWindow:   128000,
        CodingTier:      CodingStrong,
        SupportsTools:   true,
        KnowledgeBudget: 8000,
        MaxSections:     6,
        SearchTiers:     []int{0, 1, 2},
        PruneTrigger:    0.80,
        PruneUrgent:     0.90,
        PreserveRecent:  15,
        EditTool:        "file_edit",
    },
    "qwen3-coder-30b": {
        ContextWindow:   32000,
        CodingTier:      CodingModerate,
        SupportsTools:   true,
        KnowledgeBudget: 2000,
        MaxSections:     2,
        SearchTiers:     []int{0},
        PruneTrigger:    0.60,
        PruneUrgent:     0.75,
        PreserveRecent:  8,
        EditTool:        "hashline_edit",
    },
}

// 미등록 모델 → 보수적 기본값
var defaultCapability = ModelCapability{
    ContextWindow:   32000,
    CodingTier:      CodingModerate,
    KnowledgeBudget: 2000,
    MaxSections:     2,
    SearchTiers:     []int{0},
    PruneTrigger:    0.60,
    PruneUrgent:     0.75,
    PreserveRecent:  8,
    EditTool:        "hashline_edit",
}
```

**연동 포인트**:
- `sendMessage()`: 모델 능력에 따라 도구 목록 필터링
- `injector.go`: `KnowledgeBudget`, `MaxSections` 참조
- `compaction.go`: `PruneTrigger`, `PruneUrgent`, `PreserveRecent` 참조

**구현 파일**:
```
internal/llm/
└── capabilities.go    (신규, hanimo 복사+적응, ~110줄)
```

---

### 1-4. 자율 모드

**소스**: hanimo `internal/agents/auto.go` (26줄)

**현재 문제**: 매 도구 실행마다 사용자 입력 대기. 파일 10개 수정 같은 반복 작업에서 비효율.

**구현**: 시스템 프롬프트에 마커 주입 → LLM이 완료/중단 판단.

```go
const (
    MaxAutoIterations  = 20
    AutoCompleteMarker = "[AUTO_COMPLETE]"
    AutoPauseMarker    = "[AUTO_PAUSE]"
)
```

**슬래시 명령**: `/auto` — 자율 모드 토글. 상태바에 `[AUTO]` 표시.

**구현 파일**:
```
internal/agents/
└── auto.go    (신규, hanimo 복사, ~30줄)
```

---

### Phase 1 통합 효과

```
Before (v0.5.0):
  긴 대화 → API 에러 크래시
  토큰 바 없음
  모든 모델 동일 취급
  매번 수동 Enter

After (v0.6.0):
  긴 대화 → 자동 압축 → 안정 유지
  ■■■■■□□□ 45,231/128,000 tok (35%)
  120B: 풀 도구 + 넓은 지식 / 30B: 제한 도구 + 핵심 지식
  /auto로 반복 작업 자동화
```

---

## Phase 2: 신뢰성 (v0.7.0)

> **테마**: "편집이 안 깨지게"
> **의존성 추가**: `agnivade/levenshtein` (+3KB), `sergi/go-diff` (+200KB)
> **바이너리**: ~16.5MB
> **예상 기간**: ~2주

### 2-1. 퍼지 file_edit (4단계 fallback)

**설계 스펙 참고**: `next-features-analysis.md` Phase 2-A

**현재 문제**: `strings.Replace(content, oldText, newText, 1)` — 정확일치만. LLM이 공백/들여쓰기를 약간 다르게 출력하면 실패.

**4단계 매칭**:
```
1. ExactReplacer         ← 정확일치 (현재와 동일)
   실패 시 ↓
2. LineTrimmedReplacer   ← 각 줄 앞뒤 공백 제거 후 매칭
   실패 시 ↓
3. IndentFlexReplacer    ← 들여쓰기 레벨 무시, 대상 indent 유지
   실패 시 ↓
4. FuzzyReplacer         ← Levenshtein 유사도 85%+ 매칭
   실패 시 → 에러 반환
```

**안전장치**: 퍼지 매칭 시 `[FUZZY MATCH 92%]` 로그. 85% 미만 거부.

**모델별 전략**:
- GPT-OSS-120B: `file_edit` (퍼지 4단계) 사용
- qwen3-coder-30b: `hashline_edit` (해시 앵커) 우선 제공, `file_edit`도 fallback으로 등록

**구현 파일**:
```
internal/tools/
├── edit.go        (기존 file_edit 수정, 4단계 fallback 추가)
└── edit_test.go   (신규, fallback 단계별 테스트)
```

---

### 2-2. 해시 앵커 편집

**소스**: hanimo `internal/tools/hashline.go` (116줄)

**현재 문제**: 소형 모델(30B)은 정확한 문자열 복사 능력이 약함. 줄 번호만으로는 stale edit 위험.

**해시 앵커 방식**: 줄마다 MD5 해시 4자리 앵커 부여.

```
읽기 출력 (hashline_read):
1#a3f1| function hello() {
2#e4d9|   return 42
3#b2c1| }

편집 요청 (hashline_edit):
HashlineEdit(path, startAnchor="2#e4d9", endAnchor="2#e4d9", newContent="  return 99")

→ 해시 일치 확인 후 해당 줄만 교체
→ 해시 불일치 시 에러 + 최신 해시 안내 (stale edit 방지)
```

**새 도구 2개**: `hashline_read`, `hashline_edit`

**구현 파일**:
```
internal/tools/
└── hashline.go    (신규, hanimo 복사, ~120줄)
```

---

### 2-3. Git 도구

**소스**: hanimo `internal/tools/git.go` (52줄)

**5개 함수 래퍼** (10초 타임아웃):
```go
GitStatus(path) string   // git status --short
GitDiff(path)   string   // git diff
GitLog(path, n) string   // git log --oneline -n
GitCommit(path, msg)     // git commit -m
GitBranch(path) string   // git rev-parse --abbrev-ref HEAD
```

**새 도구 3개**: `git_status`, `git_diff`, `git_log` (읽기 전용, 안전)

**시스템 프롬프트 자동 주입** (`GatherFullContext()` 강화):
```
## Git Context
- Branch: feature/bxm-bean
- Status: 3 modified, 1 untracked
- Last commit: "feat: add bean template" (2h ago)
```

**구현 파일**:
```
internal/tools/
└── git.go    (신규, hanimo 복사, ~55줄)
```

---

### 2-4. Diff 뷰

**설계 스펙 참고**: `next-features-analysis.md` Phase 2-C

**해결**: `file_edit` 실행 시 편집 전후 unified diff를 빨강/초록 컬러로 표시.

```
[file_edit] /src/main.go
  - func main() {
  -     fmt.Println("hello")
  + func main() {
  +     fmt.Println("hello, world!")
      }
```

**의존성**: `github.com/sergi/go-diff` (pure Go, ~200KB)

**구현 파일**:
```
internal/tools/
└── diff.go    (신규, ~80줄)
```

---

### 2-5. 다국어 진단

**소스**: hanimo `internal/tools/diagnostics.go` (272줄)

**4개 언어 자동 감지 + 린팅**:

| 언어 | 린터 | 감지 파일 | 타임아웃 |
|------|------|----------|---------|
| Go | `go vet ./...` | `go.mod` | 30초 |
| TypeScript | `npx tsc --noEmit` | `tsconfig.json` | 60초 |
| JavaScript | `npx eslint --format compact` | `package.json` | 30초 |
| Python | `ruff check` | `pyproject.toml` | 30초 |

**슬래시 명령**: `/diagnostics` (도구가 아닌 명령으로 제공)

**폐쇄망 적응**: 린터 미설치 시 graceful skip + 안내 메시지.

**구현 파일**:
```
internal/tools/
└── diagnostics.go    (신규, hanimo 복사+적응, ~280줄)
```

---

### 2-6. 파일 스냅샷 / Undo

**설계 스펙 참고**: `next-features-analysis.md` Phase 3-A

**현재 문제**: AI가 `file_edit`/`file_write`로 파일을 망가뜨리면 복구 불가.

**해결**: 모든 파일 수정 전에 원본을 자동 백업. `/undo` 명령으로 복원.

```
> /undo                  # 마지막 편집 되돌리기
  ✓ /src/utils.ts 복원됨

> /undo all              # 이 세션의 모든 편집 되돌리기
  ✓ /src/App.tsx 복원됨
  ✓ /src/utils.ts 복원됨
```

**저장 구조**:
```
~/.tgc/snapshots/
└── {session_id}/
    └── 2026-04-10T16:41:17/
        └── src/App.tsx       ← 수정 전 원본 복사본
```

**자동 정리**: 7일 이상 된 스냅샷 자동 삭제 (앱 시작 시 체크). 총 500MB 초과 시 오래된 것부터 삭제.

**구현 파일**:
```
internal/snapshot/
├── snapshot.go    (신규, ~100줄)
└── cleanup.go     (신규, ~30줄)
```

---

### 2-7. 계층형 트리 인덱스 (PageIndex 개념)

**영감**: [PageIndex](https://github.com/VectifyAI/PageIndex) — 벡터 DB 없이 문서 구조로 검색

**현재 문제**: 키워드 매칭으로 **문서 전체** 주입. "BXM Bean 작성법" → BXM 전체 40KB 주입. 정작 필요한 건 Bean 섹션 2KB.

**PageIndex 개념의 택가이코드 적응**:
- PageIndex는 Python + 런타임 LLM 호출
- 택가이코드는 **빌드 타임에 트리를 미리 생성**하여 JSON으로 embed
- 런타임 LLM 호출 불필요 (폐쇄망 친화)

#### 트리 인덱스란?

MD 파일의 `#` 헤딩 구조를 파싱하여 트리로 만드는 것:

```markdown
# BXM 개요                    ← 노드 0001 (루트)
## Bean 구조                   ← 노드 0002 (자식)
### IO/DBIO 패턴                ← 노드 0003 (손자)
### Bean 필드 매핑              ← 노드 0004
## Service 계층                ← 노드 0005
```

#### 파일 크기별 적용 효과

| 파일 유형 | 예시 | 효과 |
|-----------|------|------|
| 큰 파일 (다수 `#` 섹션) | `bxm-overview.md` (40KB) | 관련 섹션만 → **80%+ 절감** |
| 중간 파일 (2-3 섹션) | `react-patterns.md` (8KB) | 1-2 섹션 → **40-60% 절감** |
| 작은 파일 (단일 주제) | `bxm-bean.md` (2KB) | leaf 노드 = 통째로 → 절감 없음 (이미 작음) |

#### 트리 노드 구조

```go
type KnowledgeNode struct {
    Title    string           `json:"title"`
    NodeID   string           `json:"node_id"`
    LineNum  int              `json:"line_num"`    // 원본 MD 시작 줄
    EndLine  int              `json:"end_line"`    // 원본 MD 끝 줄
    Summary  string           `json:"summary"`     // 빌드 타임 생성 1-2문장
    Keywords []string         `json:"keywords"`    // 기존 키워드 (하위호환)
    Tier     int              `json:"tier"`
    Children []KnowledgeNode  `json:"children,omitempty"`
}
```

#### 빌드 파이프라인

```
knowledge/docs/**/*.md
       ↓  (cmd/build-index/ 실행: # 헤딩 파싱 → 트리 구축)
knowledge/tree-index.json   ← //go:embed로 바이너리 내장
       ↓
런타임: 트리 로드 → 키워드 매칭 + 트리 탐색 병행
```

#### 검색 전략 (2단계 하이브리드)

```
쿼리: "BXM Bean 작성법"

1단계 (키워드 매칭 — 0ms, 기존 방식):
   키워드 추출: ["bxm", "bean"]
   → 매칭 문서: bxm-overview.md, bxm-bean.md

2단계 (트리 탐색 — 정밀):
   매칭 문서의 트리에서 키워드 포함 노드만 선택
   → bxm-overview.md → "Bean 구조" 섹션 (line 45-119) 만
   → bxm-bean.md → 전체 (단일 주제, leaf 노드)

결과: 40KB → ~5KB (관련 섹션만 주입)
```

#### 모델별 차등 주입

| 설정 | GPT-OSS-120B | qwen3-coder-30b |
|------|-------------|-----------------|
| 토큰 버짓 | 8,000 | 2,000 |
| 최대 섹션 수 | 6 | 2 |
| 검색 Tier | 0 + 1 + 2 | 0만 (BXM 핵심) |

#### tree-index.json 예시

```json
{
  "docs": [
    {
      "doc_name": "bxm-overview.md",
      "doc_description": "BXM 프레임워크 전체 개요",
      "tier": 0,
      "structure": [
        {
          "title": "BXM 개요",
          "node_id": "0001",
          "line_num": 1,
          "end_line": 44,
          "summary": "BXM 프레임워크의 목적과 전체 구조",
          "keywords": ["bxm", "프레임워크"],
          "children": [
            {
              "title": "Bean 구조",
              "node_id": "0002",
              "line_num": 45,
              "end_line": 119,
              "summary": "IO/DBIO/Bean 3계층과 Bean 작성 규칙",
              "keywords": ["bean", "io", "dbio"]
            },
            {
              "title": "Service 계층",
              "node_id": "0003",
              "line_num": 120,
              "end_line": 200,
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

#### 구현 파일

```
internal/knowledge/
├── store.go       (기존) + TreeIndex 필드 추가
├── tree.go        (신규, ~120줄) — 트리 로딩, 노드 검색
├── extractor.go   (기존) 그대로
└── injector.go    (기존) + 섹션 단위 주입 로직

cmd/build-index/
└── main.go        (기존) + 트리 인덱스 생성 로직 추가
```

---

### Phase 2 통합 효과

```
Before (v0.6.0):
  file_edit 정확일치 실패 → 재시도 → 느린 폐쇄망에서 시간 낭비
  파일 망가뜨리면 복구 불가
  Git은 shell_exec 수동
  지식 = 문서 전체 주입 (~8K tokens)

After (v0.7.0):
  file_edit 퍼지 4단계 → 편집 성공률 대폭 향상
  30B 모델은 hashline_edit로 안전 편집
  /undo로 언제든 복구
  git_status/diff/log 구조화 도구
  /diagnostics로 코드 에러 자동 감지
  트리 인덱스 → 섹션 단위 주입 (~1.3K tokens, 84% 절감)
```

**도구 목록 변화**: 7개 → 12개

| # | 도구 | Phase | 용도 |
|---|------|-------|------|
| 1-7 | (기존) | — | file_read/write/edit, list_files, shell_exec, grep/glob_search |
| 8 | `git_status` | 2 | Git 상태 조회 |
| 9 | `git_diff` | 2 | Git diff 조회 |
| 10 | `git_log` | 2 | Git 커밋 히스토리 |
| 11 | `hashline_read` | 2 | 해시 앵커 파일 읽기 |
| 12 | `hashline_edit` | 2 | 해시 앵커 파일 편집 |

---

## Phase 3: 영속성 (v0.8.0)

> **테마**: "재시작해도 기억"
> **의존성 추가**: `modernc.org/sqlite` (+5MB), `google/uuid` (+50KB)
> **바이너리**: ~21.5MB
> **예상 기간**: ~2주

### 3-1. SQLite 영구 세션

**소스**: hanimo `internal/session/` (466줄 → ~400줄 적응)

**현재 문제**: 대화가 인메모리. 터미널 닫으면 소멸.

**스키마**:
```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,              -- UUID
    name TEXT,                        -- 자동 생성 or 사용자 지정
    project_dir TEXT,                 -- CWD
    model TEXT,                       -- 사용 모델 ID
    mode INTEGER DEFAULT 0,           -- 0=super, 1=dev, 2=plan
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT REFERENCES sessions(id) ON DELETE CASCADE,
    role TEXT NOT NULL,                -- system/user/assistant/tool
    content TEXT NOT NULL,
    tool_calls TEXT,                   -- JSON (tool call 요청)
    tool_result TEXT,                  -- JSON (tool 실행 결과)
    input_tokens INTEGER DEFAULT 0,
    output_tokens INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_session ON messages(session_id, created_at);
```

**SQLite 설정 (WAL 모드)**:
```go
pragma journal_mode = WAL;
pragma synchronous = NORMAL;
pragma busy_timeout = 5000;
pragma foreign_keys = ON;
```

**슬래시 명령**:
- `/sessions` — 세션 목록
- `/resume <id>` — 이전 세션 이어가기
- `/save [name]` — 세션 이름 지정
- `/search <query>` — 세션 내용 검색
- `/fork` — 현재 세션 분기

**왜 `modernc.org/sqlite`인가**:

| 비교 | `modernc.org/sqlite` | `mattn/go-sqlite3` |
|------|---------------------|-------------------|
| CGo 필요 | **불필요** (pure Go) | 필요 |
| 단일 바이너리 | **유지** | 깨짐 |
| 크로스 컴파일 | **그대로** | 복잡 |
| 폐쇄망 빌드 | **`go mod vendor` OK** | C 툴체인 필요 |
| 바이너리 증가 | ~5MB | ~3MB |

**구현 파일**:
```
internal/storage/
├── db.go          (신규, ~120줄) — SQLite 초기화, 마이그레이션
├── session.go     (신규, ~180줄) — sessions CRUD
└── message.go     (신규, ~100줄) — messages CRUD
```

---

### 3-2. 개인화 메모리 시스템

**2가지 스코프**: 프로젝트별 + 글로벌

#### 메모리 DB 스키마

```sql
CREATE TABLE memories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scope TEXT NOT NULL,              -- 'project:/path/to/dir' 또는 'global'
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    source TEXT DEFAULT 'user',       -- 'user' (명시적) | 'auto' (추후 자동 학습)
    hit_count INTEGER DEFAULT 0,     -- 실제 시스템 프롬프트에 주입된 횟수
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(scope, key)
);
```

#### 슬래시 명령

```
/remember <내용>                  # 현재 프로젝트에 메모리 저장
/remember -g <내용>               # 글로벌 메모리 저장
/memories                         # 현재 프로젝트 메모리 목록
/memories -g                      # 글로벌 메모리 목록
/forget <key>                     # 메모리 삭제
```

#### 사용 예시

```
> /remember "이 프로젝트는 Bean 이름 뒤에 항상 Svc 붙인다"
  ✓ 메모리 저장: "Bean 이름 뒤에 항상 Svc 붙인다" (프로젝트: /home/user/bxm-app)

> /remember -g "커밋 메시지는 한국어로 쓴다"
  ✓ 글로벌 메모리 저장: "커밋 메시지는 한국어로 쓴다"

> /memories
  [프로젝트: /home/user/bxm-app]
  1. Bean 이름 뒤에 항상 Svc 붙인다 (참조 12회)
  2. SQL 쿼리는 세미콜론으로 끝낸다 (참조 8회)

  [글로벌]
  1. 커밋 메시지는 한국어로 쓴다 (참조 47회)
  2. 들여쓰기는 스페이스 4칸 (참조 33회)
```

#### 시스템 프롬프트 주입 흐름

```
세션 시작
  ↓
1. memories WHERE scope = 'project:{cwd}' → 프로젝트 메모리 로드
2. memories WHERE scope = 'global' → 글로벌 메모리 로드
3. hit_count 높은 순 정렬
4. 모델 토큰 버짓 내에서 상위 N개 선택
  ↓
시스템 프롬프트에 삽입:

## 사용자 지침 (프로젝트)
- Bean 이름 뒤에 항상 Svc 붙인다
- SQL 쿼리는 세미콜론으로 끝낸다

## 사용자 지침 (글로벌)
- 커밋 메시지는 한국어로 쓴다
- 들여쓰기는 스페이스 4칸
```

#### 모델별 메모리 버짓

| 설정 | GPT-OSS-120B | qwen3-coder-30b |
|------|-------------|-----------------|
| 프로젝트 메모리 | 상위 10개 | 상위 5개 |
| 글로벌 메모리 | 상위 5개 | 상위 3개 |
| 최대 토큰 | ~500 | ~200 |

#### 구현 파일

```
internal/storage/
└── memory.go    (신규, ~90줄) — memories CRUD + 주입 쿼리

internal/llm/
└── context.go   (기존, GatherFullContext에 메모리 주입 추가)
```

---

### 3-3. 멀티 모델 / 프로바이더 설정

**소스**: hanimo 포팅 (312줄 → ~250줄 적응)

**현재 문제**: 2개 모델이 `models.go`에 하드코딩.

**config.yaml 확장**:
```yaml
api:
  base_url: "https://techai-web-prod.shinhan.com/v1"
  api_key: "sk-internal-..."

models:
  super:
    id: "gpt-oss-120b"
    display_name: "GPT-OSS 120B"
    context_window: 128000
  dev:
    id: "qwen3-coder-30b"
    display_name: "Qwen3-Coder 30B"
    context_window: 32000
  plan:
    id: "gpt-oss-120b"
    display_name: "GPT-OSS 120B (플랜)"
    context_window: 128000
```

**단일 서버 구성**: `api.base_url` 하나로 두 모델 모두 서빙. 모드 전환 시 모델 ID만 변경.

**구현 파일**:
```
internal/config/
└── config.go    (기존, models 섹션 파싱 추가)

internal/llm/
└── client.go    (기존, 모드별 모델 ID 동적 전환)
```

---

### 3-4. LLM 추론 기반 트리 검색

**영감**: PageIndex의 Reasoning-based Retrieval

**Phase 2의 키워드 매칭 한계**:
- "BXM에서 페이징 처리할 때 성능 이슈 해결법" → 키워드 "bxm", "페이징" 매칭
- 정작 답은 `bxm-select.md`의 "대용량 조회 최적화" 섹션 → 키워드에 "성능" 없으면 놓침

**해결**: 트리 인덱스(제목+요약만)를 LLM에 보내서 관련 노드를 추론하게 함.

**2단계 하이브리드 검색 (최종)**:

```
쿼리 입력
    ↓
[Stage A: 키워드 매칭 — 0ms, LLM 호출 없음]
  키워드 추출 → 매칭 문서 + 섹션 후보 선별
  → 5개 이하 섹션 매칭: Stage A 결과만 사용 (충분)
  → 5개 초과 or 0건: Stage B 진행
    ↓
[Stage B: LLM 트리 추론 — 1회 API 호출]
  트리 인덱스(제목+요약만, ~2K 토큰) → LLM에 전송
  → LLM이 관련 node_id 목록 반환
  → 해당 노드의 텍스트만 추출하여 주입
```

**검색 프롬프트**:
```
당신은 내장 기술 문서의 검색 도우미입니다.
아래 트리 구조에서 사용자 질문에 답할 수 있는 노드를 찾으세요.

질문: {query}

문서 트리:
{tree_index_json_titles_and_summaries_only}

다음 JSON으로 답하세요:
{"thinking": "추론 과정", "nodes": ["node_id_1", "node_id_2"]}
```

**토큰 효율 비교**:
```
v0.5.0 (현재):  문서 전체 주입     ~8K tokens
v0.7.0 트리:    섹션 단위 주입     ~1.3K tokens (84% 절감)
v0.8.0 LLM:     추론 기반 정밀     ~750 tokens (91% 절감)
                 + 검색 1회          ~2K tokens (트리 전송 비용)
```

**폐쇄망 안전**: LLM 서버 미연결 시 Phase 2의 키워드 매칭으로 자동 fallback.

**구현 파일**:
```
internal/knowledge/
├── tree.go      (Phase 2에서 생성) + LLM 검색 함수 추가
└── search.go    (신규, ~100줄) — LLM 프롬프트 구성, JSON 파싱
```

---

### Phase 3 통합 효과

```
Before (v0.7.0):
  재시작 → 대화 소멸
  모델 설정 하드코딩
  개인 지침 매번 수동 입력
  지식 검색 = 키워드 매칭만

After (v0.8.0):
  /sessions, /resume → 대화 영구 저장
  config.yaml에서 모드별 모델/엔드포인트 설정
  /remember → 개인 지침 자동 주입 (프로젝트별 + 글로벌)
  LLM 추론 트리 검색 → 토큰 91% 절감
```

---

## Phase 4: UX 완성 (v0.9.0)

> **테마**: "편하게"
> **의존성 추가**: 없음
> **바이너리**: 21.5MB 유지
> **예상 기간**: ~1주

### 4-1. 커맨드 팔레트

**소스**: hanimo `internal/ui/palette.go` (154줄)

`Ctrl+K`로 팔레트 열기 → 퍼지 검색 → Enter 실행.

```
┌─────────────────────────────────┐
│ > 검색...                        │
├─────────────────────────────────┤
│ ▶ 세션 저장      /save          │
│   세션 불러오기   /resume        │
│   진단 실행      /diagnostics   │
│   테마 변경      /theme         │
│   도움말         /help          │
├─────────────────────────────────┤
│ ↑↓ 이동  Enter 선택  Esc 닫기   │
└─────────────────────────────────┘
```

### 4-2. 메뉴 오버레이

**소스**: hanimo `internal/ui/menu.go` (98줄)

`Esc`으로 플로팅 메뉴 표시. 모델 전환, 사용량 통계 등.

### 4-3. 한/영 전환

**소스**: hanimo `internal/ui/i18n.go` (133줄)

27개 UI 문자열의 한/영 번역. `/lang` 토글.

### 4-4. 테마 시스템

**소스**: hanimo `internal/ui/styles.go` (192줄)

5개 프리셋: honey, ocean, dracula, nord, forest. `/theme <name>` 전환. `config.yaml`에 저장.

### 4-5. Obsidian 스타일 지식 탐색

**영감**: Obsidian의 `[[wiki-link]]` + 태그 + 백링크

`/knowledge` 명령으로 내장 지식 트리를 대화형으로 탐색:

```
> /knowledge

┌─ 내장 지식 (38 문서) ────────────────────────┐
│                                               │
│ [Tier 0 — BXM]  13 문서                       │
│  ├─ bxm-overview     BXM 프레임워크 전체 개요  │
│  │   ├─ Bean 구조                             │
│  │   ├─ Service 계층                          │
│  │   └─ Select/Paging                         │
│  └─ ...                                       │
│                                               │
│ ↑↓ 이동  Enter 펼치기  / 검색  Esc 닫기       │
└───────────────────────────────────────────────┘
```

**구현 파일**:
```
internal/knowledge/
├── links.go      (신규, ~80줄) — [[wiki-link]] 파싱, 백링크 그래프
└── browser.go    (신규, ~70줄) — /knowledge TUI 렌더링

knowledge/
└── link-graph.json    (신규, 빌드 타임 생성)
```

---

## 추후 계획 (Phase 5+)

Phase 1-4 완료 후, 사용자 피드백을 받아 결정.

### 5-A. 사용자 커스텀 지식 추가

사용자가 `~/.tgc/knowledge/`에 `.md` 파일을 넣으면 자동 인덱싱.

**동작 방식**:
```
앱 시작
  ↓
1. 내장 지식 로드 (go:embed, 항상 존재)
2. ~/.tgc/knowledge/ 디렉토리 스캔
3. 사용자 MD 파일 발견 → # 헤딩 파싱 → index.json 자동 생성
4. 내장 트리 + 사용자 트리 병합 → 통합 검색
```

**사용법**: 파일 복사만. 슬래시 명령이나 앱 내 에디터 없음.
```bash
cp my-api-guide.md ~/.tgc/knowledge/
cp team-conventions.md ~/.tgc/knowledge/
# 다음 앱 시작 시 자동 인덱싱
```

**우선순위 가중치**:
```
사용자 지식  ×1.5  (내가 추가한 건 더 중요)
Tier 0 (BXM) ×1.2  (회사 핵심)
Tier 1-3     ×1.0  (일반 레퍼런스)
```

### 5-B. 자동 학습 (패턴 감지)

LLM이 사용자의 반복 패턴을 감지하여 자동 제안:

```
"3번 연속 같은 import 스타일을 사용하셨습니다.
 기억할까요? (Y/n)"

→ Y 입력 시 memories에 source='auto'로 저장
```

### 5-C. 브라우저 컴패니언

**설계 스펙 참고**: `next-features-analysis.md` Phase 4

`/companion` → localhost:8787 대시보드. SSE 실시간 스트리밍. Mermaid 다이어그램.
바이너리 +1.5MB.

### 5-D. LSP 통합

**설계 스펙 참고**: `next-features-analysis.md` Phase 5

gopls → typescript-language-server 순서. 폐쇄망에서 언어 서버 사전 설치 필요.
바이너리 +1MB.

### 5-E. Vim 키바인딩

**설계 스펙 참고**: `next-features-analysis.md` Phase 6

Normal/Insert/Visual 모드. config.yaml에서 `editor.mode: "vim"`.

---

## 전체 아키텍처 다이어그램

```
┌─────────────────── techai 바이너리 (~21.5MB) ──────────────────┐
│                                                                 │
│  ┌─ 모델 레이어 ────────────────────────────────────────────┐  │
│  │  GPT-OSS-120B              qwen3-coder-30b               │  │
│  │  ┌─────────────┐           ┌─────────────┐               │  │
│  │  │ 128K 컨텍스트 │         │ 32K 컨텍스트  │               │  │
│  │  │ 퍼지 edit    │          │ 해시 edit    │               │  │
│  │  │ 지식 8K tok  │          │ 지식 2K tok  │               │  │
│  │  │ 압축 80%     │          │ 압축 60%     │               │  │
│  │  └─────────────┘           └─────────────┘               │  │
│  │                                                          │  │
│  │  capabilities.go — 모델별 자동 프로필 적용                │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                 │
│  ┌─ 지식 레이어 ────────────────────────────────────────────┐  │
│  │                                                          │  │
│  │  [빌드 내장]                  [런타임 사용자 — 추후]       │  │
│  │  knowledge/docs/*.md          ~/.tgc/knowledge/*.md       │  │
│  │  tree-index.json              index.json (자동 빌드)      │  │
│  │       ↓                              ↓                    │  │
│  │       └──────── 통합 트리 검색 ────────┘                  │  │
│  │                     ↓                                     │  │
│  │    1. 키워드 매칭 (0ms)                                   │  │
│  │    2. 트리 탐색 → 섹션 선택 (Phase 2)                     │  │
│  │    3. LLM 추론 검색 (Phase 3, fallback)                   │  │
│  │                     ↓                                     │  │
│  │         모델별 토큰 버짓 내에서 주입                       │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                 │
│  ┌─ 개인화 레이어 ──────────────────────────────────────────┐  │
│  │                                                          │  │
│  │  sessions.db                                              │  │
│  │  ├─ sessions    — 대화 영구 저장 (Phase 3)                │  │
│  │  ├─ messages    — 메시지 + 토큰 카운트                    │  │
│  │  └─ memories    — 개인화 지침                             │  │
│  │      ├─ scope='project:/path'  — 프로젝트별               │  │
│  │      └─ scope='global'         — 전역                     │  │
│  │                                                          │  │
│  │  hit_count 기반 상위 N개 → 시스템 프롬프트 주입            │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                 │
│  ┌─ 도구 레이어 ────────────────────────────────────────────┐  │
│  │                                                          │  │
│  │  [공통] file_read, file_write, list_files,               │  │
│  │         shell_exec, grep_search, glob_search             │  │
│  │         git_status, git_diff, git_log                    │  │
│  │                                                          │  │
│  │  [120B 전용] file_edit (퍼지 4단계)                       │  │
│  │  [30B 전용]  hashline_read, hashline_edit (해시 앵커)     │  │
│  │                                                          │  │
│  │  /diagnostics  — 프로젝트 린팅 (슬래시 명령)              │  │
│  │  /undo         — 파일 복원 (슬래시 명령)                  │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                 │
│  ┌─ UI 레이어 ──────────────────────────────────────────────┐  │
│  │  Bubble Tea v2 TUI                                        │  │
│  │  ├─ 3개 모드 탭 (슈퍼택가이/개발/플랜)                    │  │
│  │  ├─ 상태바 (모델명 + 토큰 바 + [AUTO])                   │  │
│  │  ├─ 커맨드 팔레트 (Ctrl+K) — Phase 4                     │  │
│  │  ├─ 메뉴 오버레이 (Esc) — Phase 4                        │  │
│  │  ├─ 한/영 전환 — Phase 4                                 │  │
│  │  └─ 테마 시스템 (5 프리셋) — Phase 4                     │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
         ↕
    사내 GPU 서버 (단일)
    ├─ GPT-OSS-120B  ← 슈퍼택가이/플랜 모드
    └─ qwen3-coder-30b  ← 개발 모드
```

---

## 바이너리 크기 변화

| Phase | 버전 | 크기 | 변화 | 원인 |
|-------|------|------|------|------|
| — | v0.5.0 (현재) | 16MB | — | TUI + LLM + 7도구 + 지식 38문서 |
| 1 | v0.6.0 | 16MB | ±0 | Go 코드만 (의존성 0) |
| 2 | v0.7.0 | ~16.5MB | +0.5MB | levenshtein + go-diff + tree-index.json |
| 3 | v0.8.0 | ~21.5MB | +5MB | modernc.org/sqlite + uuid |
| 4 | v0.9.0 | 21.5MB | ±0 | Go 코드만 (lipgloss 이미 사용 중) |
| 추후 | v1.0+ | ~24MB | +2.5MB | 브라우저 컴패니언 + LSP (선택) |

---

## 슬래시 명령 전체 목록

| 명령 | Phase | 설명 |
|------|-------|------|
| `/clear` | 기존 | 대화 초기화 |
| `/help` | 기존 | 도움말 |
| `/auto` | 1 | 자율 모드 토글 |
| `/diagnostics` | 2 | 프로젝트 린팅 실행 |
| `/undo` | 2 | 마지막 파일 편집 되돌리기 |
| `/undo all` | 2 | 세션 내 모든 편집 되돌리기 |
| `/sessions` | 3 | 세션 목록 |
| `/resume <id>` | 3 | 이전 세션 이어가기 |
| `/save [name]` | 3 | 세션 이름 지정 |
| `/search <query>` | 3 | 세션 내용 검색 |
| `/fork` | 3 | 현재 세션 분기 |
| `/remember <내용>` | 3 | 프로젝트 메모리 저장 |
| `/remember -g <내용>` | 3 | 글로벌 메모리 저장 |
| `/memories` | 3 | 프로젝트 메모리 목록 |
| `/memories -g` | 3 | 글로벌 메모리 목록 |
| `/forget <key>` | 3 | 메모리 삭제 |
| `/usage` | 3 | 토큰 사용량 통계 |
| `/lang` | 4 | 한/영 전환 |
| `/theme <name>` | 4 | 테마 변경 |
| `/knowledge` | 4 | 내장 지식 트리 탐색 |
| `/knowledge <query>` | 4 | 지식 검색 미리보기 |

---

## 핵심 제약 사항 (반드시 준수)

1. **DEBUG_TRANSPORT_FREEZE**: HTTP transport 래핑 금지. 모든 HTTP 통신은 별도 고루틴에서 직접 사용. LLM SSE 스트리밍과 절대 공유 불가.

2. **구 app.go 기반 유지**: `app.go` 변경은 **필드 추가 + 메서드 호출**만. Update/View 로직의 대규모 리팩토링 금지. 새 기능은 새 패키지로 격리.

3. **단일 바이너리**: 모든 에셋은 `//go:embed`로 내장. 외부 파일 런타임 의존 금지.

4. **폐쇄망 호환**: 빌드 타임에만 네트워크 사용 (`go mod download`). 런타임에 LLM 엔드포인트 외 네트워크 의존 금지.

5. **pure Go only**: CGo 필요 라이브러리 사용 금지. `GOOS=windows go build` 크로스 컴파일 항상 가능해야 함.

6. **기존 빌드 호환**: `make build`, `make build-all`, `make build-onprem` 명령 유지. 새 Phase 추가 시 Makefile 수정 최소화.

---

## 구현 순서 총정리

```
v0.5.0 (현재) — Embedded Knowledge Store

v0.6.0 — Phase 1: 생존 (~1주)
  ├─ 1-1 컨텍스트 압축 (hanimo 137줄 포팅)
  ├─ 1-2 실제 토큰 카운팅 (API usage 파싱)
  ├─ 1-3 모델 능력 레지스트리 (hanimo 99줄 포팅)
  └─ 1-4 자율 모드 (hanimo 26줄 포팅)

v0.7.0 — Phase 2: 신뢰성 (~2주)
  ├─ 2-1 퍼지 file_edit (4단계 fallback)
  ├─ 2-2 해시 앵커 편집 (hanimo 116줄 포팅)
  ├─ 2-3 Git 도구 (hanimo 52줄 포팅)
  ├─ 2-4 Diff 뷰 (go-diff)
  ├─ 2-5 다국어 진단 (hanimo 272줄 포팅)
  ├─ 2-6 파일 스냅샷/Undo
  └─ 2-7 계층형 트리 인덱스 (PageIndex 개념)

v0.8.0 — Phase 3: 영속성 (~2주)
  ├─ 3-1 SQLite 영구 세션 (hanimo 466줄 포팅)
  ├─ 3-2 개인화 메모리 (프로젝트별 + 글로벌)
  ├─ 3-3 멀티 모델/프로바이더 설정
  └─ 3-4 LLM 추론 트리 검색 (PageIndex 핵심)

v0.9.0 — Phase 4: UX 완성 (~1주)
  ├─ 4-1 커맨드 팔레트 (Ctrl+K)
  ├─ 4-2 메뉴 오버레이 (Esc)
  ├─ 4-3 한/영 전환
  ├─ 4-4 테마 시스템
  └─ 4-5 Obsidian 지식 탐색

v1.0+ — 추후 (사용자 피드백 후)
  ├─ 사용자 커스텀 지식 (~/.tgc/knowledge/)
  ├─ 자동 학습 (패턴 감지)
  ├─ 브라우저 컴패니언
  ├─ LSP 통합
  └─ Vim 키바인딩
```

**총 예상: Phase 1-4 완료까지 ~6주, 폐쇄망 프로덕션 레디까지(Phase 1-3) ~5주**

---

## 기존 문서와의 관계

| 문서 | 역할 | 상태 |
|------|------|------|
| `next-features-analysis.md` | "무엇을 만들 것인가" (OpenCode/gstack 참고 설계) | 참고용 유지 |
| `hanimo-porting-guide.md` | "어디서 가져올 것인가" (hanimo 소스 매핑) | 참고용 유지 |
| **이 문서** (`unified-roadmap-design.md`) | **마스터 로드맵** — 통합 Phase + 2모델 최적화 + 개인화 | **구현 시 이 문서 참조** |

기존 두 문서에서 상세 구현 참고가 필요한 경우:
- 코드 수준 상세 → `hanimo-porting-guide.md` (소스 파일 매핑)
- 아키텍처 상세 → `next-features-analysis.md` (설계 다이어그램)
