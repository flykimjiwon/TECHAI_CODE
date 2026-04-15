# 택가이코드 기능 목록

> 최종 업데이트: v0.6.0-rc1 (2026-04-10)
> 빌드 변형: Default (Novita) / Onprem (신한) / Gemma

---

## 빌드 & 배포

| 변형 | Super/Plan 모델 | Dev 모델 | API 엔드포인트 | 설정 디렉토리 |
|------|----------------|----------|---------------|-------------|
| **Default** | openai/gpt-oss-120b | qwen/qwen3-coder-30b | api.novita.ai | `~/.tgc/` |
| **Onprem** | openai/gpt-oss-120b | qwen/qwen3-coder-30b | techai-web-prod.shinhan.com | `~/.tgc-onprem/` |
| **Gemma** | google/gemma-4-31b-it | google/gemma-4-31b-it | api.novita.ai | `~/.tgc-gemma/` |

**플랫폼**: darwin-arm64, darwin-amd64, windows-amd64, linux-amd64, linux-arm64 (변형당 5개, 총 15 바이너리)

**단일 바이너리**: 외부 의존성 제로. 복사 후 실행. (~20MB)

---

## 핵심 기능 (v0.5.0 기존)

### 1. 3-모드 시스템

| 모드 | 이름 | 모델 | 도구 | 설명 |
|------|------|------|------|------|
| Super | 슈퍼택가이 | gpt-oss-120b | 13개 (전체) | 만능 — 의도 감지, 코드+대화+분석 |
| Dev | 개발 | qwen3-coder-30b | 13개 (전체) | 코딩 특화 — 파일 CRUD, 코드 생성/수정 |
| Plan | 플랜 | gpt-oss-120b | 9개 (읽기전용) | 분석/계획 — 읽기 전용, 구조 파악 |

- `Tab` 키로 모드 전환
- 모드별 시스템 프롬프트 자동 교체
- 모드 전환 시 대화 히스토리 유지 (프롬프트만 교체)

### 2. 도구 시스템 — 기본 7개

| 도구 | 기능 | 모드 |
|------|------|------|
| `file_read` | 파일 읽기 (50KB 제한) | Super/Dev/Plan |
| `file_write` | 파일 생성/덮어쓰기 | Super/Dev |
| `file_edit` | 문자열 치환 편집 | Super/Dev |
| `list_files` | 디렉토리 목록 (재귀 지원) | Super/Dev/Plan |
| `shell_exec` | 셸 명령 실행 (30초 타임아웃, 위험 명령 차단) | Super/Dev/Plan |
| `grep_search` | 정규식 파일 내용 검색 | Super/Dev/Plan |
| `glob_search` | 글로브 패턴 파일 찾기 | Super/Dev/Plan |

- 도구 루프 최대 20회 (무한 루프 방지)
- 출력 30,000자 제한 (자동 truncate)
- `.git`, `node_modules`, `dist` 자동 제외

### 3. SSE 스트리밍

- OpenAI-compatible SSE 프로토콜
- 실시간 토큰 수신 (tok/s 표시)
- 연결중/수신중/응답지연/응답없음 상태 표시
- `Esc` 또는 `Ctrl+C`로 스트리밍 중단
- 스트리밍 중 메시지 큐잉 (대기중 표시)

### 4. 마크다운 렌더링

- Glamour v2 (dark 테마) 기반 터미널 마크다운
- 코드 블록, 리스트, 제목 등 완전 지원
- 20줄 초과 메시지에 `[N lines]` 카운트 표시
- CJK 2바이트 문자 올바른 줄바꿈

### 5. Embedded Knowledge Store

- 38개 문서 빌드타임 임베딩 (`knowledge/` 디렉토리)
- BXM Tier 0 키워드 매칭
- 모드별/쿼리별 자동 주입 (8,192 토큰 버짓)
- `cmd/build-index` — 인덱스 생성 도구
- `cmd/scrape-bxm` — BXM 문서 수집 도구

### 6. 프로젝트 컨텍스트

