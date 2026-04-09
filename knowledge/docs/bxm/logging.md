# 예외 처리

예외 처리한 예외 객체(Exception)를 catch하여 핸들링하는 방법을 말한다.

## 1. 업무상 어떤 로직의 처리 결과에 상관 없이 프로그램이 진행되어야 하는 경우

```
try {
    ...
} catch (DefaultApplicationException e) {
    // 처리
}
```

## 2. 에러 메시지를 변경하는 경우

예외를 catch하였으나 다시 throw하고 있다. 새로운 Exception 생성 시 catch한 Exception을 파라미터로 넣어준다.

```
try {
    ...
} catch (DefaultApplicationException e) {
    throw new DefaultApplicationException("BXME30007", null, e);
}
```

## 3. 특정 예외를 catch하여 메시지 처리하는 경우

```
try {
    ...
} catch (DasDuplicateKeyException e) {
    throw new DefaultApplicationException("BXME30011", new Object[] {empNo}, e);
}
```