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
  super: "gpt-oss-120b"
  plan: "gpt-oss-120b"
  dev: "gpt-oss-120b"
```

환경변수 오버라이드:

```bash
export TGC_API_BASE_URL=https://your-api.com/v1
export TGC_API_KEY=tg-...
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

| 모드 | 설명 | 특징 |
|------|------|------|
| **Super** (슈퍼택가이) | 만능 모드. 코드, 분석, 대화 자동 감지 | 14개 전체 도구 사용 |
| **Deep Agent** | 자율 코딩. 최대 100회 반복, 자동 검증 | 자율 실행, `[TASK_COMPLETE]` 마커 |
| **Plan** | 계획 우선. 단계별 계획 → 승인 후 실행 | 전체 도구 (쓰기 포함) |

## 사용자 지식 문서 (User Knowledge Docs)

`.tgc/knowledge/` 폴더에 `.md` 또는 `.txt` 파일을 넣으면, AI가 자동으로 인덱싱하고 질문에 관련된 문서를 검색해서 참고합니다.

사내 개발 가이드, API 문서, 코딩 규칙, 온보딩 자료 등을 넣어두면 AI가 프로젝트 맥락을 이해하고 더 정확한 답변을 줍니다.

### 공통

**지원 파일 형식**: `.md` (마크다운), `.txt` (텍스트)

**폴더 우선순위**: 프로젝트 로컬 > 글로벌 (둘 다 있으면 로컬 우선)

**동작 방식**:
1. `techai` 시작 시 knowledge 폴더를 자동 스캔 (~1ms)
2. 파일별 제목, 헤더, 첫 단락을 인덱싱
3. 시스템 프롬프트에 문서 목차 자동 주입
4. AI가 질문과 관련된 문서를 `knowledge_search` 도구로 자동 검색

**예시 파일 구조**:
```
.tgc/knowledge/
├── deploy-guide.md        # 배포 가이드
├── coding-rules.txt       # 코딩 규칙
├── api-reference.md       # API 레퍼런스
└── onboarding.md          # 신규 입사자 온보딩
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
