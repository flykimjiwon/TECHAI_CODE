# Spring Boot Ops — 설정, Actuator, 로깅, 빌드, 배포

## application.yml 프로파일 분리

```yaml
# application.yml (공통 기본값)
spring:
  application:
    name: my-service
  jackson:
    time-zone: Asia/Seoul
    default-property-inclusion: non_null

server:
  port: 8080
  shutdown: graceful  # 요청 처리 완료 후 종료
  tomcat:
    max-threads: 200
    accept-count: 100

management:
  endpoints:
    web:
      exposure:
        include: health,info,metrics
  endpoint:
    health:
      show-details: when-authorized

---
spring:
  config:
    activate:
      on-profile: local
  datasource:
    url: jdbc:h2:mem:testdb;MODE=MySQL
    driver-class-name: org.h2.Driver
  h2:
    console:
      enabled: true
  jpa:
    show-sql: true
    properties:
      hibernate:
        format_sql: true
logging:
  level:
    com.example: DEBUG
    org.hibernate.SQL: DEBUG

---
spring:
  config:
    activate:
      on-profile: dev
  datasource:
    url: jdbc:mysql://dev-db:3306/mydb
    username: ${DB_USERNAME}
    password: ${DB_PASSWORD}
    hikari:
      maximum-pool-size: 10
  jpa:
    show-sql: false
logging:
  level:
    com.example: DEBUG

---
spring:
  config:
    activate:
      on-profile: prod
  datasource:
    url: ${DB_URL}
    username: ${DB_USERNAME}
    password: ${DB_PASSWORD}
    hikari:
      maximum-pool-size: 50
      minimum-idle: 10
      connection-timeout: 30000
      idle-timeout: 600000
      max-lifetime: 1800000
  jpa:
    show-sql: false
management:
  endpoints:
    web:
      exposure:
        include: health,info  # prod에서는 최소 노출
logging:
  level:
    root: WARN
    com.example: INFO
```

---

## @ConfigurationProperties 타입 안전 설정

```java
// 설정 클래스
@ConfigurationProperties(prefix = "app")
@Component
@Validated  // JSR-303 검증 활성화
public class AppProperties {

    @NotBlank
    private String name;

    private Security security = new Security();
    private Database database = new Database();

    @Data
    public static class Security {
        @NotBlank
        private String jwtSecret;
        private long jwtExpiration = 3600000L;
        private List<String> allowedOrigins = new ArrayList<>();
    }

    @Data
    public static class Database {
        private int poolSize = 10;
        private Duration connectionTimeout = Duration.ofSeconds(30);
    }
    // getters/setters ...
}

// Spring Boot 3.x: record 사용 가능
@ConfigurationProperties(prefix = "app.security")
public record SecurityProperties(
    @NotBlank String jwtSecret,
    @DurationUnit(ChronoUnit.MILLIS) Duration jwtExpiration,
    List<String> allowedOrigins
) {}
```

```yaml
app:
  name: my-service
  security:
    jwt-secret: ${JWT_SECRET}
    jwt-expiration: 3600000
    allowed-origins:
      - https://app.example.com
      - http://localhost:3000
  database:
    pool-size: 20
    connection-timeout: 30s
```

---

## Actuator

```yaml
management:
  endpoints:
    web:
      base-path: /actuator
      exposure:
        include: "*"  # 개발용 전체 노출
  endpoint:
    health:
      show-details: always
      show-components: always
    shutdown:
      enabled: true  # POST /actuator/shutdown 으로 종료
  info:
    env:
      enabled: true
    git:
      mode: full  # git.properties 빌드 정보 노출

# info 엔드포인트 커스텀 데이터
info:
  app:
    name: ${spring.application.name}
    version: "@project.version@"  # Maven/Gradle 버전 자동 삽입
    description: 주문 관리 서비스
```

