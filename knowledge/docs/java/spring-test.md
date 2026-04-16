# Spring Test — 통합/슬라이스 테스트, MockMvc, Testcontainers

## 테스트 계층 개요

| 어노테이션 | 범위 | 속도 | 용도 |
|-----------|------|------|------|
| `@SpringBootTest` | 전체 컨텍스트 | 느림 | 통합 테스트 |
| `@WebMvcTest` | Web 계층만 | 빠름 | 컨트롤러 단위 |
| `@DataJpaTest` | JPA 계층만 | 빠름 | 레포지토리 단위 |
| `@RestClientTest` | RestTemplate/WebClient | 빠름 | HTTP 클라이언트 |
| 없음 (순수 JUnit) | 없음 | 매우 빠름 | 서비스/유틸 단위 |

---

## @SpringBootTest (통합 테스트)

```java
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
@ActiveProfiles("test")
class OrderIntegrationTest {

    @Autowired
    private TestRestTemplate restTemplate;

    @Autowired
    private OrderRepository orderRepository;

    @BeforeEach
    void setUp() {
        orderRepository.deleteAll();
    }

    @Test
    void 주문_생성_후_조회_성공() {
        // given
        CreateOrderRequest request = new CreateOrderRequest(1L, List.of(
                new OrderItemRequest(100L, 2)
        ));

        // when
        ResponseEntity<OrderDto> response = restTemplate
                .withBasicAuth("user@example.com", "password")
                .postForEntity("/api/orders", request, OrderDto.class);

        // then
        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.CREATED);
        assertThat(response.getBody()).isNotNull();
        assertThat(response.getBody().getStatus()).isEqualTo(OrderStatus.PENDING);
    }

    // MOCK: 실제 서버 없이 MockMvc 사용
    @SpringBootTest
    @AutoConfigureMockMvc
    class WithMockMvc {
        @Autowired MockMvc mockMvc;
        // ...
    }
}
```

---

## @WebMvcTest (컨트롤러 슬라이스)

```java
@WebMvcTest(OrderController.class)  // 지정 컨트롤러만 로드
class OrderControllerTest {

    @Autowired
    private MockMvc mockMvc;

    @Autowired
    private ObjectMapper objectMapper;

    @MockBean  // Spring 컨텍스트에 Mock 빈 등록
    private OrderService orderService;

    @MockBean
    private JwtTokenProvider tokenProvider;

    @Test
    @WithMockUser(username = "user@example.com", roles = "USER")  // 인증 우회
    void 주문_목록_조회() throws Exception {
        // given
        List<OrderDto> orders = List.of(
                new OrderDto(1L, OrderStatus.PENDING, 50000),
                new OrderDto(2L, OrderStatus.COMPLETED, 30000)
        );
        given(orderService.findByCurrentUser(any(Pageable.class)))
                .willReturn(new PageImpl<>(orders));

        // when & then
        mockMvc.perform(get("/api/orders")
                        .param("page", "0")
                        .param("size", "20")
                        .contentType(MediaType.APPLICATION_JSON))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.content").isArray())
                .andExpect(jsonPath("$.content.length()").value(2))
                .andExpect(jsonPath("$.content[0].id").value(1))
                .andExpect(jsonPath("$.content[0].status").value("PENDING"))
                .andDo(print());  // 요청/응답 콘솔 출력
    }

    @Test
    @WithMockUser
    void 주문_생성_검증_실패_400() throws Exception {
        // 빈 요청 바디 → @Valid 실패
        mockMvc.perform(post("/api/orders")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{}"))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.code").value("VALIDATION_FAILED"));
    }

    @Test
    @WithMockUser
    void 주문_생성_성공_201() throws Exception {
        CreateOrderRequest request = new CreateOrderRequest(1L,
                List.of(new OrderItemRequest(100L, 2)));
        OrderDto created = new OrderDto(10L, OrderStatus.PENDING, 20000);
        given(orderService.create(any())).willReturn(created);

        mockMvc.perform(post("/api/orders")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isCreated())
                .andExpect(header().exists("Location"))
                .andExpect(jsonPath("$.id").value(10));
    }
}
```

---

## @DataJpaTest (JPA 슬라이스)

