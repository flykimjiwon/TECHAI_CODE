# 트랜잭션 처리

표준 프레임워크를 사용하는 어플리케이션 개발환경은 사용자가 트랜잭션에 관련된 세부 제어를 관여하지 않는 것을 원칙으로 한다. 이를 위해 사용자는 메소드 단위(오퍼레이션 포함)로 속성을 부여하고 프레임워크가 실행환경에서 트랜잭션을 제어하는 Declarative Transaction Management(선언적 트랜잭션 정의)모델을 사용한다.

일부 업무 특성에 따라 사용자가 트랜잭션 관리 코드를 이용하여 상세하게 트랜잭션을 분리하거나 예외에 따른 트랜잭션 제어를 수행하여야 하는 경우 프레임워크에서 제공하는 API를 이용하여 Programmatic Transaction Management방법으로 트랜잭션을 제어할 수 있다.

## 1. Transactional Operation

업무로직을 처리하는 과정에서 수행된 모든 데이터 처리가 하나의 묶음으로 조회, 반영되어야 하는 단위를 하나의 트랜잭션으로 정의한다.

(1) 다음의 예에서처럼 트랜잭션의 단위는 오퍼레이션의 전체 실행구간이거나 세부적인 요건에 따라 일부 구간일 수 있다.

Figure 1. 서비스 오퍼레이션 전체가 트랜잭션 구간

```
@BxmService("SSMP1001A")
@BxmCategory(logicalName = "단건 처리")
public class SSMP1001A {
    final Logger logger = LoggerFactory.getLogger(this.getClass());
    private MSmpEmpInfMng mSmpEmpInfMng;

    @BxmServiceOperation("ssmp1001a004")
    @BxmCategory(logicalName = "단건 삭제")
    @TransactionalOperation
    public SSMP1001A004OutDto ssmp1001a004(SSMP1001A004InDto input) throws DefaultApplicationException {

        logger.debug("============== SERVICE START ==============");
        logger.debug("input = {}", input);

        mSmpEmpInfMng = DefaultApplicationContext.getBean(mSmpEmpInfMng, MSmpEmpInfMng.class);

        SSMP1001A004OutDto output = new SSMP1001A004OutDto();

        /**
         * 입력값 매핑
         */
        DSmpEmpTst000Dto beanInput = new DSmpEmpTst000Dto();
        beanInput.setFeduEmpNo(input.getFeduEmpNo());

        /**
         * 샘플직원정보관리 bean 조회 호출
         */
        int removeCnt = mSmpEmpInfMng.removeEmpInf(beanInput);

        /**
         * 출력값 매핑
         */
        output.setSuccYn(removeCnt == 1 ? "Y" : "N");

        /**
         * 정상처리 종료
         */
        DefaultApplicationContext.addMessage("BXMI60000", null, new Object[] { });

        logger.debug("output = {}", output);
        logger.debug("============== SERVICE END ==============");

        return output;
    }
}
```

(2) 서비스 오퍼레이션에서 호출된 일부 Bean 메소드가 트렌잭션 구간이 될 수 있다.

Figure 2. 서비스 오퍼레이션에서 호출된 일부 빈(Bean) 메소드가 트랜잭션 구간

```
@BxmBean
@BxmCategory(logicalName = "샘플직원정보관리")
public class MSmpEmpInfMng {
    final Logger logger = LoggerFactory.getLogger(this.getClass());

    private DSmpEmpTst000 dSmpEmpTst000; /*샘플직원정보TST */

    /**
     * 직원정보를 단건으로 조회한다.
     *
     * @param    input    DSmpEmpTst000Dto
     * @return DSmpEmpTst000Dto
     * @throws DefaultApplicationException
     */
    @BxmCategory(logicalName = "단건 조회")
    @TransactionalOperation
    public DSmpEmpTst000Dto getEmpInf(DSmpEmpTst000Dto input) throws DefaultApplicationException {

        logger.debug("============== START ==============");
        logger.debug("input = {}", input);

        dSmpEmpTst000 = DefaultApplicationContext.getBean(dSmpEmpTst000, DSmpEmpTst000.class);

        DSmpEmpTst000Dto output = null;

        /**
         * 샘플직원번호 selectOne
         */
        output = dSmpEmpTst000.selectOne00(input);

        logger.debug("output = {}", output);
        logger.debug("============== END ==============");

        return output;
    }
}
```

업무 개발자는 업무요건에 따라 트랜젝션 구간을 파악하여 서비스-오퍼레이션 메소드 또는 빈 메소드에 "@TransactionalOperation"어노테이션을 지정한다.

