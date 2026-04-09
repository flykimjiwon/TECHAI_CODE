# Embedded Knowledge Store — 설계 스펙

> 폐쇄망 환경에서 LLM 성능을 극대화하기 위한 바이너리 내장 지식 시스템

## 목표

택갈이코드 바이너리에 개발 레퍼런스 문서 + 스킬 템플릿을 Go `embed`로 내장하여,
폐쇄망 LLM(gpt-oss-120b, qwen3-coder-30b)이 프레임워크/라이브러리를 정확하게 이해하고
코드를 생성할 수 있도록 한다.

## 핵심 원칙

1. **단일 바이너리** — 외부 파일 의존 없음, USB 반입 한 번으로 끝
2. **OS 인식** — `runtime.GOOS`로 Windows/Linux/macOS별 명령어 자동 필터링
3. **토큰 효율** — 키워드 매칭 + 모드 기반 하이브리드로 필요한 문서만 주입
4. **순수 Go** — CGO 없음, 크로스 컴파일 유지

## 아키텍처

```
┌─────────────────────────────────────────────┐
│  사용자 질문                                 │
│  "BXM Bean에서 다건 조회 패턴 알려줘"        │
└──────────────┬──────────────────────────────┘
               │
       ┌───────▼────────┐
       │ Keyword Extractor│  질문에서 키워드 추출
       │ [bxm, bean, 다건]│  (정규식 + 사전 매칭)
       └───────┬─────────┘
               │
       ┌───────▼────────┐
       │  Index Lookup   │  index.json에서 관련 문서 검색
       │  bxm → 5 docs  │  우선순위 정렬 (relevance score)
       └───────┬─────────┘
               │
       ┌───────▼────────┐
       │  Token Budget   │  8K 토큰 버짓 내에서 잘라서
       │  Manager        │  시스템 프롬프트에 추가
       └───────┬─────────┘
               │
       ┌───────▼────────┐
       │  LLM Request    │  기존 프롬프트 + 지식 컨텍스트
       └────────────────┘
```

## 지식 Tier 구조

| Tier | 분류 | 대상 | 주입 방식 |
|------|------|------|----------|
| **0** | 회사 핵심 | BXM (IO, DBIO, Bean, Service, 배치, Config, Studio) | 키워드 매칭 시 최우선 |
| **1** | 매일 사용 | Tailwind v4, shadcn/ui, React 19, Next.js 15, Recharts, Chart.js, TypeScript, Node.js | 키워드 매칭 |
| **2** | 자주 사용 | Vue 3, Bootstrap 5, ECharts, D3.js, Spring Boot, Vite | 키워드 매칭 |
| **3** | 참고 | Svelte, Nivo, Tremor, Python/FastAPI, Django | 키워드 매칭 |
| **OS** | 터미널 | Windows(주력), Linux, macOS, Git | OS 자동 필터링 |
| **Skills** | 작업 패턴 | TDD, 디버깅, 코드리뷰, Git workflow, 보안, 리팩토링 | 모드 기반 자동 |

## 디렉토리 구조

