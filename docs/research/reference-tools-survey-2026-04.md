# Reference Tools Survey — 2026-04

> **Purpose**: hanimo / TECHAI_CODE 강화 방향을 정하기 위한 외부 레퍼런스 조사.
> 조사 대상: opencode(sst), OpenAI Codex CLI, Claude Code, Gemini CLI, Hermes, Crush(Charm), Aider, Goose, Continue CLI, Cursor Agent.
> 작성일: 2026-04-10

---

## 0. 한 줄 요약

2026년 터미널 코딩 에이전트는 다섯 가지 축으로 경쟁 중이다:

1. **Context Engineering** — AGENTS.md / CLAUDE.md + repo-map + 동적 압축
2. **Tool Contract 엄격화** — shell-first + apply_patch / SEARCH-REPLACE + hash anchor
3. **Sub-agent 병렬화** — 격리된 context로 탐색/연구를 위임
4. **Hooks 기반 확장성** — PreToolUse/PostToolUse 훅으로 비기능 요구 분리
5. **Skill 시스템** — 재사용 가능한 워크플로우 자산화 (SKILL.md 표준)

hanimo는 이미 (1),(2) 일부, (3) 맹아를 갖췄다. 부족한 것은 **Skill / Hooks / repo-map / worktree**.
TECHAI_CODE는 (2),(3),(4),(5) 모두 신규. 먼저 **hash-anchor edit, MCP, 다중 provider**를 hanimo에서 포팅해야 한다.

---

## 1. 도구별 핵심 인사이트

### 1.1 opencode (sst/opencode)

**아키텍처**: server + multi-client (TUI / Desktop / VS Code / Web) 분리.
이벤트 버스(`{entity}.{action}`)로 SSE 스트리밍 → 멀티 클라이언트 동기화.

**주목할 기능**:
- **Plan vs Build 이원화 에이전트**: Plan은 read-only, Build는 실행. 탐색과 실행 분리.
- **Hash-anchored edit**: `LINE#HASH` 포맷으로 "내가 읽은 정확한 버전"을 지정. (hanimo도 이미 구현됨 ✅)
- **Rules / Harness 시스템**: `.opencode/rules/*.md` 에 glob/keyword/branch 조건을 걸어 동적 주입.
- **Dynamic Context Pruning**: 모델이 직접 `compress` 도구를 호출 → 오래된 tool call을 요약문으로 치환 (임계값 기반이 아님).
- **SummaryMessageID**: 압축 지점을 링크로 저장해 "압축 이전"으로 되돌릴 수 있게.
- **75+ provider** (AI SDK + models.dev) — BYOK.

**hanimo가 참고할 것**: Rules 엔진(harness), Dynamic Pruning 도구, Plan/Build 실제 권한 분리(현재는 tool list만 다름).

**보안 주의**: opencode의 approval system은 "UX 기능일 뿐 보안 경계가 아님" — Seatbelt/Docker 같은 OS-level sandbox가 진짜 방어선. GHSA-vxw4-wv6m-9hhh 참고.

---

### 1.2 OpenAI Codex CLI

**아키텍처**: 단일 에이전트 ReAct 루프 + Responses API 스트리밍. LLM-agnostic wrapper.

**주목할 기능**:
- **Tight tool contract**: shell + `apply_patch` 두 개가 거의 전부.
  - 파일 뮤테이션은 **apply_patch (unified diff)** 로만 허용.
  - 읽기/검색/테스트는 `shell` (cat/grep/find/pytest)로.
  - 모델이 surgical diff를 내도록 유도 → whole-file rewrite 억제.
- **Prompt caching으로 O(n²) → O(n)**: 장시간 세션 비용이 선형화.
- **macOS Seatbelt / Linux Docker 격리**: 프로젝트 디렉토리 밖 접근·네트워크 차단.
- **Subagents (2026-03)**: `/root/agent_a` 경로 주소. read-only QA/연구에만 권장(토큰 비용 큼).
- **Custom system prompt**: `~/.codex/systemprompts/` 에 프로젝트별 top-level 프롬프트 추가.

**hanimo가 참고할 것**: `apply_patch` 스타일 unified-diff 도구 (hash-anchor와 별개 옵션), Seatbelt/Docker sandbox, system prompt 파일 계층화.

---

### 1.3 Claude Code

**아키텍처**: React + Ink TUI (React in terminal). Anthropic 자체 모델 고정.

