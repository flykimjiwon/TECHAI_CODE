# 택가이코드 (TECHAI CODE)

OpenAI-compatible API 기반 CLI AI 코딩 어시스턴트. Go + Bubble Tea v2 단일 바이너리.

## 특징

- **3개 모드**: Super(만능) / Deep Agent(자율 코딩 100회) / Plan(계획) — Tab 전환
- **멀티 에이전트**: 메인+서브 에이전트 병렬 동작 (Review/Consensus/Scan) + LLM 확인 게이트
- **27개 슬래시 명령어**: /init, /remember, /compact, /copy, /export, /diff, /undo, /commands 등
- **Fuzzy 파일 편집**: 4단계 매칭 (ExactMatch → LineTrimmed → IndentFlex → Levenshtein 95%)
- **파일 스냅샷 + /undo**: 수정 전 자동 백업, /undo로 즉시 복구
- **메모리 시스템**: /remember로 프로젝트별·글로벌 메모리 저장, AI 컨텍스트에 자동 주입
- **커스텀 명령어**: `.tgc/commands/*.md` 파일로 나만의 슬래시 명령어 생성
- **MCP 지원**: Model Context Protocol 클라이언트 (stdio/sse) — 사내 도구 통합
- **안전 설계**: 시크릿 감지, 셸 명령 차단/경고, Git 안전 규칙, file_edit diff 미리보기
- **.gitignore 존중**: grep/glob 검색 시 .gitignore 패턴 자동 적용
- **지식 문서 RAG**: `.tgc/knowledge/` 자동 인덱싱 + 81개 내장 문서 + 3단계 검색 파이프라인
- **텍스트 선택**: 기본 마우스 OFF — 드래그로 자유 복사, Ctrl+B로 스크롤 전환
- **세션 영속화**: SQLite 기반, 재시작 후에도 대화 복원
- **입력 히스토리**: ↑/↓ 화살표로 이전 입력 100개 탐색
- **단일 바이너리**: Node.js, Python 등 외부 의존성 없음 (~25MB)
- **크로스 플랫폼**: macOS (ARM/Intel) · Windows · Linux 전 플랫폼 지원

## 설치

### macOS (Apple Silicon — M1/M2/M3/M4)

```bash
sudo cp dist/techai-darwin-arm64 /usr/local/bin/techai
sudo chmod +x /usr/local/bin/techai

# Gatekeeper 경고 시
xattr -d com.apple.quarantine /usr/local/bin/techai

# 실행
techai
```

### macOS (Intel)

```bash
sudo cp dist/techai-darwin-amd64 /usr/local/bin/techai
sudo chmod +x /usr/local/bin/techai
techai
```

### Windows 10/11

**방법 1: GUI (가장 확실)**

1. `techai-windows-amd64.exe` 파일을 `C:\techai\techai.exe`로 복사
2. `Win + R` → `sysdm.cpl` → 고급 → **환경 변수** 클릭
3. **사용자 변수** → `Path` → **편집** → **새로 만들기** → `C:\techai` 입력 → **확인**
4. **모든 터미널 창을 닫고 새로 열기** (중요!)
5. 새 터미널에서 실행:

```powershell
techai
```

**방법 2: PowerShell (관리자 권한으로 실행)**

```powershell
# 1. 폴더 생성 + exe 복사
New-Item -ItemType Directory -Force -Path C:\techai
Copy-Item techai-windows-amd64.exe C:\techai\techai.exe

# 2. PATH 환경변수 추가
[System.Environment]::SetEnvironmentVariable("Path",
  $env:Path + ";C:\techai",
  [System.EnvironmentVariableTarget]::User)

# 3. 반드시 터미널을 완전히 닫고 새로 열기!
# 4. 새 터미널에서 실행
techai
```

**안 되면 확인:**

```powershell
# exe 파일 있는지 확인
Test-Path C:\techai\techai.exe

# PATH에 등록됐는지 확인
$env:Path -split ";" | Select-String "techai"

# 전체 경로로 직접 실행 (PATH 무관하게 동작)
C:\techai\techai.exe
```

> **Windows Terminal** (Microsoft Store 무료) 사용 권장 — 색상, 마크다운, 마우스 스크롤 지원이 우수합니다.
> CMD(명령 프롬프트)보다 PowerShell 또는 Windows Terminal을 사용하세요.

**VSCode 터미널에서 `techai` 실행하기:**

VSCode 내장 터미널은 PATH 변경이 바로 반영되지 않습니다. 아래 중 하나를 선택하세요:

```powershell
# 방법 1: 현재 세션에서 PATH 수동 갱신
$env:Path = [System.Environment]::GetEnvironmentVariable("Path", "User") + ";" + [System.Environment]::GetEnvironmentVariable("Path", "Machine")
techai
```

```json
// 방법 2: VSCode settings.json에 추가 (영구적, 추천)
// Ctrl+Shift+P → "Preferences: Open User Settings (JSON)"
{
  "terminal.integrated.env.windows": {
    "PATH": "${env:PATH};C:\\techai"
  }
}
```

> 모든 방법은 `C:\techai`가 시스템 PATH에 등록되어 있어야 합니다.

### Linux

```bash
sudo cp dist/techai-linux-amd64 /usr/local/bin/techai
sudo chmod +x /usr/local/bin/techai
techai
```

