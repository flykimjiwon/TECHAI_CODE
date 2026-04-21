# Framework Config 파일

BX Framework는 동작에 필요한 기본 설정 파일들을 xml및 properties 파일로 관리를 하고 있다. Config 파일에는 System Property, DataSource, Loader, Context, Data-Acces, Log 와 관련된 내용이 정의가 되어 있다.

## 1. Framework 인스턴스 설정 파일

인스턴스 설정화일을 이용하여 설정하는 항목은 아래와 같다.

Framework runtime이 사용할 DataBase Connection

Framework classloader 환경설정

Endpoint, ServiceExecutor에 대한 환경설정

DBIO를 수행하는 das context에 대한 환경설정

메시지코드를 메시지로 변환하는 message source에 대한 환경설정

### 1.1. 인스턴스 설정 항목

설정 파일의 각 주요 항목에 대하여 설명한다.

BX Framework instance 설정 - environment

| 항목 
| 설명 
| 비고 

system-mode

현재 프레임워크의 Mode 값 설정. (D : 개발, T : 테스트, O : 운영)

|  

datasource

프레임워크에서 사용하는 DB Data Source 정의한다. jndi-name 속성에 WAS의 datasource 명을 설정한다.

|  

BX Framework instance 설정 - environment > datasource > jdbc-datasource

| 항목 
| 설명 
| 비고 

driver-classname

JDBC 드라이버 클래스이름(fully qualified Java class name)

|  

uri

JDBC 드라이버에 전달할 connection url

|  

username

JDBC 드라이버에 전달할 데이터베이스 사용자이름

|  

password

JDBC 드라이버에 전달할 데이터베이스 password

|  

enc-password

JDBC 드라이버에 전달할 데이터베이스 password(암호화된 형태)

|  

maxTotal

동시에 사용할 수 있는 최대 커넥션 개수(기본값: 8)

|  

initialiSize

Connection Pool을 처음 만들 때 생성할 Connection의 갯수

|  

maxIdle

커넥션 풀에 반납할 때 최대로 유지될 수 있는 커넥션 개수(기본값: 8)

|  

validationQuery

connection에 대한 유효성을 검사할 때 사용하는 SQL. 보통의 경우 Pool에서 Connection을 꺼낼 때 유효성 검사를 수행한다. SQL은 적어도 1개이상 row를 반환하는 SELECT문이어야 한다.

|  

validationQueryTimeout

Connection 유효성 검사 SQL 수행시간에 대한 timeout. 0보다 큰 값이 지정되면 JDBC 드라이버의 setQueryTimeout메소드를 통해 타임아웃 값을 전달한다.

|  

BX Framework instance 설정 - environment > system-properties

| 항목 
| 설명 
| 비고 

batch.node.no

배치인스턴스가 여러 노드에서 구동되는 경우 노드를 구분하는 번호.(일반배치에서만 사용)

|  

beantype.usemetadata

getBean메소드 최적화를 위해 Spring Metadata를 사용하지 않고 프레임워크의 metadata를 사용한다. 디폴트값은 false

온라인 인스턴스의 경우 true로 설정한다.

|  

accrue.data.accesstime

요청서비스에서 수행한 DBIO수행시간을 Context에 누적하여 저장할 지 여부. 디폴트값은 false

|  

accrue.beanfactory.accesstime

요청서비스에서 수행한 beanfactory.getBean수행시간을 Context에 누적하여 저장할 지 여부. 디폴트값은 false

|  

admin.system.main.key

웹어드민을 실행한 프레임워크 인스턴스의 시스템ID

|  

admin.use.trx.cd

거래코드 사용여부

|  

admin.image.log.system.header

이미지로그조회시 전문해석을 위해 사용할 시스템헤더클래스

|  

admin.batch.type.no.select

배치작업정보 등록시 배치작업유형의 선택가능 여부

|  

file.upload.dir

화일 업로드시 업로드한 화일을 임시 저장할 디렉토리

|  

bxm.dbio.max.resultset.size

DBIO 수행 시 max record 크기 값 설정. max 값 초과할 경우 에러, 경고 처리

|  

bxm.dbio.connected.batch.max.resultset.size

CONNECTED_BATCH 로 DBIO 수행 시 max record 크기 값 설정. Max 값 초과할 경우 에러, 경고 처리

|  

bxm.dbio.max.resultset.size,
bxm.dbio.connected.batch.max.resultset.size사용시 주의사항

resultset 크기를 계산하기위해서 Predictable.predictMessageLength()를 호출하여 크기를 계산한다. predictMessageLength메소드는 OMM의 length필드 값을 사용하여 길이계산을 한다. 즉, 실제DTO객체크기를 계산하는 것이 아니고 OMM length필드값을 사용하여 계산하므로 사용시 주의해야 한다.

BX Framework instance 설정 - environment > loader

| 항목 
| 설명 
| 비고 

mode

프레임워크의 classloader 모드를 정의.

hot : 변경된 Class를 로드한 후 즉시 어플리케이션에 반영한다. 주로 개발시점에 사용한다.

tepid : 변경된 Class를 로드한 후 신규 어플리케이션을 생성한다. 기존 어플리케이션, 신규 어플리케이션 2개를 유지하며 반영명령에 따라 신규어플리케이션을 기존어플리케이션을 대체하는 모드이다. 주로 운영환경에서 사용한다.

cold : 변경된 Class를 로드하지 않는 모드이다. 변경된 Class를 반영할려면 JVM 인스턴스를 재시작해야 한다. 변경된 클래스 반영을 위한 클래스의 중복 로딩을 하지 않으므로 메모리 사용량이 가장 적다.