## 2. Declarative Transaction Management

위의 코드 예에서처럼 트랜잭션의 단위를 명시적으로 선언하여 처리하는 방법을 트랜잭션에 대한 선언적정의 모델로 규정한다. 이러한 모델에 따라 업무 개발자는 메소드 단위로 어노테이션을 이용하여 트랜잭션 속성을 부여하게 되는데 이 때 부여할 수 있는 상세 속성들은 다음과 같다.

Figure 3. Transaction Attributes for Using Annotation

propagation : 트랜잭션 처리 단위를 묶기 위한 트랜잭션 전파에 관련된 옵션

REQUIRED : 기존에 시작된 트랜잭션이 있는 경우 이에 참여하고, 없는 경우에는 새로운 트랜잭션을 시작한다.

REQUIRES_NEW : 항상 새로운 트랜잭션을 시작한다. 이미 진행중인 트랜잭션이 있다면 기존의 트랜잭션은 일시정지된다. 이 옵션을 이용하여 데이터 처리를 진행중인 트랜잭션과 분리할 수 있다.

SUPPORTS : 기존에 시작된 트랜잭션이 있는 경우 이에 참여하고, 없는 경우에는 트랜잭션과 관계없이(no transactional) 처리된다.

MANDATORY : 기존에 시작된 트랜잭션이 있는 경우 이에 참여하고, 없는 경우 예외를 발생시킨다.

NOT_SUPPORTED : 기존에 시작된 트랜잭션이 있는 경우 이를 일시정지한다. 항상 트랜잭션과 관계없이(no transactional) 처리된다.

NEVER : 항상 트랜잭션과 관계없이(no transactional)처리된다. 기존에 시작된 트랜잭션이 있는 경우 예외를 발생시킨다.

NESTED : 기존에 시작된 트랜잭션이 있는 경우 기존 트랜잭션의 nested transaction으로 동작하고 없는 경우 REQUIRED와 동일하게 동작한다.

isolation : 하나의 트랜잭션이 다른 트랜잭션들로부터 격리되는 수준을 정의하는 옵션

DEFAULT : 플랫폼에서 제공하는 기본 격리레벨을 사용한다. 일반적으로 아래의 목록중 READ_COMMITTED를 제공한다.

READ_COMMITTED : 조회 대상 데이터가 다른 트랜잭션에 의해 변경되었으나 커밋되지 않은 상태라면 이를 읽을 수 없다.

READ_UNCOMMITTED : 조회 대상 데이터가 다른 트랜잭션에 의해 변경되고 커밋되지 않은 상태라도 이를 읽을 수 없다.

REPEATABLE_READ : 트랜잭션 구간에서 한번 조회된 데이터는 다른 트랜잭션에 의해 변경되지 않는다. 즉 트랜잭션 구간에서 같은 조건의 데이터를 반복해서 조회해도 항상 같은 데이터를 가져오도록 외부에서의 변경가능성을 차단한다. 그러나 다른 트랜잭션에서 데이터를 추가할 수 는 있다.

SERIALIZABLE : 하나의 트랜잭션에서 조회된 데이터들은 독점적으로 점유되어 다른 트랜잭션에서는 일체의 변경과 추가를 할 수 없도록 한다.

timeout : 트랜잭션의 유효시간으로 지정된 시간동안 트랜잭션이 완료되지 않은 경우 TransactionManager에 의해 Exception발생 후 롤백

readonly : 트랜잭션을 읽기전용을 위한 처리로 지정

rollbackFor : 트랜잭션을 롤백시키는 예외 유형을 예외클래스 목록으로 정의

noRollbackFor : 트랜잭션을 롤백시키지 않고 정상커밋 처리하는 예외 유형을 예외클래스 목록으로 정의

tmDataSource : DataSource TransactionManager를 사용할 때 TM에서 사용할 DataSource Alias를 지정. 온라인 어플리케이션에서는 사용하지 않는다.

다음은 설정 가능한 모든 트랜잭션 옵션을 지정한 어노테이션 예제코드이다.

```
@TransactionalOperation(
        propagation= Propagation.REQUIRED,                                                              //  (1)
        isolation= Isolation.DEFAULT,                                                                   //  (2)
        timeout= 30,                                                                                    //  (3)
        readOnly= false,                                                                                //  (4)
        rollbackFor= { CustomerBizException.class, ProductBizException.class},                          //  (5)
        noRollbackFor= { DasDuplicateKeyException.class, NumberFormatException.class})                  //  (6)
public SSMP1001A001OutDto ssmp1001a001( SSMP1001A001InDto input) throws ApplicationException
{
    ... ...
}
```