**주목할 기능**:
- **Skills (Agent Skills 표준)**: `SKILL.md` + YAML frontmatter.
  - `disable-model-invocation`, `user-invocable`, `allowed-tools`, `effort` 필드.
  - **Lazy loading**: description만 항상 context에, 본문은 호출 시만.
  - `` !`shell command` `` 으로 렌더 전 동적 주입.
  - `$ARGUMENTS[N]`, `${CLAUDE_SESSION_ID}` 치환.
  - `.claude/skills/` nested 자동 탐색 (모노레포 지원).
- **Subagents**: 독립 system prompt + tool 권한 + context fork.
  - 부모에 반환되는 건 **500-token 요약**만. 전체 transcript X.
- **Hooks (25개 이벤트)**: PreToolUse / PostToolUse / PermissionRequest / PostCompact 등.
  - 4가지 핸들러: command / http / prompt / agent.
  - PreToolUse 는 exit 2 로 차단 가능.
- **Permission 5 modes**: Normal / Auto-accept / Plan / Don't Ask / Bypass. Shift+Tab 순환.
- **CLAUDE.md 영속**: 압축 후에도 디스크에서 재로드.
- **1M context window** (2026 GA).

**hanimo가 참고할 것**: Skill 시스템 전부, Hooks 25 이벤트, Permission 5-mode 순환, CLAUDE.md 압축-생존.

---

### 1.4 Gemini CLI

**아키텍처**: ReAct 루프 + Event-driven scheduler (v0.37.0, 2026-04).

**주목할 기능**:
- **Session Browser**: `~/.gemini/tmp/<project_hash>/chats/` 에 자동 저장. `/resume`, `gemini --resume <UUID>`.
- **Git worktree 통합**: worktree별로 session 격리 → 병렬 agentic development.
- **Dynamic sandbox expansion**: 권한 위반 시점에 모달로 확장 요청.
- **Output formats**: `--output-format json|stream-json`, headless 모드.
- **Skill activation UI**: 사용자 확인 후 활성화.
- **JIT context injection**: 도구 실행 시점에 context 주입.

**hanimo가 참고할 것**: Session Browser UI, worktree 통합, JSON output mode, headless scripting 모드.

---

### 1.5 Aider

**아키텍처**: Python + Git-native + tree-sitter repo-map.

**주목할 기능**:
- **Repo-map (⭐ 가장 복제 가치 높음)**:
  - tree-sitter로 심볼(def/ref) 추출 → PageRank로 중요도 정렬 → 1K 토큰 예산.
  - 130+ 언어 지원. 60K+ 레포지토리가 AGENTS.md 표준 채택.
  - incremental: mtime 체크 후 변경된 파일만 재파싱.
- **SEARCH/REPLACE blocks**: 유니파이드 diff 기반. GPT-4 laziness 3x 감소.
- **자동 git commit**: 각 변경마다 descriptive commit message.
- **테스트 루프**: 실패 → 자동 복구.

**hanimo가 참고할 것**: Repo-map을 SQLite나 bbolt로 영속화해 cold start <100ms 달성. PageRank 심볼 스코어링.

---

### 1.6 Crush (Charmbracelet)

**hanimo의 가장 직접적 경쟁자** — Go + Bubble Tea v2.

**주목할 기능**:
- **LSP 통합**: 언어 서버로 diagnostics / references / completions.
- **Self-documenting tools**: `.go` 구현 + `.md` 설명 쌍. 시스템 프롬프트에 Go template으로 주입.
- **Async tool construction** (errgroup `readyWg` 패턴).
- **<200ms startup**, 단일 바이너리.

**hanimo가 참고할 것**: LSP 통합 (`gopls`, `tsserver` 연결), tool 설명을 마크다운 파일로 분리.

---

### 1.7 Goose (Block)

**주목할 기능**:
- **MCP-native**: 확장이 전부 MCP 서버. JSON-RPC stdio/HTTP/SSE.
- **70+ extensions** 생태계.
- Desktop + CLI + HTTP API 3-way interface.

**hanimo가 참고할 것**: MCP를 first-class로 대우 (hanimo는 이미 stdio transport 있음 ✅). HTTP transport 추가 필요.

---

### 1.8 Continue CLI (cn)

**주목할 기능**:
- **2-tier permission**: read-only (auto) vs write (approval).
- `~/.continue/permissions.yaml` 학습형 저장.
- Headless 모드 + API key auth (CI 친화).
- `@file`, `/command` 참조 구문.

