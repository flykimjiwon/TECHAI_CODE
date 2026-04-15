# Go 핵심 학습자료 — Next.js 풀스택 개발자용

> 택가이코드(tgc) + hanimo 실제 코드 기반.
> "이미 아는 것"에 비유해서 Go를 빠르게 익히자.

---

## -1. JS/TS 생태계 vs Go 생태계 — 전체 그림

JS/TS 개발자의 전형적인 여정과 Go 세계를 1:1로 대응시킨 표.

### 언어 & 런타임

```
JS/TS 세계                         Go 세계
───────────────────────────────────────────────────────
JavaScript (브라우저 태생)      →   Go (서버/시스템 태생)
TypeScript (JS + 타입)          →   Go는 처음부터 정적 타입
Node.js (JS를 서버에서 실행)    →   Go는 자체 런타임 내장 (별도 설치 불필요)
V8 엔진 (JIT 컴파일)            →   Go 컴파일러 (AOT → 네이티브 바이너리)
```

### 웹 프레임워크 흐름

```
JS/TS 여정                         Go 대응
───────────────────────────────────────────────────────
HTML/CSS/JS (순수 프론트)       →   Go는 프론트엔드 없음 (서버/CLI 전문)
React (컴포넌트 UI)             →   Bubble Tea (터미널 TUI 프레임워크) ← 택가이코드가 이것
Express.js (백엔드 첫 입문)     →   net/http (Go 표준 라이브러리, 프레임워크 없이 가능!)
                                    또는 Gin, Echo, Fiber (Express급 프레임워크)
Next.js (풀스택 = React+API)    →   Go는 프론트/백 분리가 기본
                                    백엔드: Go (API 서버)
                                    프론트: 그냥 Next.js 쓰면 됨
```

**핵심 차이**: JS는 "프론트 → 백엔드로 확장"하는 여정이지만,
Go는 처음부터 **서버/CLI/인프라** 전문. 프론트엔드는 안 함.

### 개발 도구

```
JS/TS 도구                         Go 대응
───────────────────────────────────────────────────────
VS Code                        →   VS Code (동일! + Go 확장 설치)
                                    또는 GoLand (JetBrains, 유료지만 최강)
ESLint (린트)                   →   go vet (내장), golangci-lint (외부)
Prettier (포맷팅)               →   gofmt / goimports (내장! 설정 불필요)
                                    → 모든 Go 코드가 동일 포맷 (논쟁 제로)
Jest / Vitest (테스트)          →   go test (내장! 외부 설치 불필요)
ts-node (실행)                  →   go run main.go (빌드+실행 원스텝)
npm run build                   →   go build (결과물: 단일 바이너리)
npm run dev (핫리로드)          →   air (외부 도구, 파일 감시+재빌드)
Webpack / Vite (번들러)         →   불필요! 컴파일러가 알아서 함
babel (트랜스파일)              →   불필요! 하위호환성 문제 없음
```

### 패키지 관리

```
JS/TS                              Go
───────────────────────────────────────────────────────
npm / yarn / pnpm               →   go mod (내장! 별도 설치 불필요)
package.json                    →   go.mod
package-lock.json               →   go.sum
node_modules/ (수백MB)          →   $GOPATH/pkg/mod/ (전역 캐시, 프로젝트별 복사 없음)
npm install                     →   go mod tidy (또는 go get)
npx                             →   go run github.com/xxx/yyy@latest
npm publish                     →   git tag + git push (그냥 Git이 패키지 저장소)
npmjs.com                       →   pkg.go.dev (문서) + GitHub (코드)
```

### 배포 & 실행 환경

```
JS/TS                              Go
───────────────────────────────────────────────────────
Node.js 설치 필요               →   아무것도 불필요 (바이너리만 복사)
Docker (Node 이미지 ~1GB)       →   Docker (scratch 이미지 ~20MB) 또는 바이너리만
Vercel / Netlify (서버리스)     →   AWS Lambda, Cloud Run (Go 네이티브 지원)
PM2 (프로세스 관리)             →   systemd, 또는 그냥 실행 (Go는 안 죽음)
.env / dotenv                   →   환경변수 직접 + YAML 설정 (택가이코드 방식)
```

### VS Code에서 Go 개발 시작하기