```
택갈이코드/
├── knowledge/                          # //go:embed 대상
│   ├── docs/
│   │   ├── bxm/                        # Tier 0 — 회사 핵심
│   │   │   ├── overview.md             # BXM 아키텍처 개요
│   │   │   ├── io.md                   # IO 정의 가이드
│   │   │   ├── dbio.md                 # DBIO 작성 (조회/등록/수정/삭제)
│   │   │   ├── bean.md                 # Bean (@BxmBean, @BxmCategory)
│   │   │   ├── service.md              # Service 작성, 트랜잭션
│   │   │   ├── batch.md               # 배치 처리 패턴 (일반/온디맨드/데몬)
│   │   │   ├── centercut.md           # Center-Cut 아키텍처
│   │   │   ├── config.md              # Framework Config
│   │   │   ├── studio.md              # BXM Studio 사용법
│   │   │   ├── naming.md              # 네이밍 규칙
│   │   │   ├── exception.md           # 예외 처리
│   │   │   ├── logging.md             # 로깅
│   │   │   └── patterns.md            # 다건 Select, Paging, Lock 등
│   │   ├── go/                         # Tier 1
│   │   │   ├── stdlib.md
│   │   │   ├── concurrency.md
│   │   │   ├── testing.md
│   │   │   └── modules.md
│   │   ├── javascript/                 # Tier 1
│   │   │   ├── es2024.md
│   │   │   ├── node.md
│   │   │   └── patterns.md
│   │   ├── typescript/                 # Tier 1
│   │   │   ├── types.md
│   │   │   ├── utility-types.md
│   │   │   └── config.md
│   │   ├── react/                      # Tier 1
│   │   │   ├── hooks.md
│   │   │   ├── patterns.md
│   │   │   ├── nextjs.md
│   │   │   └── state.md
│   │   ├── css/                        # Tier 1
│   │   │   ├── tailwind-v4.md
│   │   │   ├── shadcn-ui.md
│   │   │   ├── bootstrap.md
│   │   │   └── responsive.md
│   │   ├── charts/                     # Tier 1-2
│   │   │   ├── recharts.md
│   │   │   ├── chartjs.md
│   │   │   ├── d3.md
│   │   │   ├── echarts.md
│   │   │   └── nivo-tremor.md
│   │   ├── vue/                        # Tier 2
│   │   │   ├── composition.md
│   │   │   ├── nuxt.md
│   │   │   └── pinia.md
│   │   ├── java/                       # Tier 2
│   │   │   ├── core.md
│   │   │   ├── spring.md
│   │   │   └── build.md
│   │   ├── python/                     # Tier 3
│   │   │   ├── core.md
│   │   │   ├── fastapi.md
│   │   │   └── django.md
│   │   ├── terminal/                   # OS별
│   │   │   ├── windows.md
│   │   │   ├── linux.md
│   │   │   ├── macos.md
│   │   │   └── git.md
│   │   └── tools/                      # 공통
│   │       ├── vite.md
│   │       ├── docker.md
│   │       └── sql.md
│   ├── skills/                         # 작업 패턴
│   │   ├── tdd.md
│   │   ├── debugging.md
│   │   ├── code-review.md
│   │   ├── git-workflow.md
│   │   ├── refactoring.md
│   │   ├── performance.md
│   │   └── security.md
│   └── index.json                      # 키워드 → 문서 매핑
├── internal/knowledge/                 # Go 패키지
│   ├── store.go                        # embed.FS + 문서 로드
│   ├── index.go                        # 키워드 인덱스 로더/검색
│   ├── injector.go                     # 프롬프트 주입 (모드+키워드 하이브리드)
│   └── extractor.go                    # 키워드 추출기
└── cmd/build-index/                    # 빌드 도구
    └── main.go                         # index.json 생성 스크립트
```

## 핵심 컴포넌트 설계

### 1. store.go — 지식 저장소

```go
package knowledge

import "embed"

//go:embed knowledge/*
var knowledgeFS embed.FS

type Document struct {
    Path     string   // "docs/bxm/bean.md"
    Title    string   // 첫 번째 # 헤더에서 추출
    Content  string   // 마크다운 원문
    Tier     int      // 0-3
    OS       string   // "" | "windows" | "linux" | "darwin"
    Keywords []string // 인덱스에서 로드
}

type Store struct {
    docs  []Document
    index map[string][]int // keyword → doc indices
    osTag string           // runtime.GOOS
}

func NewStore() *Store
func (s *Store) Search(keywords []string, budget int) []Document
func (s *Store) ForMode(mode int) []Document
func (s *Store) ForOS() []Document
```

### 2. extractor.go — 키워드 추출기

```go
// 사용자 질문에서 기술 키워드를 추출
// 방식: 사전 매칭 (keyword dictionary) + 정규식
func ExtractKeywords(query string) []string
```

키워드 사전 예시:
```
"tailwind", "tw" → css/tailwind-v4.md
"bxm", "빈", "bean", "dbio", "서비스" → docs/bxm/*.md
"recharts", "차트", "chart", "그래프" → charts/recharts.md
"react", "리액트", "useState", "useEffect" → react/*.md
"ip", "아이피", "네트워크" → terminal/{os}.md
```

### 3. injector.go — 프롬프트 주입기

```go
type Injector struct {
    store       *Store
    tokenBudget int // 기본 8192
}

// Inject: 모드 기본 문서 + 키워드 매칭 문서를 합쳐서 반환
func (inj *Injector) Inject(mode int, userQuery string) string
```

주입 우선순위:
1. Tier 0 매칭 문서 (BXM) — 항상 최우선
2. OS별 터미널 문서 — OS 관련 질문 시
3. 모드별 기본 문서 — Dev: 코딩 패턴, Plan: 아키텍처
4. 키워드 매칭 문서 — Tier 1 → 2 → 3 순서
5. 토큰 버짓 초과 시 하위 Tier부터 제거

### 4. index.json — 키워드 인덱스

```json
{
  "documents": [
    {
      "path": "docs/bxm/bean.md",
      "title": "BXM Bean 작성 가이드",
      "tier": 0,
      "os": "",
      "keywords": ["bxm", "bean", "빈", "@BxmBean", "@BxmCategory", "비즈니스로직", "POJO"],
      "tokens": 1200
    },
    {
      "path": "docs/css/tailwind-v4.md",
      "title": "Tailwind CSS v4 레퍼런스",
      "tier": 1,
      "os": "",
      "keywords": ["tailwind", "tw", "css", "유틸리티", "반응형", "다크모드"],
      "tokens": 2000
    }
  ]
}
```

