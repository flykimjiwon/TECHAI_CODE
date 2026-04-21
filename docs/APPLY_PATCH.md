# apply_patch — 멀티파일 패치 도구

## 개요

하나의 tool call로 여러 파일을 동시에 생성/수정/삭제/이동할 수 있는 패치 도구.
기존 `file_edit`가 파일당 1회 호출이 필요했던 것과 달리, `apply_patch`는 단일 호출로
모든 변경을 원자적으로 처리합니다.

Codex CLI의 apply_patch 포맷에서 영감을 받아 구현했으며, TGC의 기존 스냅샷/undo
시스템과 통합되어 있습니다.

## 패치 포맷

```
*** Begin Patch
*** Add File: path/to/new.go
+package main
+
+func main() {}

*** Update File: src/app.go
*** Move to: src/main.go
@@ func Run() {
-    log.Println("old")
+    log.Println("new")

*** Delete File: deprecated/old.go
*** End Patch
```

## 연산 타입

### Add File
새 파일을 생성합니다. 이미 존재하면 에러를 반환합니다.
각 줄은 `+` 접두사로 시작합니다.

```
*** Add File: internal/hooks/hooks.go
+package hooks
+
+func NewManager() *Manager {
+    return &Manager{}
+}
```

### Update File
기존 파일을 수정합니다. `@@ 앵커`로 변경 위치를 지정하고,
`-`로 제거할 줄, `+`로 추가할 줄을 명시합니다.

```
*** Update File: cmd/main.go
@@ func main() {
-    fmt.Println("hello")
+    fmt.Println("hello, world")
```

**@@ 앵커 매칭 규칙:**
1. 정확 매칭: 파일에서 앵커 문자열을 포함하는 첫 번째 줄
2. Fuzzy 매칭: 공백 정리 후 비교 (탭/스페이스 차이 허용)
3. 앵커 없으면 파일 처음부터 탐색

**Remove 라인 매칭:**
1. 정확 매칭 (trailing whitespace 무시)
2. Fuzzy 매칭 (leading/trailing whitespace 전부 무시)

### Delete File
파일을 삭제합니다. 삭제 전 자동으로 스냅샷이 생성됩니다.

```
*** Delete File: old/deprecated.go
```

### Move to (Update와 조합)
파일 수정 후 이동(이름 변경)합니다.

```
*** Update File: pkg/old_name.go
*** Move to: pkg/new_name.go
@@ func Handler() {
-    return nil
+    return &handler{}
```

## file_edit와의 비교

| | file_edit | apply_patch |
|---|---|---|
| 파일 수 | 1개/호출 | N개/호출 |
| 파일 생성 | file_write 별도 호출 | Add File로 통합 |
| 파일 삭제 | shell_exec rm 필요 | Delete File로 통합 |
| 파일 이동 | shell_exec mv 필요 | Move to로 통합 |
| 위치 지정 | old_string 전체 매칭 | @@ 함수 앵커 (줄번호 밀려도 OK) |
| 스냅샷 | 자동 | 자동 |
| Fuzzy 매칭 | 4단계 | 앵커+라인 2단계 |

## 사용 시점

- 2개 이상 파일을 동시에 수정할 때
- 파일 생성 + 수정 + 삭제가 함께 필요할 때
- 리팩토링으로 파일을 이동할 때

1개 파일의 간단한 수정은 기존 `file_edit`가 여전히 적합합니다.

## 구현 파일

- `internal/tools/apply_patch.go` — 파서 + 적용 로직
- `internal/tools/registry.go` — 도구 등록 및 디스패치