```bash
# 1. Go 설치
brew install go                   # Mac
# 또는 https://go.dev/dl/ 에서 다운로드

# 2. VS Code 확장 설치
# VS Code → Extensions → "Go" 검색 → Go Team at Google 설치

# 3. Go 도구 설치 (VS Code에서 Cmd+Shift+P → "Go: Install/Update Tools")
# → gopls (LSP 서버), dlv (디버거), goimports 등 자동 설치

# 4. 프로젝트 시작
mkdir myproject && cd myproject
go mod init myproject             # npm init 같은 것
code .                            # VS Code 열기
```

**VS Code Go 확장이 해주는 것:**
- 자동완성 (TypeScript 수준)
- 정의로 이동 (F12)
- 에러 실시간 표시 (빨간 밑줄)
- 저장 시 자동 포맷 + import 정리
- 디버거 (F5로 브레이크포인트 디버깅)
- 테스트 실행 (함수 위에 "run test" 버튼)

### 왜 Go를 쓰나? (JS/TS 개발자 관점)

```
JS/TS의 불편함                     Go에서 해결되는 것
───────────────────────────────────────────────────────
node_modules 지옥               →   의존성 최소, 표준 라이브러리 풍부
"works on my machine"           →   단일 바이너리, 어디서든 실행
콜백/Promise 지옥               →   goroutine으로 동시성 쉽게
런타임 타입 에러                →   컴파일 타임에 다 잡힘
빌드 도구 설정 지옥             →   go build 하나면 끝
서버 메모리 1GB+                →   Go 서버 50MB로 동일 트래픽 처리
cold start 느림 (서버리스)      →   Go cold start ~100ms
```

**Go가 약한 것:**
- 프론트엔드 (React/Next.js가 압도적)
- 빠른 프로토타이핑 (JS가 더 빠름)
- 제네릭 (1.18에서 추가됐지만 아직 TS만큼 유연하지 않음)
- 에러 핸들링 반복 (`if err != nil` 지옥)

---

## 0. 핵심 마인드셋 전환

```
Next.js (TypeScript)          →    Go
─────────────────────────────────────────
npm / package.json            →    go.mod / go.sum
node_modules/                 →    $GOPATH/pkg/mod/ (자동 관리)
import { x } from 'y'        →    import "y"
interface { }                 →    interface { } (구현 선언 없이 자동 충족)
async/await                   →    goroutine + channel
try/catch                     →    if err != nil { }
any                           →    interface{} (또는 any)
컴파일 없음 (인터프리터)       →    컴파일 필수 → 단일 바이너리
런타임 에러 많음              →    컴파일 타임에 대부분 잡힘
```

---

## 1. 프로젝트 구조 — Next.js App Router vs Go

### Next.js (익숙한 것)
```
app/
  layout.tsx          ← 전역 레이아웃
  page.tsx            ← 메인 페이지
  api/analyze/route.ts ← API Route
components/
  Header.tsx
lib/
  utils.ts
```

### Go (택가이코드 구조)
```
cmd/tgc/main.go         ← 진입점 (Next.js의 layout.tsx + 서버 시작)
internal/
  app/app.go            ← 앱 상태머신 (React의 useReducer 거대 버전)
  ui/                   ← 컴포넌트들 (React 컴포넌트 = Go 렌더 함수)
    chat.go             ← 메시지 렌더링
    styles.go           ← tailwind 대신 lipgloss 스타일
    super.go            ← 로고 + 모드 박스
  llm/                  ← API 클라이언트 (lib/api.ts 같은 것)
    client.go           ← fetch() 대신 openai.Client
    models.go           ← 모델 정의
    prompt.go           ← 시스템 프롬프트
  config/config.go      ← .env.local 대신 YAML + 환경변수
  tools/                ← API Route 핸들러 같은 것
    registry.go         ← 도구 등록
    file.go, shell.go   ← 각 도구 실행
```

**핵심**: `internal/` 폴더 = 외부 패키지에서 import 불가 (Go 컨벤션)

---

## 2. 변수 & 타입 — TypeScript와 비교

### TypeScript
```typescript
const name: string = "techai"
let count: number = 0
const user: User = { name: "kim", age: 30 }
type Role = "user" | "assistant" | "system"
```