(1) propagation을 지정한다. 트랜잭션이 분리되어야 하는 경우 Propagation.REQUIRES_NEW 값을 적용하고 일반적으로 생략된다.

(2) 일반적으로 생략하여 default level을 적용한다.

(3) 트랜잭션 타임아웃을 초단위로 지정하며 특별한 요건이 없는 경우 생략하여 타임아웃을 적용하지 않는다.

(4) 읽기 전용 트랜잭션 속성을 부여한다. 일반적으로 생략하여 false value를 적용한다.

(5) 등록된 타입의 예외가 발생했을 때 자동으로 롤백되도록 한다. CheckedException[4]이 발생하였을 때는 사용자에 의해서 예외처리하는 것이 기본이므로 자동롤백되지 않으나, 위의 예에서처럼 롤백처리하도록 예외유형을 등록하였다면 해당 예외 발생시 프레임워크는 자동으로 롤백처리한다.

(6) 등록된 타입의 예외에 대해서는 자동 롤백처리하지 않는다. RuntimeException과 Error가 발생하였을 때는 자동으로 롤백처리되는 것이 기본동작이나 위의 예에서처럼 롤백처리 예외 유형으로 등록하였다면 해당 예외 발생시 프레임워크는 자동으로 롤백처리하지 않는다.

## 3. 트랜잭션의 커밋과 롤백

사용자에 의해 지정된 Transactional Operation은 트랜잭션 시작이 필요한 경우 프레임워크에 의해 트랜잭션이 시작되고 처리가 완료된 후에는 정상처리 여부와 발생한 예외의 유형에 따라 트랜잭션이 커밋되거나 롤백된다.

### 3.1. Transactional Operation의 트랜잭션이 커밋되는 경우

트랜잭션 구간의 처리가 예외없이 정상적으로 수행되었다.

예외가 발생하였으나 발생한 예외가 CheckedException이다.

UncheckedException이 발생하였으나 트랜잭션 noRollbackFor 속성에 해당 타입이 포함되어 있다.

### 3.2. Transactional Operation의 트랜잭션이 롤백되는 경우

트랜잭션 구간의 처리 중 UncheckedException이 발생하였다.

CkeckedException이 발생하였으나 트랜잭션 rollbackFor 속성에 해당 타입이 포함되어 있다.

CheckedException 유형으로 정의된 DefaultApplicationException이 발생하였다. (설정에 따라 반영)

## 4. 트랜잭션 분리 유형

다음은 처리유형의 서비스-오퍼레이션을 실행하는 전체 트랜잭션 구간중 일부영역에서의 데이터 처리만 트랜잭션이 분리되어 처리되어야 하는 경우 이를 구현하기 위한 설명이다.

다음은 트랜잭션 분리처리가 필요한 업무요건의 예시이다.

고객관리 서비스는 고객 신규등록을 위한 오퍼레이션을 제공한다.

고객 신규등록 오퍼레이션은 다음의 절차를 진행한다.

고객정보를 등록하는 작업자 정보를 입력한다.

고객관리 영역에서 유일하게 관리할 수 있는 고객번호를 채번한다.

채번된 고객번호를 이용하여 고객정보 기본 사항을 입력한다.

고객의 주소정보 목록을 입력한다.

고객의 연락처정보 목록을 입력한다.

다음은 위의 업무요건에 따른 기술적 고려사항이다.

고객등록과정에서 발생한 데이터 입력은 작업의 성공여부에 따라 전체 내용이 DBMS에 반영되거나 취소되어야 한다.

데이터 반영단위에 작업자 정보에 대한 입력이 포함되어야한다.

시스템에서 유일한 채번을 위해 채번처리를 위한 공통 빈(Bean)을 제공하여 사용한다.

채번과정에서 발생한 데이터 입력과 변경은 전체 오퍼레이션이 실패하더라도 개별적으로 DBMS에 반영되는 것이 보장되어야 한다.

다음은 위의 업무 요건을 구현하기 위한 처리 플로우이다.

Figure 4. Flow Example

다음은 위의 기술적 고려사항을 구현하기 위한 지정한 트랜잭션 속성의 예시이다.

