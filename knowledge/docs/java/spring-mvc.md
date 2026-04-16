# Spring MVC — Controllers, Requests, Responses, Exception Handling

## Controller 기본

```java
// REST API 전용
@RestController  // @Controller + @ResponseBody
@RequestMapping("/api/v1/users")
public class UserController {

    private final UserService userService;

    public UserController(UserService userService) {
        this.userService = userService;
    }

    @GetMapping
    public List<UserDto> getAll() {
        return userService.findAll();
    }

    @GetMapping("/{id}")
    public ResponseEntity<UserDto> getById(@PathVariable Long id) {
        return userService.findById(id)
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }

    @PostMapping
    public ResponseEntity<UserDto> create(@Valid @RequestBody CreateUserRequest request,
                                          UriComponentsBuilder uriBuilder) {
        UserDto created = userService.create(request);
        URI location = uriBuilder.path("/api/v1/users/{id}")
                .buildAndExpand(created.getId()).toUri();
        return ResponseEntity.created(location).body(created);
    }

    @PutMapping("/{id}")
    public ResponseEntity<UserDto> update(@PathVariable Long id,
                                          @Valid @RequestBody UpdateUserRequest request) {
        return ResponseEntity.ok(userService.update(id, request));
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> delete(@PathVariable Long id) {
        userService.delete(id);
        return ResponseEntity.noContent().build();
    }
}

// 전통 MVC (뷰 반환)
@Controller
@RequestMapping("/users")
public class UserViewController {

    @GetMapping("/{id}")
    public String detail(@PathVariable Long id, Model model) {
        model.addAttribute("user", userService.findById(id).orElseThrow());
        return "users/detail";  // templates/users/detail.html (Thymeleaf)
    }

    @PostMapping
    public String create(@Valid @ModelAttribute CreateUserForm form,
                         BindingResult bindingResult) {
        if (bindingResult.hasErrors()) {
            return "users/create";  // 폼 재렌더링
        }
        userService.create(form);
        return "redirect:/users";
    }
}
```

---

## 요청 파라미터 바인딩

```java
@RestController
@RequestMapping("/api/products")
public class ProductController {

    // @PathVariable: /api/products/42
    @GetMapping("/{id}")
    public Product getById(@PathVariable Long id) { ... }

    // 경로 변수명과 파라미터명 다를 때
    @GetMapping("/code/{productCode}")
    public Product getByCode(@PathVariable("productCode") String code) { ... }

    // @RequestParam: /api/products?page=0&size=20&sort=name
    @GetMapping
    public Page<Product> list(
            @RequestParam(defaultValue = "0") int page,
            @RequestParam(defaultValue = "20") int size,
            @RequestParam(required = false) String sort) { ... }

    // @RequestParam으로 Map 수집
    @GetMapping("/search")
    public List<Product> search(@RequestParam Map<String, String> filters) { ... }

    // @RequestBody: JSON → 객체
    @PostMapping
    public Product create(@RequestBody @Valid CreateProductRequest request) { ... }

    // @RequestHeader
    @GetMapping("/secure")
    public String secure(@RequestHeader("X-API-Key") String apiKey) { ... }

    // @CookieValue
    @GetMapping("/profile")
    public String profile(@CookieValue(value = "sessionId", required = false) String sessionId) { ... }

    // @ModelAttribute: form data / query string → 객체 (자동 바인딩)
    @GetMapping("/filter")
    public List<Product> filter(@ModelAttribute ProductFilter filter) { ... }
}
```

---

## ResponseEntity 활용

