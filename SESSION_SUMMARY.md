# TECHAI CODE — 세션 작업 요약 (2026-04-23 ~ 04-24)

---

## 1. TUI 엔진 강화 (v0.9.6)

### Qwen3-Coder 온프레미스 전환
- Makefile 온프레미스 모델: GPT-OSS-120B → **Qwen3-Coder-30B**
- `MigrateModelsIfNeeded()`: 최초 실행 시 config.yaml 자동 마이그레이션
- `.model-migrated` 마커로 1회만 실행

### text tool_call 파싱 대폭 강화
사내망 프록시가 OpenAI tool_calls를 변환 못 할 때 텍스트에서 추출:

| 포맷 | 예시 |
|------|------|
| `<tool_call>` | `<tool_call>{"name":"...","arguments":{...}}</tool_call>` |
| `<function=name>` | `<function=file_read>{"path":"..."}</function>` |
| `<parameter=key>` | `<function=list_files> <parameter=path> . </tool_call>` |

추가 방어:
- `<think>` 태그 제거 (오탐 방지)
- `<|tool_call|>` 파이프 변형 지원
- arguments escaped string 자동 변환
- 미닫힌 태그 복구
- 텍스트+JSON 혼합 preamble 스킵
- 스트리밍 중 tool_call 태그 UI 노출 차단
- 부분 태그 청크 분리 대응 (`<function ` + `=name>`)
- tcMap + 텍스트 tool_call 혼재 시 병합

### 테스트
- **53개** 자동 테스트 (internal/llm/client_test.go)
- 기존 테스트 전부 통과 확인

### 릴리스
- v0.9.6 태그 + GitHub 릴리스 (16개 바이너리)
- macOS ARM/Intel + Windows + Linux + 온프레미스 + Gemma + Kimi

---

## 2. IDE 설계 & 디자인 (designs-v1 → v3)

### 디자인 진화
| 버전 | 특징 |
|------|------|
| designs.html (v1) | 6테마, 에디터+채팅, 기본 레이아웃 |
| designs-v2.html | 11테마, 터미널 추가, 도구 로그, 실제 코드 |
| **designs-v3.html** | Activity Bar + 버블 채팅 + 깊은 블랙 톤 + Geist 폰트 + 11테마 |

### 기술 스택 결정
- **Wails v2** (Go + 시스템 WebView) — Electron 대비 1/15 크기
- React + TypeScript + Vite
- CodeMirror 6 (구문 강조)
- xterm.js (터미널)
- Lucide React (아이콘)

---

## 3. IDE 구현 (v0.1.0 → v0.2.0)

### 프로젝트 생성 → 실행까지 과정
1. `wails init -n techai-ide -t react-ts`
2. Go 백엔드: app.go (파일 API) + chat.go (LLM 엔진) 작성
3. React 프론트엔드: designs-v3 스타일 적용
4. 첫 빌드 → 18MB (개발) → 10MB (프로덕션)
5. 채팅 연결 → 스트리밍 중복 텍스트 버그 수정 (이벤트 리스너 cleanup)
6. 기능 순차 추가 → v0.1.0 (91개) → v0.2.0 (116개)

### Go 백엔드 파일 (10개)
```
app.go          — 파일 시스템 API, 경로 보안, LiveServer, 최근 프로젝트
chat.go         — LLM 스트리밍, 도구 실행 (10개), 히스토리 관리
config.go       — .tgc/config.yaml 로더 (TUI와 공유)
git.go          — Git status/diff/graph/stage/commit/branch/checkout/pull/push
terminal.go     — PTY 터미널, 다중 세션, 쉘 선택, 리사이즈
knowledge.go    — Knowledge Packs 74개 동적 스캔, 토글, 시스템 프롬프트 주입
toolparse.go    — text tool_call 파싱 (3포맷, TUI에서 포팅)
settings.go     — API 설정 관리 (Get/Save + chat 재초기화)
session.go      — 채팅 세션 저장/불러오기/삭제
main.go         — Wails 앱 + 시스템 메뉴 (File/Edit/View/Terminal/Help)
```

### React 컴포넌트 (17개)
```
App.tsx              — 메인 레이아웃, 단축키, 패널 상태, 드래그앤드롭
Editor.tsx           — 에디터 탭, 저장, 브레드크럼, 자동저장, 마크다운 미리보기
CodeEditor.tsx       — CodeMirror 6 래퍼 (15언어, 다중커서, 줄 조작, 주석 토글)
ChatPanel.tsx        — AI 채팅, 스트리밍, 마크다운, Knowledge, 슬래시 명령, 히스토리
FileTree.tsx         — 파일 트리, 60종 아이콘, 컨텍스트 메뉴, 필터, 드래그
Terminal.tsx         — xterm.js, 다중 탭, 쉘 선택, ANSI 컬러
GitPanel.tsx         — Git 패널, Stage, Commit, Inline Diff
GitGraph.tsx         — 커밋 시각화, 브랜치 checkout, Pull/Push
SearchPanel.tsx      — 실시간 검색, Include/Exclude 필터, 하이라이트
DiffView.tsx         — side-by-side Diff (+초록/-빨강)
ThemePicker.tsx      — 11개 테마 모달
SettingsPanel.tsx    — API URL/Key/Model 설정
QuickOpen.tsx        — Cmd+P 파일 퍼지 검색
CommandPalette.tsx   — Cmd+Shift+P 명령어 팔레트
Toast.tsx            — 알림 토스트 (우하단 슬라이드)
AboutDialog.tsx      — About 다이얼로그
ResizeHandle.tsx     — 패널 리사이즈 드래그
```