### 5. app.go 통합 지점

```go
// sendMessage() 내에서, LLM 호출 직전:
knowledgeCtx := m.injector.Inject(m.activeTab, userMessage)
systemPrompt = basePrompt + projectCtx + knowledgeCtx
```

## BXM 문서 수집 (SWLab 스크래핑)

### 스크래핑 대상 URL 목록

```
https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/
├── concepts/overview.html                          → overview.md
├── bxm-developer-guide-online/
│   ├── ch02/ch02_1.html (개발 프로세스)            → patterns.md
│   ├── ch02/ch02_2.html (프로젝트 생성)            → studio.md (일부)
│   ├── ch02/ch02_3.html (DBIO 작성)                → dbio.md
│   ├── ch02/ch02_4.html (Bean 작성)                → bean.md
│   ├── ch02/ch02_5.html (Service 작성)             → service.md
│   ├── ch02/ch02_6.html (테스트)                   → patterns.md (일부)
│   ├── ch03/ch03_1.html (다건 Select)              → patterns.md
│   ├── ch03/ch03_2.html (Paging Select)            → patterns.md
│   ├── ch03/ch03_3.html (Lock Select)              → patterns.md
│   ├── ch03/ch03_4.html (Insert/Update)            → patterns.md
│   ├── ch03/ch03_5.html (메시지 처리)              → patterns.md
│   ├── ch03/ch03_6.html (트랜잭션)                 → service.md (일부)
│   ├── ch03/ch03_7.html (예외 처리)                → exception.md
│   ├── ch03/ch03_8.html (로깅)                     → logging.md
│   └── ch03/ch03_9.html (서비스 호출)              → service.md (일부)
├── bxm-developer-guide-batch/
│   ├── ch01/ (배치 개요)                           → batch.md
│   ├── ch02/ (배치 개발 프로세스)                  → batch.md
│   └── ch03/ (배치 DBIO/Bean)                      → batch.md
├── bxm-configuration-reference-server/
│   └── ch04/ (Framework Config)                    → config.md
├── bxm-user-guide-studio/
│   ├── ch01/ (시작하기)                            → studio.md
│   ├── ch03/ (프로젝트 생성)                       → studio.md
│   └── ch19/ (로컬 개발환경)                       → studio.md
└── bxm-installation-guide-centercut/               → centercut.md
```

### 스크래핑 → 마크다운 변환 프로세스

1. `cmd/scrape-bxm/main.go` 스크래핑 도구 작성 (또는 빌드 전 수동 수집)
2. HTML → Markdown 변환 (코드 블록, 테이블 보존)
3. 페이지별로 분류하여 `knowledge/docs/bxm/` 에 저장
4. `cmd/build-index/main.go` 로 index.json 자동 생성

## 토큰 버짓 관리

| 구분 | 토큰 할당 |
|------|----------|
| 시스템 프롬프트 (기본) | ~2K |
| 시스템 컨텍스트 (OS, CWD 등) | ~200 |
| 모드별 기본 문서 | ~2K |
| 키워드 매칭 문서 | ~4K (동적) |
| **총 지식 버짓** | **~8K tokens** |

qwen3-coder-30b: 256K context → 8K 지식 = 3.1% 사용
gpt-oss-120b: 충분한 context → 8K 지식 = 미미

## 예상 사이즈

| Tier | 파일 수 | 예상 크기 |
|------|---------|----------|
| Tier 0 (BXM) | ~13 | ~2MB |
| Tier 1 | ~18 | ~3.5MB |
| Tier 2 | ~10 | ~2MB |
| Tier 3 | ~6 | ~1MB |
| OS/Terminal | ~4 | ~0.5MB |
| Skills | ~7 | ~0.5MB |
| Index | ~1 | ~50KB |
| **총계** | **~59** | **~9.5MB** |

현재 바이너리 ~15MB → 내장 후 ~25MB (단일 파일)

## OS별 동작

빌드명:
- `techai-windows-amd64.exe` / `techai-windows-arm64.exe`
- `techai-linux-amd64` / `techai-darwin-arm64` / `techai-darwin-amd64`

런타임:
```go
// runtime.GOOS에 따라 터미널 문서 자동 선택
// "IP 주소 알려줘" →
//   windows: ipconfig, Get-NetIPAddress
//   linux:   ip addr, hostname -I
//   darwin:  ifconfig, networksetup
```

## 검증 기준

1. `make build` 후 바이너리에 knowledge/ 포함 확인
2. BXM 관련 질문 → BXM 문서가 컨텍스트에 주입되는지 debug.log 확인
3. Tailwind 관련 질문 → tailwind-v4.md 주입 확인
4. OS별 터미널 질문 → 해당 OS 문서만 주입 확인
5. 토큰 버짓 초과 시 하위 Tier 제거 확인
6. 크로스 플랫폼 빌드 정상 (make build-all)
