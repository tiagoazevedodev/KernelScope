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
	Config     *cli.Config
	Executor   *executor.Executor
	Monitor    *monitor.Monitor
	UsedCpuTime float64
	Stats      *monitor.Stats
}

// NewLoopController creates a new loop controller
func NewLoopController(config *cli.Config, exec *executor.Executor, mon *monitor.Monitor) *LoopController {
	return &LoopController{
		Config:     config,
		Executor:   exec,
		Monitor:    mon,
		UsedCpuTime: 0,
		Stats:      &monitor.Stats{},
	}
}

// StartLoop starts the main execution loop
func (lc *LoopController) StartLoop() {
	fmt.Println("Starting main execution loop...")
	
	// Initialize stats
	lc.Stats.StartTime = time.Now()
	lc.Stats.LoopCount = 0
	lc.Stats.SuccessCount = 0
	
	// Define loop condition based on mode
	for lc.shouldContinue() {
		// Report progress
		reporter.ReportProgress(lc.Stats)
		
		// Start a new process
		process, err := lc.Executor.StartProcess()
		if err != nil {
			fmt.Printf("Failed to start process: %v\n", err)
			lc.Stats.TermReason = "Process start failure"
			break
		}
		
		// Start monitoring the new process
		lc.Monitor.StartMonitoring(process)

		// Wait for process to complete
		result := lc.Monitor.WaitForCompletion()
		
		// Update CPU time used
		lc.UsedCpuTime += result.CpuTimeUsed
		
		// Update overall stats
		lc.updateStats(result)
		
		// Record loop iteration
		success := result.ExitCode == 0 && result.TermReason == ""
		lc.Monitor.RecordLoopIteration(success)
		
		// If we're in prepaid mode, only deduct CPU time if process was successful
		if lc.Config.PrePaidMode && success {
			lc.UsedCpuTime += result.CpuTimeUsed
		} else if !lc.Config.PrePaidMode {
			// In postpaid mode, always count CPU time
			lc.UsedCpuTime += result.CpuTimeUsed
		}
		
		// Small delay between iterations
		time.Sleep(100 * time.Millisecond)
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