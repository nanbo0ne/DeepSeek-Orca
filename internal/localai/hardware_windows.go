//go:build windows

package localai

import (
	"encoding/csv"
	"os/exec"
	"runtime"
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
	p.GPUs = detectNVIDIAGPUs()
	if len(p.GPUs) == 0 {
		p.GPUs = detectWindowsVideoControllers()
	}
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
	script := `Get-CimInstance Win32_VideoController | Select-Object Name,AdapterRAM,DriverVersion | ConvertTo-Json -Compress`
	out, err := hiddenCommand("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script).Output()
	if err != nil || len(out) == 0 {
		return nil
	}
	// Keep this fallback intentionally conservative: WMI's AdapterRAM is only a
	// 32-bit estimate on some drivers, so unknown adapters are not over-promoted.
	text := string(out)
	var result []GPUAdapter
	for _, vendor := range []struct{ key, label string }{{"AMD", "AMD"}, {"Radeon", "AMD"}, {"Intel", "Intel"}} {
		if strings.Contains(text, vendor.key) {
			result = append(result, GPUAdapter{Name: vendor.label + " GPU", Vendor: vendor.label, Backend: "vulkan"})
			break
		}
	}
	return result
}