**hanimo가 참고할 것**: permissions.yaml 학습형 저장, `@`/`/` 레퍼런스 구문.

---

### 1.9 Cursor Agent (Cursor 3, 2026-01)

**주목할 기능**:
- **Cloud Agents**: 10개 병렬 워커.
- **Subagent 기본 제공**: research, terminals, specialized.
- **Plan/Ask/Cloud 모드**: Plan은 설계, Ask는 탐색 (변경 없음), Cloud는 푸시.

**hanimo가 참고할 것**: Ask 모드 (explicit "답만 하고 변경 안 함") 신설. Plan과는 다름.

---

### 1.10 Hermes Agent (NousResearch)

**주목할 기능**:
- **Skill 자동 추출**: 성공적 워크플로우를 마크다운으로 저장 → 차후 semantic search로 재사용.
- **Multi-provider memory**: Honcho, Mem0, Hindsight, Holographic.
- Hermes-3 모델은 Atropos RL로 tool-calling 학습.
- **Cross-session learning**: 세션 간 학습.

**hanimo가 참고할 것**: "성공 패턴 자동 추출 → skill로 저장" 루프. hanimo knowledge store와 결합하면 강력함.

---

## 2. 2026년 고변별(high-differentiation) 아이디어 Top 10

우선순위는 "구현 난이도 × 체감 효과"로 정렬.

### #1. Repo-map with PageRank + 영속 캐시 ⭐⭐⭐⭐⭐
- Aider 스타일 tree-sitter 기반 symbol graph.
- SQLite 또는 bbolt에 캐시, mtime 기반 incremental.
- Go: `github.com/smacker/go-tree-sitter` + tree-sitter-{go,ts,py,...}.
- 1K 토큰 예산 내 "이 프로젝트에서 중요한 심볼" 자동 주입.
- 기대 효과: cold start <100ms, context 품질 대폭 상승.

### #2. Skill 시스템 (SKILL.md 표준 채택) ⭐⭐⭐⭐⭐
- `.hanimo/skills/*/SKILL.md` + YAML frontmatter.
- description만 시스템 프롬프트 주입, 본문은 lazy.
- `!`shell`` pre-rendering, `$ARGUMENTS[N]` 치환.
- nested 탐색 (모노레포).
- Hermes식 "성공 패턴 → skill 자동 추출" 까지 가면 차별화 최상.

### #3. Hooks 시스템 ⭐⭐⭐⭐
- 최소 8 이벤트: PreToolUse, PostToolUse, PostToolUseFailure, PermissionRequest, SessionStart/End, PostCompact, TaskComplete.
- 4 handler: command / http / prompt / subagent.
- PreToolUse exit 2 = block.
- `.hanimo/hooks/*.{sh,yaml}` 에 선언.

### #4. Context Fork Subagent ⭐⭐⭐⭐
- 현재 hanimo `/auto` 와 `agents/auto.go`는 main context 내에서 돈다. 대신:
- `spawn_subagent(task, tools, model)` 도구로 goroutine 분기.
- 500-token 요약만 부모에 반환.
- 탐색/리서치/테스트 전담 subagent에 유용.

### #5. `apply_patch` 스타일 unified-diff 도구 ⭐⭐⭐⭐
- `file_edit`, `hashline_edit` 에 더해 **diff 기반 도구** 추가.
- 여러 hunks를 한 번에 적용. surgical edit 유도.
- tree-sitter validator로 적용 후 syntax 검사 → 깨지면 rollback.

### #6. Git Worktree 병렬화 ⭐⭐⭐
- `/parallel <task1> <task2>` 로 worktree 자동 생성.
- 각 worktree에 독립 session / subagent.
- 머지 시 충돌 검사.
- Go: `github.com/go-git/go-git` 이용.

### #7. Permission 5-mode + Rule 엔진 ⭐⭐⭐⭐
- normal / auto-accept / plan / dont-ask / bypass. Shift+Tab 순환.
- `Bash(git add *)`, `Read /path/*`, `Skill(deploy *)` 와일드카드 rule.
- `~/.hanimo/permissions.yaml` 학습형 저장.

### #8. Session Browser + resume UX ⭐⭐⭐
- 이미 SQLite session store 있음 ✅. 전용 TUI 브라우저 필요.
- `/resume` 시 최근 10개 표 + 검색.
- `hanimo --resume <id>` CLI.

