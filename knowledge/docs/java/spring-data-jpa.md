# Spring Data JPA — Entity, Repository, Query, N+1, Transaction

## Entity 기본

```java
@Entity
@Table(name = "users",
       uniqueConstraints = @UniqueConstraint(columnNames = {"email"}),
       indexes = @Index(name = "idx_users_email", columnList = "email"))
@Getter
@NoArgsConstructor(access = AccessLevel.PROTECTED)  // JPA 기본 생성자 (직접 호출 방지)
public class User extends BaseEntity {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)  // AUTO_INCREMENT
    private Long id;

    @Column(nullable = false, length = 50)
    private String name;

    @Column(nullable = false, unique = true)
    private String email;

    @Enumerated(EnumType.STRING)  // DB에 문자열로 저장 (ORDINAL 금지)
    private UserRole role;

    @Embedded
    private Address address;  // 값 타입

    // 정적 팩토리 메서드 패턴 (직접 생성자 노출 지양)
    public static User create(String name, String email) {
        User user = new User();
        user.name = name;
        user.email = email;
        user.role = UserRole.USER;
        return user;
    }

    public void updateName(String name) {
        this.name = name;
    }
}

// 값 타입 (Embeddable)
@Embeddable
public class Address {
    @Column(name = "city") private String city;
    @Column(name = "street") private String street;
    @Column(name = "zipcode") private String zipcode;
}

// Auditing 기반 클래스
@MappedSuperclass
@EntityListeners(AuditingEntityListener.class)
public abstract class BaseEntity {
    @CreatedDate
    @Column(updatable = false)
    private LocalDateTime createdAt;

    @LastModifiedDate
    private LocalDateTime updatedAt;

    @CreatedBy
    @Column(updatable = false)
    private String createdBy;

    @LastModifiedBy
    private String updatedBy;
}

// Auditing 활성화
@SpringBootApplication
@EnableJpaAuditing(auditorAwareRef = "auditorProvider")
public class Application { ... }

@Bean
public AuditorAware<String> auditorProvider() {
    return () -> Optional.ofNullable(SecurityContextHolder.getContext().getAuthentication())
            .map(Authentication::getName);
}
```

---

## 연관관계

```java
// @ManyToOne (다대일 — 외래키 보유 쪽)
@Entity
public class Order {
    @Id @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @ManyToOne(fetch = FetchType.LAZY)  // 기본 EAGER → 항상 LAZY로 명시
    @JoinColumn(name = "user_id")
    private User user;

    @OneToMany(mappedBy = "order", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<OrderItem> items = new ArrayList<>();

    public void addItem(OrderItem item) {
        items.add(item);
        item.setOrder(this);  // 양방향 연관관계 편의 메서드
    }
}

// @OneToOne
@Entity
public class UserProfile {
    @Id @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @OneToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "user_id")
    private User user;
}

// @ManyToMany → 중간 엔티티로 풀기 (실무에서 직접 사용 지양)
@Entity
public class ProductTag {
    @Id @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @ManyToOne(fetch = FetchType.LAZY)
    private Product product;

    @ManyToOne(fetch = FetchType.LAZY)
    private Tag tag;

    private LocalDateTime taggedAt;  // 중간 엔티티 전용 필드 추가 가능
}
```

---

## Repository

```java
// JpaRepository<엔티티, PK타입> 상속
public interface UserRepository extends JpaRepository<User, Long> {

    // 쿼리 메서드 네이밍 규칙
    Optional<User> findByEmail(String email);
    List<User> findByRoleAndActiveTrue(UserRole role);
    List<User> findByNameContainingIgnoreCase(String keyword);
    List<User> findByCreatedAtBetween(LocalDateTime from, LocalDateTime to);
    long countByRole(UserRole role);
    boolean existsByEmail(String email);
    void deleteByEmail(String email);

    // 정렬
    List<User> findByRoleOrderByCreatedAtDesc(UserRole role);

    // 페이징
    Page<User> findByRole(UserRole role, Pageable pageable);
    Slice<User> findByActiveTrue(Pageable pageable);  // 다음 페이지 존재 여부만

    // Top/First
    Optional<User> findFirstByRoleOrderByCreatedAtDesc(UserRole role);
    List<User> findTop5ByRoleOrderByCreatedAtDesc(UserRole role);
}
```

---