|  

autodeploy

자동 Deploy 여부(Hot/Tepid 모드이면 true, Cold 이면 false로 설정)

|  

interval-millis

Hot Deploy 감지 Interval Time

|  

lazy-init

Bean 생성을 거래가 호출 시 해당 거래에 해당하는 Bean들을 생성하는 옵션 여부

|  

deferred-init

초기 Application Loading후에 Bean 생성을 Deferred(후처리)로 처리 여부

|  

deferred-init-interval

Deferred(후처리)로 Bean 생성 Interval 주기

|  

application-home

Framework에서 deploy할 Application이 위치하는 Home 디렉토리

|  

BX Framework instance 설정 - environment > scanner-support

| 항목 
| 설명 
| 비고 

candidate-addition.resource-pattern

bean candidate로 추가할 인터페이스에 대한 정규 표현식

|  

candidate-addition.resource-addition

bean candidate로 추가할 인터페이스를 구현한 Bean class에 대한 정규 표현식

|  

BX Framework instance 설정 - on-demand-batch

| 항목 
| 설명 
| 비고 

init-workers

Ondemand 배치 처리 쓰레드 수 초기값

|  

max-workers

Ondemand 배치 처리 쓰레드 수 최대값

|  

worker-wait-timeout

유휴상태의 Ondemand 배치 처리 쓰레드가 없을 때 대기하는 시간. (단위 ms)

|  

logging-home

Ondemand 배치 처리시 로그를 남기는 디렉토리. Ondemand배치 요청마다 별개의 로그화일을 생성한다.

|  

BX Framework instance 설정 - context

| 항목 
| 설명 
| 비고 

forced-timeout

서비스 수행 시간에 대한 Timeout 설정 값(ms) 정의.

|  

preprocessor

시스템선처리 클래스를 지정한다.

|  

postprocessor

시스템후처리 클래스를 지정한다.

|  

transaction-synchronization

Commit 전/후 에 Transaction과 관련하여 처리할 Class 정의

|  

transaction-rollbackfors

Exception발생시 Transaction을 롤백할 Exception을 지정한다.

(지정하지 않으면 RuntimeException이 발생했을때만 롤백한다.)

|  

BX Framework instance 설정 - context > request

| 항목 
| 설명 
| 비고 

header-classname

시스템헤더 클래스를 지정한다.

|  

imagelogging

이미지로그를 저장할 지 여부를 지정한다.

|  

request-resolver-classname

요청전문을 서비스입력객체로 변환작업을 수행할 클래스 정의

|  

response-resolver-classname

서비스응답객체를 응답전문으로 변환작업을 수행할 클래스 정의

|  

request-trace-classname

이미지로그를 저장하는 작업을 처리하는 Class 정의

|  

interceptor-classname

ServiceEndpoint에서 서비스를 처리하기 전/후 혹은 응답전문을 보내기 직전에 처리할 작업을 수행하는 Class 정의

|  

service-executor-interceptor-classname

ServiceExecutor에서 서비스를 호출하기 전/후에 처리할 작업을 수행하는 Class 정의

|  

BX Framework instance 설정 - data-access

| 항목 
| 설명 
| 비고 

trace-inject

SQL에 GUID 삽입 여부

GUID를 SQL에 삽입하게 되면 SQL이 실행시마다 달라져서 DB성능이 나빠질 수 있으므로 사용하지 않는 것을 권장한다.

|  

monitor-inject

SQL에 Monitor Hint 삽입여부. bxmAdmin 1.X을 위한 옵션이므로 사용하지 않는 것을 권장한다.

|  

identify-inject

SQL에 SQLID(DBIO Class.Method) 삽입여부

|  

identifier-prefix

identify-inject시 삽입되는 문자열의 prefix로 사용할 문자열 정의 (기본값은 "/*")

|  

identifier-suffix

identify-inject시 삽입되는 문자열의 suffix로 사용할 문자열 정의 (기본값은 "*/")

|  

oracle-setmodule

DB session에 module_name, action_name을 지정하기 위하여 DBMS_APPLICATION_INFO.SET_MODULE procedure을 호출할지 여부. module_name, action_name의 값은 Context객체의 module, action 값을 사용한다.

|  

oracle-clearmodule

oracle-setmodule 옵션에 의해 설정된 module_name, action_name을 SQL 수행후 null로 초기화 할 지 여부

|  

max-resultset

DBIO를 이용하여 조회할 수 있는 최대 Result 건수 정의

|  

limit-exceeded-action

DBIO를 이용하여 조회한 건수가 max-resultset에 지정된 값보다 많을 때 수행할 Action 정의

Warn : 경고메시지를 log로 출력한다.

Error : 즉시 LimitExceededResultsException을 발생시킨다.

|  

fetch-size

기본 Fetch 크기값 정의

DBIO SQL의 fechSize가 존재하지 않으면 fetch-size를 fetchSize값으로 사용한다.

|  

min-fetch-size

최소 Fetch 크기값 정의

DBIO SQL의 fechSize가 있으나 min-fetch-size보다 작은 경우 min-fetch-size를 fetchSize값으로 사용한다.

|  

forced-logging

SQL, Resultset을 log로 강제 출력 여부 설정

true인 경우 bxm.context.das.logging에 대한 로그레벨과 상관없이 log로 SQL, ResultSet을 출력한다.

|  

jndi-datasource-aliases > alias

Application에 사용하는 WAS의 Datasource의 Alias 정의한다. bxm-application.xml의 Datasource 이름이 WAS Datasource의 JNDI이름과 다를 때 설정하여 사용할 수 있다.

