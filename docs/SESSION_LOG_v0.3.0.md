# 택갈이코드 v0.3.0 세션 로그

> 2026-04-08 작업 기록 — v0.2.0 → v0.3.0 변경사항 및 논의 전체 정리

---

## 1. techai v1 프록시 Tool Calling 문제 분석

### 발견 경위
사내망에서 tool calling curl 테스트 시 `finish_reason: "stop"` 반환. 인프라팀은 활성화했다고 하지만 tool_calls가 안 오는 상황.

### 실제 원인: techai `web/app/api/v1` 프록시 3중 차단

| 단계 | 파일 위치 | 문제 |
|------|----------|------|
| **요청** | `route.js:728-752` | manual provider 템플릿에 `tools` 필드 없음 → vLLM에 tools 안 보냄 |
| **비스트리밍 응답** | `route.js:1086-1106` | `content`만 추출, `tool_calls` 무시. `finish_reason` 항상 `"stop"` 강제 |
| **스트리밍 응답** | `route.js:962-965` | `delta.content`만 처리, `delta.tool_calls` 무시 |

### 실제 API 환경
```
엔드포인트: http://apigtw.gapdev.shinhan.com/openapi/model/5304af32-36f0-458c-8b94-a16f2d046b9d/chat/completions
인증: Bearer 없이 토큰만 (2025121610370549312UD3TCZWVB7W045SHCVH6209L2935FA0)
프로토콜: HTTP (TLS 없음, 사내망)
모델: GPT-OSS-120B
```

onprem 빌드의 `base_url`(`https://techai-web-prod.shinhan.com/v1`)은 위 API Gateway로 프록시하는 구조. 변경 불필요.

### 해결 방향
1. GPT-OSS-120B를 `openai-compatible` provider로 등록 (가장 빠름)
2. 또는 manual 템플릿에 tools/tool_choice 추가 + 응답 처리 수정

---

## 2. v0.3.0 코드 변경사항

### 2-1. Cmd+V / Ctrl+V 붙여넣기 수정
- **원인**: Bubble Tea v2에서 붙여넣기는 `tea.PasteMsg` 타입으로 들어오지만, `switch msg.(type)`에 해당 case 없어서 무시됨
- **수정**: `tea.PasteMsg` → `InsertString()` 직접 삽입 + `tea.PasteStartMsg`/`tea.PasteEndMsg` 무시 처리
- **파일**: `internal/app/app.go:295-312`

### 2-2. 입력창 줄 수 표시
- 10줄 초과 텍스트 붙여넣기 시 `[15줄]` 인디케이터 표시
- **파일**: `internal/app/app.go` View() 메서드

### 2-3. 스트리밍 상태 실시간 표시
- 기존: `thinking... (3.2s · 15tok)` — 청크 올 때만 갱신 (멈춰 보임)
- 변경:
  - `⠋ 연결중... (0.3s)` — 첫 청크 대기 (0.1초 단위 실시간)
  - `⠋ 수신중 (15tok · 8tok/s · 3.2s)` — 정상 수신
  - `⚠ 응답 지연 8초... (15tok · 8.2s)` — 5초 이상 청크 없음
  - `⚠ 응답 없음 20초 — 연결 확인 필요` — 15초 이상
- **100ms 틱** 추가로 UI 실시간 갱신
- **파일**: `internal/app/app.go` — `streamStatus()`, `streamTickMsg`

### 2-4. 60초 자동 재시도
- `waitForNextChunk()`에 `time.NewTimer(60s)` + `select` 추가
- 60초 무응답 시 최대 3회 자동 재전송
- `[응답 지연 — 자동 재시도 1/3]` 메시지 표시
- 3회 실패 시 `[응답 없음 — 3회 재시도 실패. 네트워크를 확인하세요]`
- **버그 수정**: 재시도 시 `streamBuf` 미초기화 → 응답 중복. 부분 응답을 history에 저장하고 buf 초기화

