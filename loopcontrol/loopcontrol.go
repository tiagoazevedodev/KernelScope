package loopcontrol

import (
	"fmt"
	"kernelscope/cli"
	"kernelscope/executor"
	"kernelscope/monitor"
	"kernelscope/reporter"
	"time"
)

// LoopController manages the main execution loop
type LoopController struct {
	Config      *cli.Config
	Executor    *executor.Executor
	Monitor     *monitor.Monitor
	UsedCpuTime float64
	Stats       *monitor.Stats
}

func NewLoopController(config *cli.Config, exec *executor.Executor, mon *monitor.Monitor) *LoopController {
	return &LoopController{
		Config:      config,
		Executor:    exec,
		Monitor:     mon,
		UsedCpuTime: 0,
		Stats:       &monitor.Stats{},
	}
}

// StartLoop starts the main execution loop
func (lc *LoopController) StartLoop() {
	fmt.Println("Starting process execution and monitoring...")

	// Initialize stats
	lc.Stats.StartTime = time.Now()
	lc.Stats.LoopCount = 1
	lc.Stats.SuccessCount = 0

	// Start process once
	process, err := lc.Executor.StartProcess()
	if err != nil {
		fmt.Printf("Failed to start process: %v\n", err)
		lc.Stats.TermReason = "Process start failure"

		// Generate report even if process failed to start
		lc.Stats.EndTime = time.Now()
		reporter.GenerateReport(lc.Stats, lc.Stats)
		return
	}

	// Start monitoring the process
	lc.Monitor.StartMonitoring(process)

	// Use a separate goroutine to properly wait for the process
	waitDone := make(chan int)
	go func() {
		exitCode, err := lc.Executor.WaitForProcess(process)
		if err != nil {
			fmt.Printf("Error waiting for process: %v\n", err)
		}
		waitDone <- exitCode
	}()

	// Wait for process to complete or reach resource limits
	processRunning := true
	for processRunning && lc.shouldContinue() {
		// Report progress
		reporter.ReportProgress(lc.Monitor.Stats)

		// Check if process has completed via the wait channel
		select {
		case exitCode := <-waitDone:
			lc.Stats.ExitCode = exitCode
			processRunning = false
			fmt.Printf("Process exited with code: %d\n", exitCode)
		case <-time.After(500 * time.Millisecond):
			// Continue monitoring
		}
	}

	// If we broke out of the loop due to resource limits but process is still running
	if processRunning {
		fmt.Println("Resource limits reached, terminating process...")
		lc.Executor.KillProcess(process)

		// Wait for the process to be fully terminated
		select {
		case exitCode := <-waitDone:
			lc.Stats.ExitCode = exitCode
		case <-time.After(2 * time.Second):
			fmt.Println("Warning: Process did not terminate gracefully")
		}
	}

	// Wait for final process stats
	result := lc.Monitor.WaitForCompletion()

	// Update CPU time used
	lc.UsedCpuTime = result.CpuTimeUsed

	// Update overall stats
	lc.updateStats(result)

	// Record success
	success := lc.Stats.ExitCode == 0 && lc.Stats.TermReason == ""
	if success {
		lc.Stats.SuccessCount = 1
	}

	lc.Stats.EndTime = time.Now()
	lc.Stats.CpuTimeUsed = lc.UsedCpuTime

	// Generate final report
	reporter.GenerateReport(lc.Stats, lc.Stats)
}

// shouldContinue determines if the loop should continue
func (lc *LoopController) shouldContinue() bool {
	if lc.Config.PrePaidMode {
		// In prepaid mode, continue until CPU credits are exhausted
		return lc.UsedCpuTime < lc.Config.CpuCredit
	} else {
		// In postpaid mode, continue until CPU limit is reached
		return lc.UsedCpuTime < float64(lc.Config.CpuLimit)
	}
}

// updateStats updates the overall statistics
func (lc *LoopController) updateStats(result *monitor.Stats) {
	// Update max memory usage
	if result.MaxMemoryKB > lc.Stats.MaxMemoryKB {
		lc.Stats.MaxMemoryKB = result.MaxMemoryKB
	}

	// Update termination reason if set
	if result.TermReason != "" {
		lc.Stats.TermReason = result.TermReason
	}

	// Update exit code if non-zero
	if result.ExitCode != 0 {
		lc.Stats.ExitCode = result.ExitCode
	}
}