```java
// 상태코드 + 헤더 + 바디 세밀 제어
@GetMapping("/{id}/download")
public ResponseEntity<Resource> download(@PathVariable Long id) {
    Resource file = fileService.load(id);
    return ResponseEntity.ok()
            .header(HttpHeaders.CONTENT_DISPOSITION,
                    "attachment; filename=\"" + file.getFilename() + "\"")
            .contentType(MediaType.APPLICATION_OCTET_STREAM)
            .body(file);
}

// 조건부 응답
@GetMapping("/{id}")
public ResponseEntity<UserDto> get(@PathVariable Long id,
                                    @RequestHeader(value = "If-None-Match", required = false) String etag) {
    UserDto user = userService.findById(id).orElseThrow();
    String currentEtag = "\"" + user.getVersion() + "\"";
    if (currentEtag.equals(etag)) {
        return ResponseEntity.status(HttpStatus.NOT_MODIFIED).build();
    }
    return ResponseEntity.ok()
            .eTag(currentEtag)
            .body(user);
}
```

---

## Validation (@Valid + BindingResult)

```java
// DTO에 제약 정의
public class CreateUserRequest {
    @NotBlank(message = "이름은 필수입니다")
    @Size(max = 50)
    private String name;

    @Email(message = "이메일 형식이 올바르지 않습니다")
    @NotBlank
    private String email;

    @Min(0) @Max(150)
    private int age;

    @Pattern(regexp = "^\\d{10,11}$", message = "전화번호 형식 오류")
    private String phone;
}

// REST: 검증 실패 시 400 자동 (MethodArgumentNotValidException)
@PostMapping
public UserDto create(@RequestBody @Valid CreateUserRequest request) { ... }

// MVC 폼: BindingResult로 에러 수동 처리
@PostMapping
public String create(@Valid @ModelAttribute CreateUserForm form,
                     BindingResult result, Model model) {
    if (result.hasErrors()) {
        result.getAllErrors().forEach(e -> System.out.println(e.getDefaultMessage()));
        return "form";
    }
    return "redirect:/success";
}

// 커스텀 Validator
@Component
public class UserValidator implements Validator {
    @Override
    public boolean supports(Class<?> clazz) {
        return CreateUserRequest.class.isAssignableFrom(clazz);
    }

    @Override
    public void validate(Object target, Errors errors) {
        CreateUserRequest req = (CreateUserRequest) target;
        if (req.getAge() < 14) {
            errors.rejectValue("age", "age.tooYoung", "14세 미만은 가입 불가");
        }
    }
}
```

---

## ExceptionHandler

```java
// 전역 예외 처리
@RestControllerAdvice  // @ControllerAdvice + @ResponseBody
public class GlobalExceptionHandler {

    // 도메인 예외
    @ExceptionHandler(EntityNotFoundException.class)
    public ResponseEntity<ErrorResponse> handleNotFound(EntityNotFoundException ex) {
        return ResponseEntity.status(HttpStatus.NOT_FOUND)
                .body(new ErrorResponse("NOT_FOUND", ex.getMessage()));
    }

    // 검증 실패 (@Valid)
    @ExceptionHandler(MethodArgumentNotValidException.class)
    public ResponseEntity<ErrorResponse> handleValidation(MethodArgumentNotValidException ex) {
        List<String> errors = ex.getBindingResult().getFieldErrors().stream()
                .map(fe -> fe.getField() + ": " + fe.getDefaultMessage())
                .collect(Collectors.toList());
        return ResponseEntity.badRequest()
                .body(new ErrorResponse("VALIDATION_FAILED", errors.toString()));
    }

    // 접근 거부
    @ExceptionHandler(AccessDeniedException.class)
    public ResponseEntity<ErrorResponse> handleAccessDenied(AccessDeniedException ex) {
        return ResponseEntity.status(HttpStatus.FORBIDDEN)
                .body(new ErrorResponse("FORBIDDEN", "접근 권한이 없습니다"));
    }

    // 모든 미처리 예외 (catch-all)
    @ExceptionHandler(Exception.class)
    public ResponseEntity<ErrorResponse> handleGeneral(Exception ex, HttpServletRequest req) {
        log.error("Unhandled exception at {}: {}", req.getRequestURI(), ex.getMessage(), ex);
        return ResponseEntity.internalServerError()
                .body(new ErrorResponse("INTERNAL_ERROR", "서버 오류가 발생했습니다"));
    }
}

// 에러 응답 DTO
public record ErrorResponse(String code, String message) {}
```

