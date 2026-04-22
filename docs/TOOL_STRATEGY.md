# 도구 전략 (Tool Strategy)

> 모델 능력에 따른 파일 편집 도구 선택 전략.

---

## 파일 편집 도구 3종

| 도구 | 용도 | 장점 | 단점 |
|------|------|------|------|
| `file_edit` | 한 군데 교체 (old_string → new_string) | 단순, 빠름 | 여러 군데 수정 불가 |
| `apply_patch` | 여러 파일/여러 위치 동시 수정 | 효율적, 다중 hunk | 포맷이 복잡 (@@앵커 + -/+ 라인) |
| `file_write` | 전체 파일 재작성 | 100% 확실 | 토큰 낭비 (전체 파일 출력) |

---

## 모델별 권장 전략

### GPT-OSS-120B (현재 기본 모델)

**문제점:**
- `apply_patch` 2+ hunk 생성 시 50% 확률로 첫 hunk만 생성
- `file_edit`의 old_string 매칭이 불안정 (들여쓰기 오차)
- 빈 응답(thinking만 하고 content 없음) 빈발

**적용 전략:**
- 프롬프트에 파일 경로가 있으면 → auto-prefetch → **도구 목록에서 apply_patch, file_edit, hashline_edit 제거**
- `file_write`만 남겨서 전체 파일 재작성 강제
- 300줄 이하 파일에 최적화

```
# 코드 위치: internal/app/app.go, startStream()
if m.pendingPrefetch != "" {
    // apply_patch, file_edit, hashline_edit 제거
    // file_write만 남김
}
```

### DeepSeek V3.2 / Kimi K2 (향후 업그레이드)

**예상:**
- `apply_patch` 다중 hunk 안정적 생성 가능
- `file_edit` 정확한 매칭

**적용 전략:**
- 도구 제거 없이 전체 제공
- 모델이 상황에 맞게 자동 선택
- apply_patch로 효율적 수정 (토큰 절감)

### Claude Sonnet / Opus급

**예상:**
- 모든 도구 완벽 사용

**적용 전략:**
- 전체 제공 + apply_patch 우선 사용 권장

---

## Auto-Prefetch 시스템

프롬프트에서 파일 경로 감지 → 자동 읽기 → 컨텍스트에 주입.

### 작동 조건
- 사용자 입력에 `/`가 포함된 파일 경로가 있을 때 (예: `src/views/HomePage.tsx`)
- URL (`http://...`)은 제외
- 최대 4개 파일, 누적 50KB 캡
- 300줄 초과 파일은 생략 (file_read 안내)

### 주입 위치
- `startStream()`에서 히스토리 **복사본**의 마지막 user 메시지에 주입
- 원본 `m.history`는 깨끗하게 유지 → 대화 진행 시 컨텍스트 비대 방지
- 도구 반복(tool iteration) 동안 유지, 다음 `sendMessage`에서 교체

### 도구 필터링
- prefetch 활성화 시 apply_patch, file_edit, hashline_edit 제거
- 모델은 file_write (전체 재작성)만 사용 가능
- file_read, grep_search 등 읽기 도구는 유지 (추가 탐색 가능)

### 효과
```
# 이전 (prefetch 없음)
list_files → file_read → file_read → file_read → grep_search
→ apply_patch(1 hunk만) → file_read → grep → apply_patch(실패)
→ ... (10~20 iterations, 30초+)

# 이후 (prefetch 활성화)
[파일 자동 주입] → file_write (전체 교체)
→ (1~3 iterations, 5~15초)
```

---

## apply_patch 안전장치

### 1. 앵커 없는 순수 삽입 차단
- `@@` 앵커 없이 AddLines만 있는 hunk → 파일 맨 위에 삽입됨 → 차단
- 단, 빈 파일(1줄 이하)은 허용

### 2. 첫 줄 변경 감지
- 원래 첫 줄이 directive (`"use client"`, `package main`, `import`, `#!`)인데
- 새 첫 줄이 코드(directive가 아닌 것)면 → 거부
- 정당한 변경 (`"use client"` → `"use server"`)은 허용

### 3. 부분 적용 + 경고
- 2 hunk 중 1개만 성공 시 → 성공한 것은 적용
- 실패한 것은 경고 메시지 + "file_write를 대신 사용하세요" 안내
- 전체 실패(0 hunk 성공) 시 → 에러 반환

---

## 향후 개선

- [ ] 모델 capability 기반 자동 도구 선택 (CodingTier에 따라)
- [ ] apply_patch 성공률 모니터링 (모델별 통계)
- [ ] 300줄 초과 파일에 대한 apply_patch 자동 fallback → file_write
- [ ] temperature=0 모드 추가 (시연용 결정적 출력)
