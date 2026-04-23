# hanimo MASTER OVERVIEW — 2026-04

> **한 문서로 보는 hanimo 전체상**
> 작성일: 2026-04-11
> 통합 대상: reference-tools-survey / enhancement-plan / sync-tracker / Claude Code 대화 기록 (v0.2.x 브랜딩·Plan·Deep Agent·Certified Model 전략·리서치 디스패치)

---

## 0. TL;DR (30초 요약)

- **hanimo** = Go + Bubble Tea 기반, 한국어 친화 터미널 AI 코딩 에이전트. 현재 **v0.2.x**.
- **차별화 자산 3가지**: ① hash-anchored edit(`hashline_*`), ② 14+ provider 추상화, ③ 한국어 의도 감지 + clarify-first.
- **결정적 약점 3가지**: ① Repo-map 없음, ② Skill/Hooks 없음, ③ Subagent(context fork) 없음.
- **전략 전환**: "무한 호환(14+ provider)" → "**Certified Model System** (Tier 1/2/3, 10~12개)" 로 피벗.
- **페어 프로젝트 TECHAI_CODE**: hanimo가 upstream, TECHAI_CODE는 downstream + 실험 playground. Plan 모드 read-only 깨짐 → 🔥핫픽스 필요.
- **다음 밀수**: ①Plan read-only 패치(TECHAI) ②hashline 포팅 ③clarify-first/prompt embed ④doom-loop detector ⑤Repo-map PoC.

---

## 1. 프로젝트 정체성 & 브랜딩

| 항목 | 내용 |
|---|---|
| 이름 | **hanimo** (HANIMO CODE) |
| 포지셔닝 | 한국어 친화, 다중 provider, 로컬 우선, 단일 바이너리 Go agent |
| 페어 | **TECHAI_CODE** (택가이코드) — downstream/실험판 |
| TUI | Bubble Tea v2 + Lip Gloss, 5 테마, i18n(ko/en) |
| 라이선스·배포 | 단일 Go 바이너리, CGO free 지향 |
| 로고/브랜딩 | Claude Code 대화에서 20종 목업 제작·선정 진행 |
| 모드 체계 | Super / Plan / Deep Agent (구명: Super/Dev/Plan) |

---

## 2. 현재 구현 상태 (v0.2.x as-is)

### 2.1 Core — ✅ 이미 있는 것

| 영역 | 상태 | 근거 파일 |
|---|:---:|---|
| Provider 추상화 (14+: ollama, openai, anthropic, google, openai-compat, novita, openrouter, deepseek, groq, together, fireworks, mistral, vllm, lm studio, custom) | ✅ | `internal/llm/providers/` |
| **Hash-anchored edit** (`hashline_read`/`hashline_edit`) | ✅ | `internal/tools/hashline.go` ⭐ 업계 드문 자산 |
| MCP stdio + SSE transport | ✅ | `internal/mcp/` |
| 3-stage compaction (snip → micro → LLM summary) | ✅ | `internal/llm/compaction.go` |
| Session SQLite (save/load/fork/search) | ✅ | `internal/session/` |
| Knowledge store (extractor + injector) | ✅ | `internal/knowledge/` (Hermes memory 초기형) |
| Agents (Super/Plan/Auto/Intent/AskUser) | ✅ | `internal/agents/` |
| Command palette Ctrl+K | ✅ | `internal/ui/palette.go` |
| i18n (ko/en), 5 themes | ✅ | `internal/ui/` |
| Deep Agent 100-iter loop | ✅ | `ModeDev` |
| 한국어 의도 감지 + clarify-first directive | ✅ | `agents/intent.go`, prompt.go |
| ASK_USER 구조화 질문 | ✅ | `agents/askuser.go` |
| Tool loop detection (기본) | ✅ v0.2.1 |
| Dangerous regex 기본 세트 | ✅ | `internal/tools/shell.go` |

### 2.2 v0.2.1 패치 완료 항목