```
// 서비스-오퍼레이션
@BxmfServiceOperation( "ssmp1034a021")
@TransactionalOperation                                                                                // (1)
public SSMP1034A021OutDto ssmp1034a021( SSMP1034A021InDto input) throws ApplicationException
{
    ... ...
}

// 채번을 위한 공통 빈 메소드
@TransactionalOperation( propagation= Propagation.REQUIRES_NEW)                                         // (2)
public synchronized BigDecimal makeBizNumber( String firstCategory, String secondCategory)
{
    ... ...
}
```

(1) 서비스-오퍼레이션을 트랜잭션 단위로 처리하기 위해 오퍼레이션 선언위치에 @TransactionalOperation 어노테이션을 지정

(2) 채번을 위한 공통 빈의 채번 메소드에 트랜잭션이 개별적으로 처리되도록 하기 위해 메소드 선언 위치에 @TransactionalOperation 어노테이션과 개별 트랜잭션을 위한 속성(propagation= Propagation.REQUIRES_NEW)를 지정

## 5. Programmatic Transaction Management

프레임워크에서는 메소드 단위로 트랜잭션 속성을 지정하여 트랜잭션과 관련된 세부 처리를 프레임워크에 위임하는 선언적정의 모델과 함께 트랜잭션 API를 이용하여 업무 개발자가 세부 동작을 제어할 수 있는 방법을 제공한다. 일반적으로 트랜잭션과 관련된 동작은 선언적정의 모델의 사용으로 제어가 가능하다. 그러나 데이터처리의 결과에 따른 예외처리가 필요한 경우 API를 이용한 세부제어가 필요할 수 있다.

다음은 트랜잭션 API를 이용한 구현 예를 통해 업무 요건에 따른 트랜잭션의 세부제어를 설명한다. 아래의 예들은 트랜잭션 세부제어를 설명하기 위해 작성된 것으로 실제 업무에 적용하기에는 적합하지 않다.

### 5.1. [Example] 트랜잭션 예외처리 유형 1 - 예외발생시 처리를 변경하여 트랜잭션 커밋, 정상 리턴

아래 유형에서는 다음의 처리 플로우를 구현한다.

이력정보를 입력한다.

연락처 정보를 입력한다.

1과 2가 정상처리되었으면 트랜잭션이 커밋되고 처리결과를 리턴한다.

1, 2 처리 중 예외가 발생한 경우 onException에서 예외 상황을 처리한다.

onException에서는 연락처 정보를 수정한다. 예외 상황에 따라 연락처 정보를 수정하는 것으로 처리를 변경하였으며 트랜잭션이 롤백되지 않는다. 즉 trnasactionalProcess에서 이력정보 입력에는 성공하였을 경우 변경사항은 DBMS에 반영된다.

```
public TESTCONTACT createContactInfo( final TESTCONTACTIo contact)
{
    TransactionContext context= TransactionContextFactory.createJtaTransactionContext( Propagation.REQUIRED); // (1)
    TESTCONTACTIo result= context.execute( new TransactionalProcessor<TESTCONTACTIo>() {                   // (2)

        @Override
        public TESTCONTACTIo transactionalProcess() throws Exception                                          // (3)
        {
            contactExe.insertHistory( contact);
            contactExe.insertContact( contact);
            return contact;
        }

        @Override
        public TESTCONTACTIo onException( TransactionStatus status, Throwable th) throws Throwable            // (4)
        {
            contactExe.updateContact( contact);
            return contact;
        }
    });

    return result;
}
```

(1) 트랜잭션 처리를 위해 TransactionContext를 생성, 생성을 위한 createTransactionContext는 트랜잭션 속성을 부여할 수 있도록 다음의 호출인자를 가지는 메소드들을 제공한다.

public static TransactionContext createJtaTransactionContext( Propagation propagation)

public static TransactionContext createJtaTransactionContext( Propagation propagation, Isolation isolation)

public static TransactionContext createJtaTransactionContext( Propagation propagation, Isolation isolation, boolean readOnly)

public static TransactionContext createJtaTransactionContext( Propagation propagation, Isolation isolation, boolean readOnly, int timeout)

public static TransactionContext createJtaTransactionContext( Propagation propagation, int timeout)

public static TransactionContext createJtaTransactionContext( int timeout)

(2) 트랜잭션 컨텍스트에서 수행할 업무처리 로직을 구현하는 TransactionalProcessor innerClass를 생성

(3) 트랜잭션 컨텍스트에서 수행할 업무처리 로직을 구현

(4) 트랜잭션 컨텍스트에서 수행중인 업무처리 로직(transactionalProcess method body)에서 예외가 발생하였을 때 이를 처리하기 위한 로직을 구현