- `.techai.md` 파일 자동 로드 (프로젝트 루트)
- 시스템 프롬프트에 자동 삽입
- OS, 아키텍처, Go 버전, 호스트명 등 시스템 정보 주입

### 7. 설정 시스템

- `~/.tgc/config.yaml` YAML 설정
- 환경변수 오버라이드: `TGC_API_BASE_URL`, `TGC_API_KEY`, `TGC_MODEL_SUPER`, `TGC_MODEL_DEV`
- 빌드타임 ldflags로 기본값 주입 (변형별 분리)
- 첫 실행 시 인터랙티브 셋업 위저드 (`/setup`)

### 8. 디버그 모드

- 빌드타임 `DebugMode=true`로 활성화
- `~/.tgc/debug.log` 타임스탬프 로그
- 스트림 시작/종료, 도구 호출/결과, 세션 이벤트 전부 기록
- 상태바에 `[DEBUG]` 표시

---

## Phase 1 기능 (v0.6.0 추가)

### 9. SQLite 세션 영속화

- `modernc.org/sqlite` (순수 Go, CGO 불필요)
- `~/.tgc/sessions.db` 자동 생성
- 슬래시 명령:
  - `/new` — 새 세션 시작
  - `/sessions` — 최근 10개 세션 목록 (ID, 제목, 모델, 날짜)
  - `/session <id>` — 이전 세션 복원 (히스토리 + UI 재구성)
- 첫 사용자 메시지가 자동으로 세션 제목이 됨
- 세션 저장 실패 시 자동 폴백 (메모리 전용, 비차단)

### 10. Git HUD 통합

- `internal/gitinfo` 패키지
- 상태바에 브랜치명 + dirty 표시 (`⎇ main*`)
  - 클린: 초록색, 더티: 노란색 + `*`
- 스트림 시작/도구 완료 시 자동 새로고침
- `/git` 명령으로 상세 상태 (branch, status, recent commits)
- git 미사용 디렉토리에서 조용히 비활성화

### 11. 스마트 컨텍스트 압축

3단계 자동 압축:

| 단계 | 트리거 | 동작 |
|------|--------|------|
| Stage 1 — Snip | 40+ 메시지 | 오래된 도구 출력(200자+)을 `[snipped: N lines]`로 교체 |
| Stage 2 — Micro | 항상 | 4,000자 초과 메시지를 head(2K)+tail(2K) 트렁케이트 |
| Stage 3 — LLM Summary | ctx 90%+ | 중간 히스토리를 LLM 요약으로 압축 (최근 10개 보존) |

- 상태바에 `ctx:XX%` 사용률 표시 (초록/노랑/빨강)
- 모델별 컨텍스트 윈도우 자동 인식:
  - gpt-oss-120b: 128K
  - qwen3-coder-30b: 262K

### 12. 모델 능력 레지스트리

- `internal/llm/capabilities.go`
- 모델별 컨텍스트 윈도우, 코딩 능력(Strong/Moderate/Weak/None), 역할(Agent/Assistant/Chat), 도구 지원 여부
- `provider/model-name` 형식 자동 파싱 (suffix 매칭)
- 미등록 모델은 32K/Moderate/Assistant 기본값

---

## Tier 1 포팅 기능 (hanimo → 택가이코드)

> 커밋: `0b6361f` — 6개 기능, +1,141줄

### 13. Hash-Anchored Editing (`hashline_read`, `hashline_edit`)

- **파일**: `internal/tools/hashline.go` (116줄)
- **기능**: 파일 읽기 시 각 줄에 MD5 4자리 해시 앵커 부여 (`1#a3f1| code`)
- **편집**: 시작/끝 앵커로 범위 지정, 해시 불일치 시 거부 (stale-edit 방지)
- **용도**: LLM이 오래된 파일 내용으로 편집하는 것을 방지
- **도구**:
  - `hashline_read` — 해시 앵커 포맷으로 파일 읽기
  - `hashline_edit` — 앵커 기반 범위 편집

### 14. 자율 모드 (`/auto`)

