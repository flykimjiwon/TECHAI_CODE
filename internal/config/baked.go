package config

import "fmt"

// Build-time baked configuration for 택가이코드.
//
// 세 가지 배포 모드를 지원합니다:
//
//   1. vanilla  — 아무것도 고정하지 않음. 사용자가 setup 위저드를 통해 설정.
//   2. distro   — 엔드포인트/프로바이더/모델을 고정, API 키만 사용자 입력.
//   3. sealed   — API 키 포함 모든 것을 고정. onprem 빌드에 사용.
//
// Build mode는 -ldflags로 BakedMode를 설정하여 선택합니다.
//
// Example (onprem/sealed):
//
//   go build -ldflags "\
//     -X 'github.com/kimjiwon/tgc/internal/config.BakedMode=sealed' \
//     -X 'github.com/kimjiwon/tgc/internal/config.BakedBaseURL=https://techai-web-prod.shinhan.com/v1' \
//     -X 'github.com/kimjiwon/tgc/internal/config.BakedModel=gpt-oss-120b' \
//     -X 'github.com/kimjiwon/tgc/internal/config.BakedAPIKey=sk-…'" \
//     -o techai ./cmd/tgc
var (
	BakedMode     = ""
	BakedBaseURL  = ""
	BakedProvider = ""
	BakedModel    = ""
	BakedDevModel = ""
	BakedAPIKey   = ""
	BakedBrand    = ""
	BakedNoStream = ""
)

// ValidateBakedMode panics at process start if baked mode fields
// are internally inconsistent.
func ValidateBakedMode() {
	switch BakedMode {
	case "", "distro", "sealed":
	default:
		panic(fmt.Sprintf("config: unknown BakedMode %q (expected \"\", \"distro\", or \"sealed\")", BakedMode))
	}
	if BakedMode == "sealed" && BakedAPIKey == "" {
		panic("config: BakedMode=sealed requires BakedAPIKey to be set at build time")
	}
	if BakedMode == "sealed" && BakedBaseURL == "" {
		panic("config: BakedMode=sealed requires BakedBaseURL to be set at build time")
	}
	if BakedMode == "distro" && BakedBaseURL == "" {
		panic("config: BakedMode=distro requires BakedBaseURL to be set at build time")
	}
	if BakedAPIKey != "" && BakedMode != "sealed" {
		panic("config: BakedAPIKey must only be set when BakedMode=sealed")
	}
}

// BakeInfo returns a human-readable summary of the baked configuration.
func BakeInfo() string {
	switch BakedMode {
	case "distro":
		return "distro — 엔드포인트/모델 고정; API 키 사용자 입력"
	case "sealed":
		return "sealed — 모두 고정 (이 바이너리를 재배포하지 마세요)"
	default:
		return "vanilla — 사용자 설정 가능"
	}
}

// IsSealed reports whether this binary was built with a baked API key.
func IsSealed() bool { return BakedMode == "sealed" && BakedAPIKey != "" }

// IsDistro reports whether this binary was built with a baked endpoint.
func IsDistro() bool { return BakedMode == "distro" }

// BakedBrandName returns the product name.
func BakedBrandName() string {
	if BakedBrand != "" {
		return BakedBrand
	}
	return "택가이코드"
}

// applyBaked merges baked values into cfg.
func applyBaked(cfg Config) Config {
	switch BakedMode {
	case "sealed":
		if BakedBaseURL != "" {
			cfg.API.BaseURL = BakedBaseURL
		}
		if BakedAPIKey != "" {
			cfg.API.APIKey = BakedAPIKey
		}
		if BakedModel != "" {
			cfg.Models.Super = BakedModel
			if BakedDevModel == "" {
				cfg.Models.Dev = BakedModel
			}
		}
		if BakedDevModel != "" {
			cfg.Models.Dev = BakedDevModel
		}
	case "distro":
		if BakedBaseURL != "" {
			cfg.API.BaseURL = BakedBaseURL
		}
		if !distroUserHasConfig {
			if BakedModel != "" {
				cfg.Models.Super = BakedModel
			}
			devBake := BakedDevModel
			if devBake == "" {
				devBake = BakedModel
			}
			if devBake != "" {
				cfg.Models.Dev = devBake
			}
		}
	}
	return cfg
}

// distroUserHasConfig is set by Load() when a config.yaml was read.
var distroUserHasConfig bool
