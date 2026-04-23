# hanimo ↔ TECHAI_CODE Sync Tracker — 2026-04

> **목적**: 두 프로젝트 간 기능 포팅 상태를 한 눈에. 양쪽 동일 파일로 유지(쌍방 복사).
> **업데이트**: 항목 이동 시 즉시. PR/커밋 단위로 tick.

---

## Legend

- ✅ 구현 완료
- 🚧 작업 중
- 📋 계획됨
- ❌ 미정
- ⚠️ 취약점/불일치

---

## 1. Core LLM & Tools

| 항목 | hanimo | TECHAI_CODE | 방향 | 비고 |
|---|:---:|:---:|:---:|---|
| OpenAI-compatible provider | ✅ | ✅ | — | 양쪽 기본 |
| Ollama provider | ✅ | 📋 | →TECHAI | T1-5 |
| Anthropic provider | ✅ | 📋 | →TECHAI | T1-5 |
| Google provider | ✅ | 📋 | →TECHAI | T1-5 |
| Provider interface 추상화 | ✅ | 📋 | →TECHAI | T1-5 |
| 3-stage compaction | ✅ | ✅ | ⇄ | 둘 다 유사 |
| Context window tracker | ✅ | ✅ | ⇄ | |
| Prompt caching (anthropic) | 📋 | ❌ | hanimo 먼저 | C2 |
| Tool: file_read/write/edit | ✅ | ✅ | ⇄ | |
| Tool: grep/glob | ✅ | ✅ | ⇄ | |
| Tool: shell_exec | ✅ | ✅ | ⇄ | dangerous regex 강화 필요 |
| Tool: hashline_read/edit | ✅ | 📋 | →TECHAI | T1-1 ⭐ |
| Tool: git_status/diff/log/commit | ✅ | 📋 | →TECHAI | T1-3 |
| Tool: diagnostics | ✅ | 📋 | →TECHAI | T1-4 |
| Tool: apply_patch (unified diff) | 📋 | 📋 | TECHAI 먼저 PoC | T2-4 |
| Read-before-write enforcement | 📋 | 📋 | 양쪽 동시 | IW |

---

## 2. Agent System

| 항목 | hanimo | TECHAI_CODE | 방향 | 비고 |
|---|:---:|:---:|:---:|---|
| Super / Dev / Plan 3-mode | ✅ | ✅ | ⇄ | |
| Plan 실제 read-only | ✅ | ⚠️ | →TECHAI | T1-2 🔥 |
| Agents 패키지 분리 | ✅ | 📋 | →TECHAI | T1-7 |
| `/auto` deep agent | ✅ | ⚠️ 부분 | →TECHAI | T1-7 |
| `intent.go` 의도 감지 | ✅ | ❌ | →TECHAI | T1-7 |
| `askuser.go` ASK_USER | ✅ | ⚠️ 미구조화 | →TECHAI | |
| Doom loop detector | 📋 | 📋 | 양쪽 동시 | IW |
| Iteration cap 환경변수 | ⚠️ 하드코딩 | ⚠️ 하드코딩 | 양쪽 동시 | |
| Subagent (context fork) | 📋 | 📋 | hanimo 먼저 | B1 |
| Git worktree 병렬화 | 📋 | ❌ | hanimo 먼저 | B2 |

---

## 3. Context Engineering

| 항목 | hanimo | TECHAI_CODE | 방향 | 비고 |
|---|:---:|:---:|:---:|---|
| `.hanimo.md` / `.techai.md` 로드 | ✅ | ⚠️ | →TECHAI | |
| clarifyFirstDirective | ✅ | 📋 | →TECHAI | IW-3 |
| 계층형 프롬프트 (embed 분리) | 📋 | 📋 | 양쪽 동시 | IW-2 |
| Repo-map (tree-sitter) | 📋 | 📋 | TECHAI PoC → hanimo | A1 / T2-3 |
| Knowledge store (SQLite) | ✅ | ✅ | ⇄ | |
| Skill 시스템 | 📋 | 📋 | TECHAI PoC → hanimo | A2 / T2-1 |
| Learned-skill 자동 추출 | 📋 | ❌ | hanimo 먼저 | A2 |
| AGENTS.md 표준 지원 | 📋 | 📋 | 양쪽 동시 | |

---

## 4. Extensibility

| 항목 | hanimo | TECHAI_CODE | 방향 | 비고 |
|---|:---:|:---:|:---:|---|
| MCP stdio transport | ✅ | 📋 | →TECHAI | T1-6 |
| MCP SSE transport | ✅ | ❌ | →TECHAI | |
| MCP HTTP streaming | 📋 | ❌ | hanimo 먼저 | |
| Hooks 시스템 | 📋 | 📋 | TECHAI PoC → hanimo | A3 / T2-2 |
| Permission rule 엔진 | 📋 | 📋 | hanimo 먼저 | A4 |
| Permission 5-mode 순환 | 📋 | 📋 | hanimo 먼저 | A4 |
| permissions.yaml 학습형 | 📋 | ❌ | hanimo 먼저 | A4 |

