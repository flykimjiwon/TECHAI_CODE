# TECHAI IDE — Feature List (v0.1.0)

> 빌드: macOS 12MB / Windows 12MB
> 기술: Go + Wails v2 + React + TypeScript + CodeMirror 6
> 테스트: 30개 (Go 백엔드) — 전체 PASS

---

## 전체 84개 기능

### Core (7)
1. Activity Bar (Explorer/Search/Git/Settings/Account)
2. 파일 트리 (실제 프로젝트, 3단계 깊이)
3. 파일 트리 자동 새로고침 (AI file_write 시)
4. 파일 트리 컨텍스트 메뉴 (새파일/이름변경/삭제)
5. 폴더 열기 다이얼로그 (IDE + 시스템 메뉴)
6. 파일 아이콘 60종+
7. 폴더 아이콘 15종 색상 구분

### Editor (18)
8. CodeMirror 6 구문 강조 (15개 언어)
9. Go/TS/JS/Python/Rust/Java/PHP/C++/SQL/YAML/JSON/CSS/HTML/Markdown/Vue
10. 파일 탭 (다중, 닫기, 수정 표시)
11. Cmd+S 저장 (토스트)
12. Cmd+W 탭 닫기 (미저장 확인)
13. Cmd+F 파일 내 검색
14. Cmd+P 퍼지 파일 열기
15. 브레드크럼
16. Ln/Col + 언어 감지
17. 줄번호 + 활성줄 하이라이트
18. 괄호 매칭 + 자동 닫기
19. 코드 접기 (foldGutter)
20. 들여쓰기 자동 정리
21. Undo/Redo
22. Tab → 4 spaces
23. AI 파일 수정 시 에디터 새로고침
24. 스플릿 에디터 (Cmd+\\)
25. 자동 저장 (1.5초 디바운스)

### Preview (3)
26. 이미지 미리보기
27. 마크다운 미리보기 (Preview/Edit)
28. Live Server (HTML → 브라우저)

### AI Chat (11)
29. AI 채팅 (Qwen3 스트리밍)
30. 마크다운 렌더링 (코드블록 + 인라인)
31. 도구 10개
32. text tool_call 파싱 (3포맷)
33. .techai.md 자동 로드
34. Knowledge Packs 74개 (체크박스)
35. 슬래시 명령어 7개
36. 채팅 내보내기 (markdown)
37. 채팅 세션 저장/불러오기
38. 도구 호출 포매팅
39. Open in Browser

### Terminal (6)
40. 실제 PTY 터미널
41. 다중 탭
42. 쉘 선택 (zsh/bash/fish)
43. Ctrl+C/D
44. ANSI 제거
45. 명령어 히스토리 (↑↓)

### Git (10)
46. Git status 패널
47. Stage (+)
48. Commit
49. Inline Diff
50. Git Graph 시각화
51. 브랜치 목록 + 칩
52. 브랜치 checkout (클릭)
53. 새 브랜치 생성
54. Pull / Push
55. 알림 뱃지 (5초 갱신)

### Search (6)
56. 실시간 검색 (300ms)
57. Include 필터
58. Exclude 필터
59. 파일별 그룹핑 + 접기
60. 다중 매치 하이라이트
61. 결과 카운트

### UI/UX (13)
62. 11개 테마
63. Settings (API/Key/Model)
64. 패널 리사이즈 (드래그)
65. 알림 토스트
66. Welcome (단축키 + 최근 프로젝트)
67. About
68. 커맨드 팔레트 (Cmd+Shift+P)
69. 최근 프로젝트 (10개)
70. 앱 아이콘 (macOS + Windows)
71-74. 단축키 (Cmd+J/B/\\/트래픽라이트)

### System Menu (5)
75-79. File/Edit/View/Terminal/Help

### Shortcuts (15)
Cmd+S/W/P/Shift+P/F/Shift+F/B/J/\\/1/2/3/,/O/↑↓

---

**Total: 84 features | 30 tests | 12MB**
