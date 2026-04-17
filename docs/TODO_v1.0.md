# v1.0 TODO — Knowledge Pack System

## 현재 문제
- 81개 문서가 전부 바이너리에 내장 (~5MB)
- 프로젝트와 무관한 문서도 검색 대상 (Go 프로젝트에서 Spring 문서 등)
- 문서 수정 시 재빌드 필요

## 개선 방향: 선택식 지식 팩

```
바이너리 [코어 엔진만] + .tgc/knowledge/packs/ [선택한 팩만]
```

### 구조
```
.tgc/knowledge/
├── packs/              ← 지식 팩 (선택식)
│   ├── spring.md
│   ├── react.md
│   ├── bxm.md
│   ├── go.md
│   ├── sql.md
│   └── shell.md
└── custom/             ← 사용자 직접 추가
    └── my-api-docs.md
```

### /init 시 자동 선택
- 프로젝트 타입 감지 (go.mod → Go 팩, package.json → React 팩 등)
- 또는 수동 선택: `/knowledge add spring` `/knowledge remove bxm`

### 기대 효과
- 바이너리 ~25MB → ~20MB
- 프롬프트 토큰 절감 (불필요한 문서 검색 제거)
- 문서 업데이트 = 파일 교체만 (재빌드 불필요)
- 사내 팀별 커스텀 지식 팩 배포 가능
- USB에 바이너리 + 지식 팩 따로 전달

### 구현 순서
1. 지식 팩 포맷 정의 (카테고리별 단일 .md)
2. /init에서 프로젝트 타입 기반 자동 선택
3. /knowledge 명령어 (add/remove/list)
4. 바이너리에서 내장 문서 제거, 외부 팩으로 이전
5. 첫 실행 시 기본 팩 자동 다운로드 또는 번들
