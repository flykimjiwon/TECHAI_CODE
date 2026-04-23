# GPT-OSS-120B vs Qwen3-Coder-30B — 택가이코드 실전 비교

> 2026.04.23 기준 · 택가이코드 v0.9.4 · 시스템 프롬프트 최적화 후 테스트

---

## 핵심 한 줄

**GPT-OSS-120B는 "추론으로 코딩 문제를 푸는 제너럴리스트", Qwen3-Coder-30B는 "에이전트 루프에서 손발이 빠른 코딩 스페셜리스트".**

---

## 1. 기본 스펙

| 항목 | GPT-OSS-120B | Qwen3-Coder-30B-A3B |
|------|:---:|:---:|
| 전체/활성 파라미터 | 117B / 5.1B (MoE) | 30.5B / 3.3B (MoE, 128 experts, 8 active) |
| 네이티브 컨텍스트 | 131K | **256K** (YaRN으로 1M 확장) |
| 학습 특성 | MMLU-Pro 90%, AIME 97.9% — 범용 추론 | 7.5T 토큰 (코드 70%), long-horizon Agent RL |
| Reasoning 모드 | reasoning_content (low/med/high) | **없음** — 의도적으로 thinking 제거 |
| 가격 (in/out per M) | $0.04 / $0.19 | $0.07 / $0.27 |
| Tool Calling | OpenAI 표준 | OpenAI 호환 (Novita 변환) |

**설계 철학 차이**: Qwen3-Coder는 에이전트 루프에서 매 턴 reasoning하면 레이턴시가 폭발하므로 thinking을 의도적으로 뺌. GPT-OSS-120B는 reasoning이 핵심 무기.

---

## 2. 코딩 벤치마크

| 벤치마크 | GPT-OSS-120B | Qwen3-Coder-30B | 해석 |
|----------|:---:|:---:|------|
| SWE-bench Verified | ~45~55% (추정) | 51.6% (OpenHands 100턴) | 거의 동급 |
| CodeForces | **82.1%** | 미공개 | GPT-OSS 우세 |
| AA Coding Index | 우위 | Intelligence Index 20 | GPT-OSS 우세 |
| Agentic Tool Use | 양호 | **강함** | Qwen 우세 |
| Repository-scale 이해 | 131K 제한 | **256K~1M** | **Qwen 압승** |
| Aider Polyglot | 41.8% | 미공개 | - |

---

## 3. 택가이코드 실측 결과 (2026.04.23)

### 테스트 조건
- demo-supersol 프로젝트 (Next.js + TypeScript)
- 시스템 프롬프트: 1,493B (최적화 후)
- auto-prefetch 활성
- 동일 시나리오 3개 순차 실행

### 시나리오별 결과

| | GPT-OSS-120B | Qwen3-Coder-30B |
|---|---|---|
| **S1: 홈 거래내역** | ◎ file_write 1회 | ◎ file_read 3 + grep 2 + file_write 1 |
| **S2: 금융 계좌목록** | ◎ file_write 1회 | ◎ file_write 1회 |
| **S3: 혜택 프로그레스바** | ◎ file_write 1회 | ◎ file_write 1회 |
| **총 토큰** | 18,050 | 23,167 |
| **총 비용** | $0.005 | $0.007 |
| **체감 속도** | 기준 (1x) | **1.3~1.5x 빠름** |
| **한국어 품질** | 5/10 | **8/10** |
| **총점** | **82/100** | **88/100** |

### 시스템 프롬프트 최적화 효과

| | 최적화 전 | 최적화 후 |
|---|---|---|
| 프롬프트 크기 | 8,902B (~2,500tok) | **1,493B (~430tok)** |
| GPT-OSS 성공률 | 60% (빈 응답, 루프) | **95%+** |
| Qwen3 도구 호출 | **실패** (텍스트로 흉내) | **정상** |

---

## 4. 작업 유형별 적합도

### GPT-OSS-120B가 강한 영역
- **알고리즘/경쟁 프로그래밍** — reasoning으로 단계별 사고
- **새 라이브러리 문서 읽고 적용** — 범용 추론 강함
- **디버깅 "왜 틀렸는지" 분석** — reasoning이 원인 추적에 도움
- **아키텍처 설계/계획 수립** — 구조적 thinking

