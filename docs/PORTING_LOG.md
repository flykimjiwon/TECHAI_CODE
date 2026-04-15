# 포팅 기록: hanimo → 택가이코드

> 원본: `github.com/flykimjiwon/dev_anywhere` (hanimo)
> 대상: `github.com/kimjiwon/tgc` (택가이코드)
> 작성자: 동일인 (같은 Go + Bubble Tea v2 스택)

---

## Tier 1 포팅 (2026-04-10, 커밋 `0b6361f`)

6개 기능, +1,141줄 추가.

### 1. Hash-Anchored Editing
- **원본**: `hanimo/internal/tools/hashline.go` (116줄)
- **포팅**: `택가이코드/internal/tools/hashline.go` (116줄)
- **변경**: 임포트 경로 `hanimo` → `tgc` 교체만. 코드 동일.
- **기능**: MD5 4자리 해시 앵커로 파일 줄 식별, stale-edit 방지.

### 2. 자율 모드 (Auto)
- **원본**: `hanimo/internal/agents/auto.go` (32줄)
- **포팅**: `택가이코드/internal/agents/auto.go` (32줄)
- **변경**: 코드 동일. `[AUTO_COMPLETE]`/`[AUTO_PAUSE]` 마커 기반 제어.

### 3. Git 도구
- **원본**: `hanimo/internal/tools/git.go` (52줄)
- **포팅**: `택가이코드/internal/tools/git.go` (52줄)
- **변경**: 임포트 경로만 교체. GitStatus, GitDiff, GitLog 함수.

### 4. 코드 진단 (Diagnostics)
- **원본**: `hanimo/internal/tools/diagnostics.go` (272줄)
- **포팅**: `택가이코드/internal/tools/diagnostics.go` (272줄)
- **변경**: `fileExists` → `diagFileExists` 이름 변경 (패키지 내 충돌 방지). Go/TS/JS/Python 린터 자동 감지.

### 5. 커맨드 팔레트 (Ctrl+K)
- **원본**: `hanimo/internal/ui/palette.go` (154줄)
- **포팅**: `택가이코드/internal/ui/palette.go` (144줄)
- **변경**: PaletteItems를 택가이코드 슬래시 명령으로 재정의 (한국어 라벨). i18n `T()` 의존성 제거.

### 6. 메뉴 오버레이 (Esc)
- **원본**: `hanimo/internal/ui/menu.go` (98줄)
- **포팅**: `택가이코드/internal/ui/menu.go` (107줄)
- **변경**: i18n `T()` 제거, 한국어 하드코딩. `submenu` 파라미터 제거. `MainMenuItems`/`MenuActionFromIndex` 추가. 타이틀 "hanimo" → "택가이코드".

---

## Tier 2 포팅 (2026-04-13)

3모드 시스템 재설계 + 환경 감지 + 유저 문서 RAG.

### 7. 3모드 시스템 재설계 (Super / Deep Agent / Plan)
- **원본**: `hanimo/internal/llm/prompt.go` + `hanimo/internal/llm/prompts/*.md`
- **포팅 파일들**:
  - `택가이코드/internal/llm/prompt.go` — 전면 재작성
  - `택가이코드/internal/llm/prompts/core.md` — hanimo에서 복사
  - `택가이코드/internal/llm/prompts/super.md` — hanimo 기반, 도구 목록 확장 (hashline, git, diagnostics, knowledge_search)
  - `택가이코드/internal/llm/prompts/dev.md` — hanimo 기반, 도구 목록 확장
  - `택가이코드/internal/llm/prompts/plan.md` — hanimo 기반, 읽기전용→쓰기가능으로 변경
  - `택가이코드/internal/llm/prompts/askuser.md` — hanimo에서 복사
- **주요 변경**:
  - 모드명: 슈퍼택가이→**Super**, 개발→**Deep Agent**, 플랜→**Plan**
  - Go 인라인 문자열 → `//go:embed prompts/*.md` 파일 시스템
  - Core 지침 추가: "Clarify Before Acting" + ASK_USER 프로토콜
  - `ModeDeep` 별칭 추가 (하위 호환)
  - Deep Agent: 자율 100회 반복, `[TASK_COMPLETE]` 마커
  - Plan: 읽기전용 → 전체 도구 (쓰기 가능, 승인 후 실행)
  - 모든 모드가 super 모델(gpt-oss-120b) 사용 (hanimo 동일)