name : WAS에 등록된 Datasource의 JNDI이름

alias : bxm-application.xml의 Datasource 이름

|  

oracle-setmodule, oracle-clearmodule 사용시 주의사항

SQL수행전/후에 SET_MODULE procedure를 호출하기 위하여 Callable Statement를 호출하므로 성능이 나빠질 수 있다.

Oracle Database의 procedure를 사용하므로 Oracle Database에서만 사용가능하다.

BX Framework instance 설정 - message

| 항목 
| 설명 
| 비고 

message-source-classname

메시지코드에 대한 메시지 정보를 해석하는 클래스를 지정

|  

BX Framework instance 설정 - deferred-service-executor > deferred-services > deferred-service

| 항목 
| 설명 
| 비고 

application

Deferred Service가 위치한 어플리케이션

|  

bean

Deferred Service를 구현한 Bean

|  

max-workerpool-size

Deferred Service를 수행할 쓰레드의 최대 갯수

|  

max-queue-size

Deferred Service를 수행할 쓰레드가 모두 수행중이면 대기할 Queue의 크기

|  

rake-interval

Deferred Service를 수행한 후 다시 수행할 때까지의 간격

|  

nodata-wait

Deferred Service 수행시 입력데이터가 없는 경우 대기할 지 여부. (Default : false)

|  

shutdown-mode

Framework 인스턴스 종료시 현재 수행중인 작업을 어떻게 처리할 것인지 동작방식

PROCESSING_QUEUED_AND_RUNNING_TASK : 현재 수행중인 작업 및 Queue에서 대기중인 작업을 모두 완료한 후 Deferred service executor를 종료한다.

PROCESSING_RUNNING_TASK : 현재 수행중인 작업을 모두 완료한 후 Deferred service executor를 종료한다.

IGNORE_QUEUED_RUNNING_TASK : 현재 수행중인 작업 및 Queue에서 대기중인 작업을 무시하고 곧바로 Deferred service executor를 종료한다.

|  

BX Framework instance 설정 - batch-context

| 항목 
| 설명 
| 비고 

repos-tran

배치 수행 중 commit 시 Step 수행 정보 Update시 Framework DataSource로 기록할지, Application DataSource로 기록 할지에 대한 옵션 (Default 값 true)

true : Application DataSource로 같은 Transaction으로 Step 수행 정보Update 처리 (재실행 시 정합성 보장, F_BATCH 로 시작하는 테이블에 대하여 Application DB 계정으로 시노님설정 필요)

false : Framework DataSource로 Step 수행 정보 Update 처리 (재실행 시 정합성 보장 할 수 없음)

|  

commit-interval

배치 Step 수행 시 Default로 처리할 commit-interval 값

version 3.0이후부터 배치작업 xml에서 commit-interval 속성의 정의하지 않아도 정상 동작함. step별 commit-interval 값을 배치작업 정보에 등록할 수도 있다.

|  

batch-control-param

배치 작업 제어 파라미터 (일반배치, 온디맨드배치) 클래스

|  

daemon-control-param

데몬 배치 제어 파라미터 클래스

|  

batch-preprocessor

배치 선처리 클래스. 배치선처리는 배치 Context가 생성이 되기 전에 호출됨.

|  

batch-context-preprocessor

배치 컨텍스트 선처리 클래스. 배치컨텍스트선처리는 배치 Context가 생성이 된 후 호출됨

|  

batch-postprocessor

배치 후처리 클래스. 배치후처리는 배치 수행이 완료(정상, 에러 등..)후에 호출 됨

|  

batch-batchgroup

배치 그룹제어 클래스. 배치그룹제어 프로세서는 노드, 시스템 별 배치 수행 개수에 따라 배치 제어처리.

|  

job-listener

Spring Batch의 JobExecutionListener를 등록

|  

step-listener

Spring Batch의 StepExecutionListener 를 등록.

|  

### 1.2. 온라인 인스턴스 Framework 설정화일

온라인 인스턴스 설정 파일에 대한 예제이다. 사용자환경에 맞게 경로 및 값을 수정하여 사용한다.

