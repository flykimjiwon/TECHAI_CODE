# Code Review Checklist

## 로직

- [ ] Edge cases 처리 (빈 배열, null, undefined, 0, 음수)
- [ ] Off-by-one 에러 (< vs <=, length-1)
- [ ] Null/undefined 체크 (optional chaining, null coalescing)
- [ ] Loop 종료 조건 정확성
- [ ] 비동기 처리 (await 누락, race condition)

```js
// ❌ Bad
items.map(item => item.price).reduce((a, b) => a + b)  // 빈 배열 crash

// ✅ Good
items.reduce((sum, item) => sum + item.price, 0)
```

## 보안

- [ ] SQL Injection (parameterized query 사용?)
- [ ] XSS (사용자 입력 sanitize/escape?)
- [ ] CSRF (token 검증?)
- [ ] 인증/인가 (권한 체크?)
- [ ] Secret 하드코딩 금지 (env var 사용?)

```python
# ❌ Bad
cursor.execute(f"SELECT * FROM users WHERE id = {user_id}")

# ✅ Good
cursor.execute("SELECT * FROM users WHERE id = ?", (user_id,))
```

## 성능

- [ ] N+1 쿼리 (loop 안에 DB 호출?)
- [ ] 불필요한 반복문 (filter + map → reduce?)
- [ ] Missing DB index (WHERE/JOIN 컬럼 인덱싱?)
- [ ] 메모리 누수 (listener cleanup, connection close?)
- [ ] 대용량 데이터 (pagination, streaming?)

```js
// ❌ Bad
for (const userId of userIds) {
  const user = await db.findUser(userId)  // N+1
}

// ✅ Good
const users = await db.findUsers({ id: { $in: userIds } })
```

## 가독성

- [ ] 변수명 명확성 (`x` → `userId`, `flag` → `isActive`)
- [ ] 함수 길이 (<20 lines 권장)
- [ ] Magic number → Named constant
- [ ] 주석 필요성 (코드가 설명하면 주석 불필요)
- [ ] 일관된 스타일 (팀 컨벤션 준수)

## 테스트

- [ ] 핵심 로직 커버리지 90%+
- [ ] Assertion 품질 (구체적인 값 검증, not just truthy)
- [ ] Edge case 테스트
- [ ] Mock 남용 금지 (외부 의존성만)

## Severity

**Critical**: 보안 취약점, 데이터 손실 위험
**Major**: 성능 이슈, 로직 버그
**Minor**: 가독성, 스타일
