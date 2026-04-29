package llm

type ModelInfo struct {
	ID          string
	DisplayName string
	Description string
}

var Models = map[string]ModelInfo{
	"openai/gpt-oss-120b": {
		ID:          "openai/gpt-oss-120b",
		DisplayName: "GPT-OSS 120B",
		Description: "범용 대형 모델 — Super/Plan 모드용",
	},
	"qwen/qwen3-coder-30b": {
		ID:          "qwen/qwen3-coder-30b",
		DisplayName: "Qwen3-Coder-30B",
		Description: "코딩 특화 MoE 모델 — Deep Agent 모드용 (256K context)",
	},
	"deepseek-v4-pro": {
		ID:          "deepseek-v4-pro",
		DisplayName: "DeepSeek V4 Pro",
		Description: "최고 성능 — 1M context, 75% 할인 중 (5/5까지)",
	},
	"deepseek-v4-flash": {
		ID:          "deepseek-v4-flash",
		DisplayName: "DeepSeek V4 Flash",
		Description: "고성능 경량 — 1M context, 가성비 최고",
	},
	"deepseek/deepseek-v4-pro": {
		ID:          "deepseek/deepseek-v4-pro",
		DisplayName: "DeepSeek V4 Pro",
		Description: "최고 성능 — 1M context (Novita)",
	},
	"deepseek/deepseek-v4-flash": {
		ID:          "deepseek/deepseek-v4-flash",
		DisplayName: "DeepSeek V4 Flash",
		Description: "고성능 경량 — 1M context (Novita)",
	},
}
