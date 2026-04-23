# TECHAI CODE — 워크스페이스 가이드

> 이 레포에는 5개의 독립 작업 영역이 있습니다.
> 각 영역은 별도로 빌드/실행되며, 다른 영역에 영향을 주지 않습니다.
> 새 세션에서 이어 작업할 때 이 문서를 참고하세요.

---

## 작업 영역 한눈에 보기

```
TECHAI_CODE/
│
├── 1️⃣  TUI ─────────── cmd/tgc/ + internal/
│                        Go CLI 코딩 에이전트 (Bubble Tea)
│                        현재: v0.9.6 | 빌드: make build
│
├── 2️⃣  IDE ─────────── techai-ide/
│                        데스크톱 GUI IDE (Wails + React)
│                        현재: v0.1.0 | 빌드: wails build
│
├── 3️⃣  Demo ────────── demo-supersol/
│                        SuperSOL 시연 프로젝트 (Next.js)
│                        빌드: npm run dev (localhost:3000)
│
├── 4️⃣  Presentation ── index.html, presentation.html
│                        발표 슬라이드 (HTML)
│                        브라우저에서 열기
│
├── 5️⃣  Video ───────── demo-supersol/techai-demo-video/
│                        Remotion 데모 영상
│                        빌드: npx remotion render
│
└── 📚  Docs ─────────── docs/, DEV_GUIDE.md, designs-*.html
                         문서, 디자인 목업
```

---

## 1️⃣ TUI 작업

### 이어서 작업할 때 읽을 파일
```
DEV_GUIDE.md                    ← 빌드/구조/주요파일
docs/UX_STRUCTURE.md            ← TUI 화면 구조
internal/llm/client.go          ← LLM 스트리밍 + tool_call 파싱 (핵심)
internal/app/app.go             ← Bubble Tea 메인 모델
internal/tools/registry.go      ← 15개 도구
internal/config/config.go       ← 설정 + 마이그레이션
```

### 빌드 & 실행
```bash
# 개발
make build && ./techai

# 테스트
go test ./...

# 릴리스 (16개 바이너리)
make build-release

# 온프레미스 (사내망)
make build-onprem

# GitHub 릴리스
gh release create v0.9.7 dist/* --title "..." --notes "..."
```

### 버전 태그
```bash
git tag v0.9.7 && git push origin v0.9.7
```

### 현재 상태 (v0.9.6)
- Qwen3-Coder-30B 기본 모델
- text tool_call 파싱 3포맷 (tool_call, function, parameter)
- 온프레미스 config 자동 마이그레이션 (GPT-OSS → Qwen3)
- Knowledge Packs 81개, auto-prefetch, 시스템 프롬프트 430토큰
- 53개 테스트 (internal/llm/client_test.go)

### 다음 할 일
- docs/superpowers/specs/2026-04-10-unified-roadmap-design.md 참고
- 컴패니언 브라우저 대시보드 강화 (web/ + internal/companion/)
- LSP 연동 (Phase 5)

---

## 2️⃣ IDE 작업

### 이어서 작업할 때 읽을 파일
```
DEV_GUIDE.md                          ← IDE 섹션 (빌드/구조/패턴)
techai-ide/FEATURES.md                ← 68개 기능 전체 리스트
techai-ide/TODO.md                    ← 남은 구현 계획
docs/TECHAI_IDE_PLAN.md               ← 전체 계획서 (기술스택, API, 테마)
designs-v3.html                       ← 디자인 목업 (브라우저에서 열기)
```

### 사전 요구
```bash
# 최초 1회
go install github.com/wailsapp/wails/v2/cmd/wails@latest
cd techai-ide/frontend && npm install
```

### 빌드 & 실행
```bash
cd techai-ide

# 개발 (Hot Reload)
wails dev

# 프로덕션 빌드
wails build                              # macOS ARM
wails build -platform darwin/amd64       # macOS Intel
wails build -platform windows/amd64      # Windows

# Go API 추가 후 바인딩 재생성
wails generate module
```