### Go
```go
// 짧은 선언 (:=) — 함수 안에서만 사용 가능
name := "techai"          // 타입 자동 추론 (const와 비슷)
count := 0                // int
var count int             // 초기값 없이 선언 (0으로 초기화됨)

// 구조체 = TypeScript interface/type
type User struct {
    Name string   // 대문자 = export (public)
    age  int      // 소문자 = unexported (private)
}
user := User{Name: "kim", age: 30}

// enum = iota 패턴 (택가이코드 ui/chat.go)
type Role int
const (
    RoleUser      Role = iota  // 0
    RoleAssistant              // 1 (자동 증가)
    RoleSystem                 // 2
    RoleTool                   // 3
)
```

**대문자/소문자 규칙이 핵심!**
```go
func PublicFunc() {}    // 외부 패키지에서 접근 가능 (export)
func privateFunc() {}   // 같은 패키지에서만 접근 가능
type Config struct {
    API    APIConfig    // 외부 접근 가능
    debug  bool         // 내부 전용
}
```

---

## 3. 함수 — 가장 큰 차이점

### TypeScript
```typescript
function divide(a: number, b: number): number {
  if (b === 0) throw new Error("division by zero")
  return a / b
}
// 호출
try { const result = divide(10, 0) } catch(e) { console.error(e) }
```

### Go — 에러를 리턴값으로 처리 (try/catch 없음!)
```go
func divide(a, b int) (int, error) {
    if b == 0 {
        return 0, fmt.Errorf("division by zero")
    }
    return a / b, nil
}
// 호출 — 매번 에러 체크 (Go의 철학)
result, err := divide(10, 0)
if err != nil {
    log.Fatal(err)  // 또는 return err
}
```

### 택가이코드 실제 예시 (config.go)
```go
func Load() (Config, error) {
    cfg := DefaultConfig()
    data, err := os.ReadFile(ConfigPath())
    if err == nil {                              // 에러가 nil이면 성공
        if err := yaml.Unmarshal(data, &cfg); err != nil {
            return cfg, fmt.Errorf("config parse error: %w", err)  // %w = 에러 래핑
        }
    }
    return cfg, nil
}
```

---

## 4. 구조체 메서드 — class 대신

### TypeScript (class)
```typescript
class Client {
  private baseURL: string
  constructor(url: string) { this.baseURL = url }
  async chat(msg: string): Promise<Response> { ... }
}
const client = new Client("https://api.novita.ai")
```

### Go (struct + 메서드)
```go
type Client struct {
    baseURL string
    apiKey  string
}

// 생성자 함수 (New 접두사 관례)
func NewClient(baseURL, apiKey string) *Client {
    return &Client{baseURL: baseURL, apiKey: apiKey}
}

// 메서드 = 함수 앞에 (c *Client) 리시버 붙임
func (c *Client) StreamChat(ctx context.Context, model string) <-chan StreamChunk {
    // c.baseURL, c.apiKey 접근 가능
}

// 사용
client := llm.NewClient("https://api.novita.ai", "sk-xxx")
ch := client.StreamChat(ctx, "gemma-4-31b-it")
```

**포인터 `*` 리시버**: 원본 수정 가능 (레퍼런스)
**값 리시버**: 복사본에서 동작 (이뮤터블)

```go
func (m Model) View() string     { ... }  // 값 리시버: m 읽기만
func (m *Model) updateViewport() { ... }  // 포인터: m 수정 가능
```

---

## 5. 인터페이스 — TypeScript와 근본적으로 다름

### TypeScript — 명시적 구현
```typescript
interface Animal { speak(): string }
class Dog implements Animal {  // ← "implements" 필수
    speak() { return "멍멍" }
}
```

### Go — 암시적 충족 (Duck Typing)
```go
type Animal interface {
    Speak() string
}

type Dog struct{}

func (d Dog) Speak() string { return "멍멍" }
// Dog은 자동으로 Animal 인터페이스를 충족함!
// "implements" 키워드 없음. 메서드만 맞으면 됨.

var a Animal = Dog{}  // 자동으로 됨
```

