package resource

import (
	"fmt"
	"kernelscope/cli"
	"kernelscope/utils"
	"runtime"
	"golang.org/x/sys/unix"
)

// ResourceManager handles resource limitations
type ResourceManager struct {
	Config *cli.Config
}

// NewResourceManager creates a new resource manager
func NewResourceManager(config *cli.Config) *ResourceManager {
	return &ResourceManager{
		Config: config,
	}
}

// SetProcessLimits sets resource limits for a process
func (rm *ResourceManager) SetProcessLimits(pid int) error {
	// Resource limits only work on Linux-based systems
	if runtime.GOOS != "linux" {
		fmt.Println("Warning: Resource limits only fully supported on Linux")
		return nil
	}

	// Set CPU time limit
	if rm.Config.CpuLimit > 0 {
		// Set soft and hard limits for CPU time (in seconds)
		cpuLimit := unix.Rlimit{
			Cur: uint64(rm.Config.CpuLimit),
			Max: uint64(rm.Config.CpuLimit),
		}
		
		err := unix.Prlimit(pid, unix.RLIMIT_CPU, &cpuLimit, nil)
		if err != nil {
			return fmt.Errorf("failed to set CPU limits: %v", err)
		}
	}

	// Set memory limit
	if rm.Config.MemoryLimit > 0 {
		// Convert KB to bytes for the memory limit
		memBytes := uint64(rm.Config.MemoryLimit * 1024)
		
		memLimit := unix.Rlimit{
			Cur: memBytes,
			Max: memBytes,
		}
		
		err := unix.Prlimit(pid, unix.RLIMIT_AS, &memLimit, nil)
		if err != nil {
			return fmt.Errorf("failed to set memory limits: %v", err)
		}
	}

	return nil
}

// IsCpuQuotaExceeded checks if the CPU quota has been exceeded
func (rm *ResourceManager) IsCpuQuotaExceeded(usedCpu float64) bool {
	if rm.Config.PrePaidMode {
		// In prepaid mode, check if we've exceeded our credit
		return usedCpu >= rm.Config.CpuCredit
	} else {
		// In postpaid mode, check if we've exceeded the CPU limit
		return usedCpu >= float64(rm.Config.CpuLimit)
	}
}

// GetResourceUsage gets current resource usage information for a process and its children
func (rm *ResourceManager) GetResourceUsage(pid int) (float64, uint64, error) {
	// If not on Linux, return placeholder values
	if runtime.GOOS != "linux" {
		return 0.0, 0, nil
	}
	
	// Get stats for the main process
	stats, err := utils.ReadProcStats(pid)
	if err != nil {
		return 0.0, 0, err
	}
	
	totalCpuTime := stats.CpuTime
	totalMemoryKB := stats.MemoryKB
	
	// Get all child processes
	childPids, err := utils.GetAllChildProcesses(pid)
	if err == nil && len(childPids) > 0 {
		// Sum up resource usage from all children
		for _, childPid := range childPids {
			childStats, err := utils.ReadProcStats(childPid)
			if err == nil {
				totalCpuTime += childStats.CpuTime
				totalMemoryKB += childStats.MemoryKB
			}
		}
	}
	
	return totalCpuTime, totalMemoryKB, nil
} 