- `clarifyFirstDirective` 프롬프트 도입 — 모호한 요청에 먼저 질문.
- 한국어 키워드 동작 개선.
- `list_files` skip 로직 보정.
- Shell 타임아웃 env 부분 반영.
- 중복 tool call loop 감지.
- Plan read-only 정책 (hanimo 쪽) 유지.

### 2.3 ❌ 아직 없는 것 (갭 매트릭스)

| 카테고리 | 미구현 | 영향 |
|---|---|---|
| **Context** | Repo-map (tree-sitter+PageRank) | 대형 레포 탐색 비효율 |
| **Context** | Skill 시스템 (SKILL.md) | 반복 작업 자동화 불가 |
| **Agent** | Subagent (context fork) | 토큰 폭발, 탐색·본작업 분리 불가 |
| **Agent** | Doom loop detector (고도화) | 무한루프 완전 차단 X |
| **Agent** | Git Worktree 병렬화 | 동시 실험 불가 |
| **Safety** | Permission 5-mode 엔진 + 학습형 rule | 거친 허용/거부만 존재 |
| **Safety** | Credential scrubbing (output) | 로그 유출 리스크 |
| **Safety** | Sandbox (seatbelt/unshare) | OS 보호막 없음 |
| **Ext** | Hooks 시스템 (8~25 events) | 외부 스크립트 연동 X |
| **Ext** | apply_patch (unified diff) | 다중 hunk 편집 비효율 |
| **Ext** | LSP 통합 (gopls 등) | 정적 분석 힘 부재 |
| **UX** | Session Browser UI + `--resume <id>` | 세션 재개 검색 빈약 |
| **UX** | JSON headless 출력 | CI/CD 통합 어려움 |
| **UX** | DECSET 2026 synchronized output | 스트리밍 깜빡임 |
| **Cost** | Prompt caching (Anthropic cache_control / OpenAI prompt_cache_key) | 최대 90% 입력 비용 낭비 |

---

## 3. 참고한 도구들 — 훔쳐 올 10가지 아이디어

> 출처 문서: `docs/research/reference-tools-survey-2026-04.md`

| 출처 | 아이디어 | hanimo로 가져올 형태 |
|---|---|---|
| **Aider** | Repo-map (tree-sitter + PageRank, 영속 캐시) | `internal/repomap/` — A1 |
| **Claude Code** | SKILL.md 표준 + lazy load | `internal/skills/` — A2 |
| **Claude Code** | Hooks (8~25 events) | `internal/hooks/` — A3 |
| **Claude Code / Crush** | 5-mode permission 엔진 + 학습형 rules | `internal/permission/` — A4 |
| **Claude Code** | Subagent (context fork) + 요약 반환 | `internal/agents/subagent.go` — B1 |
| **Codex CLI** | `apply_patch` unified-diff 포맷 + AST 검증 | `internal/tools/applypatch.go` — B4 |
| **Hermes** | 성공 흐름을 learned skill로 자동 추출 | `skills/extractor.go` |
| **Crush / Continue** | LSP 통합 (gopls/tsserver/pyright) | `internal/lsp/` — C3 |
| **opencode / Gemini CLI** | JSON headless mode | `--output json` — C5 |
| **Claude Code** | Prompt caching (ephemeral cache_control) | `llm/cache.go` — C2 |

### 3.1 법적/ToS 경계 (리서치 결론)

- ❌ **Claude Pro/Max OAuth 재사용**: 2026-04-04 차단, 정책 위배.
- ❌ **Codex CLI / Gemini CLI 토큰 재사용**: ToS 위반.
- ✅ 공식 API 키 (Anthropic/OpenAI/Google), 오픈 라우터 OpenRouter/Novita/DeepSeek/Groq 등은 합법.
- ✅ 로컬 (Ollama, LM Studio, vLLM) 은 자유.

---

## 4. 전략 피벗 — Certified Model System

> "14+ provider 모두 동작" ≠ "14+ provider 모두 잘 동작". 사용자 경험의 급격한 분산을 막기 위한 피벗.

### 4.1 Tier 구조