```java
@DataJpaTest  // H2 in-memory DB 자동, @Repository 빈만 로드
@ActiveProfiles("test")
class UserRepositoryTest {

    @Autowired
    private UserRepository userRepository;

    @Autowired
    private TestEntityManager entityManager;  // 테스트용 EntityManager

    @Test
    void 이메일로_사용자_조회() {
        // given
        User user = User.create("김철수", "kim@example.com");
        entityManager.persistAndFlush(user);
        entityManager.clear();  // 1차 캐시 클리어 → DB에서 조회 강제

        // when
        Optional<User> found = userRepository.findByEmail("kim@example.com");

        // then
        assertThat(found).isPresent();
        assertThat(found.get().getName()).isEqualTo("김철수");
    }

    @Test
    void 페이징_조회() {
        // given
        for (int i = 0; i < 25; i++) {
            entityManager.persist(User.create("user" + i, "user" + i + "@example.com"));
        }
        entityManager.flush();

        // when
        Pageable pageable = PageRequest.of(0, 10, Sort.by("name").ascending());
        Page<User> page = userRepository.findAll(pageable);

        // then
        assertThat(page.getTotalElements()).isEqualTo(25);
        assertThat(page.getContent()).hasSize(10);
        assertThat(page.getTotalPages()).isEqualTo(3);
    }

    // 실제 DB 사용 (H2 대신)
    @DataJpaTest
    @AutoConfigureTestDatabase(replace = AutoConfigureTestDatabase.Replace.NONE)
    class WithRealDatabase { ... }
}
```

---

## Testcontainers (실제 DB 통합 테스트)

```java
// build.gradle.kts
// testImplementation("org.testcontainers:junit-jupiter:1.19.3")
// testImplementation("org.testcontainers:mysql:1.19.3")

@SpringBootTest
@Testcontainers
@ActiveProfiles("test")
class OrderRepositoryContainerTest {

    @Container
    static MySQLContainer<?> mysql = new MySQLContainer<>("mysql:8.0")
            .withDatabaseName("testdb")
            .withUsername("test")
            .withPassword("test");

    @DynamicPropertySource  // 컨테이너 포트는 동적 → 프로퍼티 동적 등록
    static void configureProperties(DynamicPropertyRegistry registry) {
        registry.add("spring.datasource.url", mysql::getJdbcUrl);
        registry.add("spring.datasource.username", mysql::getUsername);
        registry.add("spring.datasource.password", mysql::getPassword);
    }

    @Autowired
    private OrderRepository orderRepository;

    @Test
    void MySQL에서_쿼리_동작_검증() {
        Order order = Order.create(1L);
        orderRepository.save(order);

        List<Order> found = orderRepository.findByUserId(1L);
        assertThat(found).hasSize(1);
    }
}

// 공통 베이스 클래스로 컨테이너 재사용 (속도 향상)
@SpringBootTest
@ActiveProfiles("test")
public abstract class IntegrationTestBase {

    @Container
    static final MySQLContainer<?> MYSQL = new MySQLContainer<>("mysql:8.0")
            .withReuse(true);  // testcontainers.reuse.enable=true (~/.testcontainers.properties)

    @DynamicPropertySource
    static void configureProperties(DynamicPropertyRegistry registry) {
        registry.add("spring.datasource.url", MYSQL::getJdbcUrl);
        registry.add("spring.datasource.username", MYSQL::getUsername);
        registry.add("spring.datasource.password", MYSQL::getPassword);
    }
}
```

---

## @MockBean / @SpyBean

```java
@SpringBootTest
class PaymentServiceTest {

    @Autowired
    private PaymentService paymentService;

    @MockBean  // 실제 빈 대신 Mock으로 교체 (컨텍스트 재로드 주의)
    private PgClient pgClient;

    @SpyBean  // 실제 빈 유지하되 일부 메서드만 stub
    private NotificationService notificationService;

    @Test
    void 결제_성공_시_알림_발송() {
        // MockBean stub
        given(pgClient.charge(anyString(), anyInt()))
                .willReturn(new ChargeResult("tx-123", true));

        // SpyBean: 실제 로직은 실행, verify만
        paymentService.process(new PaymentRequest("card-001", 10000));

        verify(notificationService, times(1)).sendPaymentSuccess(any());
    }

    @Test
    void PG_오류_시_PaymentException_발생() {
        given(pgClient.charge(anyString(), anyInt()))
                .willThrow(new PgException("카드 한도 초과"));

        assertThatThrownBy(() -> paymentService.process(new PaymentRequest("card-001", 99999)))
                .isInstanceOf(PaymentException.class)
                .hasMessageContaining("카드 한도 초과");
    }
}
```

---

## @Sql / @DirtiesContext

