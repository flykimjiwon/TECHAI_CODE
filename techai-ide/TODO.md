# TECHAI IDE — Remaining TODO

> 현재: 68개 기능 완료 (v0.1.0, 12MB)
> 다음: 아래 항목 순차 구현

---

## 다음 구현 (우선순위 순)

| # | 기능 | 난이도 | 설명 |
|---|------|--------|------|
| 1 | **자동 저장** | 낮 | 수정 후 1초 디바운스 자동 저장 (설정에서 토글) |
| 2 | **채팅 히스토리 저장** | 중 | 세션별 대화 파일 저장/불러오기 |
| 3 | **마크다운 미리보기** | 중 | .md 파일 렌더링 뷰 (에디터 옆에 미리보기) |
| 4 | **Git checkout/branch** | 낮 | Git Graph에서 브랜치 전환/생성 |
| 5 | **프로젝트 최근 목록** | 낮 | Welcome 화면에 최근 열었던 폴더 표시 |
| 6 | **알림 뱃지** | 낮 | Activity Bar Git 아이콘에 변경파일 수 |
| 7 | **Diff 뷰 (side-by-side)** | 중 | AI 수정 전/후 비교 뷰 |
| 8 | **파일 워치** | 중 | 외부 변경 감지 → 에디터 자동 새로고침 |
| 9 | **탭 드래그 순서 변경** | 중 | 탭 드래그로 재배치 |

## 향후 고급 기능

| # | 기능 | 난이도 | 설명 |
|---|------|--------|------|
| 10 | **인라인 AI** | 높 | 코드 선택 → 우클릭 → AI에게 설명/수정 요청 |
| 11 | **다중 커서** | 높 | Cmd+D 같은 단어 다중 선택 |
| 12 | **미니맵** | 높 | CodeMirror minimap 확장 |
| 13 | **LSP 연동** | 높 | gopls, typescript-language-server → 자동완성/에러마커 |
| 14 | **드래그 앤 드롭** | 중 | 파일트리에서 에디터로 드래그 |
| 15 | **xterm.js 교체** | 중 | 터미널 ANSI 컬러/커서 위치 지원 |
| 16 | **VS Code Extension** | 중 | 택가이코드 채팅을 VS Code 사이드바로 |

---

## 아키텍처 노트

### 현재 구조
```
택가이코드/              ← 메인 레포
├── cmd/tgc/            ← TUI (Bubble Tea)
├── internal/           ← 공유 엔진 (LLM, 도구, 지식)
├── techai-ide/         ← GUI IDE (Wails + React)
│   ├── main.go         ← Wails 앱 (시스템 메뉴 포함)
│   ├── app.go          ← 파일 시스템 API
│   ├── chat.go         ← LLM 채팅 엔진
│   ├── git.go          ← Git 연동
│   ├── terminal.go     ← PTY 터미널
│   ├── knowledge.go    ← Knowledge Packs
│   ├── toolparse.go    ← text tool_call 파싱
│   ├── settings.go     ← 설정 관리
│   ├── config.go       ← .tgc/config.yaml 로더
│   └── frontend/       ← React + TypeScript
│       └── src/components/  ← 15개 컴포넌트
└── docs/TECHAI_IDE_PLAN.md  ← 전체 계획서
```

### 향후 엔진 통합 (방식 C → B)
현재: techai-ide가 자체 LLM 클라이언트 보유
향후: internal/ 패키지를 직접 import (같은 go.mod 또는 replace)
```