### 버전 태그
```bash
git tag ide-v0.1.0 && git push origin ide-v0.1.0
```

### 주요 컴포넌트 (15개)
```
frontend/src/components/
├── App.tsx              ← 메인 레이아웃 + 단축키 + 상태
├── Editor.tsx           ← 에디터 (탭, 저장, 브레드크럼)
├── CodeEditor.tsx       ← CodeMirror 6 (15개 언어)
├── ChatPanel.tsx        ← AI 채팅 (스트리밍, 도구, Knowledge)
├── FileTree.tsx         ← 파일 트리 (60종 아이콘)
├── Terminal.tsx         ← 터미널 (다중 탭, 쉘 선택)
├── GitPanel.tsx         ← Git (Stage, Commit, Diff)
├── GitGraph.tsx         ← 커밋 히스토리 시각화
├── SearchPanel.tsx      ← 실시간 검색 + 필터
├── ThemePicker.tsx      ← 11개 테마
├── SettingsPanel.tsx    ← API 설정
├── QuickOpen.tsx        ← Cmd+P 파일 열기
├── CommandPalette.tsx   ← Cmd+Shift+P
├── Toast.tsx            ← 알림
├── ResizeHandle.tsx     ← 패널 리사이즈
└── AboutDialog.tsx      ← About
```

### Go 백엔드 (8개)
```
techai-ide/
├── main.go              ← Wails 앱 + 시스템 메뉴
├── app.go               ← 파일 API (Read/Write/Delete/Rename/OpenFolder/LiveServer)
├── chat.go              ← LLM 채팅 (스트리밍 + 도구 10개)
├── git.go               ← Git (Status/Diff/Graph/Stage/Commit/Branches)
├── terminal.go          ← PTY (Start/Write/Resize/SetShell)
├── knowledge.go         ← Knowledge Packs 74개 (스캔 + 토글)
├── toolparse.go         ← text tool_call 파싱 (3포맷)
├── settings.go          ← 설정 (Get/Save)
└── config.go            ← .tgc/config.yaml 로더
```

### 현재 상태 (v0.1.0, 68기능, 12MB)
- CodeMirror 6 구문강조 (15언어)
- AI 채팅 + 도구 실행 + text tool_call 파싱
- 실제 PTY 터미널 (다중 탭, 쉘 선택)
- Git Graph + Stage + Commit + Diff
- 실시간 검색 + Include/Exclude 필터
- 11개 테마, 커맨드 팔레트, 스플릿 에디터
- 시스템 메뉴 (File/Edit/View/Terminal/Help)
- Live Server (HTML 브라우저 열기)

### 다음 할 일 (TODO.md 참고)
1. 자동 저장
2. 채팅 히스토리 저장
3. 마크다운 미리보기
4. Git checkout/branch
5. 인라인 AI (코드 선택 → AI 요청)
6. LSP 연동

---

## 3️⃣ Demo 작업

### 이어서 작업할 때 읽을 파일
```
demo-supersol/DEMO_GUIDE.md     ← 3개 시나리오 + 프롬프트
demo-supersol/.techai.md         ← 프로젝트 컨텍스트
```

### 실행
```bash
cd demo-supersol
npm install   # 최초 1회
npm run dev   # localhost:3000
```

### 구조
```
demo-supersol/
├── src/views/
│   ├── HomePage.tsx       ← 시나리오 1: 홈 화면 수정
│   ├── FinancePage.tsx    ← 시나리오 2: 금융 페이지
│   └── BenefitsPage.tsx   ← 시나리오 3: 혜택 페이지
├── src/components/        ← AccountCard, TransactionItem, ProgressBar
├── src/data/mock.ts       ← 김지원 고객 목 데이터
└── DEMO_GUIDE.md          ← 시연 시나리오 + 프롬프트
```

### 시연 방법
1. `npm run dev`로 서버 실행
2. 택가이코드(TUI) 또는 IDE 실행
3. DEMO_GUIDE.md의 프롬프트를 순서대로 입력
4. 브라우저에서 실시간 변경 확인