```
<?xml version="1.0" encoding="UTF-8"?>
<bxm-instance xmlns="http://www.bankwareglobal.com/schema/instance" xmlns:cn="http://www.bankwareglobal.com/schema/common" name="bxm-instance">
    <description> Bxm Online Configuration</description>
    <environment>
        <system-properties>
            beantype.usemetadata=true
            accrue.data.accesstime=true
            accrue.beanfactory.accesstime=true
        </system-properties>
        <! -- 개발 : D, 운영 : O, 테스트 : T -->
        <system-mode>D</system-mode>
        <datasource>
             <jndi-datasource jndi-name="java:/comp/env/BXMNXA" />
        </datasource>
        <loader mode="hot" autodeploy="true" interval-millis="10000" lazy-init="true"
        deferred-init="true" deferred-init-interval="1000" sessionfactory-lazyinit="true">
            <application-home><<BXM_HOME>>/apps/online</application-home>
        </loader>
    </environment>
    <management>
        <connector serviceurl="service:jmx:rmi:///jndi/rmi://localhost:20001/jmxrmi">
            <environment>
            </environment>
        </connector>
        <objectname domain="JRF">
                <properties>
                        Name=bxm-management
                </properties>
        </objectname>
    </management>

    <context forced-timeout="120000">
        <preprocessor classname="bxm.dft.service.processor.DefaultSystemPreProcessor" order="1"/>
        <postprocessor classname="bxm.dft.service.processor.DefaultSystemPostProcessor" order="1"/>
        <control-parameter classname="bxm.dft.context.control.impl.DefaultControlParametersImpl"/>
        <transaction-rollbackfors>
                bxm.app.ApplicationException
                bxm.dft.app.DefaultApplicationException
        </transaction-rollbackfors>
        <request header-classname="bxm.dft.context.DefaultSystemHeader" imagelogging="true"
                request-resolver-classname="bxm.dft.request.DefaultRequestResolver"
                response-resolver-classname="bxm.dft.request.DefaultResponseResolver"
                request-trace-classname="bxm.dft.request.DefaultRequestTrace"
                interceptor-classname="bxm.dft.service.endpoint.DefaultRequestInterceptor"
                service-executor-interceptor-classname="bxm.dft.service.DefaultServiceExecutorInterceptor"
        />
    </context>

    <!-- 개발 (max-resultset 값은 프로젝트 설정에 따라 정의) -->
    <data-access trace-inject="false" monitor-inject="false" identify-inject="true"
                max-resultset="1000" fetch-size="10"
                min-fetch-size="10"  limit-exceeded-action="Warning" forced-logging="false">
            <jndi-datasource-aliases>
            <!-- name : J2EE Application Service dataSource jndiName for applications -->
            <!-- alias : jndi-datasource, jndi-xadatasource value
                     in application configuration file( bxm-application.xml) -->
            <alias name="java:/comp/env/APPNXA" alias="DSNXA"/>
            <alias name="java:/comp/env/APPXA" alias="DSXA"/>
        </jndi-datasource-aliases>
    </data-access>

    <message message-source-classname="bxm.instance.message.DataBaseMessageSource"/>
</bxm-instance>
```

### 1.3. 일반배치, 데몬배치 인스턴스 Framework 설정화일

일반배치, 데몬배치 인스턴스 설정시 다음 설정을 주의하여 설정한다.

loader의 mode를 cold로 사용한다.

scanner-support, batch-context 설정을 해줘야 한다.

context설정을 해줘야 하나 설정값은 무시한다.(context설정이 없으면 배치 시작시 에러가 발생한다.)

일반배치, 데몬배치 인스턴스 설정 파일에 대한 예제이다. 사용자환경에 맞게 경로 및 값을 수정하여 사용한다.

```
<?xml version="1.0" encoding="UTF-8"?>
<bxm-instance xmlns="http://www.bankwareglobal.com/schema/instance"
    xmlns:cn="http://www.bankwareglobal.com/schema/common" name="bxm-batch">
    <description>BXM Sample Configuration</description>
    <environment>
        <system-properties>
            batch.node.no=1
        </system-properties>
        <system-mode>D</system-mode>
        <datasource>
            <jdbc-datasource
                driver-classname="oracle.jdbc.OracleDriver"
                uri="jdbc:oracle:thin:@xxx.xxx.xx.xxx:1521:ORA"
                username="bxm"
                password="xxx"
                maxActive="10"/>
        </datasource>
        <loader mode="cold" autodeploy="false" lazy-init="true" useonly-predef-beanprocs="true" nouse-snapshot="false" registry="off">
            <application-home>/bxm/apps/batch</application-home>
        </loader>
        <scanner-support>
            <candidate-addition resource-pattern="(.*)(.share.)(.*)(.I)([a-zA-Z_$][a-zA-Z\\d_$]*)"
                resource-addition="$1.online.$3.$5"/>
            <candidate-addition resource-pattern="(.*)(.share.)(.*)(.I)([a-zA-Z_$][a-zA-Z\\d_$]*)"
                resource-addition="$1.batch.$3.$5"/>
        </scanner-support>
    </environment>
    <management>
        <objectname domain="JRF">
            <properties>
                Name=bxm-batch
            </properties>
        </objectname>
    </management>
    <context>
        <control-parameter classname="bxm.instance.context.control.SimpleControlParameters"/>
        <transaction-rollbackfors>
            bxm.app.ApplicationException
        </transaction-rollbackfors>
        <request header-classname="bxm.dft.context.DefaultSystemHeader" imagelogging="true"
            request-resolver-classname="bxm.request.DefaultXmlRequestResolver"
            response-resolver-classname="bxm.request.DefaultXmlResponseResolver"
            request-trace-classname="bxm.request.DefaultRequestTrace"/>
    </context>
    <data-access trace-inject="false" monitor-inject="false" identify-inject="true"
        max-resultset="500" fetch-size="100" min-fetch-size="100"
        limit-exceeded-action="Error" forced-logging="true">
    </data-access>
    <message message-source-classname="bxm.instance.message.DataBaseMessageSource"/>

    <batch-context repos-tran="false">
        <batch-control-param classname="bxm.batch.dft.control.DefaultBatchControlParameter"/>
        <daemon-control-param classname="bxm.batch.dft.control.DefaultDaemonControlParameter"/>
        <batch-preprocessor classname="bxm.batch.dft.context.DefaultBatchPreProcessor"/>
        <batch-context-preprocessor classname="bxm.batch.dft.context.DefaultBatchContextPreProcessor"/>
        <batch-postprocessor classname="bxm.batch.dft.context.DefaultBatchPostProcessor"/>
        <batch-batchgroup classname="bxm.batch.dft.control.DefaultBatchGroupCheckProcessor"/>
    </batch-context>
</bxm-instance>
```

### 1.4. Ondemand배치 인스턴스 Framework 설정화일

Ondemand배치 인스턴스 설정시 다음 설정을 주의하여 설정한다.

loader의 mode를 cold로 사용한다.

