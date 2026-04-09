# Spring Boot

## Web Layer

### REST Controllers

```java
@RestController
@RequestMapping("/api/users")
public class UserController {
    @GetMapping("/{id}")
    public User getUser(@PathVariable Long id) { }

    @PostMapping
    public User createUser(@RequestBody UserDTO dto) { }

    @PutMapping("/{id}")
    public User updateUser(@PathVariable Long id, @RequestBody UserDTO dto) { }

    @DeleteMapping("/{id}")
    public void deleteUser(@PathVariable Long id) { }
}
```

## Service Layer

```java
@Service
public class UserService {
    private final UserRepository repo;

    // Constructor injection (권장)
    public UserService(UserRepository repo) {
        this.repo = repo;
    }
}
```

**@Autowired** field injection은 테스트 어려움. Constructor injection 선호.

## Data Layer

### JPA Entity

```java
@Entity
@Table(name = "users")
public class User {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false, unique = true)
    private String email;

    @OneToMany(mappedBy = "user", cascade = CascadeType.ALL)
    private List<Order> orders;
}
```

### Repository

```java
@Repository
public interface UserRepository extends JpaRepository<User, Long> {
    Optional<User> findByEmail(String email);
}
```

## Configuration

**application.yml**:
```yaml
server:
  port: 8080

spring:
  datasource:
    url: jdbc:postgresql://localhost:5432/mydb
    username: user
    password: pass
  jpa:
    hibernate:
      ddl-auto: validate
    show-sql: true
```

## Exception Handling

```java
@ControllerAdvice
public class GlobalExceptionHandler {
    @ExceptionHandler(NotFoundException.class)
    public ResponseEntity<ErrorDTO> handleNotFound(NotFoundException ex) {
        return ResponseEntity.status(404).body(new ErrorDTO(ex.getMessage()));
    }
}
```

**Patterns**: Constructor injection > field injection. Use @ControllerAdvice for centralized error handling. Validate at controller boundary.