**주요 엔드포인트:**
- `GET /actuator/health` — 헬스 체크 (K8s liveness/readiness probe)
- `GET /actuator/metrics` — 메트릭 목록
- `GET /actuator/metrics/jvm.memory.used` — 특정 메트릭
- `GET /actuator/env` — 환경 변수 (민감 정보 마스킹 자동)
- `GET /actuator/loggers` — 로거 레벨 조회
- `POST /actuator/loggers/com.example` `{"configuredLevel":"DEBUG"}` — 런타임 로그 레벨 변경

```java
// 커스텀 HealthIndicator
@Component
public class ExternalApiHealthIndicator implements HealthIndicator {

    private final ExternalApiClient client;

    @Override
    public Health health() {
        try {
            boolean alive = client.ping();
            if (alive) {
                return Health.up()
                        .withDetail("url", client.getBaseUrl())
                        .withDetail("responseTime", "120ms")
                        .build();
            }
            return Health.down().withDetail("reason", "ping failed").build();
        } catch (Exception e) {
            return Health.down(e).build();
        }
    }
}

// 커스텀 InfoContributor
@Component
public class BuildInfoContributor implements InfoContributor {
    @Override
    public void contribute(Info.Builder builder) {
        builder.withDetail("build", Map.of(
            "version", "1.2.3",
            "timestamp", LocalDateTime.now().toString()
        ));
    }
}
```

---

## 로깅 — Logback + MDC

```xml
<!-- logback-spring.xml -->
<configuration>
    <springProfile name="local,dev">
        <appender name="CONSOLE" class="ch.qos.logback.core.ConsoleAppender">
            <encoder>
                <pattern>%d{HH:mm:ss.SSS} [%thread] [%X{traceId}] %-5level %logger{36} - %msg%n</pattern>
            </encoder>
        </appender>
        <root level="INFO">
            <appender-ref ref="CONSOLE"/>
        </root>
    </springProfile>

    <springProfile name="prod">
        <appender name="FILE" class="ch.qos.logback.core.rolling.RollingFileAppender">
            <file>logs/app.log</file>
            <rollingPolicy class="ch.qos.logback.core.rolling.TimeBasedRollingPolicy">
                <fileNamePattern>logs/app.%d{yyyy-MM-dd}.%i.log.gz</fileNamePattern>
                <timeBasedFileNamingAndTriggeringPolicy
                    class="ch.qos.logback.core.rolling.SizeAndTimeBasedFNATP">
                    <maxFileSize>100MB</maxFileSize>
                </timeBasedFileNamingAndTriggeringPolicy>
                <maxHistory>30</maxHistory>
            </rollingPolicy>
            <encoder class="net.logstash.logback.encoder.LogstashEncoder"/>
        </appender>
        <root level="INFO">
            <appender-ref ref="FILE"/>
        </root>
    </springProfile>
</configuration>
```

```java
// MDC로 요청 추적
@Component
public class MdcLoggingFilter extends OncePerRequestFilter {
    @Override
    protected void doFilterInternal(HttpServletRequest req, HttpServletResponse res,
                                    FilterChain chain) throws ServletException, IOException {
        String traceId = Optional.ofNullable(req.getHeader("X-Trace-Id"))
                .orElse(UUID.randomUUID().toString().substring(0, 8));
        MDC.put("traceId", traceId);
        MDC.put("method", req.getMethod());
        MDC.put("uri", req.getRequestURI());
        res.addHeader("X-Trace-Id", traceId);
        try {
            chain.doFilter(req, res);
        } finally {
            MDC.clear();  // 스레드풀 환경에서 반드시 클리어
        }
    }
}

// 로거 사용 (SLF4J)
@Slf4j  // Lombok — private static final Logger log = LoggerFactory.getLogger(...)
@Service
public class OrderService {
    public void process(Long orderId) {
        log.info("주문 처리 시작: orderId={}", orderId);
        try {
            // ...
        } catch (Exception e) {
            log.error("주문 처리 실패: orderId={}", orderId, e);
            throw e;
        }
    }
}
```

---

## 패키징 — JAR vs WAR