on-demand-batch, batch-context 설정을 해줘야 한다.

context설정을 해줘야 하나 설정값은 무시한다.(context설정이 없으면 배치 시작시 에러가 발생한다.)

Ondemand배치 인스턴스 설정 파일에 대한 예제이다. 사용자환경에 맞게 경로 및 값을 수정하여 사용한다.

```
<?xml version="1.0" encoding="UTF-8"?>
<bxm-instance xmlns="http://www.bankwareglobal.com/schema/instance"
        xmlns:cn="http://www.bankwareglobal.com/schema/common" name="bxm-batch">
    <description>BXM Ondemand Batch Sample Configuration</description>
    <environment>
        <system-properties>
        </system-properties>
        <datasource>
            <jdbc-datasource
                driver-classname="oracle.jdbc.OracleDriver"
                uri="jdbc:oracle:thin:@xxx.xxx.xx.xxx:1521:ORA"
                username="bxm"
                password="xxx"
                maxActive="5"/>
        </datasource>
        <loader mode="cold" autodeploy="false">
            <application-home>/bxm/apps/batch</application-home>
        </loader>
    </environment>
    <on-demand-batch
        init-workers="10" max-workers="30" worker-wait-timeout="2000"
        logging-home="/bxm/logs/bxm/batch"/>
    <management>
        <connector serviceurl="service:jmx:rmi:///jndi/rmi://localhost:20002/jmxrmi">
        </connector>
        <objectname domain="JRF">
            <properties>
                Name=bxm-batch
            </properties>
        </objectname>
    </management>

    <context>
        <control-parameter classname="bxm.dft.context.control.impl.DefaultControlParametersImpl"/>
        <transaction-rollbackfors>
            bxm.app.ApplicationException
            bxm.dft.app.DefaultApplicationException
        </transaction-rollbackfors>
        <request header-classname="bxm.dft.context.DefaultSystemHeader" imagelogging="true"
            request-resolver-classname="bxm.request.DefaultXmlRequestResolver"
            response-resolver-classname="bxm.request.DefaultXmlResponseResolver"
            request-trace-classname="bxm.dft.request.DefaultRequestTrace"
            interceptor-classname="bxm.dft.service.endpoint.DefaultRequestInterceptor"
            service-executor-interceptor-classname="bxm.dft.service.DefaultServiceExecutorInterceptor"
        />
    </context>

    <data-access trace-inject="false" monitor-inject="false" identify-inject="true"
        fetch-size="100">
    </data-access>

    <message message-source-classname="bxm.instance.message.DataBaseMessageSource"/>

    <batch-context batch-extend="true">
        <batch-control-param classname="bxm.batch.dft.control.DefaultBatchControlParameter"/>
        <batch-preprocessor classname="bxm.batch.dft.context.DefaultBatchPreProcessor"/>
        <batch-context-preprocessor classname="bxm.batch.dft.context.DefaultBatchContextPreProcessor"/>
        <batch-postprocessor classname="bxm.batch.dft.context.DefaultTargetBatchPostProcessor"/>
        <batch-batchgroup classname="bxm.batch.dft.control.DefaultBatchGroupCheckProcessor"/>
    </batch-context>

</bxm-instance>
```

### 1.5. admin 인스턴스 Framework 설정화일

admin 인스턴스 설정시 다음 설정을 주의하여 설정한다.

DBIO Dependency 분석을 위한 Deferred service executor를 설정. 단, 개발툴에서 DBIO Dependency 분석을 하는 경우는 설정하지 않아도 된다.

admin 인스턴스 설정 파일에 대한 예제이다. 사용자환경에 맞게 경로 및 값을 수정하여 사용한다.

```
<?xml version="1.0" encoding="UTF-8"?>
<bxm-instance xmlns="http://www.bankwareglobal.com/schema/instance" xmlns:cn="http://www.bankwareglobal.com/schema/common" name="bxm-instance">
    <description>Bxm Admin</description>
    <environment>
        <system-properties>
            beantype.usemetadata=true
            accrue.data.accesstime=true
            accrue.beanfactory.accesstime=true
            admin.system.main.key=AC
            admin.use.trx.cd=true
            admin.image.log.system.header=bxm.dft.context.DefaultSystemHeader
            admin.batch.type.classification=true
            file.upload.dir=/bxm/excel_upload
        </system-properties>
        <system-mode>D</system-mode>
        <datasource>
            <jndi-datasource jndi-name="java:/comp/env/BXMNXA" />
        </datasource>
        <loader mode="hot" autodeploy="true"
               interval-millis="8000" lazy-init="true"
               deferred-init="true" deferred-init-interval="1000"
               sessionfactory-lazyinit="true"
               use-lastmodified-cache="true">
            <application-home>/bxm/apps/admin</application-home>
        </loader>
    </environment>

    <management>
        <connector serviceurl="service:jmx:rmi:///jndi/rmi://localhost:20003/jmxrmi">
        </connector>
        <objectname domain="JRF">
            <properties>
                Name=bxm-management
            </properties>
        </objectname>
    </management>

    <context forced-timeout="120000">
        <control-parameter classname="bxm.dft.context.control.SimpleControlParameters"/>
        <transaction-rollbackfors>
            bxm.app.ApplicationException
        </transaction-rollbackfors>
        <request header-classname="bxm.dft.context.SimpleSystemHeader" imagelogging="false"
              request-resolver-classname="bxm.dft.request.DefaultRequestResolver"
              response-resolver-classname="bxm.dft.request.DefaultResponseResolver"
              request-trace-classname="bxm.dft.request.DefaultRequestTrace" />
    </context>

    <data-access
        max-resultset="10000" fetch-size="10" min-fetch-size="100"
        limit-exceeded-action="Error" forced-logging="true">
        <jndi-datasource-aliases>
            <!-- name : J2EE Application Service dataSource jndiName for applications -->
            <!-- alias : jndi-datasource, jndi-xadatasource value in application configuration file( bxm-application.xml) -->
            <alias name="java:/comp/env/APPNXA" alias="DSNXA"/>
            <alias name="java:/comp/env/APPXA" alias="DSXA"/>
        </jndi-datasource-aliases>
    </data-access>

    <message message-source-classname="bxm.instance.message.DataBaseMessageSource"/>
    <deferred-service-executor>
        <deferred-services>
          <deferred-service
            application="bxmAdmin"
             bean="QueryAnalyzeBean"
             max-workerpool-size="10"
             max-queue-size="10"
             rake-interval="180000"
             nodata-wait="true"
             shutdown-mode="PROCESSING_QUEUED_AND_RUNNING_TASK"/>
    </deferred-service-executor>
</bxm-instance>
```