---

## 4️⃣ Presentation 작업

### 파일
```
index.html               ← 메인 발표 슬라이드 (17슬라이드)
presentation.html         ← 대체 슬라이드
designs-v2.html           ← IDE 디자인 목업 (11테마)
designs-v3.html           ← IDE 최종 디자인 (Activity Bar + 버블)
```

### 실행
```bash
open index.html           # 브라우저에서 열기
```

### 수정 시 참고
- 슬라이드 내용: index.html 내 각 `<section class="slide">` 수정
- /make-slide 스킬: `.claude/skills/make-slide/SKILL.md` 참고
- PDF 생성: Puppeteer 사용 (`capture-slides.mjs`)

---

## 5️⃣ Video 작업 (Remotion)

### 파일
```
demo-supersol/techai-demo-video/
├── src/
│   ├── DemoVideo.tsx      ← 메인 영상 컴포지션
│   ├── TitleSlide.tsx     ← 타이틀
│   ├── ScenarioSection.tsx ← 시나리오별 섹션
│   ├── ModelComparison.tsx ← 모델 비교
│   └── Root.tsx           ← Remotion 루트
├── public/assets/         ← 스크린샷, 영상 소스
└── out/DemoVideo.mp4      ← 렌더링 결과
```

### 실행
```bash
cd demo-supersol/techai-demo-video
npm install   # 최초 1회

# 미리보기
npx remotion preview

# 렌더링
npx remotion render src/index.ts DemoVideo out/DemoVideo.mp4
```

---

## 커밋 컨벤션

각 작업 영역별 접두사 사용:

```
[TUI]   feat: add LSP integration
[IDE]   fix: editor tab crash on large files
[DEMO]  update: scenario 2 prompt
[DOCS]  docs: update WORKSPACE_GUIDE
[PRES]  update: slide 14 model comparison
[VIDEO] feat: add scenario 3 section
```

---

## 새 세션 시작 체크리스트

### TUI 이어서 할 때
- [ ] `DEV_GUIDE.md` 읽기
- [ ] `git log --oneline -5` 로 최근 작업 확인
- [ ] `make build && ./techai` 로 현재 상태 확인
- [ ] `go test ./...` 전체 테스트 통과 확인

### IDE 이어서 할 때
- [ ] `DEV_GUIDE.md` IDE 섹션 읽기
- [ ] `techai-ide/FEATURES.md` 현재 기능 확인
- [ ] `techai-ide/TODO.md` 다음 할 일 확인
- [ ] `cd techai-ide && wails dev` 로 실행 확인

### Demo 이어서 할 때
- [ ] `demo-supersol/DEMO_GUIDE.md` 읽기
- [ ] `cd demo-supersol && npm run dev` 실행

### 발표자료 수정할 때
- [ ] `open index.html` 로 현재 상태 확인
- [ ] `open designs-v3.html` 로 IDE 디자인 확인

### 영상 만들 때
- [ ] `cd demo-supersol/techai-demo-video && npx remotion preview` 미리보기

---

## 환경 변수

TUI와 IDE 모두 동일한 설정 사용:

```bash
# .tgc/config.yaml 대신 환경 변수로도 설정 가능
export TGC_API_BASE_URL=https://api.novita.ai/openai
export TGC_API_KEY=sk-...
export TGC_MODEL_SUPER=qwen/qwen3-coder-30b-a3b-instruct
```

---

## 빠른 명령어 모음

```bash
# TUI
make build && ./techai
make test
make build-release

# IDE
cd techai-ide && wails dev
cd techai-ide && wails build
cd techai-ide && wails generate module

# Demo
cd demo-supersol && npm run dev

# Presentation
open index.html
open designs-v3.html

# Video
cd demo-supersol/techai-demo-video && npx remotion preview

# Git
git tag v0.9.7 && git push origin v0.9.7        # TUI 버전
git tag ide-v0.1.0 && git push origin ide-v0.1.0 # IDE 버전
```
