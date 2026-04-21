# Headless Exec — TUI 없는 CLI 실행

## 개요

`techai exec`는 TUI를 띄우지 않고 프롬프트를 실행하는 헤드리스 모드입니다.
stdout으로 결과를 출력하고, 파이프로 stdin을 받을 수 있어
CI/CD, 셸 스크립트, 자동화 파이프라인에 TGC를 통합할 수 있습니다.

## 기본 사용법

```bash
# 단순 프롬프트
techai exec "이 프로젝트의 TODO를 정리해줘"

# 파이프 입력
cat error.log | techai exec "이 에러 분석해줘"

# git diff 리뷰
git diff HEAD~1 | techai exec "코드 리뷰해줘"

# 커밋 메시지 생성
git diff --staged | techai exec "커밋 메시지를 conventional commits 형식으로 써줘"
```

## 옵션

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `--ephemeral` | false | 세션 저장 안 함 (일회성 작업) |
| `--model` | config의 super | 사용할 모델 지정 |
| `--max-turns` | 20 | 최대 도구 실행 반복 횟수 |

## 동작 방식

1. config.yaml에서 API 설정 로드
2. 시스템 프롬프트 + .techai.md 프로젝트 컨텍스트 구성
3. stdin이 파이프되면 `<stdin>...</stdin>` 블록으로 프롬프트에 첨부
4. LLM에 스트리밍 요청 → 텍스트는 stdout으로 실시간 출력
5. 도구 호출이 있으면 실행 후 결과를 LLM에 전달 (agentic loop)
6. 도구 호출 없이 텍스트만 출력되면 종료

## CI/CD 통합 예시

### GitHub Actions
```yaml
- name: AI Code Review
  run: |
    git diff origin/main | techai exec --ephemeral "코드 리뷰해줘. 버그 있으면 알려줘" > review.txt
    cat review.txt
  env:
    TGC_API_KEY: ${{ secrets.TGC_API_KEY }}
    TGC_API_BASE_URL: ${{ secrets.TGC_API_BASE_URL }}
```

### Jenkins
```groovy
stage('AI Review') {
    sh '''
        git diff HEAD~1 | techai exec --ephemeral \
          "보안 취약점이나 성능 이슈가 있는지 리뷰해줘" > ai-review.txt
        cat ai-review.txt
    '''
}
```

### 셸 스크립트
```bash
#!/bin/bash
# 에러 로그 자동 분석
ERRORS=$(tail -50 /var/log/app.log | grep ERROR)
if [ -n "$ERRORS" ]; then
  echo "$ERRORS" | techai exec --ephemeral "이 에러 로그를 분석하고 해결 방안을 제시해줘"
fi
```

### 배치 처리
```bash
# 모든 Go 파일에 대해 코드 품질 체크
for f in $(find . -name "*.go" -not -path "*/vendor/*"); do
  echo "=== $f ==="
  cat "$f" | techai exec --ephemeral "이 Go 파일의 코드 품질을 점검해줘. 개선점만 간결하게."
  echo
done
```

## Hooks 연동

헤드리스 모드에서도 lifecycle hooks가 동작합니다.
`.tgc/hooks.json`이나 `~/.tgc/hooks.json`에 설정된 hook이
도구 실행 전후에 자동으로 실행됩니다.

## 환경 변수

exec 모드도 일반 TUI 모드와 동일한 환경 변수를 사용합니다:

| 변수 | 설명 |
|------|------|
| `TGC_API_BASE_URL` | API 엔드포인트 |
| `TGC_API_KEY` | API 키 |
| `TGC_MODEL_SUPER` | 사용할 모델 |

## Exit Code

| 코드 | 의미 |
|------|------|
| 0 | 성공 |
| 1 | 에러 (API 실패, config 없음 등) |

## 구현 파일

- `internal/exec/exec.go` — 헤드리스 실행 엔진
- `cmd/tgc/main.go` — `exec` 서브커맨드 파싱