### 소스에서 빌드

```bash
git clone https://github.com/kimjiwon/tgc.git
cd tgc
make build       # → ./techai 생성
make install     # → $GOPATH/bin에 설치
```

## 설정

첫 실행 시 자동으로 설정 위저드가 실행됩니다 (API 키 입력).

```bash
techai --setup     # 설정 재실행
techai --reset     # 설정 초기화 후 재설정
```

설정 파일: `~/.tgc/config.yaml`

```yaml
api:
  base_url: "https://api.novita.ai/openai"
  api_key: "tg-..."
models:
  super: "openai/gpt-oss-120b"
  dev: "qwen/qwen3-coder-30b"
multi:
  enabled: true
  strategy: "auto"    # auto | review | consensus | scan
mcp:
  servers:
    - name: "wiki"
      transport: sse
      url: "http://internal-wiki.company.com/mcp"
    - name: "jira"
      transport: stdio
      command: "mcp-jira-server"
      args: ["--project", "MYPROJ"]
      env:
        JIRA_TOKEN: "xxx"
```

환경변수 오버라이드:

```bash
export TGC_API_BASE_URL=https://your-api.com/v1
export TGC_API_KEY=tg-...
export TGC_MODEL_SUPER=openai/gpt-oss-120b   # Novita
export TGC_MODEL_DEV=qwen/qwen3-coder-30b    # Novita
export TGC_MULTI=auto    # on | off | review | consensus | scan
```

## 사용법

```bash
techai                # 기본 (슈퍼택가이 모드)
techai --mode dev     # 개발 모드로 시작
techai --mode plan    # 플랜 모드로 시작
techai --version      # 버전 출력
```

