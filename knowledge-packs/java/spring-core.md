# Spring Core — DI, Configuration, Bean Lifecycle

## 의존성 주입 (Dependency Injection)

### 컴포넌트 스테레오타입

```java
@Component          // 범용 빈
@Service            // 비즈니스 로직 계층
@Repository         // 데이터 접근 계층 (DataAccessException 변환 포함)
@Controller         // Spring MVC 컨트롤러
@RestController     // @Controller + @ResponseBody
```

### Constructor Injection (권장)

```java
@Service
public class OrderService {

    private final OrderRepository orderRepository;
    private final PaymentService paymentService;

    // Spring 4.3+ 단일 생성자는 @Autowired 생략 가능
    public OrderService(OrderRepository orderRepository,
                        PaymentService paymentService) {
        this.orderRepository = orderRepository;
        this.paymentService = paymentService;
    }
}
```

### Field Injection (지양 — 테스트 어려움, 순환의존 감지 불가)

```java
@Service
public class OrderService {
    @Autowired
    private OrderRepository orderRepository; // final 불가, 테스트 시 직접 주입 불가
}
```

### Setter Injection (선택적 의존성)

```java
@Service
public class NotificationService {
    private EmailSender emailSender;

    @Autowired(required = false)
    public void setEmailSender(EmailSender emailSender) {
        this.emailSender = emailSender;
    }
}
```

---

## @Qualifier / @Primary

```java
// 같은 타입 빈 여러 개 존재 시
@Component
@Primary  // 기본 선택 빈
public class KakaoPay implements PaymentGateway { ... }

@Component
public class TossPay implements PaymentGateway { ... }

// 주입 시 명시적 선택
@Service
public class CheckoutService {
    public CheckoutService(@Qualifier("tossPay") PaymentGateway gateway) {
        this.gateway = gateway;
    }
}

// 이름 기반 (빈 이름 = 클래스명 camelCase)
@Autowired
@Qualifier("kakaoPayImpl")
private PaymentGateway gateway;
```

---

## @Configuration / @Bean

```java
@Configuration
public class AppConfig {

    // 외부 라이브러리 등 @Component 붙이기 어려운 경우
    @Bean
    public ObjectMapper objectMapper() {
        ObjectMapper mapper = new ObjectMapper();
        mapper.disable(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES);
        mapper.registerModule(new JavaTimeModule());
        return mapper;
    }

    @Bean
    @Scope("prototype")  // 기본은 singleton
    public HttpClient httpClient() {
        return HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(5))
                .build();
    }

    // @Bean 메서드 간 의존: proxyBeanMethods=true(기본)라서 싱글톤 보장
    @Bean
    public DataSource dataSource() { ... }

    @Bean
    public JdbcTemplate jdbcTemplate() {
        return new JdbcTemplate(dataSource()); // dataSource()는 프록시로 싱글톤 반환
    }
}

// proxyBeanMethods=false: lite mode, 성능 향상 (Spring Boot AutoConfig 내부에서 자주 사용)
@Configuration(proxyBeanMethods = false)
public class LiteConfig { ... }
```

---

## Profiles

```java
@Configuration
@Profile("local")
public class LocalDataSourceConfig {
    @Bean
    public DataSource dataSource() {
        return new EmbeddedDatabaseBuilder()
                .setType(EmbeddedDatabaseType.H2)
                .build();
    }
}

@Configuration
@Profile("prod")
public class ProdDataSourceConfig {
    @Bean
    public DataSource dataSource() {
        HikariConfig config = new HikariConfig();
        config.setJdbcUrl(jdbcUrl);
        return new HikariDataSource(config);
    }
}

// 여러 프로파일 조건
@Profile({"dev", "local"})
@Profile("!prod")  // prod가 아닐 때
```

**application.yml 프로파일 분리:**

```yaml
# application.yml (공통)
spring:
  application:
    name: my-service

---
# application-local.yml
spring:
  config:
    activate:
      on-profile: local
  datasource:
    url: jdbc:h2:mem:testdb

---
# application-prod.yml
spring:
  config:
    activate:
      on-profile: prod
  datasource:
    url: ${DB_URL}
    username: ${DB_USERNAME}
    password: ${DB_PASSWORD}
```

활성화: `--spring.profiles.active=prod` 또는 환경변수 `SPRING_PROFILES_ACTIVE=prod`

---

## @Value / @ConfigurationProperties

```java
// @Value: 단순 값 주입
@Service
public class ApiClient {
    @Value("${external.api.url}")
    private String apiUrl;

    @Value("${external.api.timeout:5000}")  // 기본값
    private int timeout;

    @Value("${app.allowed-origins}")  // 콤마 구분 리스트
    private List<String> allowedOrigins;
}
```

