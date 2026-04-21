# 메시지처리 (정상, 오류)

메시지는 온라인 서비스의 처리 결과 또는 에러 내용을 호출한 채널로 전송하는 문자열 데이터를 말한다. 정상 메시지와 오류 메시지는 동일한 체계로 관리된다.

## 1. 메시지 관리

업무 프로그램에서 메시지 처리 시 코드를 세팅하며, 그 내용은 BXM Web Admin에서 관리한다.

## 2. 정상 메시지

온라인 서비스가 정상 처리되었을 때 응답할 메시지를 설정한다. 설정한 메시지는 Header에 세팅된다.

DefaultApplicationContext.addMessage() API를 사용하여 메시지를 설정한다. 가변 메시지는 Object[] 객체에 담아 처리한다.

```
/*
* BXMI60000
* 기본 메시지 : 거래요청을 정상적으로 처리했습니다.
*/
DefaultApplicationContext.addMessage("BXMI60000", null, new Object[] {}); // 거래요청을 정상적으로 처리했습니다.
```

## 3. 오류 메시지

온라인 서비스를 오류 처리할 때 응답할 메시지를 설정한다. 의도하지 않은 런타임 에러는 프레임워크에서 오류 메시지를 조립한다. 가변 메시지는 Object[] 객체에 담아 처리한다.

```
/*
* BXME30000
* 기본 메시지 : 거래요청을 정상적으로 처리하지 못했습니다.
* 상세 메시지 : 상세메시지 : {0}
*/
throw new DefaultApplicationException("BXME30000", new Object[] {}, new Object[] { "Pre-Deploy Test Exception." }); // Object[] 객체에 담긴 메시지는 {0} 위치에 파싱된다.

/*
* BXME30006
* 기본 메시지 : 요청하신 서비스 [{0}]의 오퍼레이션 [{1}] 처리중 관리자가 지정한 제한시간[{2}]ms를 초과하였습니다.
*/
throw new DefaultApplicationException("BXME30006", new Object[] {}, new Object[] {"DummyService", "getDummy", 5000}); // 요청하신 서비스 [DummyService]의 오퍼레이션 [getDummy] 처리중 관리자가 지정한 제한시간[5000]ms를 초과하였습니다.
```