### 실전: Bubble Tea 프레임워크 (택가이코드의 핵심)
```go
// Bubble Tea가 요구하는 인터페이스
type Model interface {
    Init() Cmd
    Update(Msg) (Model, Cmd)
    View() View
}

// app.go의 Model struct가 이 3개 메서드만 구현하면 자동 충족
type Model struct { ... }
func (m Model) Init() tea.Cmd { return textarea.Blink }
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { ... }
func (m Model) View() tea.View { ... }
// → tea.Model 인터페이스 자동 충족!
```

---

## 6. 고루틴 & 채널 — async/await 대신

### TypeScript (async/await)
```typescript
async function fetchData() {
  const res = await fetch("/api/chat")
  const data = await res.json()
  return data
}
```

### Go (goroutine + channel)
```go
// goroutine = 가벼운 스레드 (go 키워드 하나로 실행)
// channel = goroutine 간 데이터 통신 파이프

// 택가이코드 실제 패턴: LLM 스트리밍 (llm/client.go)
func (c *Client) StreamChat(ctx context.Context, ...) <-chan StreamChunk {
    ch := make(chan StreamChunk)       // 채널 생성 (파이프)
    go func() {                        // 별도 goroutine에서 실행
        defer close(ch)                // 함수 끝날 때 채널 닫기
        for {
            resp, err := stream.Recv() // SSE 수신 대기
            if errors.Is(err, io.EOF) {
                ch <- StreamChunk{Done: true}  // 채널에 전송
                return
            }
            ch <- StreamChunk{Content: delta}  // 실시간 전송
        }
    }()
    return ch  // 채널 리턴 (수신 전용 <-chan)
}

// 사용 측 (app.go)
chunk := <-m.streamCh  // 채널에서 수신 (← 이게 await 같은 것)
```

**비유 정리:**
```
Promise              →  channel
await promise        →  <-channel (채널에서 값 꺼내기)
async function       →  go func() { }
Promise.all()        →  sync.WaitGroup 또는 여러 goroutine
```

---

## 7. Context — 타임아웃 & 취소

Next.js에는 없는 개념. **Go에서 가장 중요한 패턴 중 하나**.

```go
// 30초 타임아웃으로 쉘 명령 실행 (tools/shell.go)
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()  // 함수 끝날 때 반드시 취소 (리소스 정리)

result, err := exec.CommandContext(ctx, "sh", "-c", command).Output()

// 스트리밍 취소 (app.go) — 사용자가 Ctrl+C 누를 때
ctx, cancel := context.WithCancel(context.Background())
m.streamCancel = cancel     // 저장해두고
// ... 나중에 ...
m.streamCancel()            // 호출하면 goroutine 즉시 중단
```

**비유**: AbortController (fetch 취소) 같은 것, 근데 모든 곳에서 씀

---

## 8. defer — 정리 코드 보장

TypeScript에 없는 Go 고유 패턴. **finally 블록 자동화**.

```go
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()  // ← 이 함수가 어떻게 끝나든 Close() 실행됨

    // 여기서 panic이 나도 f.Close()는 실행됨!
    data, err := io.ReadAll(f)
    // ...
}

// 택가이코드: 뮤텍스 잠금/해제
func DebugLog(format string, args ...interface{}) {
    debugMu.Lock()
    defer debugMu.Unlock()  // Lock → 작업 → 자동 Unlock
    fmt.Fprintf(debugFile, ...)
}
```

---

## 9. 패키지 & import — npm vs Go modules

### npm (package.json)
```bash
npm install openai yaml
# node_modules/에 설치됨
```

### Go (go.mod)
```bash
go mod init github.com/kimjiwon/tgc    # 프로젝트 초기화
go get github.com/sashabaranov/go-openai # 패키지 추가
# go.mod, go.sum 자동 업데이트
```

```go
// go.mod (package.json 같은 것)
module github.com/kimjiwon/tgc

go 1.23

require (
    charm.land/bubbletea/v2 v2.0.0
    github.com/sashabaranov/go-openai v1.38.1
    gopkg.in/yaml.v3 v3.0.1
)
```

**import 패턴:**
```go
import (
    "fmt"              // 표준 라이브러리 (내장)
    "os"               // 표준 라이브러리
    "strings"          // 표준 라이브러리

    "gopkg.in/yaml.v3" // 외부 패키지 (npm 패키지 같은 것)

    "github.com/kimjiwon/tgc/internal/config"  // 내부 패키지
)
```