## 2. logback 설정 파일(logback.xml)

BX Framework는 로그 출력을 위하여 로그 라이브러리로 logback를 사용한다. BX Framework의 로그 출력을 위해 설정 항목을 설명한다.

본 문서에서는 BX Framework을 사용하는 데 필요한 항목만을 설명하며 logback의 자세한 설명은 logback홈페이지(logback.qos.ch/)를 참고한다.

### 2.1. Appender

BX Framework이 로그 출력을 위하여 사용하는 Appender를 설명한다.

BX Framework에서 사용하는 일반적인 Appender설정은 아래와 같다.

```
    <appender name="applogfile" class="ch.qos.logback.core.rolling.RollingFileAppender">
        <file>
            ${LOGS_ABSOLUTE_PATH}/online/app_${bxm.node.name}_${bxm.instance.name}.log   (1)
        </file>
        <encoder>
            <pattern>
            [%d{yyyy-MM-dd HH:mm:ss.SSS}][%X{GUID}][%-5level][%logger{36}:%line]-%msg%n    (2)
            </pattern>
        </encoder>
        <rollingPolicy class="ch.qos.logback.core.rolling.TimeBasedRollingPolicy">
            <fileNamePattern>${LOGS_ABSOLUTE_PATH}/online/app_${bxm.node.name}_${bxm.instance.name}.%d{yyyy-MM-dd}.log</fileNamePattern>
            <maxHistory>30</maxHistory>
            <totalSizeCap>3GB</totalSizeCap>
        </rollingPolicy>
    </appender>
```

| 1 
JVM 환경변수

${bxm.node.name},${bxm.instance.name}는 JVM의 환경변수값을 가져와 logback.xml에서 사용한다.

${bxm.node.name}의 경우 -Dbxm.node.name=XXX와 같이 JVM시작시 환경변수로 지정되어 있어야 한다.

| 2 
Log format

BX Framework에서 사용하는 일반적인 log format은 아래와 같다.

BX Framework log format

| 항목 
| 설명 
| 비고 

로그시각

로그이벤트가 발생한 시각

ex) %d\{yy/MM/dd HH:mm:ss,SSS}

|  

요청GUID

요청전문의 GUID

logback MDC를 사용하여 요청전문의 GUID를 저장한 후 사용해야 한다. %X{항목명}은 logback MDC에서 항목값을 꺼내는 표기법이다.

ex) %X{GUID}

|  

로그레벨

출력중인 로그의 로그레벨

ex) %p

|  

소스 위치

로그출력이 발생한 %F는 소스파일명, %L은 소스파일내 라인을 나타낸다.

ex) %F:%L

|  

로그 메시지

로그이벤트에서 출력하는 로그 메시지 (%n은 개행문자)

ex) %m%n

|  

DB Appender

BX Framework에서 로그를 DB로 출력하기 위해서 만든 Appender이다. DB에 로그 저장할 때 비동기방식으로 저장하여 성능을 향상시킨다.

DB Appender 사용시 주의사항

로그 저장시 비동기방식으로 저장하여도 저장하는 로그량이 많은 경우 어플리케이션의 수행속도를 느리게 만들 수 있다.

```
    <appender name="applogdb" class="bxm.dft.logging.logback.DefaultAsyncTableAppender">
        <bufferSize>5</bufferSize>   (1)
        <connectionSource class="ch.qos.logback.core.db.JNDIConnectionSource">
             <jndiLocation>java:/comp/env/BXMNXA</jndiLocation>      (2)
        </connectionSource>
        <pattern>[%-5level][%d{yyyy-MM-dd HH:mm:ss.SSS}][%logger{36}:%line] - %msg%n</pattern>
    </appender>
```

| 1 
BufferSize

로그를 DB로 저장하기전에 로그정보를 보관하는 Buffer의 크기이다. BufferSize만큼 차지 않으면 로그가 DB 에 저장되지 않으므로 적절한 값을 설정할 필요가 있다.

| 2 
JndiName

로그를 DB로 저장할때 사용하는 Datasource의 jndi name이다.

### 2.2. 온라인 인스턴스를 위한 logback.xml 설정

온라인 인스턴스 로그 출력을 위하여 다음과 같은 Appender가 필요하다.

온라인 인스턴스 logback appender name설명

| 항목 
| 설명 
| 비고 

console

Console 로 출력처리하기 위한 Appender로서 보통의 경우 WAS로그에 저장된다.

|  

applogdb