```java
// @ConfigurationProperties: 타입 안전한 그룹 설정 (권장)
@ConfigurationProperties(prefix = "external.api")
@Component  // 또는 @EnableConfigurationProperties(ApiProperties.class)
public class ApiProperties {
    private String url;
    private int timeout = 5000;
    private List<String> endpoints = new ArrayList<>();

    // Spring Boot 3.x: record 사용 가능
    // getters/setters 필수 (또는 @ConstructorBinding)
}
```

```yaml
external:
  api:
    url: https://api.example.com
    timeout: 3000
    endpoints:
      - /v1/users
      - /v1/orders
```

---

## XML → Java Config 마이그레이션

```xml
<!-- 레거시 applicationContext.xml -->
<bean id="userService" class="com.example.UserService">
    <constructor-arg ref="userRepository"/>
    <property name="maxRetry" value="3"/>
</bean>

<bean id="userRepository" class="com.example.UserRepositoryImpl">
    <property name="dataSource" ref="dataSource"/>
</bean>
```

```java
// Java Config 동등 코드
@Configuration
public class ServiceConfig {
    @Bean
    public UserService userService(UserRepository userRepository) {
        UserService service = new UserService(userRepository);
        service.setMaxRetry(3);
        return service;
    }

    @Bean
    public UserRepository userRepository(DataSource dataSource) {
        UserRepositoryImpl repo = new UserRepositoryImpl();
        repo.setDataSource(dataSource);
        return repo;
    }
}

// XML과 Java Config 혼용 (점진적 마이그레이션)
@Configuration
@ImportResource("classpath:legacy-context.xml")
public class MigrationConfig {
    // 새 빈만 Java Config로
}
```

---

## Bean Lifecycle

```java
@Component
public class CacheService {

    @PostConstruct  // 빈 초기화 완료 후 (의존성 주입 완료 시점)
    public void init() {
        // 캐시 워밍, 커넥션 풀 초기화 등
        System.out.println("캐시 초기화 완료");
    }

    @PreDestroy  // 컨테이너 종료 전 (graceful shutdown)
    public void cleanup() {
        // 리소스 해제
        System.out.println("캐시 정리 완료");
    }
}

// InitializingBean / DisposableBean 인터페이스 (레거시 방식)
@Component
public class LegacyService implements InitializingBean, DisposableBean {
    @Override
    public void afterPropertiesSet() { /* 초기화 */ }

    @Override
    public void destroy() { /* 정리 */ }
}

// @Bean에서 지정
@Bean(initMethod = "start", destroyMethod = "stop")
public SomeExternalService externalService() { ... }
```

---

## Conditional 빈 등록

```java
// 프로퍼티 값 조건
@Bean
@ConditionalOnProperty(name = "feature.cache.enabled", havingValue = "true", matchIfMissing = false)
public CacheService cacheService() { ... }

// 클래스 존재 여부 조건
@Bean
@ConditionalOnClass(name = "com.redis.clients.jedis.Jedis")
public RedisCache redisCache() { ... }

// 빈 존재/부재 조건
@Bean
@ConditionalOnMissingBean(CacheService.class)
public CacheService defaultCacheService() { ... }

@Bean
@ConditionalOnBean(DataSource.class)
public JdbcTemplate jdbcTemplate(DataSource ds) { ... }

// 커스텀 Condition
public class OnKubernetesCondition implements Condition {
    @Override
    public boolean matches(ConditionContext context, AnnotatedTypeMetadata metadata) {
        return System.getenv("KUBERNETES_SERVICE_HOST") != null;
    }
}

@Bean
@Conditional(OnKubernetesCondition.class)
public KubernetesHealthIndicator k8sHealthIndicator() { ... }
```

---

## 스코프

```java
@Bean
@Scope("singleton")   // 기본: 컨테이너당 1개
@Scope("prototype")   // 요청마다 새 인스턴스
@Scope("request")     // HTTP 요청당 1개 (웹 환경)
@Scope("session")     // HTTP 세션당 1개 (웹 환경)

// singleton에 prototype 주입 시 문제 → ObjectProvider 사용
@Service
public class ReportService {
    private final ObjectProvider<ReportBuilder> builderProvider;

    public ReportService(ObjectProvider<ReportBuilder> builderProvider) {
        this.builderProvider = builderProvider;
    }

    public Report generate() {
        ReportBuilder builder = builderProvider.getObject(); // 매번 새 인스턴스
        return builder.build();
    }
}
```