### #9. Prompt Caching 활용 ⭐⭐⭐
- Anthropic/OpenAI provider에서 `cache_control` 지원.
- 고정 system prompt + 긴 CLAUDE.md-like context → cache breakpoint 삽입.
- 장시간 세션 비용 O(n²) → O(n).

### #10. Synchronized Output Mode (Bubble Tea v2) ⭐⭐
- Terminal Sync Update escape (`\e[?2026h`) 로 flicker 제거.
- Streaming markdown + spinner 동기화.

---

## 3. Prompt Engineering 개선 (즉시 반영 가능)

### 3.1 계층형 system prompt 구조 (Codex 방식)

```
[Identity] → [Operating Model] → [Tool Contract] → [Preferences]
→ [Project Rules (AGENTS.md/CLAUDE.md)] → [Session Context]
```

현재 hanimo prompt는 identity + tool + rules가 섞여있음. 분리하면 cache breakpoint도 넣기 쉬움.

### 3.2 Tool description 강화

- 각 도구 description에 "언제 쓰지 말아야 하는지"를 명시 (negative example).
- `file_edit` 은 "old_string must appear exactly once" 를 명시 (uniqueness validation).
- hash-anchor 도구는 "use this when file may have been modified since read" 를 추가.

### 3.3 In-context few-shot

- 시스템 프롬프트에 1~2개의 "좋은 tool call 예시" 포함.
- 특히 `hashline_edit` 은 anchor 포맷이 낯설기 때문에 예시 필수.

---

## 4. 보안/안전성 (둘 다 해야 함)

1. **Read-before-write 강제**: 세션 메모리에 "최근 읽은 파일 집합" 유지 → edit 도구가 해당 집합에 없으면 거부.
2. **Stale-read detection**: hanimo hash-anchor가 이미 함. TECHAI_CODE도 포팅해야 함.
3. **Dangerous command 정규식 검사**: `rm -rf /`, `sudo`, `export (AWS|OPENAI|ANTHROPIC)_*`, `curl -H 'Authorization'` 등.
4. **Doom loop detection**: 최근 3 iteration 동일 행동 반복 → abort.
5. **Iteration cap**: hanimo `/auto` 는 20, dev 모드는 100 → 환경변수로 가변화.
6. **Credential scrubbing**: shell 출력에서 API 키 마스킹 후 모델에 전달.
7. **Sandbox (선택)**: Linux `unshare`, macOS `sandbox-exec`, Windows job object. 기본은 opt-in.

---

## 5. 참고 출처

### opencode / Codex
- https://opencode.ai/docs/
- https://deepwiki.com/sst/opencode/2.4-context-management-and-compaction
- https://deepwiki.com/code-yeongyu/oh-my-opencode/9.3-hash-anchored-edit-system
- https://developers.openai.com/codex/cli
- https://developers.openai.com/codex/cookbook/examples/gpt-5/codex_prompting_guide
- https://openai.com/index/unrolling-the-codex-agent-loop/

### Claude Code / Gemini CLI
- https://code.claude.com/docs/en/skills
- https://code.claude.com/docs/en/sub-agents
- https://code.claude.com/docs/en/hooks
- https://code.claude.com/docs/en/memory
- https://code.claude.com/docs/en/permissions
- https://developers.googleblog.com/pick-up-exactly-where-you-left-off-with-session-management-in-gemini-cli/
- https://geminicli.com/docs/hooks/
- https://geminicli.com/docs/cli/skills/
- https://geminicli.com/docs/core/subagents/
- https://geminicli.com/docs/cli/sandbox/

### Aider / Crush / Goose / Others
- https://aider.chat/2023/10/22/repomap.html
- https://github.com/charmbracelet/crush
- https://github.com/block/goose
- https://hermes-agent.nousresearch.com/
- https://docs.continue.dev/cli/overview
- https://cursor.com/docs/cli/using

### Comparative / Security
- https://www.morphllm.com/comparisons/opencode-vs-claude-code
- https://effloow.com/articles/terminal-ai-coding-agents-compared-claude-code-gemini-cli-2026/
- https://github.com/advisories/GHSA-vxw4-wv6m-9hhh (opencode RCE advisory)
- https://www.docker.com/blog/docker-sandboxes-run-claude-code-and-other-coding-agents-unsupervised-but-safely/
- https://www.firecrawl.dev/blog/ai-agent-sandbox