### Qwen3-Coder-30B가 강한 영역
- **대형 코드베이스 리팩토링** — 256K~1M 컨텍스트
- **에이전트 루프 속도** — non-thinking, 매 턴 빠름
- **파일 수정 정확도** — 보수적이고 정확한 diff/edit
- **SQL/sh/코드 수정** — 코딩 전용 학습
- **한국어 응답** — 자연스러움

### 비슷한 영역
- **단일 함수 작성/버그 수정** — SWE-bench 50%대로 엇비슷
- **Knowledge 문서 검색 + 적용** — 둘 다 tool call로 가능

---

## 5. 자원 효율

| 지표 | GPT-OSS-120B | Qwen3-Coder-30B |
|------|:---:|:---:|
| 컨텍스트 | 131K | **256K** |
| Active 파라미터 | 5.1B | **3.3B** (가벼움) |
| Reasoning 오버헤드 | **30~50% 토큰 낭비** | 없음 |
| GPU 요구 (로컬) | H100 80GB | **24GB급 가능** (q4) |
| 응답 속도 | 기준 | **1.3~1.5x** |
| 사내망 배포 난이도 | 고 (H100급) | **중** (상대적 용이) |

---

## 6. 실전 배포 추천

### 단일 모델 선택 시

| 사용 패턴 | 추천 |
|----------|------|
| Cursor/Cline 스타일 에이전틱 IDE | **Qwen3-Coder** |
| "어려운 문제 하나 깊이 생각해서 풀어줘" | **GPT-OSS** |
| 모노레포 전체 이해 + 리팩토링 | **Qwen3-Coder** (컨텍스트) |
| 로컬 GPU 리소스 부족 | **Qwen3-Coder** |
| 일반 개발자 일상 사용 | **Qwen3-Coder** (80% 작업에서 우세) |

### 하이브리드 아키텍처 (추천)

```
┌─ Orchestrator ─────────┐     ┌─ Worker (병렬) ────────┐
│ GPT-OSS-120B           │ ──▶ │ Qwen3-Coder-30B       │
│ ─ high reasoning       │     │ ─ non-thinking         │
│ ─ 작업 분해/계획       │     │ ─ 빠른 실행            │
│ ─ 검증/리뷰            │     │ ─ 파일 edit/diff       │
│ ─ 소수 호출            │     │ ─ 다수 병렬 호출       │
└────────────────────────┘     └────────────────────────┘
```

**역할 분담 근거:**
- GPT-OSS는 reasoning이 필요한 **소수 호출**에만 → 속도 문제 상쇄
- Qwen3-Coder는 **다수의 병렬 코드 수정** → 속도+컨텍스트 둘 다 승리
- 각자 강한 곳에 배치해서 약점을 상호 보완

---

## 7. 주의사항

### Qwen3-Coder-30B

- SWE-bench 51.6%는 **OpenHands 100턴 스캐폴딩 + 특정 vllm 버전 + 특정 tool call parser** 조합에 의존
- 시스템 프롬프트가 길면 **3.3B active가 도구 호출을 못 함** (텍스트로 흉내)
- 택가이코드에서는 **프롬프트 1,500B 이하**를 유지해야 안정적
- Nebius 연구에서 30B는 tool call formatting에 추가 post-processing 필요 보고

### GPT-OSS-120B

- reasoning_content가 **별도 필드**로 옴 → 택가이코드에서 별도 처리 구현 완료
- Thinking이 **영어로만 진행** → 한국어 환경에서 비효율
- 같은 thinking을 **반복하는 경향** → 토큰 낭비
- 131K 컨텍스트에서 **프롬프트 + 도구 결과 + 히스토리**가 빠르게 누적

---

## 8. 향후 아키텍처 방향

시스템 프롬프트에 의존하지 않고 **프로그램 자체의 로직으로 품질 보장**.

| 현재 (프롬프트 의존) | 미래 (코드 로직) |
|---|---|
| "NEVER force push" 규칙 | shell.go에서 패턴 차단 |
| "Read before edit" 규칙 | file_edit 호출 시 자동 file_read 선행 |
| 거대 시스템 프롬프트 | 미니멀 프롬프트 + 룰베이스 |
| 단일 모델이 전부 | 마이크로 에이전트 분산 (Qwen3-Coder급) |
| 전체 문서 주입 | RAG 정밀 주입 (관련 섹션만) |

---

*택가이코드 v0.9.4 · Tech혁신Unit 개발 AX CELL*