---

## Interceptor

```java
// HandlerInterceptor 구현
@Component
public class AuthInterceptor implements HandlerInterceptor {

    @Override
    public boolean preHandle(HttpServletRequest request,
                             HttpServletResponse response,
                             Object handler) throws Exception {
        String token = request.getHeader("Authorization");
        if (token == null || !tokenService.validate(token)) {
            response.setStatus(HttpStatus.UNAUTHORIZED.value());
            response.getWriter().write("{\"error\":\"Unauthorized\"}");
            return false;  // 요청 처리 중단
        }
        return true;  // 계속 진행
    }

    @Override
    public void postHandle(HttpServletRequest req, HttpServletResponse res,
                           Object handler, ModelAndView mv) {
        // 뷰 렌더링 전 처리 (REST에서는 거의 사용 안 함)
    }

    @Override
    public void afterCompletion(HttpServletRequest req, HttpServletResponse res,
                                Object handler, Exception ex) {
        // 요청 완료 후 정리 (로깅, MDC 클리어 등)
    }
}

// Interceptor 등록
@Configuration
public class WebMvcConfig implements WebMvcConfigurer {

    private final AuthInterceptor authInterceptor;
    private final LoggingInterceptor loggingInterceptor;

    @Override
    public void addInterceptors(InterceptorRegistry registry) {
        registry.addInterceptor(loggingInterceptor)
                .addPathPatterns("/**");

        registry.addInterceptor(authInterceptor)
                .addPathPatterns("/api/**")
                .excludePathPatterns("/api/auth/**", "/api/public/**");
    }
}
```

**Filter vs Interceptor 차이:**

| 구분 | Filter (Servlet) | Interceptor (Spring) |
|------|-----------------|---------------------|
| 위치 | DispatcherServlet 외부 | DispatcherServlet 내부 |
| 관리 | Servlet 컨테이너 | Spring 컨테이너 |
| Spring 빈 접근 | 제한적 | 완전 |
| 용도 | 인코딩, 보안 헤더, CORS | 인증, 로깅, 공통 처리 |

---

## MessageConverter / Content Negotiation

```java
// JSON 기본 설정 커스터마이징
@Configuration
public class WebConfig implements WebMvcConfigurer {

    @Override
    public void configureMessageConverters(List<HttpMessageConverter<?>> converters) {
        // Jackson 커스터마이징
        Jackson2ObjectMapperBuilder builder = new Jackson2ObjectMapperBuilder()
                .featuresToDisable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS)
                .modules(new JavaTimeModule())
                .propertyNamingStrategy(PropertyNamingStrategies.SNAKE_CASE);

        converters.add(new MappingJackson2HttpMessageConverter(builder.build()));
    }

    // CORS 전역 설정
    @Override
    public void addCorsMappings(CorsRegistry registry) {
        registry.addMapping("/api/**")
                .allowedOrigins("https://app.example.com")
                .allowedMethods("GET", "POST", "PUT", "DELETE")
                .allowCredentials(true)
                .maxAge(3600);
    }
}
```

---

## Async 컨트롤러

```java
// CompletableFuture 반환
@GetMapping("/async/{id}")
public CompletableFuture<UserDto> getAsync(@PathVariable Long id) {
    return userService.findByIdAsync(id);
}

// DeferredResult (외부 이벤트 연동)
@GetMapping("/events")
public DeferredResult<String> subscribe() {
    DeferredResult<String> result = new DeferredResult<>(30_000L);
    eventBus.register(result);
    result.onTimeout(() -> result.setErrorResult(
            ResponseEntity.status(HttpStatus.REQUEST_TIMEOUT).build()));
    return result;
}

// @Async (서비스 계층)
@Service
public class EmailService {
    @Async
    public CompletableFuture<Void> sendAsync(String to, String content) {
        // 별도 스레드에서 실행
        return CompletableFuture.runAsync(() -> doSend(to, content));
    }
}
```