## Key Bindings

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Shift+Enter` | Newline |
| `Ctrl+J` | Newline (fallback for Windows CMD/PowerShell) |
| `Tab` | Switch mode (Super → Deep Agent → Plan) |
| `↑` / `↓` | Input history (browse previous 100 messages) |
| `Ctrl+K` | Command palette (fuzzy search) |
| `Esc` | Menu / Cancel streaming |
| `Ctrl+U` | Clear input field |
| `Ctrl+B` | Toggle mouse mode (scroll vs text selection) |
| `Ctrl+L` | Clear conversation |
| `Ctrl+C` | Quit |
| `Ctrl+V` / `Cmd+V` | Paste (5 lines: expand, 6+: hint above input) |
| `Alt+↑/↓` | Scroll 3 lines |
| `PgUp/PgDown` | Page scroll |

> **Mouse Mode** (default: ON):
> - `Ctrl+B` toggles mouse mode
> - **ON**: Mouse wheel scroll works, but text drag selection disabled
> - **OFF**: Text drag selection + copy enabled, scroll via keyboard only (`Alt+↑/↓`, `PgUp/PgDn`)
> - Tip: `Ctrl+B` → select text → copy → `Ctrl+B` back to scroll mode

> **Windows Newline**: `Shift+Enter` may not work on PowerShell/CMD. Use `Ctrl+J` instead.
> To enable `Shift+Enter` on Windows Terminal, add to `settings.json`:
> ```json
> { "keys": "shift+enter", "command": { "action": "sendInput", "input": "\u000A" } }
> ```

## Slash Commands (27 commands)

### Project

| Command | Action | Details |
|---------|--------|---------|
| `/init` | Scan project → generate `.techai.md` | Analyzes directory structure, dependencies, entry points, scripts, git info. Auto-creates `.tgc/knowledge/` and `.tgc/commands/` folders. Generated file is auto-loaded into AI context on every session. Run again to refresh. |

### Session

| Command | Action | Details |
|---------|--------|---------|
| `/new` | Start new session | Creates fresh SQLite-backed session |
| `/sessions` | Browse recent sessions | Picker overlay with ↑↓ navigation, shows last 20 sessions |
| `/session <id>` | Restore specific session | Full conversation history + mode restored |
| `/compact` | Compress conversation history | 3-stage: snip old tool outputs → truncate large messages → LLM summary. Targets 50% of context window. Use when `ctx:%` gets high. |
| `/clear` | Clear conversation | Keeps system prompt, resets tokens to 0 |

### AI Modes

| Command | Action | Details |
|---------|--------|---------|
| `/auto` | Toggle autonomous mode | AI works independently up to 20 iterations, stops on `[AUTO_COMPLETE]` |
| `/multi on` | Enable multi-agent | Two models (Super + Dev) work together |
| `/multi off` | Disable multi-agent | Single model only |
| `/multi review` | Strategy: Review | Agent1(Super) generates → Agent2(Dev) reviews for bugs/improvements |
| `/multi consensus` | Strategy: Consensus | Same question sent to both models → results compared and synthesized |
| `/multi scan` | Strategy: Scan | Codebase split between agents for parallel file scanning |
| `/multi auto` | Strategy: Auto-detect | Keyword matching + LLM confirmation. Skips on pastes >300 chars |

### File Operations

| Command | Action | Details |
|---------|--------|---------|
| `/undo` | Undo last file modification | Restores from `~/.tgc/snapshots/` (auto-created before every `file_write`/`file_edit`) |
| `/undo <N>` | Undo last N modifications | Batch restore |
| `/undo list` | Show snapshot history | Lists up to 20 recent snapshots with timestamps |
| `/diff` | Show git diff | Runs async (non-blocking). Truncated at 5KB |

### Clipboard & Export

| Command | Action | Details |
|---------|--------|---------|
| `/copy` | Copy last AI response | Copies to system clipboard (uses `atotto/clipboard`) |
| `/copy <N>` | Copy Nth recent AI response | `/copy 2` = second most recent |
| `/copy all` | Copy entire session | All user + AI messages |
| `/export` | Export session to `.md` file | Default: `techai-session-YYYYMMDD-HHMMSS.md` |
| `/export <name>` | Export with custom filename | Auto-appends `.md` if missing |

### Tools & Diagnostics

| Command | Action | Details |
|---------|--------|---------|
| `/diagnostics` | Run project linters | Auto-detects: Go(`go vet`), TS(`tsc`), JS(`eslint`), Python(`pylint`) |
| `/git` | Git repository status | Branch, staged/unstaged/untracked counts |
| `/mcp` | MCP server status | Shows connected servers, tool counts, errors |
| `/companion` | Browser dashboard | Opens `localhost:8787` — real-time SSE stream of AI activity |

### Memory

| Command | Action | Details |
|---------|--------|---------|
| `/remember <text>` | Save project memory | Stored in `.tgc/memories.json`, injected into AI context |
| `/remember -g <text>` | Save global memory | Stored in `~/.tgc/memories.json`, shared across all projects |
| `/remember list` | Show all memories | Project + global, with hit counts |
| `/remember edit <id> <text>` | Update memory | Edit by ID |
| `/remember delete <id>` | Delete project memory | Delete by ID |
| `/remember -g edit <id> <text>` | Update global memory | Edit global by ID |
| `/remember -g delete <id>` | Delete global memory | Delete global by ID |
| `/remember search <query>` | Search memories | Case-insensitive keyword match |
| `/forget <id>` | Delete shorthand | Same as `/remember delete <id>` |
| `/forget -g <id>` | Delete global shorthand | Same as `/remember -g delete <id>` |

### Custom Commands

| Command | Action | Details |
|---------|--------|---------|
| `/commands` | List loaded commands | Shows all `.md`-based custom commands |
| `/<name>` | Run custom command | File `.tgc/commands/<name>.md` content sent as message |
| `/<name> <args>` | Run with arguments | `$ARGUMENTS` in template replaced with args |

> Create `.tgc/commands/review.md` → type `/review src/main.go` → AI receives the template with args.

### System

| Command | Action | Details |
|---------|--------|---------|
| `/setup` | Reset API key | Re-runs setup wizard, saves to `~/.tgc/config.yaml` |
| `/version` | Show version | Build version from `git describe` |
| `/help` | Show all commands | Keyboard shortcuts + slash command reference |
| `/exit` | Quit | Same as `Ctrl+C` or `/quit` |

## Modes (Tab to switch)

| Mode | Model (Novita) | Model (Onprem) | Description |
|------|----------------|----------------|-------------|
| **Super** | `openai/gpt-oss-120b` | `GPT-OSS-120B` | All-purpose. Auto-detects intent: code, analysis, conversation. Full 14 tools + MCP. 128K context window. Optimized for complex tool chaining, multi-step reasoning, knowledge injection (8K budget, 6 sections). Default mode. |
| **Deep Agent** | `openai/gpt-oss-120b` | `GPT-OSS-120B` | Autonomous coding. AI works without user input for up to 100 iterations. Stops on `[TASK_COMPLETE]` marker. Same model as Super but with auto-continue system prompt. Best for large refactors, multi-file changes. |
| **Plan** | `openai/gpt-oss-120b` | `GPT-OSS-120B` | Plan-first approach. AI creates step-by-step plan before execution. Full tools including write access. Best for complex tasks requiring architectural decisions. |

### Multi-Agent Models

When multi-agent is enabled (`/multi`), two models work together:

| Role | Novita | Onprem | Gemma | Purpose |
|------|--------|--------|-------|---------|
| **Agent1 (Super)** | `openai/gpt-oss-120b` | `GPT-OSS-120B` | `google/gemma-4-31b-it` | Primary: generates code, full tool access |
| **Agent2 (Dev)** | `qwen/qwen3-coder-30b` | `Qwen3-Coder-30B` | `google/gemma-4-31b-it` | Secondary: reviews, compares. Read-only tools. 32K context, 2K knowledge budget |

### Multi-Agent Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  User Input                                                 │
└──────────────────────────┬──────────────────────────────────┘
                           │
                    ┌──────▼──────┐
                    │  Auto-Detect │  Keyword match (0ms)
                    │  + LLM Gate  │  + LLM confirmation (5s)
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
     ┌────────▼──┐  ┌─────▼─────┐  ┌──▼────────┐
     │  Review    │  │ Consensus │  │   Scan     │
     └────────┬──┘  └─────┬─────┘  └──┬────────┘
              │            │            │
    ┌─────────▼─────────┐  │   ┌───────▼────────┐
    │  Phase 1           │  │   │  Parallel Split │
    │  Agent1 (Super)    │  │   │  Agent1: dir/a/ │
    │  GPT-OSS-120B      │  │   │  Agent2: dir/b/ │
    │  Full tools (14)   │  │   └───────┬────────┘
    └─────────┬─────────┘  │           │
              │            │           │
    ┌─────────▼─────────┐  │           │
    │  Phase 2           │  │           │
    │  Agent2 (Dev)      │  │           │
    │  Qwen3-Coder-30B   │  │           │
    │  Read-only tools   │  │           │
    │  Reviews output    │  │           │
    └─────────┬─────────┘  │           │
              │            │           │
              │  ┌─────────▼─────────┐ │
              │  │  Both agents run   │ │
              │  │  same prompt in    │ │
              │  │  parallel          │ │
              │  │  Agent1 ──┐        │ │
              │  │  Agent2 ──┤        │ │
              │  └──────────┬┘        │ │
              │             │         │ │
              └──────┬──────┘─────────┘ │
                     │                  │
              ┌──────▼──────────────────▼──┐
              │  LLM Synthesis              │
              │  Merges outputs into one    │
              │  cohesive response          │
              │  (skips if Agent2 = no issues│
              │   or Agent2 errored)        │
              └──────────────┬─────────────┘
                             │
                      ┌──────▼──────┐
                      │  Final      │
                      │  Response   │
                      └─────────────┘
```

