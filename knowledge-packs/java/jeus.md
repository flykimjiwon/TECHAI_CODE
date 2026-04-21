# JEUS WAS (Web Application Server)

TmaxSoft 제품. Java EE / Jakarta EE 호환 WAS. 국내 금융/공공기관 표준 WAS.

## 아키텍처

```
DAS (Domain Admin Server) — 도메인 전체 관리, 포트 9736
 └── MS (Managed Server) 1 — 실제 앱 실행, 포트 8088
 └── MS (Managed Server) 2 — 클러스터 구성 시
```

## 디렉토리 구조

```
$JEUS_HOME/
├── bin/                          # 실행 스크립트
│   ├── startDomainAdminServer    # DAS 시작
│   ├── stopServer                # 서버 중지
│   ├── startManagedServer        # MS 시작
│   └── jeusadmin                 # 관리 콘솔 CLI
├── domains/
│   └── {domain}/
│       ├── config/
│       │   └── domain.xml        # 핵심 설정 파일
│       └── servers/
│           └── {server}/
│               └── logs/         # 서버 로그
├── lib/                          # 공통 라이브러리
└── logs/                         # DAS 로그
```

## 기본 명령어

```bash
# DAS 시작/중지
$JEUS_HOME/bin/startDomainAdminServer
$JEUS_HOME/bin/stopServer -host localhost:9736 -u administrator -p <password>

# MS 시작
$JEUS_HOME/bin/startManagedServer -server server1 -u administrator -p <password>

# 대화형 관리 콘솔 접속
$JEUS_HOME/bin/jeusadmin -host localhost:9736 -u administrator -p <password>

# jeusadmin 주요 명령 (콘솔 내부)
list-servers                          # 서버 목록
server-info -server server1           # 서버 상세 정보
deploy -u administrator -p <password> -path /app/myapp.war -server server1
undeploy -appname myapp -server server1
redeploy -appname myapp -server server1
thread-dump -server server1           # 스레드 덤프 출력
```

## domain.xml 주요 설정

```xml
<domain>
  <servers>
    <server>
      <name>server1</name>
      <listeners>
        <listener>
          <name>http-listener</name>
          <port>8088</port>
        </listener>
      </listeners>
      <jvm-config>
        <jvm-option>-Xms1024m</jvm-option>
        <jvm-option>-Xmx2048m</jvm-option>
        <jvm-option>-XX:+UseG1GC</jvm-option>
        <jvm-option>-XX:MaxGCPauseMillis=200</jvm-option>
        <jvm-option>-XX:+HeapDumpOnOutOfMemoryError</jvm-option>
        <jvm-option>-XX:HeapDumpPath=/logs/heapdump.hprof</jvm-option>
        <jvm-option>-Dfile.encoding=UTF-8</jvm-option>
      </jvm-config>
    </server>
  </servers>
</domain>
```

## 데이터소스 설정 (JDBC)

```xml
<!-- domain.xml 내 resources 섹션 -->
<resources>
  <data-source>
    <name>myDS</name>
    <jndi-name>jdbc/myDS</jndi-name>
    <driver>
      <!-- Tibero -->
      <class-name>com.tmax.tibero.jdbc.TbDriver</class-name>
      <url>jdbc:tibero:thin:@localhost:8629:tibero</url>

      <!-- Oracle -->
      <!-- <class-name>oracle.jdbc.OracleDriver</class-name> -->
      <!-- <url>jdbc:oracle:thin:@localhost:1521:orcl</url> -->

      <!-- MySQL -->
      <!-- <class-name>com.mysql.cj.jdbc.Driver</class-name> -->
      <!-- <url>jdbc:mysql://localhost:3306/mydb?useUnicode=true&characterEncoding=UTF-8</url> -->

      <username>dbuser</username>
      <password>dbpass</password>
    </driver>
    <pool>
      <min>5</min>
      <max>30</max>
      <step>2</step>
      <period>3000</period>           <!-- 유휴 커넥션 검사 주기(ms) -->
      <wait-timeout>10000</wait-timeout>
      <idle-timeout>300000</idle-timeout>
    </pool>
    <connection-validation>
      <check-query>SELECT 1 FROM DUAL</check-query>  <!-- Tibero/Oracle -->
      <!-- <check-query>SELECT 1</check-query> -->   <!-- MySQL -->
      <check-period>60000</check-period>
    </connection-validation>
  </data-source>
</resources>
```

Java에서 JNDI 조회:
```java
Context ctx = new InitialContext();
DataSource ds = (DataSource) ctx.lookup("jdbc/myDS");
Connection conn = ds.getConnection();
```

## 애플리케이션 배포

```bash
# 1. jeusadmin CLI 배포
deploy -u administrator -p <password> \
       -path /deploy/myapp.war \
       -server server1 \
       -appname myapp

# 2. autodeploy 폴더에 복사 (자동 감지)
cp myapp.war $JEUS_HOME/domains/{domain}/servers/{server}/autodeploy/

# 3. Hot deploy (WAR 교체 — 다운타임 없이 클래스 재로드)
redeploy -appname myapp -server server1

# 4. 전체 재배포 (Undeploy → Deploy)
undeploy -appname myapp -server server1
deploy -path /deploy/myapp.war -server server1 -appname myapp
```

