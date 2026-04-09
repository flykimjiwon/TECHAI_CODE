# Test-Driven Development

## Red-Green-Refactor Cycle

1. **Red**: 실패하는 테스트 작성
2. **Green**: 테스트를 통과하는 최소 코드 작성
3. **Refactor**: 중복 제거, 구조 개선

```python
# 1. Red
def test_calculate_total_with_discount():
    assert calculate_total([100, 200], discount=0.1) == 270

# 2. Green
def calculate_total(items, discount=0):
    return sum(items) * (1 - discount)

# 3. Refactor (if needed)
```

## Test Pyramid

```
       /\
      /e2e\      적음 (느리고 취약)
     /------\
    /integra\   중간 (실제 연동 검증)
   /----------\
  /   unit     \ 많음 (빠르고 격리)
 /--------------\
```

**Unit 90%+ > Integration 10% > E2E 5%**

## 테스트 네이밍

```
test_[기능]_[조건]_[기대결과]
```

Examples:
- `test_user_login_with_invalid_password_returns_401`
- `test_calculate_discount_when_amount_over_100_applies_10_percent`
- `test_create_order_without_items_raises_validation_error`

## Arrange-Act-Assert Pattern

```python
def test_add_item_to_cart():
    # Arrange
    cart = ShoppingCart()
    item = Item("Apple", 1.5)

    # Act
    cart.add(item)

    # Assert
    assert cart.total == 1.5
    assert len(cart.items) == 1
```

## Mocking 원칙

**외부 의존성만 mock**:
- ✅ API 호출, DB 쿼리, 파일 I/O, 시간(datetime.now)
- ❌ 내부 로직, 단순 함수, DTO/모델

```python
from unittest.mock import patch

@patch('requests.get')
def test_fetch_user_data(mock_get):
    mock_get.return_value.json.return_value = {'id': 1}
    user = fetch_user(1)
    assert user['id'] == 1
```

## Coverage 목표

**핵심 비즈니스 로직: 90%+**
**유틸/헬퍼: 70%+**
**Config/setup 코드: skip 가능**

100% 집착 금지. Edge case 우선.