### 2-5. Tool Calling 점검 (startup probe)
- 앱 시작 시 `tool_choice: "required"` + calculator tool로 테스트 요청
- 15초 타임아웃, 비동기 실행 (앱 차단 안 함)
- 결과를 채팅 영역 + 디버그 로그에 기록
- **v0.3.0 이후**: debug 전용 → 일반 모드에서도 실행 (`2c02dc4`)
- **파일**: `internal/llm/client.go` — `CheckToolSupport()`

### 2-6. 상태바 Tool:ON/OFF 표시
- `Tool:ON` (초록 #34D399) — tool calling 동작 확인
- `Tool:OFF` (빨강 #F87171) — tool calling 미지원/비활성화
- `Tool:--` (회색 #9CA3AF) — 점검 중
- **파일**: `internal/ui/chat.go` — `RenderStatusBar()` 시그니처 변경

---

## 3. 크로스플랫폼 점검 결과

### macOS / Linux
모든 기능 정상 동작. ✅

### Windows
| 항목 | Windows Terminal | cmd.exe (레거시) |
|------|:---:|:---:|
| 붙여넣기 | ✅ PasteMsg | ✅ KeyPressMsg fallback |
| 스피너 `⠋⠙⠹` | ✅ | ⚠ 깨질 수 있음 (cp437) |
| `⚠` 경고 | ✅ | ⚠ 깨질 수 있음 |
| 한국어 | ✅ | ⚠ `chcp 65001` 필요 |

> Windows 레거시 이슈는 기존 코드도 동일 (로고, 모드명 전부 한국어). 이번 변경으로 새로 생긴 문제 아님.

---

## 4. 릴리즈 v0.3.0

**커밋**: `744cc6b` → `2c02dc4`
**GitHub**: https://github.com/flykimjiwon/TECHAI_CODE/releases/tag/v0.3.0

### 빌드 목록 (11개)
| 파일 | 용도 |
|------|------|
| techai-darwin-arm64 | macOS Apple Silicon |
| techai-darwin-amd64 | macOS Intel |
| techai-windows-amd64.exe | Windows |
| techai-linux-amd64 | Linux x64 |
| techai-linux-arm64 | Linux ARM |
| techai-onprem-darwin-arm64 | 사내망 macOS Apple Silicon |
| techai-onprem-darwin-amd64 | 사내망 macOS Intel |
| techai-onprem-windows-amd64.exe | 사내망 Windows |
| techai-onprem-linux-amd64 | 사내망 Linux x64 |
| techai-onprem-linux-arm64 | 사내망 Linux ARM |
| techai-debug-onprem-windows-amd64.exe | 사내망 디버그 Windows |

---

## 5. Gemma 4 모델 분석 — 사내망 도입 검토

### 사내망 보유 모델
- GPT-OSS-120B (현재 메인)
- Qwen-Coder-30B (개발 모드)
- **Gemma 4** (신규 도입 예정)

### Gemma 4 사이즈별 아키텍처 차이

단순 크기 차이가 아니라 **아키텍처 자체가 다름**.

#### E2B / E4B — PLE(Per-Layer Embeddings) 구조
- 일반 트랜스포머는 입력 시 한 번만 임베딩 → PLE는 **레이어마다** 토큰 전용 신호 별도 주입
- E4B: 총 8B, 유효 4.5B / E2B: 총 5.1B, 유효 2.3B
- **오디오 입력** (USM 인코더, 최대 30초) — E2B/E4B만 지원
- 128K 컨텍스트
- **택갈이코드에는 너무 작음** — 에이전트 용도 부적합

#### 26B A4B — MoE (Mixture of Experts) 가성비 최강
- 128개 experts 중 토큰당 8개 + 공유 1개만 활성화
- 추론 속도 4B급, 품질 31B급 근처
- **18GB → RTX 3090/4090 한 장**
- 256K 컨텍스트
- 서버 처리량 중요 시 최적

#### 31B Dense — 가장 단순, 가장 강력
- 모든 파라미터 항상 활성화 (전통 구조)
- LMArena 오픈웨이트 #3
- **20GB → 4090 빠듯, 속도는 26B보다 느림**
- 256K 컨텍스트

#### 31B-cloud — 로컬 없음
- Google AI Studio API로 실행 (weights 다운로드 X)
- 사내망 불가

### 벤치마크 비교

| | Gemma4 31B | Gemma4 26B | GPT-OSS-120B | Qwen-Coder-30B |
|---|---|---|---|---|
| **MMLU Pro** | 85.2% | 82.6% | ? | ? |
| **AIME 2026** | 89.2% | 88.3% | ? | ? |
| **LiveCodeBench v6** | 80.0% | 77.1% | ? | 코딩 특화 |
| **Codeforces ELO** | 2150 | 1718 | ? | ? |
| **Tool Calling** | **네이티브** | **네이티브** | ⚠ 프록시 문제 | vLLM hermes 필요 |
| **Thinking 모드** | **있음** | **있음** | 없음 | 없음 |
| **컨텍스트** | **256K** | **256K** | ? | 32K~128K |
| **멀티모달** | 텍스트+이미지 | 텍스트+이미지 | 텍스트만 | 텍스트만 |

### 추천 구성

```
슈퍼택가이 (메인)  →  Gemma 4 31B   (추론+코딩+tool calling 만능)
개발 모드          →  Gemma 4 26B   (MoE라 빠름, 코딩 77% 충분)
플랜 모드          →  Gemma 4 31B   (분석/계획은 추론력 필요)
서브에이전트       →  Gemma 4 26B   (처리량 우선)
```

### Gemma 4 메인 전환 이유
1. **Tool calling 네이티브** — GPT-OSS-120B 프록시 문제 없음
2. **Thinking 모드** — `<|think|>` 토큰으로 reasoning chain 생성/제어 가능
3. **256K 컨텍스트** — 큰 파일 분석에 유리
4. **벤치마크 압도적** — AIME 89.2%, LCB 80.0%는 30B급 최상위

### 택갈이코드 추가 작업 필요
- [ ] vLLM Gemma 4 tool call parser 확인 (인프라팀)
- [ ] Thinking 모드 지원 — `<|think|>` / `<|channel>thought` 토큰 파싱, UI에서 접기/숨기기
- [ ] 모델 설정에 Gemma 4 31B / 26B 추가
- [ ] 멀티모달 (이미지) 입력 지원 검토

---

## 6. Tool Calling 동작 원리 (정리)

### 흐름
```
사용자: "파일 만들어줘"
  → 택갈이코드: tools 파라미터와 함께 /chat/completions 호출
    → LLM: finish_reason="tool_calls" + tool_calls 배열 반환
      → 택갈이코드: tool 실행 (file_write, shell_exec 등)
        → 결과를 history에 추가 → LLM에 재전송
```

### tool calling 미지원 시
```
사용자: "파일 만들어줘"
  → LLM: finish_reason="stop" + content="file_write({...})" ← 텍스트로만 출력
  → 실제 실행 안 됨
```

### finish_reason: "stop" 판별법
| 상황 | content | 의미 |
|------|---------|------|
| Tool calling 비활성화 | `file_write({...})` 텍스트 | 서버 설정 문제 |
| 모델이 도구 안 쓰기로 판단 | 자연스러운 답변 | 정상 |
| tools 파라미터 안 보낸 경우 | 답변 | 당연히 stop |

### vLLM tool calling 활성화 옵션
```bash
vllm serve [MODEL] \
  --enable-auto-tool-choice \
  --tool-call-parser hermes   # 모델에 따라 변경
```

| 모델 계열 | parser |
|----------|--------|
| Hermes / Qwen | hermes |
| Llama 3.1+ | llama3_json |
| Mistral | mistral |
| Gemma 4 | 확인 필요 |
