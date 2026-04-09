# Debugging Skill

## 절차
1. 에러 메시지 정확히 읽기
2. 재현 조건 확인 (입력값, 환경)
3. 최소 재현 케이스 만들기
4. 이분법(bisect)으로 원인 좁히기
5. 가설 → 검증 → 수정 → 테스트

## 도구
- `grep_search`: 에러 메시지로 관련 코드 검색
- `file_read`: 스택 트레이스의 파일/라인 확인
- `shell_exec`: `git log --oneline -20`으로 최근 변경 확인
