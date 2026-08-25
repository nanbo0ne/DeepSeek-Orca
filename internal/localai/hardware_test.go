package localai

import (
	"encoding/json"
	"testing"
)

func TestApplyRecommendationUsesBestAdapterInsteadOfGPUZero(t *testing.T) {
	profile := HardwareProfile{
		Supported: true,
		GPUs: []GPUAdapter{
			{Name: "Integrated GPU", Backend: "vulkan", DedicatedMiB: 1024},
			{Name: "Discrete GPU", Backend: "cuda", DedicatedMiB: 16 * 1024, AvailableMiB: 14 * 1024},
		},
	}
	applyRecommendation(&profile)
	if profile.RecommendedModel != "qwen3.8-27b-iq3-xxs" {
		t.Fatalf("recommended model = %q", profile.RecommendedModel)
	}
	if profile.RecommendedRuntime != "cuda-12.4-x64" {
		t.Fatalf("recommended runtime = %q", profile.RecommendedRuntime)
	}
}

func TestApplyRecommendationKeepsGPUsJSONArrayNonNull(t *testing.T) {
	profile := HardwareProfile{Supported: true}
	applyRecommendation(&profile)
	data, err := json.Marshal(profile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "" || profile.GPUs == nil {
		t.Fatalf("hardware profile must keep a non-nil GPU list: %s", data)
	}
}
