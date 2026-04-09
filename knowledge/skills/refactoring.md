# Refactoring Patterns

## Extract Method

**Before**:
```python
def process_order(order):
    total = 0
    for item in order.items:
        total += item.price * item.quantity
    discount = total * 0.1 if total > 100 else 0
    return total - discount
```

**After**:
```python
def calculate_total(items):
    return sum(item.price * item.quantity for item in items)

def calculate_discount(total):
    return total * 0.1 if total > 100 else 0

def process_order(order):
    total = calculate_total(order.items)
    discount = calculate_discount(total)
    return total - discount
```

## Rename Variable/Function

```js
// ❌ Bad
function calc(x, y) { return x * y * 0.9 }

// ✅ Good
function calculatePriceWithDiscount(price, quantity) {
    return price * quantity * 0.9
}
```

## Replace Conditional with Polymorphism

**Before**:
```python
def get_speed(vehicle_type):
    if vehicle_type == "car":
        return 120
    elif vehicle_type == "bike":
        return 80
    elif vehicle_type == "plane":
        return 900
```

**After**:
```python
class Car:
    def get_speed(self): return 120
class Bike:
    def get_speed(self): return 80
class Plane:
    def get_speed(self): return 900
```

## Introduce Parameter Object

**Before**:
```js
function createUser(firstName, lastName, email, phone, address) { ... }
```

**After**:
```js
function createUser(userInfo) {
    const { firstName, lastName, email, phone, address } = userInfo
    ...
}
```

## Replace Magic Number with Named Constant

```python
# ❌ Bad
if user.age < 18:
    return "Minor"

# ✅ Good
ADULT_AGE = 18
if user.age < ADULT_AGE:
    return "Minor"
```

## Remove Dead Code

```js
// Delete unused imports, functions, variables
// Use linter to detect dead code
```

## Code Smells

**Long Method**: >20 lines → Extract Method
**God Class**: >500 lines → Split responsibility
**Feature Envy**: 메서드가 다른 클래스 데이터만 사용 → Move to that class
**Duplicate Code**: Copy-paste 금지 → Extract shared logic
**Primitive Obsession**: `(string, string, int)` → Use object/struct

**원칙**: 한 번에 하나씩. 테스트 유지하며 refactor. Working software 깨지 않기.
