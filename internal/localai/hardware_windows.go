//go:build windows

package localai

import (
	"encoding/csv"
	"encoding/json"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

var (
	kernel32DLL              = windows.NewLazySystemDLL("kernel32.dll")
	globalMemoryStatusExProc = kernel32DLL.NewProc("GlobalMemoryStatusEx")
	getDiskFreeSpaceExProc   = kernel32DLL.NewProc("GetDiskFreeSpaceExW")
)

func DetectHardware(dataRoot string) HardwareProfile {
	p := HardwareProfile{Platform: "windows", Supported: true, CPULogicalCores: runtime.NumCPU(), GPUs: []GPUAdapter{}}
	var mem memoryStatusEx
	mem.Length = uint32(unsafe.Sizeof(mem))
	if ok, _, _ := globalMemoryStatusExProc.Call(uintptr(unsafe.Pointer(&mem))); ok != 0 {
		p.MemoryTotalMiB = int64(mem.TotalPhys / 1024 / 1024)
		p.MemoryFreeMiB = int64(mem.AvailPhys / 1024 / 1024)
	}
	if ptr, err := windows.UTF16PtrFromString(dataRoot); err == nil {
		var free uint64
		if ok, _, _ := getDiskFreeSpaceExProc.Call(uintptr(unsafe.Pointer(ptr)), uintptr(unsafe.Pointer(&free)), 0, 0); ok != 0 {
			p.DiskFreeBytes = int64(free)
		}
	}
	p.GPUs = mergeGPUAdapters(detectNVIDIAGPUs(), detectWindowsVideoControllers())
	p.GPUs = nonNilGPUs(p.GPUs)
	applyRecommendation(&p)
	return p
}

func hiddenCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.CREATE_NO_WINDOW}
	return cmd
}

func detectNVIDIAGPUs() []GPUAdapter {
	out, err := hiddenCommand("nvidia-smi", "--query-gpu=name,memory.total,memory.free,driver_version", "--format=csv,noheader,nounits").Output()
	if err != nil {
		return nil
	}
	rows, err := csv.NewReader(strings.NewReader(string(out))).ReadAll()
	if err != nil {
		return nil
	}
	var result []GPUAdapter
	for _, row := range rows {
		if len(row) < 4 {
			continue
		}
		total, _ := strconv.ParseInt(strings.TrimSpace(row[1]), 10, 64)
		free, _ := strconv.ParseInt(strings.TrimSpace(row[2]), 10, 64)
		result = append(result, GPUAdapter{Name: strings.TrimSpace(row[0]), Vendor: "NVIDIA", Backend: "cuda", DedicatedMiB: total, AvailableMiB: free, DriverVersion: strings.TrimSpace(row[3])})
	}
	return result
}

func detectWindowsVideoControllers() []GPUAdapter {
	script := `@(Get-CimInstance Win32_VideoController | Select-Object Name,AdapterRAM,DriverVersion) | ConvertTo-Json -Compress`
	out, err := hiddenCommand("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script).Output()
	if err != nil || len(out) == 0 {
		return nil
	}
	var controllers []struct {
		Name          string `json:"Name"`
		AdapterRAM    uint64 `json:"AdapterRAM"`
		DriverVersion string `json:"DriverVersion"`
	}
	if err := json.Unmarshal(out, &controllers); err != nil {
		return nil
	}
	result := make([]GPUAdapter, 0, len(controllers))
	for _, controller := range controllers {
		name := strings.TrimSpace(controller.Name)
		if name == "" {
			continue
		}
		vendor, backend := gpuVendorAndBackend(name)
		result = append(result, GPUAdapter{
			Name: name, Vendor: vendor, Backend: backend,
			DedicatedMiB:  int64(controller.AdapterRAM / 1024 / 1024),
			DriverVersion: strings.TrimSpace(controller.DriverVersion),
		})
	}
	return result
}

func gpuVendorAndBackend(name string) (string, string) {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "nvidia"):
		return "NVIDIA", "cuda"
	case strings.Contains(lower, "amd"), strings.Contains(lower, "radeon"):
		return "AMD", "vulkan"
	case strings.Contains(lower, "intel"):
		return "Intel", "vulkan"
	default:
		return "Unknown", "vulkan"
	}
}

func mergeGPUAdapters(primary, discovered []GPUAdapter) []GPUAdapter {
	merged := make([]GPUAdapter, 0, len(primary)+len(discovered))
	index := map[string]int{}
	add := func(gpu GPUAdapter) {
		key := strings.ToLower(strings.TrimSpace(gpu.Name))
		if key == "" {
			return
		}
		if at, ok := index[key]; ok {
			current := merged[at]
			if gpu.DedicatedMiB > current.DedicatedMiB {
				current.DedicatedMiB = gpu.DedicatedMiB
			}
			if gpu.AvailableMiB > 0 {
				current.AvailableMiB = gpu.AvailableMiB
			}
			if current.DriverVersion == "" {
				current.DriverVersion = gpu.DriverVersion
			}
			if gpuPreference(gpu) > gpuPreference(current) {
				current.Backend = gpu.Backend
				current.Vendor = gpu.Vendor
			}
			merged[at] = current
			return
		}
		index[key] = len(merged)
		merged = append(merged, gpu)
	}
	for _, gpu := range primary {
		add(gpu)
	}
	for _, gpu := range discovered {
		add(gpu)
	}
	sort.SliceStable(merged, func(i, j int) bool {
		if merged[i].DedicatedMiB != merged[j].DedicatedMiB {
			return merged[i].DedicatedMiB > merged[j].DedicatedMiB
		}
		return gpuPreference(merged[i]) > gpuPreference(merged[j])
	})
	return merged
}