### 5.2. [Example] 트랜잭션 예외처리 유형 2 - 예외발생시 트랜잭션을 롤백하고 예외를 전달

아래 유형에서는 다음의 처리 플로우를 구현한다.

이력정보를 입력한다.

연락처 정보를 입력한다.

1과 2가 정상처리되었으면 트랜잭션이 커밋되고 처리결과를 리턴한다.

1, 2 처리 중 예외가 발생한 경우 onException에서 예외 상황을 처리한다.

onException에서는 트랜잭션이 롤백되도록 지정하고 예외를 호출자에게 전달한다.

```
public TESTCONTACT createContactInfo( final TESTCONTACTIo contact)
{
    TransactionContext context= TransactionContextFactory.createJtaTransactionContext( Propagation.REQUIRED);
    TESTCONTACTIo result= context.execute( new TransactionalProcessor<TESTCONTACTIo>() {

        @Override
        public TESTCONTACTIo transactionalProcess() throws Exception
        {
            contactExe.insertHistory( contact);
            contactExe.insertContact( contact);
            return contact;
        }

        @Override
        public TESTCONTACTIo onException( TransactionStatus status, Throwable th) throws Throwable
        {
            status.setRollbackOnly();
            throw th;
        }
    });

    return result;
}
```

### 5.3. [Example] 트랜잭션 예외처리 유형 3 - 예외발생시 트랜잭션을 롤백하고 정상응답을 리턴

아래 유형에서는 다음의 처리 플로우를 구현한다.

이력정보를 입력한다.

연락처 정보를 입력한다.

1과 2가 정상처리되었으면 트랜잭션이 커밋되고 처리결과를 리턴한다.

1, 2 처리 중 예외가 발생한 경우 onException에서 예외 상황을 처리한다.

onException에서는 트랜잭션이 롤백되도록 지정하고 정상응답을 리턴하여 정상 플로우가 계속되도록 한다.

```
public TESTCONTACT createContactInfo( final TESTCONTACTIo contact)
{
    TransactionContext context= TransactionContextFactory.createJtaTransactionContext( Propagation.REQUIRED);
    TESTCONTACTIo result= context.execute( new TransactionalProcessor<TESTCONTACTIo>() {

        @Override
        public TESTCONTACTIo transactionalProcess() throws Exception
        {
            contactExe.insertHistory( contact);
            contactExe.insertContact( contact);
            return contact;
        }

        @Override
        public TESTCONTACTIo onException( TransactionStatus status, Throwable th) throws Throwable
        {
            status.setRollbackOnly();
            return contact;
        }
    });

    return result;
}
```

하나의 트랜잭션 구간에서 사용되는 모든 DBIO가 같은 DataSource를 사용할 때에는 미들웨어에서 제공하는 JTA Transaction 관리 API를 사용하지 않고 DataSource TransactionManager를 사용할 수 있다.

DataSource TransactionManager는 트랜잭션 구간에서 사용하는 모든 DB Access에 같은 세션을 전달함으로써 트랜잭션을 보장하는 방법을 구현한다. 이러한 방법은 데이터 처리에서 사용하는 DataSource가 모두 동일하다고 가정할 때 JTA TransactionManager를 사용하는 것보다 우수한 성능을 보장한다. 그러나 트랜잭션 속성을 부여하는 위치에서는 트랜잭션 구간에서 호출되는 모든 DBIO가 같은 DataSource를 사용한다는 명확한 보장을 할 수 없으며 어플리케이션의 변경에 유연하게 대응할 수 없으므로 개발표준에서는 JTA TransactionManager를 사용하도록 정하고있다.

작성자가 해당 메소드를 호출하였을 때 처리해야하는 예외를 정의하여 메소드의 throws keyword를 통해 지정한다. FileNotFoundException은 잘 알려진 CheckedException으로 메소드를 호출하는 위치에서 해당예외를 처리하지 않은 경우 컴파일러에 의해 오류가 발생하므로 try, catch를 이용한 예외처리가 강제된다. 일반적으로 사용자가 정의하는 사용자 정의 예외는 Exception을 상속받는 CheckedException으로 정의한다.

메소드를 호출하여 사용할 때 사용자가 try, catch 구문을 이용하여 예외처리를 하였는지 컴파일러가 확인하지 않는 예외 및 에러 유형이다. RuntimeException, Error, RuntimeException의 상속 예외, Error의 상속에러 등을 포함한다.