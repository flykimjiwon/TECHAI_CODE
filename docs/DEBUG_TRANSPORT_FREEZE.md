# debugTransport 스트리밍 Freeze 버그 포스트모템

> 2026-04-09 해결 | 영향: 전체 빌드에서 채팅 응답 freeze

## 증상

- `make build` 후 실행하면 채팅 입력 시 "연결중... (0.0s)"에서 멈춤
- 스피너가 정지하고 응답이 오지 않음
- 구버전 바이너리(commit 5e0c1a2)는 정상 동작

## 원인

### 핵심: `debugTransport` HTTP 래퍼가 SSE 스트리밍을 블로킹

`client.go`의 `NewClient()`에서 `DebugMode=true`일 때:

1. `http.DefaultTransport`를 Clone → `debugTransport`로 감쌈
2. `debugTransport.RoundTrip()`이 모든 요청/응답 헤더, TLS 인증서 정보를 동기적으로 로깅
3. Clone된 Transport + 새 TLSClientConfig가 SSE 스트리밍 chunk 전달을 블로킹
4. Bubble Tea UI 갱신 루프에 데이터가 도달하지 못해 freeze

### 왜 발생했나

**구버전 Makefile (정상):**
```makefile
LDFLAGS = -ldflags "-s -w -X main.version=$(VERSION)"
```
- DebugMode 미설정 → 기본값 `"false"` → go-openai 기본 HTTP 클라이언트 사용

**신버전 Makefile (freeze):**
```makefile
LDFLAGS = -ldflags "-s -w -X main.version=$(VERSION) \
    -X 'github.com/kimjiwon/tgc/internal/config.DebugMode=true'"
```
- 모든 빌드에 DebugMode 강제 → debugTransport 활성 → 스트리밍 블로킹

비유: 고속도로에 검문소(debugTransport)를 세워서 차량(chunk)마다 사진 찍고 기록하다가 도로가 막힌 것.

## 해결 (최종)

**debugTransport HTTP 래핑을 코드에서 제거**, DebugMode는 유지.

```go
// client.go NewClient() — 변경 전
if config.IsDebug() {
    // DNS pre-check ...
    origTransport := http.DefaultTransport.(*http.Transport).Clone()
    origTransport.TLSClientConfig = &tls.Config{...}
    cfg.HTTPClient = &http.Client{
        Transport: &debugTransport{inner: origTransport},
        Timeout: 0,
    }
}

// client.go NewClient() — 변경 후
if config.IsDebug() {
    // DNS pre-check ... (유지)
    // NOTE: debugTransport HTTP wrapping removed — blocks SSE streaming.
    // Application-level DebugLog calls in StreamChat are sufficient.
}
```

- `make build` → DebugMode=true (앱 레벨 로깅 O, HTTP 래핑 X) → 정상
- StreamChat 내부의 DebugLog는 goroutine 안에서 실행되므로 스트리밍에 영향 없음
- `~/.tgc/debug.log`에 chunk 타이밍, 모델, 메시지 등 유용한 로그 계속 기록됨

## 재발 방지 규칙

1. **`debugTransport`로 HTTP 클라이언트를 감싸지 말 것** — SSE 스트리밍과 비호환
2. 디버그 로깅은 앱 레벨(`DebugLog()`)에서만 수행
3. HTTP 레벨 디버깅이 필요하면 별도 CLI 도구(curl, mitmproxy)를 사용
4. go-openai의 기본 HTTP 클라이언트를 교체하지 말 것

## 디버깅 타임라인

| 시도 | 내용 | 결과 |
|------|------|------|
| 1차 | GatherSystemContext 제거 + heartbeat 추가 | freeze 지속 |
| 2차 | curl로 API 직접 테스트 → API 정상 확인 | API 문제 아님 |
| 3차 | prompt.go/client.go/registry.go 구버전 복원 | freeze 지속 |
| 4차 | MouseMode 재활성화 | freeze 지속 |
| 5차 | app.go/super.go/chat.go/styles.go 구버전 복원 | freeze 지속 |
| 6차 | models.go/config.go 구버전 복원 | freeze 지속 |
| **7차** | **Makefile 구버전 복원 (DebugMode=true 제거)** | **정상 동작** |
| **7차** | **debugTransport 래핑 코드 제거 + DebugMode=true 유지** | **정상 동작** |
| 8차 | 신버전 app.go + UI 전체 복원 (debugTransport 제거 상태) | freeze 재발 |
| 9차 | 신버전 prompt.go + registry.go만 복원 (구 app.go 유지) | **정상 동작** |
| **최종** | **구 app.go 기반 + 신 prompt/registry + Tool:ON 상태바 추가** | **정상 동작** |

## 발견: 두 가지 독립적 원인

1. **debugTransport HTTP 래핑** — SSE 스트리밍 블로킹 (해결: 래핑 코드 제거)
2. **신버전 app.go의 UI/이벤트 루프 변경** — 정확한 원인 미확정, 구 app.go 유지로 회피
   - 의심 지점: MouseMode 비활성화, textarea 메시지 포워딩, StatusBarData 구조체, streamStatus() 변경
   - prompt.go/registry.go 변경(한국어, 7도구)은 무관 확인됨

## 교훈

- 소스코드가 동일해도 **빌드 플래그(ldflags)**가 다르면 동작이 달라진다
- HTTP Transport 래핑은 일반 요청에선 문제없지만 **SSE 스트리밍에선 블로킹** 발생 가능
- 디버깅 시 소스코드뿐 아니라 Makefile/빌드 시스템도 반드시 비교할 것
- **여러 원인이 동시에 존재**할 수 있다 — 하나 고쳐도 다른 원인이 남아있을 수 있음
- app.go 같은 대형 변경은 하나씩 추가하며 테스트할 것

## 관련 파일

| 파일 | 역할 |
|------|------|
| `Makefile` | 빌드 플래그 관리 (ldflags), DebugMode=true 기본 |
| `internal/llm/client.go` | ~~debugTransport 래퍼~~ (제거됨), `NewClient()`, `StreamChat()` |
| `internal/app/app.go` | 구버전(5e0c1a2) 기반 + Tool:ON 상태바 추가 |
| `internal/llm/prompt.go` | 신버전 (한국어 프롬프트, 7도구) |
| `internal/tools/registry.go` | 신버전 (grep_search, glob_search 추가) |
| `internal/ui/chat.go` | 구버전 기반 + Tool:ON(N)/OFF 상태바 표시 추가 |
