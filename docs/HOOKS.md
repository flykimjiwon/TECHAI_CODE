# Lifecycle Hooks — 도구 실행 전후 자동화

## 개요

TGC의 도구 실행 파이프라인에 사용자 정의 셸 명령을 끼워넣는 기능입니다.
파일 수정 후 자동 lint, 셸 실행 전 감사 로그, 작업 완료 알림 등을
코드 수정 없이 설정 파일만으로 구현할 수 있습니다.

## hooks.json 위치

1. **프로젝트 로컬**: `.tgc/hooks.json` (우선)
2. **글로벌**: `~/.tgc/hooks.json` (폴백)

프로젝트 로컬 파일이 있으면 글로벌은 무시됩니다.

## 이벤트 타입

| 이벤트 | 발생 시점 | 용도 |
|--------|----------|------|
| `session_start` | 세션 시작 시 | 환경 검증, 로그 초기화 |
| `pre_tool_use` | 도구 실행 **전** | 감사, 차단 (exit 2 = abort) |
| `post_tool_use` | 도구 실행 **후** | 자동 lint/format, 로깅 |
| `stop` | 에이전트 턴 종료 시 | 알림 (Slack, Teams, 시스템) |
| `user_prompt` | 사용자 입력 시 | 입력 로깅, 필터링 |

## hooks.json 형식

```json
{
  "pre_tool_use": [
    {
      "matcher": { "tool_name": "shell_exec" },
      "command": ["python3", "scripts/audit_command.py"],
      "timeout_ms": 5000
    }
  ],
  "post_tool_use": [
    {
      "matcher": { "tool_name": "file_edit" },
      "command": ["golangci-lint", "run", "--fix"],
      "timeout_ms": 30000
    },
    {
      "matcher": { "tool_name": "apply_patch" },
      "command": ["golangci-lint", "run", "--fix"],
      "timeout_ms": 30000
    }
  ],
  "stop": [
    {
      "command": ["notify-send", "TGC", "작업 완료"],
      "timeout_ms": 3000
    }
  ],
  "session_start": [
    {
      "command": ["bash", "-c", "echo 'TGC session started' >> ~/.tgc/audit.log"],
      "timeout_ms": 2000
    }
  ]
}
```

## Matcher

`matcher.tool_name`으로 특정 도구에만 hook을 적용합니다.
비어있으면 모든 도구에 적용됩니다.

```json
{ "matcher": { "tool_name": "shell_exec" } }  // shell_exec만
{ "matcher": {} }                               // 모든 도구
```

## Hook 입력 (stdin)

Hook 명령에는 JSON 페이로드가 stdin으로 전달됩니다:

```json
{
  "event": "pre_tool_use",
  "tool_name": "shell_exec",
  "tool_args": "{\"command\":\"rm -rf dist\"}",
  "cwd": "/home/user/project"
}
```

post_tool_use에는 추가로 `tool_result` 필드가 포함됩니다.

## 차단 (Abort)

`pre_tool_use` hook이 **exit code 2**로 종료하면 해당 도구 실행이 **차단**됩니다.
LLM에게 "Aborted by pre_tool_use hook" 결과가 전달됩니다.

```bash
#!/bin/bash
# scripts/block_dangerous.sh
INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.tool_args' | jq -r '.command')

if echo "$COMMAND" | grep -qE 'rm -rf|DROP TABLE'; then
  echo "BLOCKED: dangerous command detected" >&2
  exit 2  # exit 2 = abort
fi
exit 0  # exit 0 = continue
```

## Timeout

`timeout_ms`로 hook 실행 제한 시간을 설정합니다 (기본: 10000ms).
타임아웃 시 hook은 실패하지만 도구 실행은 계속됩니다 (`HookFailContinue`).

## 템플릿 생성

```bash
# 프로젝트 로컬 hooks.json 템플릿 생성
# (향후 /hooks-init 슬래시 명령 추가 예정)
```

## 실용 예시

### 파일 수정 후 자동 format
```json
{
  "post_tool_use": [{
    "matcher": { "tool_name": "file_edit" },
    "command": ["gofmt", "-w", "."],
    "timeout_ms": 10000
  }]
}
```

### 모든 도구 실행 감사 로그
```json
{
  "pre_tool_use": [{
    "matcher": {},
    "command": ["bash", "-c", "cat >> ~/.tgc/audit.log"],
    "timeout_ms": 2000
  }]
}
```

### 작업 완료 시 macOS 알림
```json
{
  "stop": [{
    "command": ["osascript", "-e", "display notification \"TGC 작업 완료\" with title \"택가이코드\""],
    "timeout_ms": 3000
  }]
}
```

### 작업 완료 시 Teams Webhook
```json
{
  "stop": [{
    "command": ["curl", "-s", "-X", "POST",
      "https://outlook.office.com/webhook/YOUR_WEBHOOK_URL",
      "-H", "Content-Type: application/json",
      "-d", "{\"text\":\"TGC 작업이 완료되었습니다.\"}"],
    "timeout_ms": 5000
  }]
}
```

## 구현 파일

- `internal/hooks/hooks.go` — 로더, 매처, 실행기
- `internal/app/app.go` — TUI 도구 실행 루프 통합
- `internal/exec/exec.go` — 헤드리스 모드 통합