### Model Capability Differences

| Capability | GPT-OSS-120B (Super) | Qwen3-Coder-30B (Dev) |
|-----------|---------------------|----------------------|
| Context Window | 128K tokens | 32K tokens |
| Coding Tier | Strong (complex chaining) | Moderate (simple tasks) |
| Knowledge Budget | 8,000 tokens (6 sections) | 2,000 tokens (2 sections) |
| Compaction Trigger | 80% / 90% (auto) | 60% / 75% (auto) |
| Preserve Recent | 15 messages | 8 messages |
| Edit Strategy | `file_edit` (fuzzy 4-stage) | `hashline_edit` (hash anchors) |

## Tools (14 built-in + MCP)

AI automatically invokes these tools during conversation. Shown as `Tool:ON(14)` in status bar.

| Tool | Description | Safety |
|------|-------------|--------|
| `file_read` | Read file contents (max 50KB) | — |
| `file_write` | Create or overwrite file | Auto-snapshot before write |
| `file_edit` | Edit file with **4-stage fuzzy matching**: ExactMatch → LineTrimmed → IndentFlex → Levenshtein(85%) | Auto-snapshot + diff preview |
| `list_files` | List directory contents (recursive supported) | Skips `.git`, `node_modules`, `dist` |
| `shell_exec` | Execute shell command (30s timeout) | Blocks dangerous commands (`rm -rf /`, `sudo`). Warns on risky commands (`rm -r`, `git reset --hard`) |
| `grep_search` | Regex search in file contents | Respects `.gitignore`. Max 100 matches |
| `glob_search` | Find files by glob pattern (`**/*.go`) | Respects `.gitignore`. Max 2000 files |
| `hashline_read` | Read file with MD5 hash anchors per line | For stale-edit protection |
| `hashline_edit` | Edit file using hash anchors | Verifies hash before replacing |
| `git_status` | Git status (short format) | — |
| `git_diff` | Git diff (staged or unstaged) | — |
| `git_log` | Recent commits (default 10) | — |
| `diagnostics` | Auto-detect project type → run linter | Go/TS/JS/Python |
| `knowledge_search` | Search embedded + user knowledge docs | 3-stage: keyword → BM25 → LLM |
| `mcp_*` | MCP server-provided tools | Auto-registered from config. Prefixed `mcp_{server}_{tool}` |

## Glossary

| Term | Description |
|------|-------------|
| **Context Window** | Maximum tokens the LLM can process at once. Shown as `ctx:XX%` in HUD. Auto-compact at 90% |
| **Compaction** | 3-stage process to reduce conversation size: (1) Snip old tool outputs (2) Truncate large messages (3) LLM summary |
| **Multi-Agent** | Two LLMs working together. Strategy determines how: Review (generate+review), Consensus (compare), Scan (parallel search) |
| **Fuzzy Edit** | 4-stage file matching: exact → whitespace-trimmed → indent-normalized → Levenshtein similarity (85%+) |
| **Snapshot** | Auto-backup of files before AI modification. Stored in `~/.tgc/snapshots/`. Restored via `/undo` |
| **MCP** | Model Context Protocol. Standard for connecting AI to external tools (Jira, Wiki, CI/CD). Supports stdio and SSE transports |
| **Tool** | Function the AI can call during conversation (read files, run commands, search code). Count shown as `Tool:ON(14)` |
| **HUD** | Status bar at bottom: mode, model, CWD, git branch, debug flag, tool count, multi status, token count, cost estimate, context % |
| **Knowledge RAG** | 81 embedded reference docs + user `.tgc/knowledge/` docs. Auto-injected into prompts via 3-stage search pipeline |
| **Deep Agent** | Autonomous mode where AI continues working without user input until task is complete or iteration limit reached |
| **Companion** | Browser-based dashboard (`localhost:8787`) showing real-time AI activity via Server-Sent Events (SSE) |
| **Session** | Conversation persisted in SQLite (`~/.tgc/sessions.db`). Survives app restart. Browsable via `/sessions` |
| **.techai.md** | Project context file generated by `/init`. Auto-loaded into system prompt. Contains structure, deps, scripts, git info |
| **Onprem** | On-premise build variant for internal networks. Separate config dir (`~/.tgc-onprem/`), hardcoded endpoint |
| **Palette** | Command palette (`Ctrl+K`). Fuzzy search across all slash commands |
| **Memory** | Persistent facts saved via `/remember`. Project-local (`.tgc/memories.json`) or global (`~/.tgc/memories.json`). Auto-injected into system prompt |
| **Custom Command** | User-defined slash command from `.md` file in `.tgc/commands/` or `~/.tgc/commands/`. Supports `$ARGUMENTS` placeholder |