서비스로그를 DB에 저장처리하기 위한 Appender(운영서버에서 서비스로그를 남기지 않으려면 해당 설정 제거)

|  

applogfile

서비스로그를 파일로 처리하기 위한 Appender

|  

온라인 인스턴스를 위한 logback.xml이다. 사용자환경에 맞게 경로 및 appender를 수정하여 사용한다.

```
<?xml version="1.0" encoding="UTF-8"?>
<configuration scan="false">

    <property name="LOGS_ABSOLUTE_PATH" value="/bxm/logs" />

    <appender name="console" class="ch.qos.logback.core.ConsoleAppender">
        <encoder>
            <pattern>[%d{yyyy-MM-dd HH:mm:ss.SSS}][%-5level][%logger{36}:%line][%thread] - %msg%n</pattern>
        </encoder>
    </appender>

    <appender name="applogfile" class="ch.qos.logback.core.rolling.RollingFileAppender">
        <file>${LOGS_ABSOLUTE_PATH}/online/app_${bxm.node.name}_${bxm.instance.name}.log</file>
        <encoder>
            <pattern>[%d{yyyy-MM-dd HH:mm:ss.SSS}][%X{GUID}][%-5level][%logger{36}:%line] - %msg%n</pattern>
        </encoder>
        <rollingPolicy class="ch.qos.logback.core.rolling.TimeBasedRollingPolicy">
            <fileNamePattern>${LOGS_ABSOLUTE_PATH}/online/app_${bxm.node.name}_${bxm.instance.name}.%d{yyyy-MM-dd}.log</fileNamePattern>
            <maxHistory>30</maxHistory>
            <totalSizeCap>3GB</totalSizeCap>
        </rollingPolicy>
    </appender>

    <appender name="applogdb" class="bxm.dft.logging.logback.DefaultAsyncTableAppender">
        <connectionSource class="ch.qos.logback.core.db.JNDIConnectionSource">
             <jndiLocation>java:/comp/env/BXMNXA</jndiLocation>
        </connectionSource>
        <bufferSize>5</bufferSize>
        <pattern>[%-5level][%d{yyyy-MM-dd HH:mm:ss.SSS}][%logger{36}:%line] - %msg%n</pattern>
    </appender>

    <root level="debug">
        <appender-ref ref="console" />
        <appender-ref ref="applogfile" />
        <appender-ref ref="applogdb" />
    </root>

</configuration>
```

### 2.3. 일반 배치 인스턴스를 위한 logback.xml 설정

일반배치 인스턴스 로그 출력을 위하여 다음과 같은 Appender가 필요하다.

일반배치 인스턴스 logback appender name설명

| 항목 
| 설명 
| 비고 

CONSOLE

배치로그를 표준출력으로 출력처리하기 위한 Appender로서 보통의 경우 배치작업스케줄러에서 표준출력으로 출력된 로그를 읽어 사용한다.

|  

FILE

배치로그를 파일로 저장하기 위한 Appender

|  

일반배치 인스턴스를 위한 logback.xml이다. 사용자환경에 맞게 경로 및 appender를 수정하여 사용한다.

```
<?xml version="1.0" encoding="UTF-8"?>
<configuration scan="false">

    <appender name="CONSOLE" class="ch.qos.logback.core.ConsoleAppender">
        <encoder>
            <pattern>[%d{yyyy-MM-dd HH:mm:ss}][%-5level][%logger{20}:%line][%thread][%X{STEPID}] - %msg%n</pattern>
        </encoder>
    </appender>

    <appender name="FILE" class="ch.qos.logback.core.FileAppender">
        <file>${logfile}</file>
        <encoder>
            <pattern>[%d{yyyy-MM-dd HH:mm:ss.SSS}][%-5level][%logger{20}:%line][%thread][%X{STEPID}] - %msg%n</pattern>
        </encoder>
    </appender>

    <root level="info">
        <appender-ref ref="CONSOLE" />
        <appender-ref ref="FILE" />
    </root>

</configuration>
```

### 2.4. 데몬 배치 인스턴스를 위한 logback.xml 설정

데몬 배치 인스턴스 로그 출력을 위하여 다음과 같은 Appender가 필요하다.

데몬 배치 인스턴스 logback appender name설명

| 항목 
| 설명 
| 비고 

CONSOLE

데몬배치 Daemon 에서 출력하는 로그를 Console 보여주는 Appender.

|  

FILE

데몬배치 Daemon 에서 출력하는 로그를 파일로 저장하는 Appender

|  

데몬배치 인스턴스를 위한 logback.xml이다. 사용자환경에 맞게 경로 및 appender를 수정하여 사용한다.

```
<?xml version="1.0" encoding="UTF-8"?>
<configuration scan="false">

    <property name="LOGS_ABSOLUTE_PATH" value="/bxm/logs" />

    <appender name="CONSOLE" class="ch.qos.logback.core.ConsoleAppender">
        <encoder>
            <pattern>[%d{yyyy-MM-dd HH:mm:ss}][%-5level][%logger{36}:%line] - %msg%n</pattern>
        </encoder>
    </appender>

    <appender name="FILE" class="ch.qos.logback.core.rolling.RollingFileAppender">
        <file>${LOGS_ABSOLUTE_PATH}/batch/daemon.log</file>
        <encoder>
            <pattern>[%d{yyyy-MM-dd HH:mm:ss.SSS}][%-5level][%logger{36}:%line] - %msg%n</pattern>
        </encoder>
        <rollingPolicy class="ch.qos.logback.core.rolling.TimeBasedRollingPolicy">
            <fileNamePattern>${LOGS_ABSOLUTE_PATH}/batch/daemon.%d{yyyy-MM-dd}.log</fileNamePattern>
            <maxHistory>30</maxHistory>
            <totalSizeCap>3GB</totalSizeCap>
        </rollingPolicy>
    </appender>

    <root level="debug">
        <appender-ref ref="CONSOLE" />
        <appender-ref ref="FILE" />
    </root>

</configuration>
```

