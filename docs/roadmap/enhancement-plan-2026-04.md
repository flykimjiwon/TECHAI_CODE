# TECHAI_CODE (택가이코드) Enhancement Plan — 2026-04

> **작성일**: 2026-04-10
> **기반 조사**: `docs/research/reference-tools-survey-2026-04.md`
> **목표**: (1) hanimo가 이미 가진 핵심 자산을 따라잡고, (2) TECHAI_CODE를 "빠르고 단순한 실험장"으로 포지셔닝해 신기능의 초기 검증지로 삼는다.
> **페어**: `../../hanimo/` — hanimo가 upstream, TECHAI_CODE는 가벼운 downstream.

---

## 0. 현재 상태 점검

| 영역 | 상태 | 비고 |
|---|---|---|
| LLM provider | ⚠️ OpenAI-compatible 1종만 | `sashabaranov/go-openai` 고정. hanimo는 14+ |
| Tool 세트 | ⚠️ 7개 (file/shell/grep/glob) | git, diagnostics, hashline 없음 |
| Hash-anchored edit | ❌ 없음 | hanimo에는 있음 |
| MCP | ❌ 없음 | |
| 3-stage compaction | ✅ 있음 | `internal/llm/compaction.go` 185 line |
| Sessions | ✅ SQLite store | hanimo 대비 simpler |
| Knowledge store | ✅ 있음 | extractor + injector |
| Gitinfo 패키지 | ✅ 있음 | `internal/gitinfo/` |
| Agents (plan/auto/intent) | ❌ 없음 | 현재 ModeDev 자율 루프는 단순 |
| Command palette | ❌ 없음 | |
| i18n / Theme | ❌ 없음 | |
| Modes | ✅ 3개 (Super/Dev/Plan) | Tools는 모두 동일 — Plan이 read-only가 아님 ⚠️ |

---

## 1. 포지셔닝

**hanimo**: 완성형, 다다익선, 레퍼런스 구현.
**TECHAI_CODE**: 빠른 실험장, 단일 provider(gpt-oss-120b)에 집중, 새 아이디어 PoC 먼저.

이 차이를 살리면:
- hanimo에서 검증된 기능은 *안정적 포팅*.
- 새 아이디어는 TECHAI_CODE에서 PoC → 검증 후 hanimo에 머지.

---

## 2. 로드맵

### Track 1 — Catch-up from hanimo (우선, 2~3주)

Simpler downstream이 *반드시* 따라붙어야 할 항목들.

#### T1-1. Hash-anchored edit 포팅 ⭐⭐⭐⭐⭐
- **출처**: `hanimo/internal/tools/hashline.go`
- **이유**: 모델이 stale edit으로 파일 손상시키는 케이스를 구조적으로 차단. 업계에 드문 자산이므로 반드시 이식.
- **작업**:
  1. `internal/tools/hashline.go` 신설 (hanimo 파일 참조해 복사).
  2. `internal/tools/registry.go` 에 `hashline_read`, `hashline_edit` 도구 정의 추가.
  3. `SystemPrompt` 에 hash anchor 사용 가이드 추가.
  4. `Plan` 모드 읽기 전용 도구에도 `hashline_read` 포함.

#### T1-2. Plan 모드 실제 read-only 화
- 현재 `Modes[ModePlan].Tools` 가 `file_write`, `file_edit` 을 그대로 포함 ⚠️.
- `ReadOnlyTools()` 함수 신설 (hanimo 패턴 참조) → `file_read`, `list_files`, `grep_search`, `glob_search`, `shell_exec(readonly)`, `hashline_read` 만.
- `ToolsForMode(int)` 라우팅 함수 추가.

#### T1-3. Git 도구 포팅
- **출처**: `hanimo/internal/tools/git.go`
- 도구: `git_status`, `git_diff`, `git_log`, `git_commit`.
- 기존 `internal/gitinfo/` 는 유지하되, 도구 레이어는 `tools/git.go` 신설.

#### T1-4. Diagnostics 도구 포팅
- **출처**: `hanimo/internal/tools/diagnostics.go`
- `go vet`, `tsc --noEmit`, `eslint`, `ruff` auto-detect.

