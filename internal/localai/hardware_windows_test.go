//go:build windows

package localai

import "testing"

func TestMergeGPUAdaptersKeepsIntegratedAndDiscreteAdapters(t *testing.T) {
	got := mergeGPUAdapters(
		[]GPUAdapter{{Name: "NVIDIA GeForce RTX", Vendor: "NVIDIA", Backend: "cuda", DedicatedMiB: 16 * 1024, AvailableMiB: 12 * 1024}},
		[]GPUAdapter{
			{Name: "Intel Graphics", Vendor: "Intel", Backend: "vulkan", DedicatedMiB: 1024},
			{Name: "NVIDIA GeForce RTX", Vendor: "NVIDIA", Backend: "cuda", DedicatedMiB: 4095},
		},
	)
	if len(got) != 2 {
		t.Fatalf("merged adapters = %+v", got)
	}
	if got[0].Name != "NVIDIA GeForce RTX" || got[0].DedicatedMiB != 16*1024 {
		t.Fatalf("best adapter was not preserved: %+v", got)
	}
	if got[1].Name != "Intel Graphics" {
		t.Fatalf("integrated adapter was lost: %+v", got)
	}
}