### 테스트 (49개)
```
app_test.go         — 13: List/Read/Write/Delete/Rename/Search/Cwd
toolparse_test.go   — 14: tool_call 3포맷 + think + parameter
config_test.go      — 4: 설정 기본값 + maskKey
git_test.go         — 8: status/diff/graph/stage/commit/branch/checkout
knowledge_test.go   — 7: .techai.md + Go/Node 감지 + toggle
```

---

## 4. 코드 리뷰 & 보안 수정

### 발견된 29개 이슈 → 28개 수정

**Critical (4/4 수정)**
- 파일 워치 고루틴 누수 → `watcherDone` 채널로 정상 종료
- race condition → `chatMu sync.Mutex` 추가
- 경로 탐색 취약점 → `safePath()` 프로젝트 외부 접근 차단
- 터미널 고루틴 누수 → `closed` 플래그 + `cmd.Wait()`

**Medium (10/10 수정)**
- GitLog n>=10 포맷 버그 → `fmt.Sprintf` 사용
- shell injection → `filepath.WalkDir` Go native 교체
- 스트리밍 context → `app.ctx` 전달 (앱 종료 시 취소)
- 채팅 히스토리 → 80개 제한 (overflow 방지)
- Settings 동기화 → `chatMu` 잠금
- LiveServer 포트 → 5500-5600 자동 탐색
- 터미널 mutex 보강
- 기타 3건

**Low (14/15 수정)**
- 마크다운 링크 지원
- 검색 100개 제한 안내
- 기타 12건

---

## 5. 전체 기능 목록 (116개)

### 카테고리별
| 카테고리 | 수 |
|----------|-----|
| Core (파일트리, Activity Bar) | 7 |
| Editor (CodeMirror, 탭, 검색) | 18 |
| Preview (이미지, 마크다운, LiveServer) | 3 |
| AI Chat (스트리밍, 도구, Knowledge) | 11 |
| Terminal (xterm.js, PTY, 다중탭) | 6 |
| Git (Graph, Stage, Commit, Branch) | 10 |
| Search (실시간, 필터, 하이라이트) | 6 |
| UI/UX (테마, 토스트, 팔레트) | 13 |
| System Menu | 5 |
| Editor 단축키 (줄 복제/이동/삭제/주석) | 6 |
| Polish (줌, 복사, 파일크기, 타임스탬프) | 10 |
| 기타 (드래그, 전체화면, 히스토리) | 6+ |

### 키보드 단축키 (20+)
```
Cmd+S        저장              Cmd+W        탭 닫기
Cmd+P        파일 열기          Cmd+Shift+P  커맨드 팔레트
Cmd+F        찾기              Cmd+H        바꾸기
Cmd+G        줄 이동            Cmd+/        주석 토글
Cmd+Shift+D  줄 복제           Cmd+Shift+K  줄 삭제
Alt+↑↓       줄 이동            Cmd+B        사이드바
Cmd+J        터미널             Cmd+\        스플릿
Cmd+1/2/3    패널 전환          Cmd+,        테마
Cmd+O        폴더 열기          Cmd+±0       줌
F11          전체화면           Alt+Click    다중 커서
```

---

## 6. 바이너리 & 배포

| 제품 | 크기 | 플랫폼 |
|------|------|--------|
| TUI v0.9.6 | 25MB | macOS/Windows/Linux (16개) |
| IDE v0.2.0 | 12MB | macOS + Windows |

### 비교
| IDE | 크기 |
|-----|------|
| **TECHAI IDE** | **12MB** |
| VS Code | 350MB |
| Cursor | 500MB |
| Zed | 80MB |

---

## 7. 레포 구조

```
TECHAI_CODE/
├── cmd/tgc/              ← TUI 진입점
├── internal/             ← TUI 공유 엔진 (15개 패키지)
├── knowledge/            ← 81개 내장 문서 (go:embed)
├── web/                  ← 컴패니언 웹 UI
├── techai-ide/           ← GUI IDE (Wails + React)
│   ├── *.go              ← Go 백엔드 (10개)
│   ├── *_test.go         ← 테스트 (5개, 49 케이스)
│   ├── frontend/src/     ← React 컴포넌트 (17개)
│   ├── FEATURES.md       ← 116개 기능 리스트
│   ├── TODO.md           ← v0.3.0 로드맵
│   └── TEST_CHECKLIST.md ← 수동 테스트 80+ 항목
├── demo-supersol/        ← SuperSOL 데모 프로젝트
├── designs-v3.html       ← IDE 디자인 목업
├── DEV_GUIDE.md          ← TUI/IDE 개발 가이드
├── WORKSPACE_GUIDE.md    ← 5개 작업 영역 가이드
└── docs/                 ← 계획서, 로드맵, 스펙
```

---

## 8. 버전 히스토리

### TUI
| 태그 | 내용 |
|------|------|
| v0.9.6 | Qwen3-Coder 전환 + tool_call 파싱 53 테스트 |

### IDE
| 태그 | 내용 |
|------|------|
| ide-v0.1.0 | 91개 기능, 49 테스트, 12MB |
| ide-v0.2.0 | 116개 기능, xterm.js, 코드 리뷰 28건 수정 |

---

## 9. 다음 작업 (v0.3.0)

| 기능 | 난이도 |
|------|--------|
| LSP 연동 (gopls, ts-server) | 높 |
| VS Code Extension | 중 |
| 플러그인 시스템 | 높 |
| 다중 프로젝트 | 중 |
| 원격 서버 (SSH) | 높 |