```java
@DataJpaTest
@Sql("/sql/test-data.sql")  // 클래스 전체 적용
class ProductRepositoryTest {

    @Test
    @Sql(scripts = "/sql/products.sql",
         executionPhase = Sql.ExecutionPhase.BEFORE_TEST_METHOD)
    @Sql(scripts = "/sql/cleanup.sql",
         executionPhase = Sql.ExecutionPhase.AFTER_TEST_METHOD)
    void 상품_검색() { ... }
}

// @DirtiesContext: 컨텍스트 오염 시 재생성 (느리므로 최후 수단)
@Test
@DirtiesContext(methodMode = DirtiesContext.MethodMode.AFTER_METHOD)
void 싱글톤_상태_변경_테스트() { ... }
```

---

## JUnit 5 패턴

```java
class UserServiceTest {

    private UserService userService;
    private UserRepository userRepository;

    @BeforeEach
    void setUp() {
        userRepository = mock(UserRepository.class);
        userService = new UserService(userRepository);
    }

    @Test
    @DisplayName("존재하지 않는 사용자 조회 시 예외 발생")
    void findById_notFound_throws() {
        given(userRepository.findById(999L)).willReturn(Optional.empty());

        assertThatThrownBy(() -> userService.findById(999L))
                .isInstanceOf(EntityNotFoundException.class);
    }

    @Nested
    @DisplayName("사용자 생성")
    class Create {
        @Test
        void 이메일_중복_시_예외() {
            given(userRepository.existsByEmail("dup@example.com")).willReturn(true);
            assertThatThrownBy(() -> userService.create("dup@example.com"))
                    .isInstanceOf(DuplicateEmailException.class);
        }

        @Test
        void 정상_생성() {
            given(userRepository.existsByEmail("new@example.com")).willReturn(false);
            given(userRepository.save(any())).willAnswer(inv -> inv.getArgument(0));

            User result = userService.create("new@example.com");
            assertThat(result.getEmail()).isEqualTo("new@example.com");
        }
    }

    @ParameterizedTest
    @ValueSource(strings = {"", " ", "not-an-email", "a@"})
    @DisplayName("유효하지 않은 이메일 형식")
    void invalidEmail(String email) {
        assertThatThrownBy(() -> userService.create(email))
                .isInstanceOf(IllegalArgumentException.class);
    }

    @ParameterizedTest
    @CsvSource({
        "user@example.com, 김철수, USER",
        "admin@example.com, 관리자, ADMIN"
    })
    void 역할별_생성(String email, String name, UserRole role) {
        // ...
    }

    @ParameterizedTest
    @MethodSource("provideOrderStatuses")
    void 상태별_처리(OrderStatus status, boolean expectedComplete) { ... }

    static Stream<Arguments> provideOrderStatuses() {
        return Stream.of(
                Arguments.of(OrderStatus.COMPLETED, true),
                Arguments.of(OrderStatus.PENDING, false),
                Arguments.of(OrderStatus.CANCELLED, false)
        );
    }
}
```

---

## AssertJ 주요 패턴

```java
// 기본
assertThat(result).isEqualTo(expected);
assertThat(result).isNotNull().isInstanceOf(UserDto.class);
assertThat(str).isBlank();
assertThat(str).startsWith("prefix").endsWith("suffix").contains("mid");

// 숫자
assertThat(count).isGreaterThan(0).isLessThanOrEqualTo(100);
assertThat(price).isBetween(1000, 99999);

// 컬렉션
assertThat(list).hasSize(3).isNotEmpty();
assertThat(list).containsExactly(a, b, c);          // 순서 포함 정확히
assertThat(list).containsExactlyInAnyOrder(c, a, b); // 순서 무관
assertThat(list).contains(a, b);                     // 포함 여부만
assertThat(list).doesNotContain(x);
assertThat(list).allMatch(item -> item.isActive());
assertThat(list).noneMatch(item -> item.isDeleted());
assertThat(list).extracting("name").containsExactly("김철수", "이영희");
assertThat(list).extracting(User::getName, User::getEmail)
        .containsExactly(tuple("김철수", "kim@example.com"));

// 예외
assertThatThrownBy(() -> service.doSomething())
        .isInstanceOf(IllegalArgumentException.class)
        .hasMessage("유효하지 않은 입력")
        .hasMessageContaining("입력");

assertThatNoException().isThrownBy(() -> service.doSomething());

// Optional
assertThat(optional).isPresent().contains(expected);
assertThat(optional).isEmpty();

// 소수점
assertThat(0.1 + 0.2).isCloseTo(0.3, within(0.001));
```