#### T1-5. Provider 추상화
- **출처**: `hanimo/internal/llm/providers/`
- 현재 `sashabaranov/go-openai` 직결 → `Provider` interface 도입.
- 1단계: openai-compat만 구현 (기존 동작 유지).
- 2단계: ollama, anthropic 순차 추가 (hanimo 파일 그대로 복사 가능).
- 목표: hanimo와 `internal/llm/providers/` 동일 구조 → 이후 포팅 제로코스트.

#### T1-6. MCP (stdio) 포팅
- **출처**: `hanimo/internal/mcp/`
- 단일 바이너리 원칙 유지: 외부 MCP 서버는 사용자가 직접 실행.
- 설정: `~/.tgc/config.yaml` 에 `mcp_servers` 섹션 추가.

#### T1-7. Agents (plan / auto) 포팅
- **출처**: `hanimo/internal/agents/`
- 현재 `ModeDev` 자율 루프는 `app.go` 에 인라인 → 패키지 분리.
- `auto.go` 의 doom loop detector 포함 포팅.

#### T1-8. Dangerous command regex 강화
- `internal/tools/shell.go` 에 정규식 추가:
  - `rm\s+-rf\s+/`, `sudo`, `export\s+(AWS|OPENAI|ANTHROPIC|GITHUB)_.*KEY`
  - `curl\s+.*-H\s+['"]Authorization`, `:\(\)\{\s*:\|:&\s*\};:` (fork bomb)
- 기존 검사 로직과 병합.

---

### Track 2 — Forward PoC Ground (병렬 2주)

TECHAI_CODE가 **먼저** 실험할 기능들. 단순한 코드베이스라 이터레이션 빠름.

#### T2-1. Skill 시스템 PoC ⭐
- 1단계(MVP): `.techai/skills/*.md` 단일 파일 형식. frontmatter 없이 `# Name` 첫 줄 = 이름.
- 2단계: YAML frontmatter.
- 3단계: lazy loading.
- 목적: hanimo에 머지하기 전 UX 테스트.

#### T2-2. Hooks 시스템 PoC
- 1단계: `.techai/hooks.yaml` 에 PreToolUse / PostToolUse 2종만.
- handler는 `command` 타입만.
- 검증 후 hanimo에 머지.

#### T2-3. Repo-map PoC (tree-sitter)
- Go 한 언어만 타깃.
- `smacker/go-tree-sitter` + `tree-sitter-go`.
- 심볼 추출 → 메모리 map → PageRank 없이 단순 top-N.
- SQLite 캐시 없이 파일 기반 cache 먼저.
- 성공 시 hanimo에 풀 포팅.

#### T2-4. `apply_patch` 도구
- Codex CLI 포맷 (`*** Begin Patch` 등).
- 여러 hunk 파싱기 구현.
- tree-sitter validation은 PoC 단계에선 생략.

---

### Track 3 — UX Parity (선택, 1~2주)

#### T3-1. Command Palette (Ctrl+K)
- **출처**: `hanimo/internal/ui/palette.go`
- fuzzy search + 모든 슬래시 명령 접근.

#### T3-2. i18n (ko/en 토글)
- **출처**: `hanimo/internal/ui/i18n.go`
- 구조만 복사, 문자열은 점진 이관.

#### T3-3. Themes
- hanimo 5 themes 중 2개만 먼저 (honey + dracula).

#### T3-4. TAB 키 분리 (mode vs permission)
- 현재 Tab이 mode 순환만. `Shift+Tab` 에 permission 순환 매핑.

---

## 3. 즉시 개선 (Immediate, 각 <30min)

현재 코드 기준 당장 반영 가능한 수정.

### IW-1. Plan 모드에서 쓰기 도구 제거 ⚠️ 보안
```go
// internal/llm/models.go
ModePlan: {
    // Tools: []string{"grep_search", "glob_search", "file_read", "file_write", "file_edit", "list_files", "shell_exec"},
    Tools: []string{"grep_search", "glob_search", "file_read", "list_files", "shell_exec"},
},
```
+ `tools/registry.go` 에 `ReadOnlyTools()` 함수 추가.