## @Query (JPQL / Native)

```java
public interface OrderRepository extends JpaRepository<Order, Long> {

    // JPQL (객체 기반 — 엔티티명/필드명 사용)
    @Query("SELECT o FROM Order o WHERE o.user.id = :userId AND o.status = :status")
    List<Order> findByUserIdAndStatus(@Param("userId") Long userId,
                                      @Param("status") OrderStatus status);

    // fetch join으로 N+1 해결
    @Query("SELECT DISTINCT o FROM Order o JOIN FETCH o.items WHERE o.user.id = :userId")
    List<Order> findWithItemsByUserId(@Param("userId") Long userId);

    // Projection (필요한 필드만)
    @Query("SELECT new com.example.dto.OrderSummary(o.id, o.totalAmount, o.status) " +
           "FROM Order o WHERE o.user.id = :userId")
    List<OrderSummary> findSummaryByUserId(@Param("userId") Long userId);

    // Native SQL (DB 종속 — 최후 수단)
    @Query(value = "SELECT * FROM orders WHERE user_id = :userId AND YEAR(created_at) = :year",
           nativeQuery = true)
    List<Order> findByUserIdAndYear(@Param("userId") Long userId, @Param("year") int year);

    // 수정 쿼리
    @Modifying
    @Query("UPDATE User u SET u.active = false WHERE u.lastLoginAt < :cutoff")
    int deactivateInactiveUsers(@Param("cutoff") LocalDateTime cutoff);
}
```

---

## N+1 문제와 해결

```java
// 문제: Order 목록 조회 시 각 Order마다 user 쿼리 발생
List<Order> orders = orderRepository.findAll();
orders.forEach(o -> System.out.println(o.getUser().getName())); // N+1 발생

// 해결 1: fetch join (컬렉션은 DISTINCT 필요)
@Query("SELECT DISTINCT o FROM Order o JOIN FETCH o.user JOIN FETCH o.items")
List<Order> findAllWithUserAndItems();

// 해결 2: @EntityGraph (JPQL 없이 선언적)
@EntityGraph(attributePaths = {"user", "items"})
@Query("SELECT o FROM Order o")
List<Order> findAllWithGraph();

// 해결 3: @BatchSize (Hibernate — N+1을 N/batch_size+1로 줄임)
@OneToMany(mappedBy = "order")
@BatchSize(size = 100)
private List<OrderItem> items;

// 또는 글로벌 설정
// application.yml
// spring.jpa.properties.hibernate.default_batch_fetch_size: 100

// 주의: fetch join + 페이징 함께 사용 금지 (컬렉션)
// → ToOne 관계는 fetch join, 컬렉션은 @BatchSize 조합 권장
```

---

## 페이징

```java
// 컨트롤러
@GetMapping
public Page<OrderDto> list(
        @RequestParam(defaultValue = "0") int page,
        @RequestParam(defaultValue = "20") int size,
        @RequestParam(defaultValue = "createdAt") String sort) {

    Pageable pageable = PageRequest.of(page, size, Sort.by(sort).descending());
    return orderService.findAll(pageable).map(OrderDto::from);
}

// 서비스/레포지토리
Page<Order> findByStatus(OrderStatus status, Pageable pageable);

// Page vs Slice
// Page: 전체 count 쿼리 실행 → 총 페이지 수 알 수 있음 (목록 UI)
// Slice: count 없이 다음 페이지 존재 여부만 → 무한 스크롤에 적합

// count 쿼리 최적화 (복잡한 join 있을 때)
@Query(value = "SELECT o FROM Order o JOIN FETCH o.user WHERE o.status = :status",
       countQuery = "SELECT COUNT(o) FROM Order o WHERE o.status = :status")
Page<Order> findByStatus(@Param("status") OrderStatus status, Pageable pageable);
```

---

## Specification (동적 쿼리)