| Tier | 의미 | 예시 후보 | 보증 수준 |
|---|---|---|---|
| **Tier 1 — Certified** | 매 릴리스 회귀 테스트 통과 | Claude Sonnet 4.x, GPT-5(공식), Gemini 2.5 Pro, qwen3-coder-30b, gpt-oss-120b | 공식 지원 / doom-loop 방지 / 프롬프트 튜닝 적용 |
| **Tier 2 — Supported** | 동작 확인, 일부 feature 제한 | Groq llama-3.3-70b, DeepSeek V3, Mistral Large, Gemma 4 31B | 동작 보장 + 알려진 제한 문서화 |
| **Tier 3 — Experimental** | `--experimental` 플래그 필요 | Novita 신규 모델, 커스텀 endpoint, <7B 로컬 | No SLA, 사용자 책임 |

### 4.2 이 전략이 만드는 변화

- 릴리스 체크리스트에 **모델별 eval 행렬** 도입 (Intent / Plan / Edit / Shell / Compaction).
- Tier 1 모델에만 **전용 system prompt 튜닝** (negative-example, tool description override).
- Novita 429 쏟아짐 같은 문제는 **Tier 2/3 전략으로 격리** → Tier 1 경험 지킴.
- `hanimo models` 명령으로 사용자에게 Tier 시각화.

### 4.3 3-Layer Defense Architecture

1. **Pre-send**: dangerous regex 확장, credential scrubbing, tool call loop hash.
2. **Post-receive**: 모델 출력에서 secret leak 정규식, over-long tool args 차단, JSON 파싱 실패 시 retry-or-ask.
3. **Loop**: doom-loop detector (최근 3회 동일 tool hash → abort), iteration cap env.

---

## 5. 3단계 로드맵 (enhancement-plan-2026-04 요약)

### Phase A — Foundation Upgrade (2~3주)

| 코드 | 항목 | 산출 |
|---|---|---|
| **A1** | Repo-map (tree-sitter + PageRank + SQLite cache) | `internal/repomap/` — cold <300ms / warm <80ms (10K files) |
| **A2** | Skill 시스템 (SKILL.md lazy load, `$ARGUMENTS`, inline `!shell`, learned skills) | `internal/skills/` |
| **A3** | Hooks (8 events: SessionStart/End, Pre/PostToolUse, PermissionRequest, PostCompact, TaskComplete, PostToolUseFailure) | `internal/hooks/` + `.hanimo/hooks.yaml` |
| **A4** | Permission 5-mode + rule 엔진 + 학습형 yaml append | `internal/permission/` + Shift+Tab 순환 |

### Phase B — Agent Orchestration (2~3주)

| 코드 | 항목 |
|---|---|
| **B1** | Subagent (context fork, depth≤2, concurrent≤3, 500-token 요약 반환) |
| **B2** | Git Worktree 병렬 실행 `/parallel` (go-git, CGO-free) |
| **B3** | Plan ↔ Build 권한 분리 + `plan.json` → `/execute` |
| **B4** | `apply_patch` (unified diff, multi-hunk, AST 검증 rollback) |

### Phase C — Polish & Scale (2주)

| 코드 | 항목 |
|---|---|
| **C1** | Session Browser UI + `--resume <id\|-N>` |
| **C2** | Prompt caching (Anthropic `cache_control` / OpenAI `prompt_cache_key`) |
| **C3** | LSP 통합 (gopls 우선) |
| **C4** | Synchronized output mode (DECSET 2026) |
| **C5** | JSON / headless 출력 |

### Phase A 성공 지표

- 10K 파일 레포 cold start ≤ 300ms, warm ≤ 80ms.
- 기본 skill 5개: `review-pr`, `write-tests`, `refactor`, `deploy`, `debug`.
- 예제 hook 3개 동작: gofmt-on-save, block-dangerous, summarize-on-compact.
- Permission 학습 경험 1회 이상 (yaml auto-append).
- `hashline_edit` + `apply_patch` 공존.
- doom-loop detector가 실제 1회 이상 차단.

---

## 6. Immediate Wins — 오늘 당장 (각 <1h)

