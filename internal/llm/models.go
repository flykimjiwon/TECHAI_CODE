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
		Description: "범용 대형 모델 — 슈퍼택가이/플랜 모드용",
	},
	"qwen/qwen3-coder-30b": {
		ID:          "qwen/qwen3-coder-30b",
		DisplayName: "Qwen3-Coder-30B",
		Description: "코딩 특화 MoE 모델 — 개발 모드용 (256K context)",
	},
}