### IW-2. System prompt 계층화
`prompt.go` 를 3개 상수 단일 파일로 유지하지 말고:
```
internal/llm/prompts/
├── core.md         (공통 directive)
├── super.md
├── dev.md
└── plan.md
```
`//go:embed` 로 포함. 수정·diff가 훨씬 쉬워짐.

### IW-3. `clarifyFirstDirective` 포팅
hanimo `prompt.go` 68~108 라인의 "ASK_USER first" 블록을 TECHAI_CODE에도 추가. Plan 모드는 특히 필요.

### IW-4. Tool description에 negative example 추가
- `file_edit`: "DO NOT use for creating new files; use file_write instead."
- `file_write`: "DO NOT use to modify existing files; use file_edit."
- `shell_exec`: "DO NOT use grep/find here; use grep_search/glob_search."

### IW-5. Read-before-write 세션 캐시
`tools/file.go` 에 session-local `readFiles map[string]bool` 유지 → `file_edit`/`file_write` 시 경고.

### IW-6. Doom loop detector
`app.go` 자율 루프에서 최근 3회 tool call hash 비교 → 동일 시 abort.

### IW-7. Shell 30초 타임아웃 문서화 + 환경변수화
현재 `Execute` 에서 하드코딩된 30초 → `TECHAI_SHELL_TIMEOUT=60s` 환경변수로 override 가능하게.

### IW-8. Dangerous regex 목록 확장
위 T1-8 참조.

---

## 4. hanimo 대비 차이 테이블 (따라갈 순서)

우선 순위대로.

| # | 기능 | hanimo | TECHAI_CODE | 난이도 | 우선 |
|---|---|---|---|---|---|
| 1 | Plan 읽기 전용 | ✅ | ❌ (쓰기 허용) | 🟢 쉬움 | 🔥 보안 |
| 2 | Hash-anchored edit | ✅ | ❌ | 🟡 중간 | ⭐⭐⭐⭐⭐ |
| 3 | Git 도구 (status/diff/log/commit) | ✅ | ❌ | 🟢 쉬움 | ⭐⭐⭐⭐ |
| 4 | Diagnostics 도구 | ✅ | ❌ | 🟢 쉬움 | ⭐⭐⭐⭐ |
| 5 | clarifyFirstDirective | ✅ | ❌ | 🟢 쉬움 | ⭐⭐⭐⭐ |
| 6 | Provider 추상화 | ✅ | ❌ | 🟡 중간 | ⭐⭐⭐ |
| 7 | MCP (stdio) | ✅ | ❌ | 🔴 어려움 | ⭐⭐⭐ |
| 8 | Agents (plan/auto 분리) | ✅ | ❌ | 🟡 중간 | ⭐⭐⭐ |
| 9 | Command palette | ✅ | ❌ | 🟡 중간 | ⭐⭐ |
| 10 | i18n / Theme | ✅ | ❌ | 🟢 쉬움 | ⭐ |
| 11 | Deep Agent 100-iter | ✅ | ⚠️ 부분 | 🟡 중간 | ⭐⭐⭐ |

---

## 5. 새 파일/패키지 매핑 (Track 1 기준)

```
internal/
├── tools/
│   ├── hashline.go     [T1-1] hanimo에서 복사
│   ├── git.go          [T1-3] hanimo에서 복사
│   └── diagnostics.go  [T1-4] hanimo에서 복사
├── llm/
│   ├── providers/      [T1-5] 패키지 신설
│   │   ├── registry.go
│   │   └── openai_compat.go
│   └── prompts/        [IW-2] embed dir
│       ├── core.md
│       ├── super.md
│       ├── dev.md
│       └── plan.md
├── mcp/                [T1-6] stdio client
│   ├── client.go
│   └── transport_stdio.go
└── agents/             [T1-7] plan/auto 분리
    ├── plan.go
    ├── auto.go
    └── intent.go
```

---

## 6. 기존 파일 수정 지점

| 파일 | 수정 |
|---|---|
| `internal/llm/models.go` | Plan `Tools` 목록에서 쓰기 도구 제거 |
| `internal/llm/prompt.go` | embed 분리, clarifyFirstDirective 추가 |
| `internal/tools/registry.go` | ReadOnlyTools() 추가, hashline/git/diagnostics 도구 등록 |
| `internal/tools/shell.go` | dangerous regex 확장, 타임아웃 env |
| `internal/tools/file.go` | read-before-write 세션 캐시 |
| `internal/app/app.go` | agents 패키지 분리, doom loop detector |
| `internal/config/config.go` | `mcp_servers`, `skills_dir`, `hooks` 섹션 (PoC 단계) |
| `cmd/tgc/main.go` | `--resume`, `--permission` 플래그 |