---

## 10. 빌드 & 배포 — 가장 큰 장점

### Next.js
```bash
npm run build          # .next/ 폴더 생성
npm start              # Node.js 런타임 필요
# 또는 Docker 이미지 (수백MB)
```

### Go — 단일 바이너리!
```bash
go build -o techai ./cmd/tgc    # 16MB 단일 실행파일
./techai                         # 끝. 런타임 불필요!

# 크로스 컴파일 (GOOS/GOARCH만 바꾸면 됨)
GOOS=windows GOARCH=amd64 go build -o techai.exe ./cmd/tgc
GOOS=linux   GOARCH=arm64 go build -o techai-linux ./cmd/tgc
GOOS=darwin  GOARCH=arm64 go build -o techai-mac ./cmd/tgc
```

### 택가이코드 Makefile 빌드 패턴
```makefile
# ldflags로 빌드 시 변수 주입 (env 없이 바이너리에 내장)
LDFLAGS = -ldflags "-s -w \
    -X main.version=$(VERSION) \
    -X 'github.com/kimjiwon/tgc/internal/config.DefaultModel=google/gemma-4-31b-it'"

build:
    go build $(LDFLAGS) -o techai ./cmd/tgc
```

**`-ldflags`**: 빌드 타임에 Go 변수값을 주입. `.env` 파일 없이 바이너리에 설정을 구움.

---

## 11. Embed — 파일을 바이너리에 내장

Next.js의 `public/` 폴더와 비슷하지만, **파일이 바이너리 안에 들어감**.

```go
// knowledge.go (프로젝트 루트)
//go:embed knowledge/**/*.md knowledge/index.json
var KnowledgeFS embed.FS

// 사용: 파일시스템처럼 접근
data, err := KnowledgeFS.ReadFile("knowledge/index.json")
```

이게 있어서 `techai` 바이너리 하나만 배포하면 docs까지 다 포함됨.

---

## 12. 에러 핸들링 패턴 총정리

Go에서 가장 많이 쓰는 패턴들:

```go
// 1. 기본: 에러 체크
result, err := doSomething()
if err != nil {
    return fmt.Errorf("doSomething failed: %w", err)  // %w로 래핑
}

// 2. 에러 무시 (의도적)
_ = os.MkdirAll(dir, 0755)  // 실패해도 괜찮은 경우

// 3. 에러 타입 체크
if errors.Is(err, io.EOF) {     // 특정 에러인지 확인
    // 정상 종료
}

// 4. 여러 리턴값
func Load() (Config, error) {   // (결과, 에러) 쌍으로 리턴
    // ...
}
```

---

## 13. 동시성 안전 — sync.Mutex

```go
// TypeScript는 싱글스레드라 필요 없지만, Go는 멀티스레드!
var (
    debugFile *os.File
    debugMu   sync.Mutex  // 뮤텍스 = 잠금장치
)

func DebugLog(msg string) {
    debugMu.Lock()          // 잠금 (다른 goroutine 대기)
    defer debugMu.Unlock()  // 함수 끝날 때 자동 해제
    fmt.Fprintln(debugFile, msg)
}
```

---

## 14. 택가이코드 핵심 아키텍처: Bubble Tea

Bubble Tea = **Go판 React**. 택가이코드/hanimo의 UI 프레임워크.

```
React                          →   Bubble Tea
─────────────────────────────────────────────
useState / useReducer          →   Model struct
dispatch(action)               →   Update(msg) → (Model, Cmd)
render() / return JSX          →   View() → string
useEffect                      →   Cmd (사이드 이펙트)
props                          →   Msg (메시지)
```

### 실제 흐름 (app.go)
```go
// 1. 상태 (React의 state)
type Model struct {
    msgs      []ui.Message
    streaming bool
    activeTab int
}

// 2. 업데이트 (React의 reducer)
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        switch msg.String() {
        case "enter":
            // 메시지 전송
        case "tab":
            m.activeTab = (m.activeTab + 1) % 3
        case "ctrl+c":
            return m, tea.Quit
        }
    case streamChunkMsg:         // 커스텀 메시지 (LLM 스트림)
        m.streamBuf += msg.content
    case tea.WindowSizeMsg:      // 터미널 크기 변경
        m.viewport.SetSize(msg.Width, msg.Height)
    }
    return m, nil
}

// 3. 렌더링 (React의 render)
func (m Model) View() tea.View {
    content := ui.RenderMessages(m.msgs, m.streamBuf, width)
    // 문자열로 UI 구성 → 터미널에 출력
}
```

