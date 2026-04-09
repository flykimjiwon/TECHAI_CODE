# Go Standard Library Quick Reference

## fmt
- `fmt.Sprintf(format, args...)` — 포맷된 문자열 반환
- `fmt.Fprintf(w, format, args...)` — io.Writer에 출력
- `fmt.Errorf(format, args...)` — 에러 생성 (%w로 래핑)

## strings
- `strings.Contains(s, substr)` — 포함 여부
- `strings.Split(s, sep)` — 분리
- `strings.TrimSpace(s)` — 양쪽 공백 제거
- `strings.ReplaceAll(s, old, new)` — 전체 치환

## os
- `os.ReadFile(name)` — 파일 전체 읽기
- `os.WriteFile(name, data, perm)` — 파일 쓰기
- `os.Getenv(key)` — 환경변수
- `os.Getwd()` — 현재 디렉토리