| 방식 | 다운타임 | 용도 |
|------|---------|------|
| autodeploy | 없음 | 개발 환경 |
| hot deploy (redeploy) | 없음 | 소규모 변경 |
| undeploy + deploy | 있음 | 구조적 변경, 운영 배포 |

## 세션 클러스터링

```xml
<!-- domain.xml -->
<cluster>
  <name>cluster1</name>
  <session-clustering>
    <type>memory</type>             <!-- in-memory replication -->
    <servers>
      <server>server1</server>
      <server>server2</server>
    </servers>
  </session-clustering>
</cluster>
```

## 로그 경로 및 설정

```
# 주요 로그 파일
$JEUS_HOME/domains/{domain}/servers/{server}/logs/JeusServer.log
$JEUS_HOME/domains/{domain}/servers/{server}/logs/JeusServer_GC.log
$JEUS_HOME/domains/{domain}/logs/DomainAdminServer.log

# GC 로깅 JVM 옵션 (domain.xml에 추가)
-Xlog:gc*:file=/logs/gc.log:time,uptime:filecount=5,filesize=20m
```

## WebtoB 연동 (TmaxSoft 웹서버)

```
# WebtoB — JEUS 연동 구성 (wbsvr.conf)
*WEBTOB
    Port = 80

*SVRGROUP
    JeusGroup  SVGTYPE=jeus, MINPROCNT=2, MAXPROCNT=8

*SERVER
    JeusServer  SVGNAME=JeusGroup, MIN=2, MAX=5

*URI
    uri1  Uri="/", Svrtype=jeus
```

JEUS 측 `domain.xml`에서 WebtoB 커넥터 포트 일치 확인 (기본 8007).

## Spring Boot WAR on JEUS

```java
// 1. SpringBootServletInitializer 상속
@SpringBootApplication
public class MyApplication extends SpringBootServletInitializer {
    @Override
    protected SpringApplicationBuilder configure(SpringApplicationBuilder application) {
        return application.sources(MyApplication.class);
    }
    public static void main(String[] args) {
        SpringApplication.run(MyApplication.class, args);
    }
}
```

```xml
<!-- 2. pom.xml: packaging WAR, tomcat scope provided -->
<packaging>war</packaging>
<dependency>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-tomcat</artifactId>
    <scope>provided</scope>
</dependency>
```

```bash
# 3. 빌드 후 배포
mvn clean package -DskipTests
cp target/myapp.war $JEUS_HOME/domains/{domain}/servers/{server}/autodeploy/
```

주의: `application.properties`의 `server.port`는 JEUS 내에서 무시됨. JEUS 리스너 포트가 우선.

## JEUS 7 vs JEUS 8 주요 차이

| 항목 | JEUS 7 | JEUS 8 |
|------|--------|--------|
| Java EE 버전 | Java EE 6 | Java EE 7 / Jakarta EE 8 |
| 설정 파일 | `jeusconfig.xml` + `JeusMain.xml` | 통합 `domain.xml` |
| 관리 콘솔 | JEUS Manager (GUI) | WebAdmin (브라우저) + jeusadmin CLI |
| 클러스터 설정 | 별도 파일 | domain.xml 통합 |
| HTTP/2 | 미지원 | 지원 |

## 트러블슈팅

### 포트 충돌 확인
```bash
# 기본 포트: DAS 9736, HTTP 8088, WebtoB connector 8007
lsof -i :9736
lsof -i :8088
netstat -an | grep 9736
```

### OutOfMemoryError 대응
```bash
# HeapDump 생성 (JVM 옵션으로 자동 생성 권장)
jmap -dump:format=b,file=/tmp/heap.hprof <PID>

# 메모리 사용 현황
jstat -gcutil <PID> 1000 10

# domain.xml JVM 옵션에 추가
-XX:+HeapDumpOnOutOfMemoryError
-XX:HeapDumpPath=/logs/
```

### 스레드 덤프
```bash
# 방법 1: 시그널 (Linux)
kill -3 <PID>                         # JeusServer.log에 출력됨

# 방법 2: jeusadmin
thread-dump -server server1

# 방법 3: jstack
jstack <PID> > /tmp/threaddump.txt
```

### 커넥션 풀 고갈
```bash
# jeusadmin에서 모니터링
list-datasources -server server1
datasource-info -ds myDS -server server1

# 증상: "No more idle connections" 에러
# 대응:
# 1. max pool size 증가 (domain.xml pool/max)
# 2. 커넥션 누수 확인 (close() 호출 누락)
# 3. wait-timeout 조정
# 4. check-query로 dead connection 제거 주기 단축
```

### 배포 실패 시
```bash
# 로그 확인
tail -f $JEUS_HOME/domains/{domain}/servers/server1/logs/JeusServer.log

# 클래스 충돌: WEB-INF/lib와 $JEUS_HOME/lib 중복 라이브러리 제거
# context root 충돌: undeploy 후 재배포
```
