# TECHAI IDE — Feature List (v0.1.0)

> 빌드: macOS 12MB / Windows 12MB
> 기술: Go + Wails v2 + React + TypeScript + CodeMirror 6

---

## Core

| # | 기능 | 상태 |
|---|------|------|
| 1 | Activity Bar (Explorer/Search/Git/Settings/Account) | ✅ |
| 2 | 파일 트리 (실제 프로젝트, 3단계 깊이) | ✅ |
| 3 | 파일 트리 자동 새로고침 (AI file_write 시) | ✅ |
| 4 | 파일 트리 컨텍스트 메뉴 (새파일/이름변경/삭제) | ✅ |
| 5 | 폴더 열기 다이얼로그 (IDE + 시스템 메뉴) | ✅ |
| 6 | 파일 아이콘 60종+ (Go/TS/JS/Python/Rust/Java/C++/PHP 등) | ✅ |
| 7 | 폴더 아이콘 15종 색상 구분 (src/test/public/config 등) | ✅ |

## Editor

| # | 기능 | 상태 |
|---|------|------|
| 8 | CodeMirror 6 구문 강조 (15개 언어) | ✅ |
| 9 | 언어: Go, TS, JS, Python, Rust, Java, PHP, C/C++, SQL, YAML, JSON, CSS, HTML, Markdown | ✅ |
| 10 | 파일 탭 (다중 파일, 닫기, 수정 표시) | ✅ |
| 11 | Cmd+S 저장 (토스트 알림) | ✅ |
| 12 | Cmd+W 탭 닫기 (미저장 확인) | ✅ |
| 13 | Cmd+F 파일 내 검색 | ✅ |
| 14 | Cmd+P 파일 빠른 열기 (퍼지 검색) | ✅ |
| 15 | 브레드크럼 (파일 경로) | ✅ |
| 16 | Ln/Col 표시 + 언어 감지 (StatusBar) | ✅ |
| 17 | 줄번호 + 활성줄 하이라이트 | ✅ |
| 18 | 괄호 매칭 + 자동 닫기 | ✅ |
| 19 | 코드 접기 (foldGutter) | ✅ |
| 20 | 들여쓰기 자동 정리 (indentOnInput) | ✅ |
| 21 | Undo/Redo 히스토리 | ✅ |
| 22 | Tab → 4 spaces | ✅ |
| 23 | AI 파일 수정 시 에디터 자동 새로고침 | ✅ |

## AI Chat

| # | 기능 | 상태 |
|---|------|------|
| 24 | AI 채팅 (Qwen3-Coder 스트리밍) | ✅ |
| 25 | 마크다운 렌더링 (코드블록 + 인라인 코드) | ✅ |
| 26 | 도구 10개 (file_read/write, shell, grep, glob, list, git x3, apply_patch) | ✅ |
| 27 | text tool_call 파싱 (3포맷: tool_call, function, parameter) | ✅ |
| 28 | .techai.md 프로젝트 컨텍스트 자동 로드 | ✅ |
| 29 | Knowledge Packs 74개 (체크박스 ON/OFF, 프로젝트 타입 자동 감지) | ✅ |
| 30 | 슬래시 명령어 (/clear /export /model /help) | ✅ |
| 31 | 채팅 내보내기 (markdown 파일) | ✅ |
| 32 | 도구 호출 표시 (>> 파란색, << 초록색) | ✅ |

## Terminal

| # | 기능 | 상태 |
|---|------|------|
| 33 | 실제 터미널 (PTY + bash/zsh) | ✅ |
| 34 | 다중 터미널 탭 (+ 새 탭, X 닫기) | ✅ |
| 35 | 쉘 선택 드롭다운 (zsh/bash/fish/powershell) | ✅ |
| 36 | Ctrl+C / Ctrl+D 지원 | ✅ |
| 37 | ANSI escape code 제거 | ✅ |
| 38 | Cmd+J / Cmd+` 터미널 토글 | ✅ |

## Git

| # | 기능 | 상태 |
|---|------|------|
| 39 | Git 패널 (실제 git status) | ✅ |
| 40 | Git Stage (+ 버튼) | ✅ |
| 41 | Git Commit (메시지 입력 + Enter) | ✅ |
| 42 | Inline Diff (파일 클릭 → +초록/-빨강) | ✅ |
| 43 | Git Graph (커밋 히스토리 시각화) | ✅ |
| 44 | 브랜치 목록 (현재 브랜치 하이라이트) | ✅ |
| 45 | 브랜치/태그 칩 색상 구분 | ✅ |

## Search

| # | 기능 | 상태 |
|---|------|------|
| 46 | 프로젝트 전체 검색 (실시간 300ms 디바운스) | ✅ |
| 47 | Include 필터 (*.go, src/) | ✅ |
| 48 | Exclude 필터 (*.test.*, node_modules) | ✅ |
| 49 | 파일별 그룹핑 + 접기/펼치기 | ✅ |
| 50 | 다중 매치 하이라이트 (노란색) | ✅ |
| 51 | 결과 카운트 (N results in M files) | ✅ |

## UI / UX

| # | 기능 | 상태 |
|---|------|------|
| 52 | 11개 테마 (Slate/Cursor/Linear/GitHub/Dracula/Nord/OneDark/Monokai/Solarized/Claude/Vercel) | ✅ |
| 53 | Settings 페이지 (API URL/Key/Model) | ✅ |
| 54 | 패널 리사이즈 (사이드바/터미널/채팅 드래그) | ✅ |
| 55 | 알림 토스트 (저장/에러, 우하단 슬라이드) | ✅ |
| 56 | Welcome 화면 (단축키 안내) | ✅ |
| 57 | About 다이얼로그 | ✅ |
| 58 | Live Server (HTML 파일 우클릭 → 브라우저) | ✅ |
| 59 | Open in Browser (HTML 파일) | ✅ |

## System Menu (macOS / Windows)

| # | 기능 | 상태 |
|---|------|------|
| 60 | File → Open Folder / Save / Save All / Close Tab / Settings | ✅ |
| 61 | Edit → Undo / Redo / Cut / Copy / Paste / Find / Find in Files | ✅ |
| 62 | View → Explorer / Search / Git / Sidebar / Terminal / Quick Open / Theme | ✅ |
| 63 | Terminal → New Terminal | ✅ |
| 64 | Help → About / Keyboard Shortcuts | ✅ |

## Keyboard Shortcuts

| 단축키 | 기능 |
|--------|------|
| Cmd+S | 저장 |
| Cmd+W | 탭 닫기 |
| Cmd+P | 파일 빠른 열기 |
| Cmd+F | 파일 내 검색 |
| Cmd+Shift+F | 프로젝트 검색 |
| Cmd+B | 사이드바 토글 |
| Cmd+J / Cmd+` | 터미널 토글 |
| Cmd+1 | Explorer |
| Cmd+2 | Search |
| Cmd+3 | Git |
| Cmd+, | 테마 선택 |
| Cmd+O | 폴더 열기 |

## Build

| 플랫폼 | 크기 |
|--------|------|
| macOS ARM/Intel | 12MB |
| Windows x64 | 12MB |

---

**Total: 64 features**
