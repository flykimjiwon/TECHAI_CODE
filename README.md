# 택가이코드 (TaekgaliCode)

OpenAI-compatible API 기반 CLI AI 코딩 어시스턴트. Go + Bubble Tea 단일 바이너리.

## 특징

- **3개 모드**: 슈퍼택가이(만능) / 플랜(계획) / 개발(코딩) — Tab 키로 전환
- **2개 모델**: gpt-oss-120b (범용), qwen-coder-30b (코딩 특화)
- **스트리밍**: 실시간 응답 스트리밍
- **도구**: 파일 읽기/쓰기, 셸 명령 실행 (개발 모드)
- **단일 바이너리**: 크로스 플랫폼 (macOS, Windows, Linux)

## 설치

```bash
# 소스에서 빌드
go install github.com/kimjiwon/tgc/cmd/tgc@latest

# 또는 직접 빌드
git clone https://github.com/kimjiwon/tgc.git
cd tgc
make build
```

## 설정

첫 실행 시 자동으로 설정 위저드가 실행됩니다.

```bash
# 직접 설정 실행
tgc --setup

# 또는 환경변수
export TGC_API_BASE_URL=https://your-api.com/v1
export TGC_API_KEY=sk-...
```

설정 파일: `~/.tgc/config.yaml`

```yaml
api:
  base_url: "https://api.openai.com/v1"
  api_key: "sk-..."
models:
  super: "gpt-oss-120b"
  plan: "gpt-oss-120b"
  dev: "qwen-coder-30b"
```

## 사용법

```bash
tgc                # 기본 (슈퍼택가이 모드)
tgc --mode dev     # 개발 모드로 시작
tgc --mode plan    # 플랜 모드로 시작
tgc --version      # 버전 출력
```

## 키 바인딩

| 키 | 동작 |
|---|------|
| `Tab` | 모드 전환 (슈퍼택가이 → 플랜 → 개발) |
| `Enter` | 메시지 전송 |
| `Alt+Enter` | 줄바꿈 |
| `Esc` | 스트리밍 중단 |
| `Ctrl+L` | 화면 클리어 |
| `Ctrl+C` | 종료 (스트리밍 중이면 중단) |

## 빌드

```bash
make build          # 현재 플랫폼
make build-all      # 크로스 컴파일 (5개 타겟)
make install        # go install
make test           # 테스트
make lint           # go vet
```
