# 택가이코드 디버그 모드 가이드

> 사내망(온프레미스) 환경에서 발생하는 문제를 진단하기 위한 디버그 빌드 사용법

---

## 목차

1. [사내망 문제 증상](#1-사내망-문제-증상)
2. [디버그 빌드 설치](#2-디버그-빌드-설치)
3. [디버그 빌드 실행](#3-디버그-빌드-실행)
4. [debug.log 분석 가이드](#4-debuglog-분석-가이드)
5. [vLLM Tool Calling 확인](#5-vllm-tool-calling-확인)
6. [문제별 진단 플로우차트](#6-문제별-진단-플로우차트)
7. [인프라팀 요청 템플릿](#7-인프라팀-요청-템플릿)

---

## 1. 사내망 문제 증상

| 증상 | 추정 원인 |
|------|----------|
| thinking 스피너 안 돌아감 | 프록시가 SSE(chunked transfer) 버퍼링 |
| `shell_exec({...` 텍스트 그대로 출력 | vLLM에 tool calling 미활성화 |
| 쓰기 동작 불가 (파일 생성/수정 안 됨) | tool call JSON 파싱 실패 또는 셸 권한 차단 |
| 연결 자체가 안 됨 | DNS 실패, 프록시 차단, TLS 인증서 문제 |
| 응답이 중간에 끊김 | 프록시 타임아웃, 네트워크 불안정 |

---

## 2. 디버그 빌드 설치

### Windows (사내망)

```powershell
# 1. 디버그 exe를 기존 techai 위치에 복사 (또는 별도 폴더)
Copy-Item techai-debug-onprem-windows-amd64.exe C:\techai\techai-debug.exe

# 2. 실행
C:\techai\techai-debug.exe
```

### macOS

```bash
# 로컬 디버그 빌드 실행
./dist/techai-debug
```

---

## 3. 디버그 빌드 실행

### 시작하면 보이는 것

1. **TUI 화면에 `[DEBUG MODE]` 표시** — 상태바 오른쪽에 빨간색으로 표시
2. **로그 파일 경로 안내** — TUI 상단에 `[DEBUG MODE] 로그 파일: C:\Users\사용자\.tgc-onprem\debug.log`
3. **종료 시 로그 경로 재안내** — 터미널에 로그 파일 위치 출력

### 로그 파일 위치

| 환경 | 경로 |
|------|------|
| 사내망 (온프레미스) | `C:\Users\사용자\.tgc-onprem\debug.log` |
| 일반 (인터넷) | `~/.tgc/debug.log` |

### 로그 파일 열기

```powershell
# Windows — 메모장으로 열기
notepad %USERPROFILE%\.tgc-onprem\debug.log

# Windows — PowerShell에서 실시간 모니터링
Get-Content -Path "$env:USERPROFILE\.tgc-onprem\debug.log" -Wait

# Windows — 마지막 50줄 보기
Get-Content -Path "$env:USERPROFILE\.tgc-onprem\debug.log" -Tail 50
```

```bash
# macOS/Linux
tail -f ~/.tgc/debug.log        # 실시간 모니터링
tail -50 ~/.tgc/debug.log       # 마지막 50줄
```

---

## 4. debug.log 분석 가이드

### 로그 태그 일람표

| 태그 | 구간 | 확인 내용 |
|------|------|----------|
| `[SYS]` | 시작 | OS, ARCH, Go버전, 호스트명, CWD, PID |
| `[ENV]` | 시작 | TGC_*, HTTP_PROXY, HTTPS_PROXY, NO_PROXY |
| `[NET-DNS]` | 클라이언트 초기화 | API 호스트 DNS 해석 성공/실패 + 소요시간 |
| `[NET-REQ#N]` | HTTP 요청 | 전체 요청 헤더, Content-Length |
| `[NET-RES#N]` | HTTP 응답 | 전체 응답 헤더, Status, Content-Type, Transfer-Encoding |
| `[NET-TLS#N]` | TLS | 인증서 체인 (Issuer, Subject, 만료일, Cipher, TLS버전) |
| `[NET-PROXY#N]` | 프록시 | Via, X-Forwarded-For 헤더 |
| `[NET-ERR#N]` | 네트워크 에러 | timeout 여부, 에러 메시지 |
| `[API-REQ]` | API 요청 | 모델명, 메시지수, 도구목록, 마지막 메시지 미리보기 |
| `[API-REQ-RAW]` | API 요청 원문 | JSON 요청 본문 전체 (5KB까지) |
| `[API-RES]` | API 응답 | 스트림 열린 시간 |
| `[STREAM] TTFC` | 첫 청크 | 첫 번째 청크까지 걸린 시간 (프록시 버퍼링 감지) |
| `[CHUNK#N]` | 청크별 | 내용 길이, 도구호출 여부, 이전 청크 시간 간격, 내용 미리보기 |
| `[CHUNK#N] finishReason` | 청크 | stop, tool_calls, length 등 |
| `[STREAM-DONE]` | 스트림 완료 | 총 청크수, 바이트, 도구호출 수, 총 소요시간, args 전문 |
| `[STREAM-ERR]` | 스트림 에러 | 에러 메시지, 마지막 청크 이후 경과시간 |
| `[TOOL-CALL]` | 도구 실행 | 도구명, args JSON 전체 |
| `[TOOL-RESULT]` | 도구 결과 | 결과 길이, 잘림 여부 |
| `[SHELL]` | 셸 실행 | 명령어, CWD, 타임아웃, OS, exit code, stdout/stderr 크기, 소요시간 |
| `[SHELL-BLOCK]` | 셸 차단 | 차단된 명령어, 매칭된 패턴 |
| `[APP-STREAM]` | TUI 스트림 | 모드, 히스토리 메시지수, 도구수, 루프 카운터 |
| `[APP-TOOL]` | TUI 도구 | 도구 결과 수, 각 결과 크기, 루프 이터레이션 |

### 로그 예시 (정상 작동)

```
[15:30:01.000] ========================================
[15:30:01.001] === TECHAI DEBUG MODE ===
[15:30:01.002] ========================================
[15:30:01.003] [SYS] OS=windows | ARCH=amd64 | GoVersion=go1.22.0
[15:30:01.004] [SYS] Hostname=SHINHAN-PC-001
[15:30:01.005] [SYS] ConfigDir=C:\Users\user\.tgc-onprem
[15:30:01.006] [ENV] HTTP_PROXY=http://proxy.shinhan.com:8080
[15:30:01.007] [ENV] HTTPS_PROXY=http://proxy.shinhan.com:8080
[15:30:01.008] ========================================
[15:30:01.100] [NET] API BaseURL=https://techai-web-prod.shinhan.com/v1
[15:30:01.101] [NET-DNS] resolving techai-web-prod.shinhan.com ...
[15:30:01.250] [NET-DNS] OK techai-web-prod.shinhan.com → [10.0.1.50] (took 149ms)
[15:30:05.000] [API-REQ] POST /chat/completions | model=GPT-OSS-120B | msgs=2 | tools=5
[15:30:05.100] [NET-REQ#1] POST https://techai-web-prod.shinhan.com/v1/chat/completions
[15:30:05.500] [NET-RES#1] Status=200 | elapsed=400ms
[15:30:05.501] [NET-RES#1] Content-Type=text/event-stream    ← 이게 맞아야 함
[15:30:05.600] [API-RES] stream opened in 500ms
[15:30:06.100] [STREAM] first chunk after 1.1s (TTFC)
[15:30:06.101] [CHUNK#1] content len=6 | toolCall=false | gap=500ms | preview="안녕하"
[15:30:06.200] [CHUNK#2] content len=9 | toolCall=false | gap=99ms
[15:30:06.300] [CHUNK#3] toolCall[0] START name="shell_exec" | id="call_abc" | gap=100ms
[15:30:06.500] [STREAM-DONE] chunks=10 | toolCalls=1 | totalTime=1.5s
[15:30:06.501] [STREAM-DONE] toolCall[0] name=shell_exec | args={"command":"ls -la"}
[15:30:06.502] [TOOL-CALL] shell_exec | args={"command":"ls -la"}
[15:30:06.503] [SHELL] cmd="ls -la" | cwd=cmd | timeout=30s | os=windows
[15:30:06.650] [SHELL] exitCode=0 | stdout=1234bytes | stderr=0bytes | elapsed=147ms
[15:30:06.651] [TOOL-RESULT] shell_exec | resultLen=1234 | truncated=false
```

### 로그 예시 (문제 상황)

#### 문제: tool calling 미지원

```
[15:30:06.101] [CHUNK#1] content len=45 | preview="shell_exec({\"command\": \"ls\"})"
[15:30:06.500] [STREAM-DONE] chunks=5 | toolCalls=0    ← 도구호출 0개!
```
→ **원인**: vLLM에 `--enable-auto-tool-choice` 미설정

#### 문제: 프록시 버퍼링

```
[15:30:05.600] [API-RES] stream opened in 500ms
[15:30:15.100] [STREAM] first chunk after 10.1s (TTFC)    ← 10초 대기!
[15:30:15.101] [CHUNK#1] content len=2048 | gap=10s       ← 한번에 큰 청크
```
→ **원인**: 프록시가 SSE를 버퍼링하여 한번에 전달

#### 문제: DNS 실패

```
[15:30:01.101] [NET-DNS] resolving techai-web-prod.shinhan.com ...
[15:30:06.101] [NET-DNS] FAILED: no such host (took 5s)
```
→ **원인**: 사내 DNS에서 호스트를 찾지 못함

#### 문제: TLS 인증서

```
[15:30:05.501] [NET-TLS#1] Cert[0] Subject=proxy.shinhan.com | Issuer=Shinhan-Root-CA
```
→ **원인**: 기업 프록시가 TLS를 중간에서 복호화 (MITM). 자체 CA 인증서 필요할 수 있음

---

## 5. vLLM Tool Calling 확인

### 방법 1: curl로 직접 테스트 (Windows PowerShell)

```powershell
# 기본 응답 테스트 (tool 없이)
curl.exe -X POST "https://techai-web-prod.shinhan.com/v1/chat/completions" `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer YOUR_API_KEY" `
  -d '{\"model\":\"GPT-OSS-120B\",\"messages\":[{\"role\":\"user\",\"content\":\"hello\"}],\"max_tokens\":50}'
```

```powershell
# Tool Calling 테스트
curl.exe -X POST "https://techai-web-prod.shinhan.com/v1/chat/completions" `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer YOUR_API_KEY" `
  -d '{\"model\":\"GPT-OSS-120B\",\"messages\":[{\"role\":\"user\",\"content\":\"현재 디렉토리의 파일 목록을 보여줘\"}],\"tools\":[{\"type\":\"function\",\"function\":{\"name\":\"list_files\",\"description\":\"List files in a directory\",\"parameters\":{\"type\":\"object\",\"properties\":{\"path\":{\"type\":\"string\",\"description\":\"Directory path\"}},\"required\":[\"path\"]}}}]}'
```

```powershell
# 스트리밍 테스트 (SSE 확인)
curl.exe -X POST "https://techai-web-prod.shinhan.com/v1/chat/completions" `
  -H "Content-Type: application/json" `
  -H "Accept: text/event-stream" `
  -H "Authorization: Bearer YOUR_API_KEY" `
  -d '{\"model\":\"GPT-OSS-120B\",\"messages\":[{\"role\":\"user\",\"content\":\"hello\"}],\"stream\":true,\"max_tokens\":50}' `
  --no-buffer
```

```powershell
# vLLM 모델 목록 확인
curl.exe "https://techai-web-prod.shinhan.com/v1/models" `
  -H "Authorization: Bearer YOUR_API_KEY"
```

### 방법 2: curl로 직접 테스트 (Windows CMD)

```cmd
REM 기본 응답 테스트
curl -X POST "https://techai-web-prod.shinhan.com/v1/chat/completions" -H "Content-Type: application/json" -H "Authorization: Bearer YOUR_API_KEY" -d "{\"model\":\"GPT-OSS-120B\",\"messages\":[{\"role\":\"user\",\"content\":\"hello\"}],\"max_tokens\":50}"
```

```cmd
REM Tool Calling 테스트
curl -X POST "https://techai-web-prod.shinhan.com/v1/chat/completions" -H "Content-Type: application/json" -H "Authorization: Bearer YOUR_API_KEY" -d "{\"model\":\"GPT-OSS-120B\",\"messages\":[{\"role\":\"user\",\"content\":\"파일 목록을 보여줘\"}],\"tools\":[{\"type\":\"function\",\"function\":{\"name\":\"list_files\",\"description\":\"List files\",\"parameters\":{\"type\":\"object\",\"properties\":{\"path\":{\"type\":\"string\"}},\"required\":[\"path\"]}}}]}"
```

```cmd
REM 스트리밍 테스트
curl -X POST "https://techai-web-prod.shinhan.com/v1/chat/completions" -H "Content-Type: application/json" -H "Accept: text/event-stream" -H "Authorization: Bearer YOUR_API_KEY" -d "{\"model\":\"GPT-OSS-120B\",\"messages\":[{\"role\":\"user\",\"content\":\"hello\"}],\"stream\":true,\"max_tokens\":50}" --no-buffer
```

```cmd
REM 모델 목록 확인 (vLLM 버전 정보 포함)
curl "https://techai-web-prod.shinhan.com/v1/models" -H "Authorization: Bearer YOUR_API_KEY"
```

### 방법 3: JSON 파일 사용 (이스케이프 문제 회피)

`test-tool.json` 파일을 만들고:

```json
{
  "model": "GPT-OSS-120B",
  "messages": [
    {"role": "user", "content": "현재 디렉토리의 파일 목록을 보여줘"}
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "list_files",
        "description": "List files in a directory",
        "parameters": {
          "type": "object",
          "properties": {
            "path": {"type": "string", "description": "Directory path"}
          },
          "required": ["path"]
        }
      }
    },
    {
      "type": "function",
      "function": {
        "name": "shell_exec",
        "description": "Execute a shell command",
        "parameters": {
          "type": "object",
          "properties": {
            "command": {"type": "string", "description": "Shell command"}
          },
          "required": ["command"]
        }
      }
    }
  ]
}
```

```powershell
# PowerShell — JSON 파일로 요청
curl.exe -X POST "https://techai-web-prod.shinhan.com/v1/chat/completions" `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer YOUR_API_KEY" `
  -d "@test-tool.json"
```

```cmd
REM CMD — JSON 파일로 요청
curl -X POST "https://techai-web-prod.shinhan.com/v1/chat/completions" -H "Content-Type: application/json" -H "Authorization: Bearer YOUR_API_KEY" -d @test-tool.json
```

### 응답 판별법

**Tool Calling 지원됨:**
```json
{
  "choices": [{
    "message": {
      "role": "assistant",
      "content": null,
      "tool_calls": [{
        "id": "call_abc123",
        "type": "function",
        "function": {
          "name": "list_files",
          "arguments": "{\"path\": \".\"}"
        }
      }]
    },
    "finish_reason": "tool_calls"
  }]
}
```

**Tool Calling 미지원 (텍스트로 출력):**
```json
{
  "choices": [{
    "message": {
      "role": "assistant",
      "content": "list_files({\"path\": \".\"})"
    },
    "finish_reason": "stop"
  }]
}
```

핵심 차이:
- `finish_reason: "tool_calls"` → 지원됨
- `finish_reason: "stop"` + content에 함수 텍스트 → 미지원

---

## 6. 문제별 진단 플로우차트

### A. "아무 응답이 없음 / 멈춤"

```
debug.log 확인
    ↓
[NET-DNS] 있음?
  ├─ FAILED → DNS 문제. 사내 DNS에 호스트 등록 필요
  └─ OK
    ↓
[NET-REQ#1] 있음?
  ├─ 없음 → 클라이언트 초기화 실패. 설정 확인
  └─ 있음
    ↓
[NET-RES#1] 있음?
  ├─ 없음 → 네트워크 차단. 방화벽/프록시 규칙 확인
  ├─ Status=403/401 → API Key 문제
  ├─ Status=502/503 → vLLM 서버 다운
  └─ Status=200
    ↓
[STREAM] TTFC 있음?
  ├─ 없음 → 스트림 열렸지만 청크 도착 안 함. 프록시 버퍼링 or vLLM 처리 중
  └─ 있음 (TTFC > 10초) → 프록시가 SSE 버퍼링 중
```

### B. "shell_exec({... 텍스트가 그대로 출력"

```
debug.log 확인
    ↓
[STREAM-DONE] toolCalls 확인
  ├─ toolCalls=0 → vLLM에 tool calling 미활성화
  │   → 인프라팀에 --enable-auto-tool-choice 요청
  └─ toolCalls>0 → tool calling 작동 중
    ↓
    그런데도 텍스트로 보인다면?
    → 다른 청크에서 content로 tool call 텍스트가 온 것
    → [CHUNK#N] preview 확인
```

### C. "파일 쓰기가 안 됨"

```
debug.log 확인
    ↓
[TOOL-CALL] file_write 있음?
  ├─ 없음 → tool calling 자체가 안 됨 (B 참조)
  └─ 있음
    ↓
[TOOL-RESULT] file_write 확인
  ├─ Error: permission denied → 파일 쓰기 권한 문제
  └─ OK → 정상 작동 (TUI 표시 문제일 수 있음)
```

### D. "thinking 스피너 안 돌아감"

```
debug.log에서 [CHUNK#N] gap 시간 확인
    ↓
gap이 대부분 > 3초?
  ├─ 예 → 프록시 SSE 버퍼링 (청크가 모여서 한번에 도착)
  │   → 인프라팀에 프록시 SSE passthrough 설정 요청
  └─ 아니오 (gap < 500ms) → 정상 스트리밍. TUI 렌더링 문제
```

---

## 7. 인프라팀 요청 템플릿

### Tool Calling 활성화 요청

```
제목: [techai] vLLM Tool Calling 활성화 요청

안녕하세요,

techai 코딩 도우미에서 파일 읽기/쓰기, 셸 실행 기능을 사용하려면
vLLM 서버에 Tool Calling 기능이 활성화되어야 합니다.

현재 상태:
- API 엔드포인트: https://techai-web-prod.shinhan.com/v1
- 모델: GPT-OSS-120B
- 증상: tool_calls 필드가 응답에 포함되지 않음
  (텍스트로 "shell_exec({...})" 출력)

요청 사항:
vLLM 서버 시작 옵션에 아래 파라미터 추가 부탁드립니다:

  --enable-auto-tool-choice
  --tool-call-parser hermes

참고:
- vLLM 0.5.0+ 필요 (현재 버전 확인 부탁드립니다)
- 모델이 Hermes 형식이 아닌 경우 --tool-call-parser를
  mistral, llama3_json, pythonic 등으로 변경 필요

첨부: debug.log (디버그 빌드 실행 결과)
```

### SSE 프록시 설정 요청

```
제목: [techai] SSE 스트리밍 프록시 설정 요청

안녕하세요,

techai에서 AI 응답이 실시간으로 표시되지 않고 한번에 출력되는
문제가 있습니다. 프록시가 Server-Sent Events(SSE) 응답을
버퍼링하는 것으로 보입니다.

현재 상태:
- 엔드포인트: https://techai-web-prod.shinhan.com/v1/chat/completions
- Content-Type: text/event-stream
- 증상: 첫 응답까지 10초+ 대기 후 한번에 도착 (TTFC > 10s)

요청 사항:
해당 엔드포인트에 대해 프록시 버퍼링 비활성화 부탁드립니다:
- Nginx: proxy_buffering off;
- Apache: SetEnv proxy-sendchunked 1
- HAProxy: option http-no-delay
- 기업 프록시: SSE/chunked transfer 패스스루 설정

첨부: debug.log (디버그 빌드 실행 결과)
```

---

## 부록: vLLM Tool Calling 버전별 지원

| vLLM 버전 | Tool Calling | 비고 |
|-----------|-------------|------|
| < 0.4.0 | 미지원 | 업그레이드 필요 |
| 0.4.x | 실험적 | 일부 모델만, 불안정 |
| 0.5.0 ~ 0.5.5 | 지원 | `--enable-auto-tool-choice` 필수 |
| 0.6.0+ | 안정 지원 | parser 자동 감지 개선 |

### Tool Call Parser 선택

| Parser | 대상 모델 |
|--------|----------|
| `hermes` | Hermes 계열 (NousResearch 등) |
| `mistral` | Mistral, Mixtral |
| `llama3_json` | Llama 3.x |
| `pythonic` | Qwen 2.5+ |
| `jamba` | AI21 Jamba |

GPT-OSS-120B의 기반 모델에 따라 적절한 parser를 선택해야 합니다.
기반 모델을 모르면 `hermes`부터 시도하세요.

---

## 부록: 프록시 환경에서 자체 CA 인증서 사용

사내 프록시가 TLS를 중간에서 복호화하는 경우 (debug.log에서 `[NET-TLS] Cert Issuer=회사CA` 확인):

```powershell
# Windows — 시스템 인증서 저장소에 자체 CA 추가
# (보통 사내 PC에는 이미 설치되어 있음)

# Go 바이너리에서 시스템 CA를 사용하도록 환경변수 설정
$env:SSL_CERT_DIR = "C:\path\to\certs"
# 또는
$env:SSL_CERT_FILE = "C:\path\to\company-ca.pem"
```

---

## 빌드 명령어 (개발자용)

```bash
# 로컬 디버그 빌드
make build-debug
# → dist/techai-debug

# 사내망 윈도우 디버그 빌드
make build-debug-onprem
# → dist/techai-debug-onprem-windows-amd64.exe

# 일반 빌드 (디버그 없음)
make build
make build-onprem
```