1. **System prompt 계층화**: `prompts/{core,super,dev,plan}.md` 분리, `//go:embed`.
2. **Tool description에 negative example**: `file_edit: DO NOT use for new files`, `shell_exec: DO NOT use for search`, `hashline_edit: "3#e4d9"` 예시.
3. **Doom loop detector**: 최근 3 iteration tool hash 동일 시 abort.
4. **Dangerous regex 확장**: `export\s+(AWS|OPENAI|ANTHROPIC|GITHUB)_.*KEY`, `curl -H 'Authorization`, `dd if= of=/dev/`, fork bomb.
5. **`.hanimo.md` 상위 디렉토리 탐색** 2단계.
6. **Read-before-write 강제**: 세션 local set, 미조회 파일 편집 시 경고.

---

## 7. TECHAI_CODE 와의 관계 (sync-tracker 요약)

### 7.1 방향성

- **기본 단방향**: hanimo → TECHAI_CODE (catch-up).
- **역방향 PoC**: Skills / Repo-map / Hooks / apply_patch 은 TECHAI_CODE 에서 먼저 실험 → hanimo로 흡수.

### 7.2 Sync 상태 하이라이트

| 항목 | hanimo | TECHAI | 방향 |
|---|:---:|:---:|---|
| hashline tools | ✅ | 📋 | →TECHAI (T1-1 ⭐) |
| Plan 실제 read-only | ✅ | ⚠️ 🔥 | →TECHAI (T1-2 핫픽스) |
| 다중 provider 추상화 | ✅ | 📋 | →TECHAI (T1-5) |
| git/diagnostics tools | ✅ | 📋 | →TECHAI (T1-3,4) |
| MCP stdio/SSE | ✅ | 📋/❌ | →TECHAI (T1-6) |
| Command palette / i18n / themes | ✅ | 📋 | →TECHAI (T3) |
| Repo-map / Skill / Hooks / apply_patch | 📋 | 📋 | TECHAI에서 PoC 먼저 |

### 7.3 커밋 컨벤션

- `port(hashline): from hanimo@<sha>` — 포팅
- `poc(skill): initial loader` — TECHAI 쪽 실험
- `sync(docs): update sync-tracker` — 본 문서/트래커 갱신

### 7.4 🔥 보안 이슈 — TECHAI_CODE Plan 모드

- `internal/llm/models.go` 의 Plan `Tools` 목록에 `file_write`, `file_edit` 포함 → 실제로 파일 수정 가능.
- **조치**: read-only 화이트리스트로 축소. `ReadOnlyTools()` 헬퍼 도입 후 Plan에서 이것만 반환.
- 우선순위 **IW-1 / T1-2 (최우선)**.

---

## 8. 다음 밀수 (Next Drop — 우선 처리 순서)

1. 🔥 **IW-1**: TECHAI_CODE Plan read-only 핫픽스.
2. ⭐ **T1-1**: hashline 포팅 (TECHAI 업그레이드의 상징).
3. **IW-2,3**: prompt embed 분리 + clarifyFirstDirective (양쪽 동시).
4. **T1-3,4**: git/diagnostics 도구 포팅.
5. **Doom-loop detector** — 양쪽 동시.
6. **Repo-map PoC** — TECHAI_CODE에서.
7. **Skill PoC** — TECHAI_CODE에서.
8. **hanimo A1 — Repo-map 풀 구현**.
9. **hanimo A2 — Skill 풀 구현**.
10. **hanimo A3 — Hooks**.

---

## 9. 경쟁 도구 비교표 (요약)