```java
// Specification 정의
public class OrderSpec {

    public static Specification<Order> hasStatus(OrderStatus status) {
        return (root, query, cb) ->
                status == null ? null : cb.equal(root.get("status"), status);
    }

    public static Specification<Order> hasUserId(Long userId) {
        return (root, query, cb) ->
                userId == null ? null : cb.equal(root.get("user").get("id"), userId);
    }

    public static Specification<Order> createdAfter(LocalDateTime from) {
        return (root, query, cb) ->
                from == null ? null : cb.greaterThanOrEqualTo(root.get("createdAt"), from);
    }
}

// 레포지토리
public interface OrderRepository extends JpaRepository<Order, Long>,
        JpaSpecificationExecutor<Order> { }

// 서비스
public Page<Order> search(OrderSearchDto dto, Pageable pageable) {
    Specification<Order> spec = Specification
            .where(OrderSpec.hasStatus(dto.getStatus()))
            .and(OrderSpec.hasUserId(dto.getUserId()))
            .and(OrderSpec.createdAfter(dto.getFrom()));
    return orderRepository.findAll(spec, pageable);
}
```

---

## QueryDSL

```java
// build.gradle.kts 의존성
// implementation("com.querydsl:querydsl-jpa:5.0.0:jakarta")
// annotationProcessor("com.querydsl:querydsl-apt:5.0.0:jakarta")

// QueryDSL 레포지토리
@Repository
@RequiredArgsConstructor
public class OrderQueryRepository {

    private final JPAQueryFactory queryFactory;

    public List<Order> searchOrders(OrderSearchCondition cond) {
        QOrder order = QOrder.order;
        QUser user = QUser.user;

        return queryFactory
                .selectFrom(order)
                .join(order.user, user).fetchJoin()
                .where(
                    statusEq(cond.getStatus()),
                    userIdEq(cond.getUserId()),
                    amountGoe(cond.getMinAmount())
                )
                .orderBy(order.createdAt.desc())
                .offset(cond.getOffset())
                .limit(cond.getLimit())
                .fetch();
    }

    // BooleanExpression: null 반환 시 where 조건에서 제외 (동적 쿼리 핵심)
    private BooleanExpression statusEq(OrderStatus status) {
        return status != null ? QOrder.order.status.eq(status) : null;
    }

    private BooleanExpression userIdEq(Long userId) {
        return userId != null ? QOrder.order.user.id.eq(userId) : null;
    }

    private BooleanExpression amountGoe(Integer minAmount) {
        return minAmount != null ? QOrder.order.totalAmount.goe(minAmount) : null;
    }

    // 페이징 + count 최적화
    public Page<OrderDto> searchPage(OrderSearchCondition cond, Pageable pageable) {
        QOrder order = QOrder.order;

        List<OrderDto> content = queryFactory
                .select(Projections.constructor(OrderDto.class,
                        order.id, order.totalAmount, order.status))
                .from(order)
                .where(statusEq(cond.getStatus()))
                .offset(pageable.getOffset())
                .limit(pageable.getPageSize())
                .fetch();

        Long total = queryFactory
                .select(order.count())
                .from(order)
                .where(statusEq(cond.getStatus()))
                .fetchOne();

        return new PageImpl<>(content, pageable, total != null ? total : 0);
    }
}

// JPAQueryFactory 빈 등록
@Bean
public JPAQueryFactory jpaQueryFactory(EntityManager em) {
    return new JPAQueryFactory(em);
}
```

---

## @Transactional

```java
@Service
@Transactional(readOnly = true)  // 클래스 기본: 읽기 전용 (성능 최적화)
public class OrderService {

    @Transactional  // 쓰기 작업만 오버라이드
    public Order create(CreateOrderRequest request) {
        Order order = Order.create(request);
        return orderRepository.save(order);
    }

    public Order findById(Long id) {  // readOnly 상속
        return orderRepository.findById(id).orElseThrow();
    }

    // 전파 레벨
    @Transactional(propagation = Propagation.REQUIRES_NEW)  // 새 트랜잭션 시작
    public void logOrder(Order order) { ... }

    @Transactional(propagation = Propagation.NEVER)  // 트랜잭션 있으면 예외
    public Report generateReport() { ... }

    // 격리 수준
    @Transactional(isolation = Isolation.REPEATABLE_READ)
    public void criticalUpdate(Long id) { ... }

    // 롤백 제어
    @Transactional(rollbackFor = Exception.class)  // checked exception도 롤백
    @Transactional(noRollbackFor = BusinessWarningException.class)
    public void process() { ... }
}

// 주의: self-invocation은 트랜잭션 적용 안 됨 (프록시 우회)
// → 같은 클래스 내에서 @Transactional 메서드 호출 금지
// → ApplicationContext.getBean() 또는 별도 서비스 클래스로 분리
```
