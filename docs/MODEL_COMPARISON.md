# TECHAI CODE — LLM 모델 비교 분석

> 작성일: 2026-04-29
> 기준: Novita.ai API 제공 모델 + DeepSeek 공식 API

---

## 현재 사용 모델

**DeepSeek V4 Flash** (via Novita.ai)
- Model ID: `deepseek/deepseek-v4-flash`
- 가격: $0.14/M Input, $0.28/M Output
- 컨텍스트: 1M tokens
- 선정 이유: 가성비 최고, 코딩 성능 상위

---

## 코딩 벤치마크 비교

| 모델 | SWE-bench | LiveCodeBench | 컨텍스트 | Active Params |
|------|:-:|:-:|:-:|:-:|
| **DeepSeek V4 Pro** | **80.6%** | **93.5%** | 1M | 49B |
| **DeepSeek V4 Flash** ← 현재 | **79.0%** | **91.6%** | 1M | 13B |
| Qwen3 Coder Next | 70.6% | - | 262K | 3B |
| Qwen3 Coder 30B (이전) | ~60% 추정 | ~75% 추정 | 160K | 3B |
| Claude Opus 4.7 | 97/100* | - | 200K | - |
| GPT 5.5 | 96/100* | - | 128K | - |

*커스텀 벤치마크 기준 (Rails 앱 빌드 테스트)

---

## Novita.ai 모델 전체 리스트 (코딩 추천순)

### Tier 1 — 최고 성능 (코딩 특화)

| 모델 | Input | Output | 컨텍스트 | 비고 |
|------|------:|-------:|:---:|------|
| DeepSeek V4 Pro | $1.74 | $3.48 | 1M | thinking mode, 느림 |
| **DeepSeek V4 Flash** | **$0.14** | **$0.28** | **1M** | **현재 사용 — 가성비 최고** |
| Qwen3 Coder Next | $0.20 | $1.50 | 262K | Output 비쌈 |
| Qwen3 Max | $2.11 | $8.45 | 262K | 비쌈 |

### Tier 2 — 범용 고성능

| 모델 | Input | Output | 컨텍스트 | 비고 |
|------|------:|-------:|:---:|------|
| DeepSeek V3.2 | $0.27 | $0.40 | 163K | 안정적 |
| Llama 3.3 70B | $0.14 | $0.40 | 131K | Meta, 검증됨 |
| Gemma 4 31B | $0.14 | $0.40 | 262K | Google |
| Gemma 4 26B (MoE) | $0.13 | $0.40 | 262K | 경량 |
| GLM-5.1 | $1.40 | $4.40 | 204K | 중국 최신 |
| GLM-5 | $1.00 | $3.20 | 202K | |
| MiniMax M2.7 | $0.30 | $1.20 | 204K | |

### Tier 3 — 경량/저가

| 모델 | Input | Output | 컨텍스트 | 비고 |
|------|------:|-------:|:---:|------|
| Qwen3 Coder 30B (이전) | $0.07 | $0.27 | 160K | 가성비 좋지만 성능 보통 |
| Qwen3 8B | $0.04 | $0.14 | 128K | 빠름, 가벼움 |
| Gemma 3 12B | $0.05 | $0.10 | 131K | |
| Qwen3 4B | $0.03 | $0.03 | 128K | 초저가 |
| Llama 3.1 8B | $0.02 | $0.05 | 16K | 최저가 |

---

## DeepSeek 공식 API vs Novita 가격 비교

### V4 Flash

| | DeepSeek 공식 | Novita |
|---|---:|---:|
| Input (캐시 미스) | $0.14 | $0.14 |
| Input (캐시 히트) | **$0.0028** | 없음 |
| Output | $0.28 | $0.28 |

공식 API가 캐시 히트 시 **50배 저렴**.

### V4 Pro (75% 할인 — 2026/05/05까지)

| | 공식 (할인) | 공식 (원가) | Novita |
|---|---:|---:|---:|
| Input | **$0.435** | $1.74 | $1.74 |
| Input (캐시) | **$0.003625** | $0.0145 | 없음 |
| Output | **$0.87** | $3.48 | $3.48 |

할인 중 Novita 대비 **4배 저렴**.

---

## 모델 선택 가이드

| 용도 | 추천 모델 | 이유 |
|------|----------|------|
| **일상 코딩** | DeepSeek V4 Flash | 가성비 + 1M 컨텍스트 |
| **복잡한 아키텍처** | DeepSeek V4 Pro | 최고 성능, 느림 감수 |
| **시연/데모** | DeepSeek V4 Flash | 빠른 응답, 도구 호출 안정 |
| **사내 폐쇄망** | Qwen3 Coder 30B | 자체 서버 배포 가능 |
| **초저가 테스트** | Qwen3 4B | $0.03/M |

---

## 비용 추정 (하루 사용량)

| 사용량 | V4 Flash | V4 Pro | Qwen3 Coder 30B |
|--------|----------|--------|-----------------|
| 가벼운 테스트 (10회) | ~$0.01 | ~$0.03 | ~$0.01 |
| 보통 사용 (30회) | ~$0.05 | ~$0.15 | ~$0.03 |
| 하루 종일 (100회+) | ~$0.15 | ~$0.50 | ~$0.10 |
| **월간 (매일 사용)** | **~$5** | **~$15** | **~$3** |

---

## config.yaml 설정 예시

### Novita + V4 Flash (현재)
```yaml
api:
    base_url: https://api.novita.ai/openai
    api_key: sk_...
models:
    super: deepseek/deepseek-v4-flash
    dev: deepseek/deepseek-v4-flash
```

### DeepSeek 공식 API (캐시 할인 활용)
```yaml
api:
    base_url: https://api.deepseek.com/v1
    api_key: sk-...
models:
    super: deepseek-v4-flash
    dev: deepseek-v4-flash
```

### 사내 온프렘
```yaml
api:
    base_url: https://techai-web-prod.shinhan.com/v1
    api_key: (내부키)
models:
    super: Qwen3-Coder-30B
    dev: Qwen3-Coder-30B
```

---

## 참고 자료

- [Novita.ai LLM Models](https://novita.ai/models/llm)
- [DeepSeek API Pricing](https://api-docs.deepseek.com/quick_start/pricing)
- [LLM Coding Benchmark April 2026](https://akitaonrails.com/en/2026/04/24/llm-benchmarks-parte-3-deepseek-kimi-mimo/)
- [Qwen3-Coder-Next Technical Report](https://arxiv.org/html/2603.00729v1)
- [DeepSeek V4 Flash — HuggingFace](https://huggingface.co/deepseek-ai/DeepSeek-V4-Flash)