- **파일**: `internal/agents/auto.go` (32줄)
- **기능**: AI가 도구를 사용해 독립적으로 작업 수행
- **제어**:
  - `/auto` 토글 (ON/OFF)
  - `[AUTO_COMPLETE]` — 작업 완료 시 자동 정지
  - `[AUTO_PAUSE]` — 인간 입력 필요 시 대기
  - 최대 20회 반복 후 자동 정지
- **프롬프트**: 자율 모드 시 시스템 프롬프트에 `AutoPromptSuffix` 임시 삽입 (모드 해제 시 제거)

### 15. Git 도구 (`git_status`, `git_diff`, `git_log`)

- **파일**: `internal/tools/git.go` (52줄)
- **기능**: LLM이 도구로 Git 상태를 직접 조회
- **도구**:
  - `git_status` — `git status -s` (short format)
  - `git_diff` — `git diff` (staged 옵션)
  - `git_log` — `git log --oneline` (N개 지정)
- 10초 타임아웃, Plan 모드에서도 사용 가능

### 16. 코드 진단 (`diagnostics`, `/diagnostics`)

- **파일**: `internal/tools/diagnostics.go` (272줄)
- **기능**: 프로젝트 타입 자동 감지 → 적절한 린터 실행
- **지원 언어**:
  - Go: `go vet`
  - TypeScript: `npx tsc --noEmit`
  - JavaScript: `npx eslint --format compact`
  - Python: `ruff check`
- **출력**: 파일명:줄번호:심각도:메시지 구조화 포맷
- **사용**: 도구 호출 (`diagnostics`) 또는 슬래시 명령 (`/diagnostics [파일필터]`)

### 17. 커맨드 팔레트 (`Ctrl+K`)

- **파일**: `internal/ui/palette.go` (144줄)
- **기능**: VS Code 스타일 퍼지 검색 커맨드 팔레트
- **조작**:
  - `Ctrl+K` — 팔레트 열기
  - 타이핑 — 실시간 퍼지 필터 (라벨/설명/액션)
  - `↑↓` — 항목 이동
  - `Enter` — 선택 실행
  - `Esc` — 닫기
- **항목**: 새 세션, 세션 목록, 자율 모드, 진단, Git 상태, 화면 정리, API 설정, 도움말
- **UI**: 반투명 플로팅 박스, 둥근 테두리, 선택 항목 하이라이트

### 18. 메뉴 오버레이 (`Esc`)

- **파일**: `internal/ui/menu.go` (107줄)
- **기능**: 빠른 액세스 메뉴 (스트리밍 중이 아닐 때)
- **조작**:
  - `Esc` — 메뉴 열기 (비스트리밍 시)
  - `↑↓` — 항목 이동
  - `Enter` — 선택 실행
  - `Esc` — 닫기
- **항목**: 자율 모드, 진단 실행, Git 상태, 새 세션, 세션 목록, 설정, 화면 정리, 도움말
- **UI**: 택가이코드 타이틀, 플로팅 박스, 스크롤 인디케이터

---

## 키보드 단축키

| 키 | 동작 | 조건 |
|----|------|------|
| `Enter` | 메시지 전송 | 비스트리밍 |
| `Shift+Enter` | 줄바꿈 | 항상 |
| `Tab` | 모드 전환 (Super→Dev→Plan) | 비스트리밍 |
| `Ctrl+K` | 커맨드 팔레트 열기 | 비스트리밍 |
| `Esc` | 메뉴 열기 / 스트리밍 취소 | 상황별 |
| `Ctrl+C` | 스트리밍 취소 / 종료 | 항상 |
| `Ctrl+L` | 화면 정리 | 비스트리밍 |
| `PgUp/PgDn` | 스크롤 | 항상 |
| `Alt+↑/↓` | 3줄 스크롤 | 항상 |

---

## 슬래시 명령

| 명령 | 기능 |
|------|------|
| `/new` | 새 세션 시작 |
| `/sessions` | 최근 세션 목록 |
| `/session <id>` | 이전 세션 복원 |
| `/auto` | 자율 모드 토글 |
| `/diagnostics [파일]` | 코드 진단 실행 |
| `/git` | Git 상태 상세 보기 |
| `/clear` | 화면 정리 (대화 초기화) |
| `/setup` | API 키 재설정 |
| `/help` | 키보드 단축키 도움말 |

