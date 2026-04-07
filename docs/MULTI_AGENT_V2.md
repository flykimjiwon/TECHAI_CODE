# TECHAI v2 — Multi-Agent Architecture Plan

> 현재 v1은 단일 에이전트(gpt-oss-120b)로 동작.
> v2에서 Orchestrator + Worker 멀티에이전트로 진화 예정.

---

## 현재 상태 (v1)

```
User ──▶ [gpt-oss-120b] ──▶ Tools (file_read, file_write, file_edit, list_files, shell_exec)
          단일 에이전트
          모든 모드 동일 모델
```

**결정 근거 (2026-04 벤치마크 기준)**

| 지표 | gpt-oss-120b | Qwen3-Coder-30B |
|------|-------------|-----------------|
| Intelligence (ELO) | 33 | 20 |
| Output Speed (tok/s) | 236 | 28 |
| Price ($/M tokens) | $0.30 | $1.43 |
| Context Window | 131K | 262K |
| Parameters (Active) | 5.1B (MoE) | 30B (Dense) |

gpt-oss-120b가 지능, 속도, 가격 모두 우위. 유일한 열세는 컨텍스트 윈도우(131K vs 262K).
단일 에이전트에서는 gpt-oss-120b 통일이 최적.

---

## v2 아키텍처: Orchestrator + Worker Pool

```
                    ┌─────────────────────────┐
                    │    Orchestrator          │
                    │    (gpt-oss-120b)        │
                    │                          │
                    │  - 의도 파악              │
                    │  - 태스크 분해            │
                    │  - 결과 종합/검증          │
                    └────────┬────────────────┘
                             │
              ┌──────────────┼──────────────┐
              ▼              ▼              ▼
     ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
     │  Worker A     │ │  Worker B     │ │  Worker C     │
     │  (qwen-coder) │ │  (qwen-coder) │ │  (qwen-coder) │
     │               │ │               │ │               │
     │  파일 A 수정   │ │  파일 B 수정   │ │  테스트 작성   │
     └──────┬───────┘ └──────┬───────┘ └──────┬───────┘
            │                │                │
            ▼                ▼                ▼
     ┌─────────────────────────────────────────────┐
     │              Reviewer / Verifier             │
     │              (gpt-oss-120b)                  │
     │                                              │
     │  - 코드 리뷰                                  │
     │  - 통합 검증                                  │
     │  - 충돌 해결                                  │
     └─────────────────────────────────────────────┘
```

### 역할 분담

| 역할 | 모델 | 이유 |
|------|------|------|
| **Orchestrator** | gpt-oss-120b | 높은 지능(ELO 33)으로 의도 파악, 태스크 분해, 결과 종합 |
| **Worker** | Qwen3-Coder-30B | 코딩 특화, 262K 컨텍스트로 대형 파일 처리 가능 |
| **Reviewer** | gpt-oss-120b | 넓은 시야로 전체 코드 품질 검증 |

### 워크플로우

```
1. User → "인증 시스템 구현해줘"
2. Orchestrator (gpt-oss-120b):
   - 요구사항 분석
   - 서브태스크 생성:
     a) auth/middleware.go 작성
     b) auth/jwt.go 작성
     c) auth/handler_test.go 작성
3. Worker Pool (qwen-coder × N):
   - 각 서브태스크 병렬 실행
   - 파일 읽기/쓰기/수정 도구 사용
4. Reviewer (gpt-oss-120b):
   - Worker 결과물 검토
   - 통합 충돌 해결
   - 최종 승인 or 재작업 지시
5. Orchestrator → User: 완료 보고
```

### Qwen-Coder가 Worker로 적합한 이유

1. **262K 컨텍스트**: 대형 파일 전체를 한번에 읽고 수정 가능
2. **코딩 특화**: 코드 생성/수정에 집중된 학습 데이터
3. **비용 효율**: 병렬 Worker로 쓸 때 단가보다 처리량이 중요
4. **독립 태스크**: 각 Worker가 독립 파일을 담당하므로 충돌 최소화

### gpt-oss-120b가 Orchestrator로 적합한 이유

1. **높은 지능**: 복잡한 요구사항 분해, 의존성 파악
2. **빠른 속도(236 tok/s)**: Orchestrator는 짧은 판단을 자주 해야 함
3. **저렴한 비용($0.30/M)**: 오케스트레이션은 토큰 소모 적음
4. **범용성**: 코드뿐 아니라 계획, 리뷰, 사용자 대화 모두 처리

---

## 구현 고려사항

### 필요한 인프라

```go
// internal/agent/orchestrator.go
type Orchestrator struct {
    client    *openai.Client
    model     string          // gpt-oss-120b
    workers   []*Worker
    reviewer  *Reviewer
}

// internal/agent/worker.go
type Worker struct {
    id        int
    client    *openai.Client
    model     string          // qwen-coder-30b
    task      SubTask
    tools     []Tool
}

// internal/agent/task.go
type SubTask struct {
    ID          string
    Description string
    Files       []string      // 담당 파일 목록
    DependsOn   []string      // 의존 태스크 ID
    Status      TaskStatus
}
```

### 통신 패턴

- Orchestrator → Worker: goroutine + channel
- Worker → Orchestrator: 결과 channel
- Reviewer: 모든 Worker 완료 후 일괄 검토

### 모드별 동작

| 모드 | v1 (현재) | v2 (멀티에이전트) |
|------|----------|------------------|
| 슈퍼택가이 | gpt-oss 단일 | Orchestrator(gpt-oss) + Worker(qwen) 자동 |
| 개발 | gpt-oss 단일 | Worker(qwen) 직접 실행, 복잡하면 Orchestrator 개입 |
| 플랜 | gpt-oss 단일 | Orchestrator(gpt-oss)만 (읽기 전용) |

---

## 마이그레이션 경로

### Phase 1 (현재): 단일 에이전트
- gpt-oss-120b 통일
- 도구 시스템 안정화
- 에이전틱 루프 검증

### Phase 2: Orchestrator 분리
- 태스크 분해 로직 추가
- Worker 인터페이스 정의
- 단일 Worker로 테스트 (여전히 gpt-oss)

### Phase 3: 멀티 Worker
- Qwen-Coder Worker 추가
- 병렬 실행 (goroutine)
- 결과 병합 로직

### Phase 4: Reviewer + 자동 검증
- 코드 리뷰 자동화
- 테스트 실행 후 피드백 루프
- 충돌 해결 자동화

---

## 참고

- Novita AI 엔드포인트: `https://api.novita.ai/openai`
- gpt-oss-120b: MoE 120B total / 5.1B active
- Qwen3-Coder-30B: Dense 30B / 262K context
- 벤치마크 출처: Artificial Analysis (2026-04)
