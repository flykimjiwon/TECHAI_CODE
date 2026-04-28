# Windows 붙여넣기 (Ctrl+V) 이슈 — 조사 & 대응 기록

> 작성일: 2026-04-28
> 작성자: 김지원
> 상태: 부분 해결 (Bubble Tea 업그레이드 + /paste 폴백)

---

## 증상

Windows 환경에서 특정 텍스트를 Ctrl+V로 붙여넣으면 TUI가 프리징됨.

### 재현 조건
- **OS**: Windows 10/11
- **터미널**: Windows Terminal, PowerShell, cmd.exe
- **IME**: 한글 입력 모드 활성화 상태
- **텍스트**: 영문으로 시작하는 한영 혼합 텍스트 (예: `RWA_IBS_DMB_CMM_MAS 테이블의...`)
- **줄 수**: 3줄 이상 동일 텍스트 반복

### 프리징 되는 예시
```
RWA_IBS_DMB_CMM_MAS 테이블의 DMB_K 컬럼 사용하는 프로그램 알려줘
RWA_IBS_DMB_CMM_MAS 테이블의 DMB_K 컬럼 사용하는 프로그램 알려줘
RWA_IBS_DMB_CMM_MAS 테이블의 DMB_K 컬럼 사용하는 프로그램 알려줘
```

### 프리징 안 되는 예시
```
ㄱRWA_IBS_DMB_CMM_MAS 테이블의 DMB_K 컬럼 사용하는 프로그램 알려줘ㄱ
```
(한글 문자를 앞뒤에 추가하면 정상 동작)

### macOS에서는 정상
macOS에서는 Bracket Paste 모드가 정상 동작하여 `tea.PasteMsg`로 한번에 처리됨.

---

## 근본 원인

### 1. Windows Terminal이 Ctrl+V를 가로챔
- Windows Terminal/PowerShell이 Ctrl+V 키 이벤트를 가로채서 앱에 전달하지 않음
- 대신 클립보드 내용을 **개별 문자 키 이벤트**로 변환하여 전송
- 참고: [charmbracelet/bubbletea#1301](https://github.com/charmbracelet/bubbletea/issues/1301)

### 2. Bracket Paste 미작동
- Windows Terminal은 Bracket Paste를 지원하지만, 한글 IME가 활성화된 상태에서는
  Bracket Paste escape sequence가 IME에 의해 가로채질 수 있음
- 결과: `tea.PasteMsg`가 도착하지 않음 (debug.log에 `[PASTE]` 미기록)

### 3. 한글 IME + 영문 시작 텍스트 충돌
- 한글 IME 모드에서 영문 문자가 입력되면 IME가 조합 이벤트를 비정상 처리
- textarea.Update()가 각 문자마다 호출되면서 내부 렌더링 루프 발생
- 한글로 시작하면 IME가 정상 모드로 동작하여 프리징 없음

### 4. KeyPressMsg.Text 빈 문자열 (v2.0.2 버그)
- Bubble Tea v2.0.2에서 Windows의 `KeyPressMsg.Text`가 항상 빈 문자열이던 버그
- v2.0.6에서 수정됨
- 참고: [charmbracelet/bubbletea#1453](https://github.com/charmbracelet/bubbletea/issues/1453)

---

## 시도한 해결 방법

### 1. Rapid-input 감지 (Enter → 줄바꿈 전환)
- **접근**: 빠른 키 입력 감지 시 Enter를 줄바꿈으로 처리
- **결과**: 실패 — 일반 타이핑도 오탐, 임계값 조정 불가능
- **문제**: 한글 IME 조합 지연으로 인해 정확한 임계값 설정 불가

### 2. Layout 재계산 debounce
- **접근**: 빠른 입력 중 recalcLayout() 호출 건너뛰기
- **결과**: 실패 — textarea.Update() 자체가 프리징 원인

### 3. 키 입력 버퍼링
- **접근**: 빠른 입력 시 textarea 우회, 버퍼에 모았다가 한번에 InsertString
- **결과**: 실패 — msg.String()으로 문자 추출 시 한글/특수키 처리 불완전

### 4. Ctrl+V 직접 가로채기
- **접근**: `case "ctrl+v"` 에서 clipboard.ReadAll() 호출
- **결과**: 실패 — Windows Terminal이 Ctrl+V를 먹어서 앱까지 안 옴

### 5. `/paste` 명령어 (현재 적용)
- **접근**: `/paste` 입력 시 OS 클립보드에서 직접 읽어 InsertString
- **결과**: 성공 — IME 완전 우회, 100% 동작
- **단점**: 사용자가 `/paste`를 직접 입력해야 함

### 6. Bubble Tea v2.0.6 업그레이드 (현재 적용)
- **접근**: KeyPressMsg.Text 빈 문자열 버그 수정 포함 버전으로 업그레이드
- **결과**: 테스트 필요 — 프리징 해결 가능성 있음

---

## 현재 상태 (v0.9.7)

| 환경 | Ctrl+V 붙여넣기 | /paste 명령어 |
|------|:---:|:---:|
| macOS (모든 터미널) | ✅ 정상 | ✅ 정상 |
| Windows (대부분 텍스트) | ✅ 정상 | ✅ 정상 |
| Windows (영문시작+한글IME+3줄반복) | ❌ 프리징 가능 | ✅ 정상 |

---

## 향후 개선 방안

1. **Bubble Tea v2.0.6 테스트**: 업그레이드 후 프리징 재현 여부 확인
2. **textarea 커스텀 구현**: Bubble Tea textarea 대신 자체 구현으로 IME 이벤트 제어
3. **Windows Terminal 설정 가이드**: 사용자에게 IME 관련 설정 안내
4. **OpenCode 참조**: TypeScript 기반이지만 PowerShell API로 클립보드 직접 접근하는 방식 참고

---

## 참고 자료

- [charmbracelet/bubbletea#1301 — Cannot detect Ctrl+V keypress in Windows](https://github.com/charmbracelet/bubbletea/issues/1301)
- [charmbracelet/bubbletea#1453 — KeyPressMsg text property always empty on Windows](https://github.com/charmbracelet/bubbletea/issues/1453)
- [charmbracelet/bubbletea#404 — Missing support for bracketed paste](https://github.com/charmbracelet/bubbletea/issues/404)
- [OpenCode TUI — paste handling with PowerShell clipboard API](https://github.com/anomalyco/opencode)
