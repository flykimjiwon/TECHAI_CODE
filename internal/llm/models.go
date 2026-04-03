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
	"qwen/qwen-2.5-coder-32b-instruct": {
		ID:          "qwen/qwen-2.5-coder-32b-instruct",
		DisplayName: "Qwen Coder 32B",
		Description: "코딩 특화 모델 — 개발 모드용",
	},
}
