# Search Architecture — v0.9.4

## 전체 사용자-모델-도구-출력 상세 흐름

```
┌──────────────────────────────────────────────────────────────┐
│  사용자 입력                                                  │
│  "RWA_IBS_DMB_CMM_MAS 테이블의 DMB_K 컬럼 사용하는 프로그램"  │
└────────────────────────────┬─────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────┐
│  1. 전처리 (app.go sendMessage)                            │
│                                                            │
│  ├─ ResetFailedPatterns()  ← 실패 패턴 캐시 초기화          │
│  ├─ 메모리 주입 (memories.json → 시스템 프롬프트)            │
│  ├─ .techai.md 로드 (프로젝트 컨텍스트)                      │
│  ├─ 지식 문서 검색 (81개 중 관련 문서 주입)                   │
│  ├─ 히스토리 Compact (90% 넘으면 LLM 요약)                   │
│  └─ Multi 체크 (기본 OFF → 스킵)                            │
└────────────────────────────┬───────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────┐
│  2. API 요청 (llm/client.go StreamChat)                    │
│                                                            │
│  POST https://endpoint/v1/chat/completions                 │
│  {                                                         │
│    model: "GPT-OSS-120B",                                  │
│    messages: [시스템프롬프트, 이전대화, 사용자질문],           │
│    tools: [14개 도구 정의],                                  │
│    stream: true                                            │
│  }                                                         │
│                                                            │
│  ┌─ 45초 무응답? → 자동 재시도 (최대 3회)                    │
│  └─ 응답 시작 → 청크 스트리밍                                │
└────────────────────────────┬───────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────┐
│  3. 스트리밍 수신 (app.go streamChunkMsg)                   │
│                                                            │
│  화면: ⠋ Connecting... (0.5s)                               │
│  화면: ⠹ Streaming (2.1s · 150tok · 35tok/s)                │
│                                                            │
│  모델이 판단:                                               │
│  ├─ 텍스트 응답? → streamBuf에 축적 → 완료 시 화면 출력      │
│  └─ 도구 호출? → toolCalls 감지 → 도구 실행 단계로           │
└────────────────────────────┬───────────────────────────────┘
                             │
              (도구 호출 감지)│
                             ▼
┌────────────────────────────────────────────────────────────┐
│  4. 도구 호출 표시                                          │
│                                                            │
│  >> searching DMB_K                                        │
│  >> reading file.sh                                        │
│                                                            │
│  화면: Processing... (tool 1/20)                            │
└────────────────────────────┬───────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────┐
│  5. 도구 실행 (tools/registry.go Execute)                   │
│                                                            │
│  예: grep_search("RWA_IBS_DMB_CMM_MAS.*DMB_K")             │
│                                                            │
│  ┌─ 방어1: 패턴 중복 체크 (같은 패턴 2회 실패 → 즉시 차단)   │
│  │   └─ 다른 패턴은 무제한 허용                              │
│  │                                                          │
│  ├─ ripgrep 시도 (rg 있으면 100배 빠름)                      │
│  │   └─ 결과 있으면 반환                                     │
│  │                                                          │
│  ├─ Go grep 병렬 실행 (8 goroutine, 세마포어)                │
│  │   ├─ 파일별 3초 타임아웃                                   │
│  │   └─ 결과 파일별 그룹핑 + 컨텍스트 2줄                     │
│  │                                                          │
│  ├─ A.*B 패턴 + No matches?                                  │
│  │   ├─ co_search(A, B) ← 교집합 자동 계산                   │
│  │   │   ├─ 파일1: A ✓ B ✓ → 교집합 ✓                        │
│  │   │   ├─ 파일2: A ✓ B ✗ → 제외                            │
│  │   │   └─ 결과: "2 files containing BOTH"                  │
│  │   └─ 교집합 없으면 → 개별 결과 fallback                    │
│  │                                                          │
│  └─ 전체 30초 타임아웃                                       │
└────────────────────────────┬───────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────┐
│  6. 도구 결과 표시 + 히스토리 추가                           │
│                                                            │
│  화면:                                                     │
│  << grep_search:                                           │
│    file1.sh (2 matches):                                   │
│      :22  ... RWA_IBS_DMB_CMM_MAS ...                      │
│      :208 ... DMB_K /* 담보종류 */ ...                       │
│                                                            │
│  Processing... (tool 1/20)                                 │
│                                                            │
│  ├─ 결과를 히스토리에 추가 (tool role)                       │
│  ├─ 도구 루프 20회 제한 체크                                  │
│  ├─ git 상태 갱신                                           │
│  └─ continueAfterTools() → 다시 API 호출 (스텝 2로)          │
└────────────────────────────┬───────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────┐
│  7. 모델 2차 응답 (도구 결과 기반)                           │
│                                                            │
│  모델이 도구 결과를 보고:                                    │
│  ├─ 충분한 정보? → 텍스트 답변 생성 ✅                        │
│  ├─ 더 필요? → 추가 도구 호출 (스텝 4로, 최대 20회)          │
│  │   └─ 다른 키워드, 다른 include로 검색 (무제한)            │
│  └─ file_read(offset=200, limit=30) → 특정 라인 확인        │
└────────────────────────────┬───────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────┐
│  8. 최종 답변 출력                                          │
│                                                            │
│  ▎ RWA_IBS_DMB_CMM_MAS 테이블의 DMB_K 컬럼을 사용하는       │
│  ▎ 프로그램:                                                │
│  ▎                                                         │
│  ▎ v_brwa_ibsd_cusgt33.sh                                  │
│  ▎   • line 208: INSERT ... DMB_K /* 담보종류 */            │
│  ▎                                                         │
│  ┌──────────────────────────────────────────────┐           │
│  │ Super GPT-OSS-120B Tool:ON(14) 888tok $0.003 │           │
│  └──────────────────────────────────────────────┘           │
└────────────────────────────────────────────────────────────┘
```

