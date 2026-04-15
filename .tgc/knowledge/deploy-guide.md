# 택가이코드 배포 가이드

## 빌드 방법
1. `make build` — 로컬 바이너리 생성 (macOS arm64)
2. `make build-onprem` — 온프레미스 빌드 (신한 gpt-oss-120b)
3. `make build-release` — 전체 15개 크로스 플랫폼 릴리스

## 배포 대상
- macOS (arm64, amd64)
- Windows (amd64)
- Linux (amd64, arm64)

## 온프레미스 설정
- 엔드포인트: `https://techai-web-prod.shinhan.com/v1`
- Super 모델: `openai/gpt-oss-120b`
- Dev 모델: `qwen/qwen3-coder-30b`
- 설정 디렉토리: `~/.tgc-onprem/`

## 주의사항
- API 키는 `.env`가 아닌 `~/.tgc/config.yaml`에 저장
- 온프레미스와 Novita 빌드는 설정 디렉토리가 분리됨