### 8. 환경 자동 감지 (Environment Probe)
- **원본**: `hanimo/internal/llm/environment.go` (187줄)
- **포팅**: `택가이코드/internal/llm/environment.go` (148줄)
- **변경**: 임포트 경로 교체. 이모지(✅❌) → 텍스트(Installed/Not installed)로 변경 (호환성).
- **기능**: 40+ 도구 설치 여부 감지 (node, npm, python3, go, rust, java, docker, git, aws, kubectl 등). 시스템 프롬프트에 주입하여 LLM이 미설치 도구 사용 방지.
- **동작**: 앱 시작 시 1회 실행 (~200ms), 결과 캐싱.

### 9. 유저 문서 RAG (User Knowledge Docs)
- **원본**: `hanimo/internal/knowledge/userdocs.go` (276줄)
- **포팅**: `택가이코드/internal/knowledge/userdocs.go` (221줄)
- **변경**:
  - 폴더 경로: `.hanimo/knowledge/` → `.tgc/knowledge/` (프로젝트 로컬), `~/.hanimo/knowledge/` → `~/.tgc/knowledge/` (글로벌)
  - 함수명: `parseDoc` → `parseUserDoc` (기존 knowledge 패키지와 충돌 방지)
  - 함수명: `splitTerms` → `splitSearchTerms` (충돌 방지)
  - 유니코드 `…` → ASCII `...` (호환성)
  - `—` → `--` (호환성)
- **기능**: `.tgc/knowledge/` 폴더에 `.md`/`.txt` 파일을 넣으면 자동 인덱싱. 제목, 헤더, 첫 단락 파싱. `knowledge_search` 도구로 키워드 검색. 시스템 프롬프트에 목차 자동 주입.

### 10. knowledge_search 도구 등록
- **원본**: `hanimo/internal/tools/registry.go` (knowledge_search 케이스)
- **포팅**: `택가이코드/internal/tools/registry.go`에 추가
  - `AllTools()`에 knowledge_search 도구 정의 추가
  - `executeInner()`에 knowledge_search 케이스 추가
  - `ExecuteKnowledgeSearch()` 함수 추가
  - `knowledge` 패키지 임포트 추가

### 11. Deep Agent 자율 실행 로직
- **원본**: `hanimo/internal/app/app.go` (Deep Agent 모드 자율 루프)
- **포팅**: `택가이코드/internal/app/app.go` 수정
  - `activeTab == 1` (Deep Agent)일 때 자동 자율 모드 동작
  - Deep Agent: `MaxDeepIterations` = 100회 (Super/Plan의 /auto는 20회 유지)
  - `[TASK_COMPLETE]` + `[AUTO_COMPLETE]` 양쪽 마커 인식
  - `currentModel()`: 모든 모드가 super 모델 사용 (hanimo 동일)

### 12. app.go 통합 변경
- env probe 결과를 projectCtx에 주입 (시스템 프롬프트)
- user docs 목차를 projectCtx에 주입 (시스템 프롬프트)
- `knowledge.GlobalIndex` 글로벌 변수에 인덱스 할당

### 13. UI 스타일 변경
- **원본**: `hanimo/internal/ui/styles.go` — `DeepColor`, `DevColor = DeepColor`
- **포팅**: `택가이코드/internal/ui/styles.go`에 추가
  - `DeepColor = ColorSuccess` (민트 그린)
  - `DevColor = DeepColor` (하위 호환 별칭)

---

## 포팅 통계

| 라운드 | 날짜 | 신규 파일 | 수정 파일 | 추가 줄수 | 기능 수 |
|--------|------|----------|----------|----------|--------|
| Tier 1 | 2026-04-10 | 6 | 3 | +1,141 | 6 |
| Tier 2 | 2026-04-13 | 8 | 4 | +~900 | 7 |
| **합계** | | **14** | **7** | **~2,041** | **13** |

## 미포팅 (hanimo 전용)

| 기능 | hanimo 파일 | 이유 |
|------|------------|------|
| MCP 클라이언트 | `internal/mcp/` | 택가이코드는 내부망 전용, MCP 불필요 |
| 멀티 프로바이더 | `internal/llm/providers/` | 택가이코드는 OpenAI-compat 단일 엔드포인트 |
| 테마 시스템 | `internal/ui/styles.go` Themes | 우선순위 낮음 |
| 페르소나 | `internal/ui/persona.go` | 우선순위 낮음 |
| i18n 다국어 | `internal/ui/i18n.go` | 택가이코드는 한국어 고정 |
| 스킬 로더 | `internal/skills/` | 우선순위 낮음 |
| 프롬프트 캐시 | `internal/llm/cache.go` | 차후 포팅 예정 |
| 사용량 통계 | `internal/config/cache_stats.go` | 차후 포팅 예정 |