---

## 15. 자주 쓰는 표준 라이브러리

| 패키지 | 용도 | Next.js 대응 |
|--------|------|-------------|
| `fmt` | 포맷 출력 | `console.log`, 템플릿 리터럴 |
| `os` | 파일/환경변수 | `fs`, `process.env` |
| `strings` | 문자열 조작 | `String.prototype` 메서드들 |
| `path/filepath` | 경로 처리 | `path` 모듈 |
| `encoding/json` | JSON 파싱 | `JSON.parse/stringify` |
| `net/http` | HTTP 서버/클라 | `fetch`, Express |
| `context` | 취소/타임아웃 | `AbortController` |
| `sync` | 동시성 | 없음 (싱글스레드) |
| `embed` | 파일 내장 | `public/` 폴더 |
| `time` | 시간 처리 | `Date`, `setTimeout` |
| `errors` | 에러 유틸 | `Error` 클래스 |
| `regexp` | 정규식 | `RegExp` |

---

## 16. 빠른 레퍼런스 — 매일 쓰는 명령어

```bash
# 프로젝트
go mod init myproject     # npm init
go mod tidy               # 미사용 의존성 정리
go get package@v1.2.3     # npm install package

# 빌드 & 실행
go build ./cmd/tgc        # npm run build
go run ./cmd/tgc          # npm run dev (빌드+실행)
go install ./cmd/tgc      # npm install -g (글로벌 설치)

# 테스트
go test ./...             # npm test (전체)
go test ./internal/llm/   # 특정 패키지만
go test -v -run TestName  # 특정 테스트만

# 코드 품질
go vet ./...              # 정적 분석 (eslint 비슷)
go fmt ./...              # 자동 포맷 (prettier 비슷)
```

---

## 17. 흔한 실수 & 팁

### 1. 세미콜론 없음, 중괄호 위치 고정
```go
// 컴파일 에러!
if true
{
}

// 올바름
if true {
}
```

### 2. 미사용 변수/import = 컴파일 에러
```go
import "fmt"  // fmt를 안 쓰면 컴파일 에러!
x := 5        // x를 안 쓰면 컴파일 에러!
_ = x         // 의도적 무시는 _ 사용
```

### 3. nil 체크 습관
```go
// TypeScript: optional chaining
user?.name

// Go: 명시적 nil 체크
if user != nil {
    fmt.Println(user.Name)
}
```

### 4. 슬라이스 (배열) 기본
```go
// TypeScript: const arr = ["a", "b", "c"]
arr := []string{"a", "b", "c"}          // 슬라이스 생성
arr = append(arr, "d")                   // push
arr[0]                                   // 인덱스 접근
len(arr)                                 // .length
arr[1:3]                                 // .slice(1, 3)

// for 루프
for i, item := range arr {              // forEach with index
    fmt.Println(i, item)
}
for _, item := range arr {              // index 무시
    fmt.Println(item)
}
```

### 5. map (객체)
```go
// TypeScript: const obj: Record<string, number> = {}
m := map[string]int{}                   // 빈 맵 생성
m := make(map[string]int)               // 동일
m["key"] = 42                           // 값 설정
val, ok := m["key"]                     // 값 읽기 (ok = 존재 여부)
delete(m, "key")                        // 삭제
```

---

## 요약: Next.js → Go 전환 체크리스트

- [ ] **에러 = 리턴값** (try/catch 없음, `if err != nil` 반복)
- [ ] **대문자 = public** (export 키워드 없음)
- [ ] **goroutine + channel** = async/await
- [ ] **context** = 취소/타임아웃 전파
- [ ] **defer** = finally 자동화
- [ ] **interface** = 암시적 충족 (implements 없음)
- [ ] **단일 바이너리** 배포 (런타임 불필요)
- [ ] **`go build`로 크로스 컴파일** (GOOS/GOARCH)
- [ ] **미사용 = 컴파일 에러** (깔끔한 코드 강제)