---

## 7. 성공 지표 (Track 1 완료)

- [ ] Plan 모드에서 `file_write`, `file_edit` 시도가 명시적으로 차단됨.
- [ ] `hashline_read` / `hashline_edit` 정상 동작, stale read 시 에러 메시지.
- [ ] `git_status`/`git_diff`/`git_log`/`git_commit` 4종 모두 통과.
- [ ] `diagnostics` 가 go/ts/py 프로젝트에서 자동 감지.
- [ ] Provider interface 도입, 기존 테스트 모두 통과.
- [ ] `~/.tgc/config.yaml` 에 MCP 서버 선언 시 stdio 연결 성공.
- [ ] 모든 단위 테스트 통과 + `compaction_test.go`, `context_test.go`, `capabilities_test.go`, `usage_test.go` 깨지지 않음.

---

## 8. Track 2 PoC → hanimo 머지 프로세스

TECHAI_CODE에서 먼저 검증할 기능 흐름:

1. **PoC 구현** (TECHAI_CODE, 1~2 iteration).
2. 실 사용 1주일 dogfooding.
3. 문제점/개선점 회고.
4. hanimo에 이식 (generic화, 설정화, 에러 핸들링 강화).
5. 양쪽 문서 업데이트.

우선 실험 후보:
- **Skill 시스템**: 포맷/로딩 UX 확정 전에 여러 변형 시도.
- **Repo-map**: tree-sitter 언어별 query 튜닝이 오래 걸림, 단일 언어로 시작.
- **Hooks**: PreToolUse block 타이밍이 까다로움, command handler만 먼저.
- **apply_patch**: diff 파서 엣지케이스 탐색.

---

## 9. 참고 구현 바로가기 (hanimo 파일 경로)

```
../hanimo/internal/tools/hashline.go           → T1-1
../hanimo/internal/tools/git.go                → T1-3
../hanimo/internal/tools/diagnostics.go        → T1-4
../hanimo/internal/llm/providers/              → T1-5
../hanimo/internal/llm/providers/openai_compat.go
../hanimo/internal/llm/providers/ollama.go
../hanimo/internal/llm/providers/anthropic.go
../hanimo/internal/llm/providers/google.go
../hanimo/internal/mcp/                        → T1-6
../hanimo/internal/mcp/client.go
../hanimo/internal/mcp/transport_stdio.go
../hanimo/internal/agents/                     → T1-7
../hanimo/internal/agents/plan.go
../hanimo/internal/agents/auto.go
../hanimo/internal/agents/intent.go
../hanimo/internal/agents/askuser.go
../hanimo/internal/llm/prompt.go               → IW-2, IW-3
../hanimo/internal/ui/palette.go               → T3-1
../hanimo/internal/ui/i18n.go                  → T3-2
```

---

## 10. 다음 액션 (구체 순서)

1. 본 문서 리뷰 + 우선순위 사인오프.
2. **IW-1 (Plan read-only)** 즉시 반영 — 보안.
3. **IW-2 ~ IW-8** 1일 스프린트로 마무리.
4. **T1-1 (hashline 포팅)** → 2일.
5. **T1-2 ~ T1-4 (Plan read-only function, git, diagnostics)** → 2일.
6. **T1-5 (provider interface)** → 3일.
7. **T1-6 (MCP stdio)** → 3일.
8. **T1-7 (agents 패키지 분리)** → 2일.
9. **Track 2 PoC 착수**: Skill 또는 Repo-map 중 하나 선정.

---

## 11. 동반 문서

- `docs/research/reference-tools-survey-2026-04.md` — 본 계획의 근거 조사.
- `../hanimo/docs/roadmap/enhancement-plan-2026-04.md` — 쌍(pair) 문서.
- 향후: `docs/porting/hanimo-sync-tracker.md` — 포팅 진행 체크리스트.