---

## 상태바 정보

```
슈퍼택가이  GPT-OSS 120B  ./project  ⎇ main*  [DEBUG]  Tool:ON(13)  1234tok  ctx:12%  3.2s
```

| 영역 | 내용 |
|------|------|
| 모드명 | 현재 모드 (색상 코딩) |
| 모델명 | provider 접두사 제거된 모델명 |
| 경로 | 현재 작업 디렉토리 |
| Git | 브랜치 + dirty 표시 (초록/노랑) |
| Debug | 디버그 빌드 시 `[DEBUG]` |
| Tool | 활성 도구 수 (ON/OFF + 갯수) |
| Tokens | 누적 토큰 수 |
| Context | 컨텍스트 윈도우 사용률 (초록/노랑/빨강) |
| Elapsed | 마지막 응답 소요 시간 |

---

## 아키텍처

```
cmd/tgc/main.go          — 엔트리포인트, CLI 플래그
internal/app/app.go       — Bubble Tea 모델 (UI + 이벤트 루프)
internal/llm/
  client.go               — OpenAI-compatible SSE 클라이언트
  prompt.go               — 모드별 시스템 프롬프트
  models.go               — 모델 레지스트리
  capabilities.go         — 모델 능력 레지스트리
  compaction.go           — 3단계 컨텍스트 압축
  context.go              — 시스템 컨텍스트 수집
internal/tools/
  registry.go             — 도구 정의 + 실행 라우터 (13개)
  file.go                 — file_read, file_write, file_edit
  search.go               — grep_search, glob_search
  shell.go                — shell_exec (위험 명령 차단)
  hashline.go             — hash-anchored 파일 편집
  git.go                  — git_status, git_diff, git_log
  diagnostics.go          — 멀티 언어 린터 통합
internal/agents/
  auto.go                 — 자율 모드 (마커 기반 제어)
internal/ui/
  chat.go                 — 메시지 렌더링, 상태바
  styles.go               — 색상, 스타일 상수
  palette.go              — Ctrl+K 커맨드 팔레트
  menu.go                 — Esc 메뉴 오버레이
  context.go              — 컨텍스트 사용률 계산
  super.go/dev.go/plan.go — 모드별 UI 헬퍼
  tabbar.go               — 탭바 렌더링
internal/config/
  config.go               — YAML 설정 + 환경변수 + 디버그 로그
internal/session/
  store.go                — SQLite 세션 CRUD
internal/gitinfo/
  gitinfo.go              — Git 브랜치/상태 캐시
internal/knowledge/
  store.go                — 임베디드 지식 스토어
  injector.go             — 쿼리 기반 지식 주입
  extractor.go            — 키워드 추출
knowledge/                — 38개 임베디드 문서
knowledge.go              — embed.FS 선언
```

---

## 미구현 (향후 계획)

| 기능 | 출처 | 우선순위 | 예상 규모 |
|------|------|----------|----------|
| 멀티에이전트 오케스트레이션 | 신규 설계 | Phase 4+ | ~500줄 |
| i18n 다국어 | hanimo Tier 2 | 낮음 | ~200줄 |
| 환경 자동 감지 (env probe) | hanimo Tier 2 | 중간 | ~150줄 |
| 프로젝트 메모리 | hanimo Tier 2 | 중간 | ~200줄 |
| 사용량 통계 | hanimo Tier 2 | 낮음 | ~150줄 |
| 의도 분류 | hanimo Tier 2 | 중간 | ~100줄 |
| 확인 대화 (AskUser) | hanimo Tier 2 | 중간 | ~80줄 |
| 테마 시스템 | hanimo Tier 3 | 낮음 | ~200줄 |
| 페르소나 | hanimo Tier 3 | 낮음 | ~150줄 |
| 프롬프트 캐시 | hanimo Tier 3 | 중간 | ~200줄 |
| 스킬 로더 | hanimo Tier 3 | 낮음 | ~300줄 |