---

## 7겹 방어 레이어

```
┌─────────────────────────────────────────────────┐
│              사용자 질문                          │
└──────────────────┬──────────────────────────────┘
                   │
    ┌──────────────▼──────────────┐
    │  방어 1: 패턴 중복 차단      │ 같은 패턴 2회 실패 → 즉시 차단
    │  (다른 패턴은 무제한 허용)    │ 다른 키워드/include 변경 = OK
    └──────────────┬──────────────┘
                   │
    ┌──────────────▼──────────────┐
    │  방어 2: ripgrep 우선        │ rg 있으면 100배 빠르게
    └──────────────┬──────────────┘
                   │
    ┌──────────────▼──────────────┐
    │  방어 3: 병렬 8 goroutine    │ 파일별 3초 타임아웃
    └──────────────┬──────────────┘
                   │
    ┌──────────────▼──────────────┐
    │  방어 4: A.*B → co_search    │ 교집합 자동 계산 (핵심!)
    └──────────────┬──────────────┘
                   │
    ┌──────────────▼──────────────┐
    │  방어 5: 결과 그룹핑         │ 파일별 + 컨텍스트 2줄
    └──────────────┬──────────────┘
                   │
    ┌──────────────▼──────────────┐
    │  방어 6: 도구 30초 타임아웃   │ 무한 블로킹 방지
    └──────────────┬──────────────┘
                   │
    ┌──────────────▼──────────────┐
    │  방어 7: API 45초 자동 재시도 │ 3회까지 자동 복구
    └──────────────┬──────────────┘
                   │
         ┌─────────▼─────────┐
         │    답변 출력       │
         └───────────────────┘
```

---

## Before vs After 비교

### 복합 키워드 검색 (TABLE + COLUMN)

```
Before (루프 발생):
  grep("TABLE.*COLUMN") → 실패
  grep("TABLE.*COLUMN") → 실패 (반복)
  file_read(전체) → 못 찾음
  grep("TABLE.*COLUMN") → 실패 (반복)
  → ∞ 무한 루프

After (교집합 → 확정):
  grep("TABLE.*COLUMN") → 실패
  → 자동 co_search("TABLE","COLUMN")
  → 교집합: file1.sh (TABLE: line 22, COLUMN: line 208)
  → 확정 결과 → 답변 → 끝 ✅
```

### 같은 패턴 재시도

```
Before:
  grep("TABLE.*COLUMN") → 실패
  grep("TABLE.*COLUMN") → 실패
  grep("TABLE.*COLUMN") → 실패
  → ∞

After:
  grep("TABLE.*COLUMN") → 실패 (1회)
  grep("TABLE.*COLUMN") → 차단 "Already searched — try DIFFERENT keyword"
  → 모델이 전략 변경 (file_read 또는 다른 키워드) ✅
```

### API 무응답

```
Before:
  API 호출 → 무한 대기 → 사용자가 Ctrl+C

After:
  API 호출 → 45초 → "Retrying... (1/3)" → 재연결
  → 45초 → "Retrying... (2/3)" → 재연결
  → 45초 → "Retrying... (3/3)" → 재연결
  → 45초 → "Error: no response" → 멈춤 (무한 X) ✅
```

---

## grep 결과 출력 형식

