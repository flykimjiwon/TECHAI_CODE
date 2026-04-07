# 택가이코드 (TECHAI CODE)

OpenAI-compatible API 기반 CLI AI 코딩 어시스턴트. Go + Bubble Tea v2 단일 바이너리.

## 특징

- **3개 모드**: 슈퍼택가이(만능) / 플랜(계획) / 개발(코딩) — Tab 키로 전환
- **모델**: gpt-oss-120b (전 모드 통일)
- **배치 렌더링**: 스피너 + 마크다운 일괄 출력 (Claude Code 스타일)
- **도구**: 파일 읽기/쓰기/수정/삭제, 셸 명령 실행 (개발/슈퍼택가이 모드)
- **큐 시스템**: AI 응답 중에도 미리 타이핑하고 대기열에 저장
- **단일 바이너리**: Node.js, Python 등 외부 의존성 없음 (~17MB)

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

```powershell
# 1. 폴더 생성 + exe 복사
mkdir C:\techai
copy dist\techai-windows-amd64.exe C:\techai\techai.exe

# 2. PATH 환경변수 추가 (PowerShell 관리자 권한)
[System.Environment]::SetEnvironmentVariable("Path",
  $env:Path + ";C:\techai",
  [System.EnvironmentVariableTarget]::User)

# 3. 터미널 재시작 후 실행
techai
```

**또는 GUI로 PATH 등록:**
1. `Win + R` → `sysdm.cpl` → 고급 → 환경 변수
2. 사용자 변수 → `Path` → 편집 → 새로 만들기 → `C:\techai` → 확인
3. 터미널 재시작 → `techai`

> Windows Terminal (Microsoft Store 무료) 사용 권장 — 색상, 마크다운, 마우스 스크롤 지원이 우수합니다.

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
  api_key: "sk-..."
models:
  super: "gpt-oss-120b"
  plan: "gpt-oss-120b"
  dev: "gpt-oss-120b"
```

환경변수 오버라이드:

```bash
export TGC_API_BASE_URL=https://your-api.com/v1
export TGC_API_KEY=sk-...
```

## 사용법

```bash
techai                # 기본 (슈퍼택가이 모드)
techai --mode dev     # 개발 모드로 시작
techai --mode plan    # 플랜 모드로 시작
techai --version      # 버전 출력
```

## 키 바인딩

| 키 | 동작 |
|---|------|
| `Enter` | 메시지 전송 |
| `Shift+Enter` | 줄바꿈 |
| `Tab` | 모드 전환 (슈퍼택가이 → 개발 → 플랜) |
| `Esc` | 스트리밍 중단 |
| `Ctrl+L` | 화면 클리어 |
| `Ctrl+C` | 종료 |
| `Alt+↑/↓` | 3줄 스크롤 |
| `PgUp/PgDown` | 페이지 스크롤 |
| `/clear` | 대화 삭제 |
| `/setup` | 설정 재실행 |
| `/help` | 도움말 |

## 모드별 차이

| 모드 | 설명 | 도구 |
|------|------|------|
| **슈퍼택가이** | 만능 모드. 코드 CRUD, 분석, 대화 자동 감지 | file_read, file_write, file_edit, list_files, shell_exec |
| **개발** | 코딩 특화. 파일 생성/읽기/수정/삭제 | file_read, file_write, file_edit, list_files, shell_exec |
| **플랜** | 분석/계획. 읽기 전용, 구조 파악, 리뷰 | file_read, list_files, shell_exec (읽기 전용) |

## 빌드

```bash
make build          # 현재 플랫폼 → ./techai
make build-all      # 크로스 컴파일 (macOS/Windows/Linux × amd64/arm64)
make install        # go install
make test           # 테스트
make lint           # go vet
make run            # 빌드 + 실행
make clean          # 정리
```

### 빌드 결과물

```
dist/
├── techai-darwin-arm64       # macOS Apple Silicon
├── techai-darwin-amd64       # macOS Intel
├── techai-windows-amd64.exe  # Windows
├── techai-linux-amd64        # Linux x64
└── techai-linux-arm64        # Linux ARM
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
택갈이코드/
├── cmd/tgc/main.go              # 엔트리포인트
├── internal/
│   ├── app/app.go               # 메인 TUI 앱 (Model/Update/View)
│   ├── ui/
│   │   ├── styles.go            # 색상/스타일 정의
│   │   ├── chat.go              # 메시지 렌더링, 마크다운, 상태바
│   │   ├── super.go             # 로고, 모드 정보 박스
│   │   └── tabbar.go            # 탭 바
│   ├── llm/
│   │   ├── client.go            # OpenAI-compatible 스트리밍 클라이언트
│   │   ├── models.go            # 모델 정의
│   │   └── prompt.go            # 모드별 시스템 프롬프트
│   ├── config/config.go         # 설정 로드 (YAML + env)
│   └── tools/
│       ├── registry.go          # 도구 등록/실행
│       ├── file.go              # 파일 도구
│       └── shell.go             # 셸 명령 도구
├── docs/MULTI_AGENT_V2.md       # v2 멀티에이전트 아키텍처 계획
├── Makefile                     # 빌드 스크립트
└── go.mod
```

## 라이선스

MIT
