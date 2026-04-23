# TECHAI IDE — 수동 테스트 체크리스트

> Go 테스트 (49개)는 `go test ./...`로 자동 검증.
> 아래는 UI/프론트엔드 수동 테스트 항목.

---

## 파일 트리
- [ ] 폴더 클릭 → 펼침/접기
- [ ] 파일 클릭 → 에디터에서 열림
- [ ] 파일 아이콘 색상 구분 (Go=시안, TS=파랑, JS=노랑 등)
- [ ] 폴더 아이콘 색상 구분 (src=파랑, test=초록 등)
- [ ] 우클릭 → New File / Rename / Delete
- [ ] HTML 파일 우클릭 → Open in Browser / Live Server
- [ ] 상단 📁+ 아이콘 → 다른 폴더 열기
- [ ] 상단 🔄 아이콘 → 새로고침
- [ ] AI가 file_write 후 트리 자동 갱신

## 에디터
- [ ] 파일 열기 → 구문 강조 (Go/TS/JS/Python 각각)
- [ ] 여러 파일 탭 열기 → 탭 전환
- [ ] 수정 시 탭에 노란 ● 표시
- [ ] Cmd+S → 저장 + 토스트 "Saved"
- [ ] Cmd+W → 탭 닫기 (미저장 시 확인)
- [ ] Cmd+F → 찾기 바
- [ ] 줄번호 표시 + 활성줄 하이라이트
- [ ] 브레드크럼 경로 표시
- [ ] Ln/Col 하단 표시
- [ ] 괄호 매칭 하이라이트
- [ ] 코드 접기 (함수 옆 ▾)
- [ ] Tab → 4칸 들여쓰기
- [ ] Undo (Cmd+Z) / Redo (Cmd+Shift+Z)
- [ ] 스플릿 에디터 (Cmd+\\)
- [ ] Auto/Manual 저장 토글
- [ ] .md 파일 → Preview/Edit 토글
- [ ] 이미지 파일 → 미리보기

## AI 채팅
- [ ] 메시지 입력 → AI 스트리밍 응답
- [ ] 응답 중 커서 깜빡임
- [ ] 코드블록 구문 강조 (``` 마크다운)
- [ ] 인라인 코드 (`code`) 렌더링
- [ ] >> 도구 호출 파란색 표시
- [ ] << 도구 결과 초록색 표시
- [ ] 📖 Knowledge Packs 체크박스 토글
- [ ] /clear → 대화 초기화
- [ ] /export → markdown 파일 저장
- [ ] /save test → 세션 저장
- [ ] /sessions → 저장 목록
- [ ] /load [id] → 세션 불러오기
- [ ] /model → 현재 모델 표시
- [ ] /help → 명령어 목록
- [ ] ⬇ 내보내기 버튼
- [ ] 🗑 초기화 버튼

## 터미널
- [ ] 명령어 입력 → 실행 결과 표시
- [ ] ls, pwd, echo 등 기본 명령
- [ ] Ctrl+C → 프로세스 중단
- [ ] ↑/↓ → 명령어 히스토리
- [ ] + 탭 → 새 터미널
- [ ] X → 탭 닫기
- [ ] 쉘 드롭다운 → zsh/bash 전환
- [ ] ANSI 코드 안 보임

## Git
- [ ] Cmd+3 → Git Graph 표시
- [ ] 커밋 히스토리 목록
- [ ] 브랜치 칩 표시 + HEAD 하이라이트
- [ ] 브랜치 클릭 → checkout + 토스트
- [ ] + branch → 새 브랜치 생성
- [ ] Pull / Push 버튼
- [ ] 좌측 Git 패널 → 변경 파일 목록
- [ ] Stage (+) 버튼
- [ ] Commit 메시지 입력 + Enter
- [ ] 파일 클릭 → Inline Diff 표시
- [ ] Activity Bar Git 뱃지 (변경파일 수)

## 검색
- [ ] Cmd+2 → Search 패널
- [ ] 2글자 이상 → 실시간 검색
- [ ] 파일별 그룹핑 + 카운트
- [ ] 매치 노란색 하이라이트
- [ ] 파일명 클릭 → 에디터에서 열기
- [ ] 🔽 필터 → Include/Exclude
- [ ] ▾/▸ 파일 접기/펼치기

## UI/UX
- [ ] 톱니바퀴 → 11개 테마 선택 (각각 확인)
- [ ] 사람 아이콘 → Settings (API URL/Key/Model)
- [ ] Cmd+Shift+P → 커맨드 팔레트
- [ ] Cmd+P → 파일 빠른 열기
- [ ] Cmd+B → 사이드바 토글
- [ ] Cmd+J → 터미널 토글
- [ ] 패널 리사이즈 (사이드바/터미널/채팅 경계 드래그)
- [ ] 토스트 알림 (저장/에러 시 우하단)
- [ ] Welcome 화면 → 단축키 + 최근 프로젝트

## 시스템 메뉴 (macOS)
- [ ] File → Open Folder
- [ ] File → Save (Cmd+S)
- [ ] File → Close Tab (Cmd+W)
- [ ] Edit → Find (Cmd+F)
- [ ] Edit → Find in Files (Cmd+Shift+F)
- [ ] View → Quick Open (Cmd+P)
- [ ] View → Toggle Terminal (Cmd+J)
- [ ] Help → About

---

> 전부 체크 완료 후 FEATURES.md의 기능 수와 일치하는지 확인
