# TODO: 도구 비동기 실행 + 실시간 진행 표시

## 현재 문제
- 도구가 `func() tea.Msg` 클로저 안에서 **동기 실행**
- 실행 중 TUI가 완전히 멈춤 (스피너 안 돌고, 화면 안 그려짐)
- "Processing..." 메시지는 도구 실행 **전**에만 보이고, 실행 **중**에는 화면 갱신 없음
- 사용자 입장: "멈췄나?" → 같은 질문 다시 입력

## 해결 방향

### 방법 1: 도구별 비동기 실행 (추천)
```go
// 현재 (동기, 블로킹)
return m, func() tea.Msg {
    for _, tc := range calls {
        output := tools.Execute(tc.Name, tc.Arguments)  // 블로킹
        results = append(results, ...)
    }
    return toolResultMsg{results}
}

// 개선 (비동기, 도구 하나씩 결과 반환)
// 도구 1개 실행 → 결과 표시 → 다음 도구 실행 → 결과 표시
return m, func() tea.Msg {
    output := tools.Execute(calls[0].Name, calls[0].Arguments)
    return singleToolResultMsg{result: ..., remaining: calls[1:]}
}
```

### 방법 2: 3초 tick 기반 진행 표시
```go
// 도구 실행을 goroutine으로
// 3초마다 tickMsg로 "Still searching... (15s elapsed)" 표시
// 완료 시 toolResultMsg 전송
```

### 방법 3: 도구 실행 중 스피너 유지
- streamStatus()처럼 도구 실행 중에도 tick 기반 스피너 표시
- ">> running grep_search... (5.2s)" 실시간 갱신

## 구현 시 주의
- Bubble Tea의 Update는 순차적이라 goroutine에서 msg 전송 필요
- tea.Cmd 체이닝으로 도구 하나씩 실행
- Esc 누르면 도구 실행 취소 (context 전파)

## 예상 효과
- 사용자가 항상 "진행 중"임을 알 수 있음
- "멈춤" 느낌 제거
- 도구 실행 중 Esc로 취소 가능