### 2.5. Ondemand 배치 인스턴스를 위한 logback.xml 설정

Ondemand 배치 인스턴스 로그 출력을 위하여 다음과 같은 Appender가 필요하다.

Ondemand 배치 인스턴스 logback appender name설명

| 항목 
| 설명 
| 비고 

FILE

프레임워크에서 생성하는 컨테이너 로딩 및 온디맨드 호출 정보를 로깅하는 Appender

|  

Ondemand배치 인스턴스를 위한 logback.xml이다. 사용자환경에 맞게 경로 및 appender를 수정하여 사용한다.

```
<?xml version="1.0" encoding="UTF-8"?>
<configuration scan="false">

    <property name="LOGS_ABSOLUTE_PATH" value="/bxm/logs" />

    <appender name="FILE" class="ch.qos.logback.core.rolling.RollingFileAppender">
        <file>${LOGS_ABSOLUTE_PATH}/batch/ondemand.log</file>
        <encoder>
            <pattern>[%d{yyyy-MM-dd HH:mm:ss.SSS}][%-5level][%logger{36}:%line] - %msg%n</pattern>
        </encoder>
        <rollingPolicy class="ch.qos.logback.core.rolling.TimeBasedRollingPolicy">
            <fileNamePattern>${LOGS_ABSOLUTE_PATH}/batch/ondemand.%d{yyyy-MM-dd}.log</fileNamePattern>
            <maxHistory>30</maxHistory>
            <totalSizeCap>3GB</totalSizeCap>
        </rollingPolicy>
    </appender>

    <root level="debug">
         <appender-ref ref="FILE" />
    </root>

</configuration>
```

### 2.6.  admin 인스턴스를 위한 logback.xml 설정

admin 인스턴스 로그 출력을 위하여 다음과 같은 Appender가 필요하다.

admin 인스턴스 logback appender name설명

| 항목 
| 설명 
| 비고 

console

Console 로 출력처리하기 위한 Appender로서 보통의 경우 WAS로그에 저장된다.

|  

applogfile

서비스로그를 파일로 처리하기 위한 Appender

|  

admin 인스턴스를 위한 logback.xml이다. 사용자환경에 맞게 경로 및 appender를 수정하여 사용한다.

```
<?xml version="1.0" encoding="UTF-8"?>
<configuration scan="false" >

    <property name="LOGS_ABSOLUTE_PATH" value="/data1/prod/bxm400/bxm/logs" />

    <appender name="console" class="ch.qos.logback.core.ConsoleAppender">
        <encoder>
            <pattern>[%d{yyyy-MM-dd HH:mm:ss.SSS}][%-5level][%logger{36}:%line] - %msg%n</pattern>
        </encoder>
    </appender>

    <appender name="applogfile" class="ch.qos.logback.core.rolling.RollingFileAppender">
        <file>${LOGS_ABSOLUTE_PATH}/admin/app_${bxm.node.name}_${bxm.instance.name}.log</file>
        <encoder>
            <pattern>[%d{yyyy-MM-dd HH:mm:ss.SSS}][%-5level][%logger{36}:%line] - %msg%n</pattern>
        </encoder>
        <rollingPolicy class="ch.qos.logback.core.rolling.TimeBasedRollingPolicy">
            <fileNamePattern>${LOGS_ABSOLUTE_PATH}/admin/app_${bxm.node.name}_${bxm.instance.name}.%d{yyyy-MM-dd}.log</fileNamePattern>
            <maxHistory>30</maxHistory>
            <totalSizeCap>3GB</totalSizeCap>
        </rollingPolicy>
    </appender>

    <root level="debug">
        <appender-ref ref="console" />
        <appender-ref ref="applogfile" />
    </root>

</configuration>
```

## 3.  배치 Database 연결설정(datasource.properties)

배치 어플리케이션에서 DB Connection을 생성하기 위한 DataSource 설정 화일이다.

다음과 같이 bxm-application.xml에 jdbc dataSource가 설정되어 있는 경우를 가정하여 datasource.properties의 프러퍼티를 이용하여 설정한다.

Figure 1. bxm-application.xml의 jdbc datasource설정

datasource.properties 설정항목

| 항목 
| 설명 
| 비고 

dirver-classname

Jdbc Driver Class Name

ex)oracle.jdbc.OracleDriver

|  

uri

jdbc DB 접속 URL

ex) jdbc:oracle:thin:@xx.xx.xxx.xx:1521:ora

|  

username

DB 접속 User명

|  

password

DB 접속 password.

암호화된 password를 사용할 지 평문 password를 사용할 지는 BXM_BUILDER_CONFIG의 설정값(namespace:'DBConnectionConfiguration', property:'jdbcUseEncPassword')을 참조한다.

평문password를 암호화된 password로 변경하는 방법은 "부록 A."를 참고한다.

|  

다음은 datasource.properties의 예제이다. 사용자환경에 맞게 <<&#8230;&#8203;>>로 표시된 값을 수정하여 사용한다.

```
dirver-classname=oracle.jdbc.OracleDriver
uri=<<DB JDBC URL>>
username=bxm
password=<<암호화된 bxm계정 password>>
```