```kotlin
// build.gradle.kts — 실행 가능한 FAT JAR (Spring Boot 기본)
plugins {
    id("org.springframework.boot") version "3.2.0"
    id("io.spring.dependency-management") version "1.1.4"
    kotlin("jvm") version "1.9.21"
    kotlin("plugin.spring") version "1.9.21"
    kotlin("plugin.jpa") version "1.9.21"
}

dependencies {
    implementation("org.springframework.boot:spring-boot-starter-web")
    implementation("org.springframework.boot:spring-boot-starter-data-jpa")
    implementation("org.springframework.boot:spring-boot-starter-security")
    runtimeOnly("com.mysql:mysql-connector-j")
    testImplementation("org.springframework.boot:spring-boot-starter-test")
}

tasks.bootJar {
    archiveFileName.set("app.jar")  // 빌드 결과물 이름
    layered { enabled.set(true) }   // 레이어드 JAR (Docker 캐시 최적화)
}
```

```xml
<!-- pom.xml — 외장 톰캣 WAR 배포 -->
<packaging>war</packaging>
<dependencies>
    <dependency>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-tomcat</artifactId>
        <scope>provided</scope>  <!-- 외장 톰캣 사용 시 -->
    </dependency>
</dependencies>
```

```java
// WAR 배포 시 필수
@SpringBootApplication
public class Application extends SpringBootServletInitializer {
    @Override
    protected SpringApplicationBuilder configure(SpringApplicationBuilder builder) {
        return builder.sources(Application.class);
    }
    public static void main(String[] args) {
        SpringApplication.run(Application.class, args);
    }
}
```

---

## Docker — Multi-stage Build

```dockerfile
# Dockerfile (Layered JAR 활용)
FROM eclipse-temurin:21-jdk-alpine AS builder
WORKDIR /app
COPY gradlew .
COPY gradle gradle
COPY build.gradle.kts settings.gradle.kts .
RUN ./gradlew dependencies --no-daemon  # 의존성 캐시 레이어
COPY src src
RUN ./gradlew bootJar --no-daemon

# 레이어 추출
FROM eclipse-temurin:21-jdk-alpine AS extractor
WORKDIR /app
COPY --from=builder /app/build/libs/app.jar app.jar
RUN java -Djarmode=layertools -jar app.jar extract

# 최종 이미지
FROM eclipse-temurin:21-jre-alpine
WORKDIR /app
COPY --from=extractor /app/dependencies/ ./
COPY --from=extractor /app/spring-boot-loader/ ./
COPY --from=extractor /app/snapshot-dependencies/ ./
COPY --from=extractor /app/application/ ./

ENV SPRING_PROFILES_ACTIVE=prod
ENV JAVA_OPTS="-Xms512m -Xmx1024m -XX:+UseContainerSupport"

EXPOSE 8080
ENTRYPOINT ["sh", "-c", "java $JAVA_OPTS org.springframework.boot.loader.launch.JarLauncher"]
```

---

## 환경변수 바인딩

Spring Boot는 환경변수를 자동으로 프로퍼티로 변환:

```
SPRING_DATASOURCE_URL        → spring.datasource.url
SPRING_DATASOURCE_USERNAME   → spring.datasource.username
APP_SECURITY_JWT_SECRET      → app.security.jwt-secret
SERVER_PORT                  → server.port
SPRING_PROFILES_ACTIVE       → spring.profiles.active
```

```bash
# 실행 시 환경변수 전달
java -jar app.jar \
  --spring.profiles.active=prod \
  --server.port=9090

# 시스템 프로퍼티 우선순위 (높은 순)
# 1. CLI 인수 (--key=value)
# 2. SPRING_APPLICATION_JSON
# 3. 환경변수
# 4. application-{profile}.yml
# 5. application.yml
```

---

## Graceful Shutdown

```yaml
server:
  shutdown: graceful  # 기본: immediate

spring:
  lifecycle:
    timeout-per-shutdown-phase: 30s  # 최대 대기 시간
```

```java
// K8s readiness/liveness probe 설정
management:
  endpoint:
    health:
      probes:
        enabled: true  # /actuator/health/liveness, /actuator/health/readiness
  health:
    livenessstate:
      enabled: true
    readinessstate:
      enabled: true
```