### Before
```
file.sh:15:DMB_K
file.sh:208:FROM RWA_IBS_DMB_CMM_MAS
other.sh:42:DMB_K
```

### After (그룹핑 + 컨텍스트)
```
file.sh (2 matches):
  :13   ... SELECT
  :14   ... T1.CMN_K,
  :15 > ... T1.DMB_K /* 담보종류 */,
  :16   ... T1.GDS_C
  :17   ... FROM MSD_TABLE
  :206  ... DELETE FROM
  :207  ... ${SCHEMA}.RWA_IBS_DMB_CMM_MAS
  :208> ... WHERE DW_BAS_DDT = '${V_BAS_DT}'
  :209  ... AND GRPCO_C = '${GRPCO_C}'
  :210  ... ;

other.sh (1 match):
  :40   ... INSERT INTO
  :41   ... ${SCHEMA}.RWA_IBS_DMB_CMM_MAS
  :42 > ... (DMB_K, CMN_K, ...)
  :43   ... SELECT
  :44   ... T1.DMB_K

(3 matches in 2 files, 247 files scanned)
```

---

## 도구 목록 (모델에 보이는 14개)

| # | Tool | Description |
|---|------|-------------|
| 1 | file_read | Read file (offset/limit for targeted sections) |
| 2 | file_write | Create/overwrite file (auto snapshot) |
| 3 | file_edit | Fuzzy 4-stage edit (auto snapshot + diff) |
| 4 | list_files | Directory listing |
| 5 | shell_exec | Shell command (30s timeout, risky warnings) |
| 6 | grep_search | Regex search (ripgrep, parallel, grouped, auto co_search) |
| 7 | glob_search | File pattern finder (mtime sorted) |
| 8 | hashline_read | Hash-anchored line reading |
| 9 | hashline_edit | Hash-anchored safe editing |
| 10 | git_status | Git status |
| 11 | git_diff | Git diff |
| 12 | git_log | Git log |
| 13 | diagnostics | Project linter (Go/TS/JS/Python) |
| 14 | knowledge_search | 81 embedded docs + user docs |

### 숨겨진 내부 도구 (모델에 안 보임, 프로그램이 자동 사용)

| Tool | 자동 발동 조건 |
|------|---------------|
| co_search | grep("A.*B") 실패 시 자동 교집합 |
| symbol_search | (내부 호출 가능) |
| fuzzy_find | (내부 호출 가능) |

---

## 패턴 중복 차단 규칙

```
캐시 키 = pattern 문자열

grep("TABLE.*COLUMN")          → 1회 실패 → 기록
grep("TABLE.*COLUMN")          → 2회 차단 ✅ "Already searched"

grep("DMB_K")                  → OK ✅ (다른 키)
grep("DMB_K", include="*.sh")  → OK ✅ (다른 키)
grep("TABLE")                  → OK ✅ (다른 키)
grep("TABLE", include="*.sql") → OK ✅ (다른 키)

→ 같은 패턴만 차단. 다른 키워드/include 조합은 무제한.
→ 새 사용자 메시지마다 캐시 초기화.
```

---

## 각 단계 소요 시간 (사내망 기준)

| 단계 | 동작 | 예상 시간 |
|------|------|----------|
| 1 | 전처리 (메모리, 컨텍스트) | ~100ms |
| 2 | API 요청 전송 | ~200ms |
| 3 | 첫 청크 수신 | 1-5초 |
| 4 | 도구 호출 감지 | 즉시 |
| 5 | 도구 실행 (grep+co_search) | 1-3초 |
| 6 | 결과 표시 | 즉시 |
| 7 | 2차 API 응답 | 3-10초 |
| 8 | 최종 답변 | 즉시 |
| **총** | | **5-20초** |

---

## 파일 위치

| 파일 | 역할 |
|------|------|
| `internal/app/app.go` | TUI, 스트리밍, 도구 실행, 재시도 |
| `internal/tools/registry.go` | 도구 등록, 실행, 패턴 차단, auto co_search |
| `internal/tools/search.go` | 병렬 grep, 그룹핑, 파일 타임아웃 |
| `internal/tools/cosearch.go` | 교집합 검색 (multi-line) |
| `internal/tools/ripgrep.go` | ripgrep 폴백 |
| `internal/tools/fuzzy.go` | 퍼지 파일명 매칭 |
| `internal/tools/symbols.go` | 함수/클래스 검색 |
| `internal/llm/client.go` | API 스트리밍 |
| `internal/llm/compaction.go` | 히스토리 압축 |
| `internal/llm/prompts/super.md` | 시스템 프롬프트 (최소화) |
| `internal/llm/prompts/core.md` | 핵심 지시 |
