# 택가이코드 Knowledge Packs

내장 지식 문서를 프로젝트별로 선택하여 사용하는 시스템.

바이너리에 74개 문서가 포함되어 있지만, **프로젝트에서 필요한 팩만 활성화**하여
시스템 프롬프트를 경량화하고 검색 정확도를 높입니다.

---

## 작동 방식

```
바이너리 빌드 시: 74개 .md 파일이 embed.FS로 포함
                       ↓
/init 실행:      프로젝트 타입 자동 감지 (go/node/python/java)
                       ↓
.techai.md 생성:  knowledge_packs: react, css, typescript  ← 자동 추천
                       ↓
앱 시작:          .techai.md의 팩만 인덱싱 (나머지 무시)
                       ↓
사용자 질문:      knowledge_search 도구로 활성화된 팩에서만 검색
```

---

## 사용법

### 1. 자동 설정 (`/init`)

프로젝트 폴더에서 택가이코드를 실행하고 `/init`을 입력하면:

```
$ cd ~/my-nextjs-project
$ techai
> /init
```

`.techai.md`가 자동 생성되며, 프로젝트 타입에 맞는 팩이 추천됩니다:

```markdown
## Knowledge Packs

knowledge_packs: javascript, typescript, react, css, terminal

사용 가능한 팩: auth, bxm, charts, css, database, go, java, javascript,
python, react, terminal, testing, tooling, typescript, utils, validation, vue
```

### 2. 수동 편집

`.techai.md`를 열고 `knowledge_packs:` 줄을 직접 수정:

```markdown
# 예: Spring + Oracle + BXM 프로젝트
knowledge_packs: java, database, bxm, terminal

# 예: React + PostgreSQL 프로젝트
knowledge_packs: react, css, typescript, database, terminal

# 예: 팩 비활성화 (지식 없이 사용)
knowledge_packs: none

# 예: 줄 자체를 삭제하면 → 전체 74개 로드 (이전 동작)
```

### 3. 프로젝트 타입별 자동 추천 규칙

| 프로젝트 타입 | 감지 기준 | 자동 추천 팩 |
|-------------|----------|------------|
| **Next.js** | package.json + next | javascript, typescript, react, css, terminal |
| **React (Vite)** | package.json + react | javascript, typescript, react, css, terminal |
| **Vue/Nuxt** | package.json + vue | javascript, typescript, vue, css, terminal |
| **Go** | go.mod | go, terminal |
| **Python** | requirements.txt | python, terminal |
| **Java/Spring** | pom.xml / build.gradle | java, terminal |
| **미감지** | - | terminal |

### 4. 추가 팩을 수동으로 넣고 싶을 때

자동 추천에 포함 안 된 팩을 `.techai.md`에 추가하면 됩니다:

```markdown
# 자동 추천: javascript, typescript, react, css, terminal
# 수동 추가: database, testing, charts
knowledge_packs: javascript, typescript, react, css, terminal, database, testing, charts
```

---

## 팩 목록 (전체 17종, 74개 문서)

### 프레임워크 / 언어

| 팩 | 파일 수 | 내용 |
|----|--------|------|
| `react` | 11 | React 19 Hooks, Next.js (App/Pages Router), React Hook Form, TanStack Query, Zustand, 성능 최적화, 패턴 |
| `vue` | 1 | Vue 3 Composition API |
| `go` | 1 | Go 표준 라이브러리 (net/http, context, sync, io, encoding 등) |
| `java` | 7 | Spring Core/MVC/Security/Data JPA/Boot Ops/Test, JEUS WAS |
| `python` | 1 | FastAPI |
| `javascript` | 1 | ES2024+ (구조 분해, Optional Chaining, Top-level Await 등) |
| `typescript` | 1 | TypeScript 타입 시스템 (Generics, Utility Types, 고급 패턴) |

### 스타일 / UI

| 팩 | 파일 수 | 내용 |
|----|--------|------|
| `css` | 5 | Tailwind CSS v4 (기본+고급), Framer Motion, Radix UI, shadcn/ui |
| `charts` | 4 | Chart.js, D3.js, ECharts, Recharts |

### 데이터베이스

| 팩 | 파일 수 | 내용 |
|----|--------|------|
| `database` | 9 | SQL Core (JOIN/CTE/윈도우함수), DDL/설계, PostgreSQL, MySQL, Oracle, Tibero, Prisma, Drizzle, Supabase |

### 인프라 / 도구

| 팩 | 파일 수 | 내용 |
|----|--------|------|
| `terminal` | 11 | Bash 스크립팅, Shell 도구 (sed/awk/jq/curl), Git, Linux/macOS/Windows, 크로스플랫폼 |
| `tooling` | 3 | ESLint (Flat Config), Turborepo, Vite |
| `testing` | 3 | Vitest, Testing Library, Playwright |

### 기업용 / 특수

| 팩 | 파일 수 | 내용 |
|----|--------|------|
| `bxm` | 13 | BXM 프레임워크 (배치, Bean, Config, DBIO, Service, 트랜잭션 등) |
| `auth` | 1 | Auth.js (NextAuth v5) |
| `validation` | 1 | Zod |
| `utils` | 1 | date-fns |

---

## 기존 사용자 지식 (`.tgc/knowledge/`)

Knowledge Packs는 **바이너리에 내장된** 문서입니다.
사용자가 직접 작성한 문서는 별도로 `.tgc/knowledge/` 폴더에 넣으면 됩니다.

```
.tgc/knowledge/
├── my-api-guide.md       ← 사용자가 직접 작성한 문서
├── team-conventions.md   ← 팀 코딩 컨벤션
└── db-schema-notes.md    ← DB 스키마 메모
```

이 문서들은 팩 선택과 무관하게 **항상 인덱싱**됩니다.
`knowledge_search` 도구로 내장 팩 + 사용자 문서 모두 검색됩니다.

---

## FAQ

**Q: knowledge_packs 줄이 없으면?**
A: 전체 74개 문서가 로드됩니다 (이전 동작과 동일).

**Q: `none`으로 설정하면?**
A: 내장 팩이 아무것도 로드되지 않습니다. 사용자 문서(.tgc/knowledge/)만 사용.

**Q: 팩을 바꾸면 바로 반영되나?**
A: 택가이코드를 재시작해야 합니다 (팩은 시작 시 인덱싱).

**Q: 팩에 없는 기술 스택을 물어보면?**
A: 모델의 기본 학습 지식으로 답변합니다. 팩은 보충 자료일 뿐, 없어도 기본적인 답변은 가능합니다.
