package monitor

import (
	"fmt"
	"kernelscope/cli"
	"kernelscope/executor"
	"kernelscope/resource"
	"time"
)

// Stats holds process statistics
type Stats struct {
	StartTime    time.Time
	EndTime      time.Time
	CpuTimeUsed  float64
	MaxMemoryKB  uint64
	ExitCode     int
	TermReason   string
	LoopCount    int
	SuccessCount int
}

// Monitor handles process monitoring
type Monitor struct {
	Config         *cli.Config
	ResourceMgr    *resource.ResourceManager
	Stats          *Stats
	stopMonitoring chan bool
}

// NewMonitor creates a new process monitor
func NewMonitor(config *cli.Config) *Monitor {
	return &Monitor{
		Config:         config,
		ResourceMgr:    resource.NewResourceManager(config),
		Stats:          &Stats{},
		stopMonitoring: make(chan bool),
	}
}

// StartMonitoring begins monitoring the specified process
func (m *Monitor) StartMonitoring(process *executor.Process) *Stats {
	fmt.Printf("Starting to monitor process PID: %d\n", process.Pid)
	
	// Set process resource limits
	err := m.ResourceMgr.SetProcessLimits(process.Pid)
	if err != nil {
		fmt.Printf("Warning: Failed to set resource limits: %v\n", err)
	}
	
	// Initialize stats
	m.Stats.StartTime = time.Now()
	m.Stats.LoopCount = 0
	m.Stats.SuccessCount = 0
	m.Stats.CpuTimeUsed = 0
	m.Stats.MaxMemoryKB = 0
	
	// Start monitoring goroutine
	go m.monitorProcess(process)
	
	// Start timeout goroutine
	if m.Config.Timeout > 0 {
		go m.enforceTimeout(process)
	}
	
	return m.Stats
}

// monitorProcess continuously monitors a process's resource usage
func (m *Monitor) monitorProcess(process *executor.Process) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			// Get current resource usage
			cpuTime, memoryKB, err := m.ResourceMgr.GetResourceUsage(process.Pid)
			if err != nil {
				// Process may have terminated
				continue
			}
			
			// Update stats
			m.Stats.CpuTimeUsed = cpuTime
			if memoryKB > m.Stats.MaxMemoryKB {
				m.Stats.MaxMemoryKB = memoryKB
			}
			
			// Check memory limit
			if m.Config.MemoryLimit > 0 && memoryKB > uint64(m.Config.MemoryLimit) {
				fmt.Printf("Memory limit exceeded: %d KB > %d KB\n", memoryKB, m.Config.MemoryLimit)
				m.Stats.TermReason = "Memory limit exceeded"
				m.terminateProcess(process)
				return
			}
			
			// Check CPU quota
			if m.ResourceMgr.IsCpuQuotaExceeded(cpuTime) {
				fmt.Printf("CPU quota exceeded: %.2f seconds\n", cpuTime)
				m.Stats.TermReason = "CPU quota exceeded"
				m.terminateProcess(process)
				return
			}
			
			// Output current stats
			fmt.Printf("PID: %d | CPU: %.2fs | Memory: %d KB\n", process.Pid, cpuTime, memoryKB)
			
		case <-m.stopMonitoring:
			fmt.Println("Stopping monitoring")
			return
		}
	}
}

// enforceTimeout enforces the process timeout
func (m *Monitor) enforceTimeout(process *executor.Process) {
	select {
	case <-time.After(time.Duration(m.Config.Timeout) * time.Second):
		fmt.Printf("Process timeout after %d seconds\n", m.Config.Timeout)
		m.Stats.TermReason = "Timeout"
		m.terminateProcess(process)
	case <-m.stopMonitoring:
		return
	}
}

// terminateProcess terminates the specified process
func (m *Monitor) terminateProcess(process *executor.Process) {
	executor := executor.NewExecutor(m.Config)
	err := executor.KillProcess(process)
	if err != nil {
		fmt.Printf("Error killing process: %v\n", err)
	}
	
	m.stopMonitoring <- true
}

// WaitForCompletion waits for the process to complete and returns the exit code
func (m *Monitor) WaitForCompletion() *Stats {
	// Wait for the process to complete
	// This would be implemented in a real system to wait for the process
	// to finish or be terminated
	
	m.Stats.EndTime = time.Now()
	return m.Stats
}

// RecordLoopIteration records a loop iteration
func (m *Monitor) RecordLoopIteration(success bool) {
	m.Stats.LoopCount++
	if success {
		m.Stats.SuccessCount++
	}
} 