---

## 5. UX / TUI

| 항목 | hanimo | TECHAI_CODE | 방향 | 비고 |
|---|:---:|:---:|:---:|---|
| Bubble Tea 기반 | ✅ | ✅ | ⇄ | |
| Command palette (Ctrl+K) | ✅ | 📋 | →TECHAI | T3-1 |
| i18n (ko/en) | ✅ | 📋 | →TECHAI | T3-2 |
| Themes (5개) | ✅ | 📋 | →TECHAI | T3-3 |
| Session 저장 (SQLite) | ✅ | ✅ | ⇄ | |
| Session Browser UI | 📋 | ❌ | hanimo 먼저 | C1 |
| `--resume <id>` CLI | 📋 | 📋 | 양쪽 동시 | |
| JSON / headless 출력 | 📋 | 📋 | 양쪽 동시 | C5 |
| Synchronized Output (DECSET 2026) | 📋 | 📋 | 양쪽 동시 | C4 |
| LSP 통합 (gopls 등) | 📋 | ❌ | hanimo 먼저 | C3 |
| Tab / Shift+Tab 분리 | ⚠️ | ⚠️ | 양쪽 동시 | A4 |

---

## 6. Safety

| 항목 | hanimo | TECHAI_CODE | 방향 | 비고 |
|---|:---:|:---:|:---:|---|
| Dangerous regex 기본 | ✅ | ⚠️ 부분 | →TECHAI | T1-8 |
| 확장 dangerous regex (env leak 등) | 📋 | 📋 | 양쪽 동시 | IW-8 / T1-8 |
| Credential scrubbing (output) | 📋 | 📋 | hanimo 먼저 | |
| Sandbox (seatbelt/unshare) | ❌ | ❌ | hanimo 먼저 | C-future |
| Shell timeout env 가변 | ⚠️ 하드 | ⚠️ 하드 | 양쪽 동시 | IW-7 |

---

## 7. 포팅 규칙

1. **단방향 기본**: hanimo → TECHAI_CODE. 코드 베이스 스타일 일치를 위해 파일 이름/패키지명 동일 유지.
2. **PoC 역방향**: Skill, Repo-map, Hooks, apply_patch는 TECHAI_CODE에서 먼저 PoC → 검증 후 hanimo 포팅. 포팅 시 generic화 + 에러 핸들링 + 설정화 추가.
3. **양쪽 커밋 메시지 컨벤션**:
   - `port(hashline): from hanimo@<sha>` — 포팅 시
   - `poc(skill): initial loader` — PoC 시
   - `sync(docs): update sync-tracker` — 본 문서 업데이트
4. **테스트 동반**: 포팅 시 원본 프로젝트의 테스트 파일도 함께 복사·수정.
5. **문서 동기화**: 본 문서는 양쪽 `docs/porting/` 에 동일 내용으로 복사.

---

## 8. 다음 밀수 (Next Drop) — 우선 처리

순서대로:

1. 🔥 **IW-1 (TECHAI_CODE Plan read-only)** — 보안 핫픽스.
2. ⭐ **T1-1 (hashline 포팅)** — TECHAI_CODE 업그레이드의 상징적 아이템.
3. **IW-2,3 (prompt embed + clarifyFirstDirective)** — 양쪽 동시.
4. **T1-3,4 (git, diagnostics 도구)** — TECHAI_CODE.
5. **Doom loop detector** — 양쪽 동시.
6. **Repo-map PoC** — TECHAI_CODE에서.
7. **Skill PoC** — TECHAI_CODE에서.
8. **hanimo A1 (Repo-map 풀 구현)**.
9. **hanimo A2 (Skill 풀 구현)**.
10. **hanimo A3 (Hooks)**.

---

## 9. 파일 대조표 (심볼릭 링크 대용)

포팅 시 이 표로 1:1 대응.

| hanimo 경로 | TECHAI_CODE 경로 | 상태 |
|---|---|---|
| `internal/tools/hashline.go` | `internal/tools/hashline.go` | 📋 |
| `internal/tools/git.go` | `internal/tools/git.go` | 📋 |
| `internal/tools/diagnostics.go` | `internal/tools/diagnostics.go` | 📋 |
| `internal/llm/providers/` | `internal/llm/providers/` | 📋 |
| `internal/mcp/` | `internal/mcp/` | 📋 |
| `internal/agents/plan.go` | `internal/agents/plan.go` | 📋 |
| `internal/agents/auto.go` | `internal/agents/auto.go` | 📋 |
| `internal/agents/intent.go` | `internal/agents/intent.go` | 📋 |
| `internal/ui/palette.go` | `internal/ui/palette.go` | 📋 |
| `internal/ui/i18n.go` | `internal/ui/i18n.go` | 📋 |

---

_Last updated: 2026-04-10_