| 기능 | hanimo | Claude Code | Aider | Crush | Codex CLI | opencode | Gemini CLI |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| Multi-provider | ✅(14+) | ⚠️ | ✅ | ✅ | ❌ | ✅ | ❌ |
| Hash-anchored edit | ✅⭐ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Repo-map (tree-sitter) | ❌ | ⚠️ | ✅ | ⚠️ | ❌ | ⚠️ | ⚠️ |
| Skill system | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Hooks | ❌ | ✅ | ❌ | ⚠️ | ❌ | ❌ | ❌ |
| Subagent / context fork | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| apply_patch unified-diff | ❌ | ⚠️ | ✅ | ⚠️ | ✅ | ⚠️ | ⚠️ |
| MCP | ✅ | ✅ | ❌ | ✅ | ❌ | ✅ | ⚠️ |
| 한국어 intent/clarify | ✅⭐ | ⚠️ | ❌ | ❌ | ❌ | ❌ | ⚠️ |
| TUI 품질 (Bubble Tea) | ✅ | ✅ | ⚠️ | ✅ | ⚠️ | ✅ | ⚠️ |
| Permission 5-mode | ❌ | ✅ | ❌ | ⚠️ | ⚠️ | ⚠️ | ❌ |
| LSP 통합 | ❌ | ⚠️ | ❌ | ✅ | ❌ | ⚠️ | ❌ |
| Prompt caching | ❌ | ✅ | ⚠️ | ⚠️ | ⚠️ | ⚠️ | ⚠️ |

**결론**: hanimo의 유일무이한 자산은 **hashline_edit + 한국어 친화** 둘. Repo-map/Skill/Hooks/Subagent 4 종 세트를 잡으면 상위 티어 진입.

---

## 10. 즉시 가치 Top 5 (투입 대비 효과)

1. **Repo-map (A1)** — 대형 레포 UX 결정적, 비용 절감, 한 번 해두면 모든 세션 이득.
2. **Skill 시스템 (A2)** — 사용자 락인, 반복 작업 자동화, Claude Code 표준 호환.
3. **Hooks (A3)** — 에이전트를 조직 워크플로에 결합, gofmt/test/lint 자동 연동.
4. **Subagent (B1)** — 토큰 폭발 억제, Deep Agent 품질 2배.
5. **Prompt caching (C2)** — 비용 90% 절감, 1일 안에 구현 가능.

---

## 11. 파일 / 패키지 매핑 (구현 가이드)

```
hanimo/internal/
├── repomap/       [A1] parser.go graph.go rank.go cache.go inject.go
├── skills/        [A2] loader.go frontmatter.go registry.go invoke.go extractor.go
├── hooks/         [A3] events.go loader.go executor.go handlers/{command,http,prompt,skill}.go
├── permission/    [A4] engine.go rules.go modes.go
├── agents/
│   ├── subagent.go    [B1]
│   └── summarizer.go
├── worktree/      [B2] manager.go
├── tools/
│   └── applypatch.go  [B4]
└── lsp/           [C3] client.go gopls.go
```

### 기존 파일 수정 지점

| 파일 | 수정 |
|---|---|
| `internal/app/app.go` | hook 훅 5곳, subagent spawn 라우팅, permission 주입 |
| `internal/llm/prompt.go` | embed 분리, repo-map/skill description 주입 |
| `internal/llm/compaction.go` | PostCompact hook + cache breakpoint |
| `internal/tools/registry.go` | 새 도구 등록 (repomap_*, spawn_subagent, apply_patch) |
| `internal/tools/shell.go` | dangerous regex 확장 + credential scrub |
| `internal/tools/file.go` | read-before-write 강제 |
| `internal/session/store.go` | fork, worktree 링크 |
| `internal/config/config.go` | hooks / permissions / skills_dir / repomap 섹션 |
| `internal/ui/palette.go` | skill / hook / session browser 엔트리 |
| `cmd/hanimo/main.go` | `--resume`, `--output json`, `--permission` 플래그 |

---

## 12. 참고 문서 (이 MASTER의 원본)

- `docs/research/reference-tools-survey-2026-04.md` — 10개 도구 상세 조사 + 출처.
- `docs/roadmap/enhancement-plan-2026-04.md` — Phase A/B/C 풀 플랜.
- `docs/porting/sync-tracker-2026-04.md` — hanimo ↔ TECHAI_CODE 1:1 매핑.
- (TECHAI_CODE 쪽) `TECHAI_CODE/docs/roadmap/enhancement-plan-2026-04.md` — Track 1 catch-up / Track 2 PoC / Track 3 UX parity.

---

_Last updated: 2026-04-11 · hanimo v0.2.x 기준_