## File Paths

All paths use `os.UserHomeDir()` — `~` on macOS/Linux, `C:\Users\<username>` on Windows.

### Global (per user, all projects)

| File | macOS / Linux | Windows |
|------|---------------|---------|
| Config | `~/.tgc/config.yaml` | `C:\Users\<user>\.tgc\config.yaml` |
| Sessions DB | `~/.tgc/sessions.db` | `C:\Users\<user>\.tgc\sessions.db` |
| Debug Log | `~/.tgc/debug.log` | `C:\Users\<user>\.tgc\debug.log` |
| Snapshots | `~/.tgc/snapshots/` | `C:\Users\<user>\.tgc\snapshots\` |
| Global Memory | `~/.tgc/memories.json` | `C:\Users\<user>\.tgc\memories.json` |
| Global Commands | `~/.tgc/commands/*.md` | `C:\Users\<user>\.tgc\commands\*.md` |
| Global Knowledge | `~/.tgc/knowledge/*.md` | `C:\Users\<user>\.tgc\knowledge\*.md` |

### Project-local (per project, in CWD)

| File | Path |
|------|------|
| Project Context | `.techai.md` (generated by `/init`) |
| Project Memory | `.tgc/memories.json` |
| Project Commands | `.tgc/commands/*.md` |
| Project Knowledge | `.tgc/knowledge/*.md` |

> **Priority**: Project-local overrides global when both exist (commands, knowledge).

### Build Variants (separate config dirs)

| Build | Config Dir | macOS / Linux | Windows |
|-------|-----------|---------------|---------|
| Default (Novita) | `.tgc` | `~/.tgc/` | `C:\Users\<user>\.tgc\` |
| Onprem (Shinhan) | `.tgc-onprem` | `~/.tgc-onprem/` | `C:\Users\<user>\.tgc-onprem\` |
| Gemma (Novita) | `.tgc-gemma` | `~/.tgc-gemma/` | `C:\Users\<user>\.tgc-gemma\` |

> Three builds can be installed side-by-side without interference.

## 지식 시스템 (Knowledge System)

택가이코드는 **일반 LLM의 한계를 넘어** 81개의 내장 전문 문서를 자동으로 주입하는 지식 시스템을 갖추고 있습니다. [Karpathy의 LLM-Wiki 개념](https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f)에서 영감을 받아, **영속적 지식 저장소**를 LLM 컨텍스트에 자동으로 연결합니다.

### LLM-Wiki 개념 차용

| 개념 | Karpathy LLM-Wiki | 택가이코드 |
|------|-------------------|-----------|
| **지식 저장** | LLM이 위키를 직접 작성/관리 | 사람이 문서를 관리, 시스템이 자동 주입 |
| **검색 방식** | 인덱스 파일 기반 탐색 | 3단계 검색 파이프라인 (키워드→BM25→LLM 판단) |
| **지식 축적** | 소스 추가 시 위키 전체 업데이트 | 문서 추가만 하면 다음 질문부터 자동 반영 |
| **인프라** | Obsidian + 마크다운 | 바이너리 내장 + `.tgc/knowledge/` 폴더 |

**핵심 차이**: RAG 시스템은 매번 원문에서 정보를 재추출하지만, 택가이코드는 **사전 구조화된 전문 레퍼런스 문서**를 키워드 매칭으로 즉시 주입합니다. 벡터 DB 없이 ~1ms로 동작합니다.

### 3단계 검색 파이프라인

사용자가 질문하면, 아래 3단계를 순차적으로 거쳐 가장 관련 높은 문서를 찾습니다.

```
사용자: "스프링에서 인증 처리 어떻게 해?"

┌─ Level 1: 키워드 추출 + 유의어 확장 ──────────────────────┐
│                                                          │
│  "스프링" → techDictionary → "spring-core"               │
│  "인증"   → synonymGroups["spring-security"] 매칭 ✓      │
│                                                          │
│  결과: spring-core.md + spring-security.md (2개)         │
│  → 충분! 바로 주입                                        │
└──────────────────────────────────────────────────────────┘

                    │ 결과 부족 (0~1개)?
                    ▼

┌─ Level 2: BM25 본문 검색 (fallback) ─────────────────────┐
│                                                          │
│  전체 81개 문서의 본문을 대상으로 BM25 스코어링            │
│  TF-IDF 기반 관련도 계산 → 상위 문서 반환                 │
│                                                          │
│  비용: 0원, 지연: <1ms (메모리 내 검색)                   │
└──────────────────────────────────────────────────────────┘

                    │ 결과 0개?
                    ▼

┌─ Level 3: LLM 판단 매칭 (최후 수단) ─────────────────────┐
│                                                          │
│  LLM에게 문서 목록을 보여주고 관련 문서를 고르게 함        │
│  "아래 문서 중 질문에 답하는 데 필요한 것을 골라주세요"     │
│                                                          │
│  비용: API 1회 (~$0.0001), 지연: ~200ms                  │
│  가장 정확하지만, L1+L2로 해결되면 호출 안 함              │
└──────────────────────────────────────────────────────────┘

                    ▼

┌─ 시스템 프롬프트 주입 ────────────────────────────────────┐
│                                                          │
│  [기본 프롬프트] + [프로젝트 컨텍스트] + [환경 정보]      │
│  + ★ [매칭된 지식 문서 전문] ← 매 질문마다 동적 교체      │
│                                                          │
│  → LLM은 "원래 알고 있던 것처럼" 정확한 코드를 답변       │
└──────────────────────────────────────────────────────────┘
```

### 내장 지식 문서 (81개)

바이너리에 컴파일되어 설치 즉시 사용 가능합니다.

| 카테고리 | 문서 수 | 내용 |
|----------|:-------:|------|
| **BXM 프레임워크** | 13 | 배치, Bean, Config, DBIO, 예외처리, 트랜잭션, Service, Studio 등 |
| **Spring 생태계** | 7 | Core(DI/Bean), MVC(컨트롤러), Security(JWT/OAuth2), Data JPA(QueryDSL/N+1), Boot Ops(Actuator/배포), Test(MockMvc/Testcontainers), JEUS WAS |
| **React/Next.js** | 11 | Hooks, 클래스 컴포넌트(레거시), 패턴(Compound/HOC), 성능최적화, App Router, Pages Router(레거시), Hook Form, TanStack Query, Zustand |
| **DB/SQL** | 9 | SQL Core(JOIN/CTE/윈도우함수), DDL/설계(인덱스/정규화), PostgreSQL, MySQL, Oracle(PL/SQL), Tibero(TmaxSoft), Prisma, Drizzle, Supabase |
| **Shell/터미널** | 11 | Bash 스크립팅(변수/함수/trap), Shell 도구(sed/awk/jq/curl), Shell 운영(cron/systemd/SSH), Git, Linux, macOS, Windows, 크로스플랫폼 |
| **프론트엔드** | 6 | Chart.js, D3.js, ECharts, Recharts, shadcn/ui, Tailwind CSS v4 |
| **백엔드/언어** | 5 | Go, JavaScript ES2024+, TypeScript, FastAPI, Vue 3 |
| **도구/테스트** | 6 | ESLint, Turborepo, Vite, Vitest, Testing Library, Playwright |
| **개발 실무** | 7 | 코드 리뷰, 디버깅, Git 워크플로, 성능 최적화, 리팩토링, 보안, TDD |
| **기타** | 6 | Auth.js, Framer Motion, Radix UI, date-fns, Zod |

### 사용자 지식 문서 (User Knowledge)

`.tgc/knowledge/` 폴더에 `.md` 또는 `.txt` 파일을 넣으면, AI가 자동으로 인덱싱하고 질문에 관련된 문서를 검색해서 참고합니다.

사내 개발 가이드, API 문서, 코딩 규칙, 온보딩 자료 등을 넣어두면 AI가 프로젝트 맥락을 이해하고 더 정확한 답변을 줍니다.

**지원 파일 형식**: `.md` (마크다운), `.txt` (텍스트)

**폴더 우선순위**: 프로젝트 로컬 > 글로벌 (둘 다 있으면 로컬 우선)

**사용자 지식 관리**:
- **추가**: `.md`/`.txt` 파일을 폴더에 드롭 → 다음 실행 시 자동 인덱싱
- **수정**: 파일을 직접 편집 → 다음 실행 시 반영
- **삭제**: 파일 삭제 → 다음 실행 시 제거
- 별도 등록/설정 없이 **파일 존재만으로 작동**

**예시 파일 구조**:
```
.tgc/knowledge/
├── company/
│   ├── coding-rules.md      # 사내 코딩 규칙
│   └── deploy-guide.md      # 배포 가이드
├── project/
│   ├── api-reference.md     # API 레퍼런스
│   └── db-schema.md         # DB 스키마 문서
└── onboarding.md            # 신규 입사자 온보딩
```

**검색 방식**: 키워드 AND 매칭. "배포 가이드"를 검색하면 "배포"와 "가이드" 모두 포함된 문서만 반환.

---

### macOS

**프로젝트 로컬** (해당 프로젝트에서만 참조):
```bash
# 프로젝트 루트에서
mkdir -p .tgc/knowledge
cp ~/Documents/my-guide.md .tgc/knowledge/
```

**글로벌** (모든 프로젝트에서 참조):
```bash
mkdir -p ~/.tgc/knowledge
cp ~/Documents/company-rules.md ~/.tgc/knowledge/
```

**확인**:
```bash
# 파일 목록 확인
ls -la .tgc/knowledge/
ls -la ~/.tgc/knowledge/

# techai 실행 후 디버그 로그에서 인덱싱 확인
# [USERDOCS] indexed 3 user documents from /path/to/.tgc/knowledge
```

> `.tgc/` 폴더를 `.gitignore`에 추가하면 개인 문서가 커밋되지 않습니다.

---

### Windows

**프로젝트 로컬** (해당 프로젝트에서만 참조):

```powershell
# 프로젝트 루트에서 (PowerShell)
New-Item -ItemType Directory -Force -Path .tgc\knowledge

# 파일 복사
Copy-Item C:\Users\사용자\Documents\my-guide.md .tgc\knowledge\
```

또는 파일 탐색기에서:
1. 프로젝트 폴더 열기
2. `.tgc` 폴더 생성 (숨김 폴더이므로 보기 → 숨긴 항목 체크)
3. `.tgc` 안에 `knowledge` 폴더 생성
4. `.md` / `.txt` 파일 복사

**글로벌** (모든 프로젝트에서 참조):

```powershell
# PowerShell
New-Item -ItemType Directory -Force -Path $HOME\.tgc\knowledge

# 파일 복사
Copy-Item C:\Users\사용자\Documents\company-rules.md $HOME\.tgc\knowledge\
```

또는 파일 탐색기에서:
1. `Win + R` → `%USERPROFILE%` → Enter
2. `.tgc` 폴더 생성 → 안에 `knowledge` 폴더 생성
3. 파일 복사

**확인**:
```powershell
# 파일 확인
Get-ChildItem .tgc\knowledge\
Get-ChildItem $HOME\.tgc\knowledge\
```

> **CMD 사용 시**: `mkdir .tgc\knowledge` 그리고 `copy 파일.md .tgc\knowledge\`

---

### Linux

**프로젝트 로컬** (해당 프로젝트에서만 참조):
```bash
# 프로젝트 루트에서
mkdir -p .tgc/knowledge
cp ~/docs/my-guide.md .tgc/knowledge/
```

**글로벌** (모든 프로젝트에서 참조):
```bash
mkdir -p ~/.tgc/knowledge
cp ~/docs/company-rules.md ~/.tgc/knowledge/
```

**확인**:
```bash
ls -la .tgc/knowledge/
ls -la ~/.tgc/knowledge/
```

> 서버 환경에서는 글로벌(`~/.tgc/knowledge/`)에 공통 문서를 넣어두면 어느 디렉토리에서 실행해도 참조됩니다.

---

## 온프레미스 (On-Premise) 버전

사내망 전용 빌드. API 엔드포인트와 모델이 고정되어 있고, 개인 API Key만 입력하면 사용 가능합니다.

- **API 엔드포인트**: `https://techai-web-prod.shinhan.com/v1`
- **모델**: `GPT-OSS-120B` (슈퍼택가이 / 플랜 / 개발 전 모드 동일)
- **설정 파일**: `~/.tgc-onprem/config.yaml` (일반 버전과 분리)

### 온프레미스 설치

#### macOS (Apple Silicon — M1/M2/M3/M4)

```bash
sudo cp dist/techai-onprem-darwin-arm64 /usr/local/bin/techai
sudo chmod +x /usr/local/bin/techai
xattr -d com.apple.quarantine /usr/local/bin/techai   # Gatekeeper 경고 시
techai
```

#### macOS (Intel)

```bash
sudo cp dist/techai-onprem-darwin-amd64 /usr/local/bin/techai
sudo chmod +x /usr/local/bin/techai
techai
```

#### Windows 10/11

1. `techai-onprem-windows-amd64.exe`를 `C:\techai\techai.exe`로 복사
2. `Win + R` → `sysdm.cpl` → 고급 → **환경 변수** → 사용자 변수 `Path` → **편집** → **새로 만들기** → `C:\techai` → **확인**
3. **모든 터미널 창을 닫고 새로 열기** (중요!)

```powershell
# 또는 PowerShell (관리자 권한)
New-Item -ItemType Directory -Force -Path C:\techai
Copy-Item techai-onprem-windows-amd64.exe C:\techai\techai.exe
[System.Environment]::SetEnvironmentVariable("Path",
  $env:Path + ";C:\techai",
  [System.EnvironmentVariableTarget]::User)

# 터미널 완전히 닫고 새로 열기 후 실행
techai
```

#### Linux

```bash
sudo cp dist/techai-onprem-linux-amd64 /usr/local/bin/techai
sudo chmod +x /usr/local/bin/techai
techai
```

### 온프레미스 첫 실행

첫 실행 시 자동으로 API Key 입력 위저드가 실행됩니다:

```
  택가이코드 설정
  API Base URL [https://techai-web-prod.shinhan.com/v1]:    ← 엔터 (기본값 사용)
  API Key: tg-your-api-key-here                             ← 발급받은 키 입력
```

설정은 `~/.tgc-onprem/config.yaml`에 저장됩니다.

### API Key 변경

```bash
# 방법 1: 설정 위저드 다시 실행
techai --setup

# 방법 2: 설정 초기화 후 재설정
techai --reset

# 방법 3: 실행 중 명령어
/setup

# 방법 4: 직접 파일 수정
vi ~/.tgc-onprem/config.yaml      # macOS/Linux
notepad %USERPROFILE%\.tgc-onprem\config.yaml   # Windows
```

### 온프레미스 설정 파일

```yaml
api:
  base_url: "https://techai-web-prod.shinhan.com/v1"
  api_key: "tg-your-api-key"
models:
  super: "GPT-OSS-120B"
  dev: "GPT-OSS-120B"
```

### 온프레미스 빌드 결과물

```
dist/
├── techai-onprem-darwin-arm64       # macOS Apple Silicon
├── techai-onprem-darwin-amd64       # macOS Intel
├── techai-onprem-windows-amd64.exe  # Windows
├── techai-onprem-linux-amd64        # Linux x64
└── techai-onprem-linux-arm64        # Linux ARM
```

## 빌드

```bash
make build          # 현재 플랫폼 → ./techai
make build-all      # 크로스 컴파일 (macOS/Windows/Linux × amd64/arm64)
make build-onprem   # 온프레미스 크로스 컴파일 (5개 플랫폼)
make install        # go install
make test           # 테스트
make lint           # go vet
make run            # 빌드 + 실행
make clean          # 정리
```

### 빌드 결과물

```
dist/
├── techai-darwin-arm64              # macOS Apple Silicon
├── techai-darwin-amd64              # macOS Intel
├── techai-windows-amd64.exe         # Windows
├── techai-linux-amd64               # Linux x64
├── techai-linux-arm64               # Linux ARM
├── techai-onprem-darwin-arm64       # 온프레미스 macOS Apple Silicon
├── techai-onprem-darwin-amd64       # 온프레미스 macOS Intel
├── techai-onprem-windows-amd64.exe  # 온프레미스 Windows
├── techai-onprem-linux-amd64        # 온프레미스 Linux x64
└── techai-onprem-linux-arm64        # 온프레미스 Linux ARM
```

## 기술 스택

| 패키지 | 용도 |
|--------|------|
| `charm.land/bubbletea/v2` | TUI 프레임워크 (Kitty keyboard protocol) |
| `charm.land/lipgloss/v2` | 터미널 스타일링 |
| `charm.land/bubbles/v2` | 텍스트 입력, 뷰포트 컴포넌트 |
| `charm.land/glamour/v2` | 마크다운 렌더링 |
| `sashabaranov/go-openai` | OpenAI-compatible API 클라이언트 |
| `gopkg.in/yaml.v3` | YAML 설정 파싱 |

## 프로젝트 구조

```
택가이코드/
├── cmd/tgc/main.go              # 엔트리포인트
├── internal/
│   ├── app/app.go               # 메인 TUI 앱 (Model/Update/View)
│   ├── ui/                      # UI 컴포넌트
│   │   ├── styles.go            # 색상/스타일 정의
│   │   ├── chat.go              # 메시지 렌더링, 마크다운, 상태바
│   │   ├── palette.go           # 커맨드 팔레트 (Ctrl+K)
│   │   ├── menu.go              # 메뉴 오버레이 (Esc)
│   │   ├── super.go             # 로고, 모드 정보 박스
│   │   └── tabbar.go            # 탭 바
│   ├── llm/                     # LLM 통신
│   │   ├── client.go            # OpenAI-compatible 스트리밍
│   │   ├── models.go            # 모델 정의 + 컨텍스트 윈도우
│   │   ├── prompt.go            # 모드별 시스템 프롬프트
│   │   ├── compaction.go        # 히스토리 압축 (90% 자동)
│   │   └── environment.go       # 환경 프로브 (40+ 도구 감지)
│   ├── tools/                   # AI 도구
│   │   ├── registry.go          # 도구 등록/실행 (14개 + MCP)
│   │   ├── file.go              # 파일 도구 (Fuzzy 4단계 편집)
│   │   ├── snapshot.go          # 파일 스냅샷 + /undo
│   │   ├── search.go            # grep/glob (.gitignore 존중)
│   │   ├── gitignore.go         # .gitignore 파서
│   │   ├── shell.go             # 셸 명령 도구
│   │   ├── git.go               # Git 도구 (status/diff/log)
│   │   ├── hashline.go          # 해시 앵커 편집
│   │   └── diagnostics.go       # 코드 진단
│   ├── mcp/                     # MCP 클라이언트
│   │   ├── types.go             # JSON-RPC 프로토콜 타입
│   │   ├── client.go            # stdio/sse 트랜스포트
│   │   └── manager.go           # 멀티 서버 관리
│   ├── multi/                   # 멀티 에이전트 시스템
│   │   ├── orchestrator.go      # 전략 실행기
│   │   └── strategy.go          # Review/Consensus/Scan
│   ├── knowledge/               # 지식 문서 RAG (3단계 검색)
│   │   ├── store.go             # 임베디드 문서 로드 + BM25 검색
│   │   ├── extractor.go         # 키워드 추출 + 유의어 확장
│   │   ├── injector.go          # 3단계 파이프라인 오케스트레이션
│   │   └── userdocs.go          # 사용자 문서 스캔
│   ├── session/store.go         # SQLite 세션 영속화
│   ├── companion/               # 브라우저 대시보드 (SSE)
│   ├── config/config.go         # 설정 (YAML + env + MCP)
│   ├── gitinfo/gitinfo.go       # Git 브랜치/dirty HUD
│   └── agents/auto.go           # 자율 모드 로직
├── knowledge/docs/              # 내장 지식 문서 (81개)
├── web/                         # 컴패니언 웹 UI
├── frontend/                    # Vite + React 프론트엔드
├── Makefile                     # 빌드 스크립트
└── go.mod
```

## 라이선스

MIT
