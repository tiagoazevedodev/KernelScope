package reporter

import (
	"fmt"
	"kernelscope/monitor"
	"time"
)

// GenerateReport generates a report of process execution
func GenerateReport(stats *monitor.Stats, finalStats *monitor.Stats) {
	duration := finalStats.EndTime.Sub(finalStats.StartTime)

	fmt.Println("\n=========== KernelScope Execution Report ===========")
	fmt.Printf("Execution Duration: %v\n", duration.Round(time.Millisecond))
	fmt.Printf("CPU Time Used: %.2f seconds\n", finalStats.CpuTimeUsed)
	fmt.Printf("Peak Memory Usage: %d KB\n", finalStats.MaxMemoryKB)

	if finalStats.TermReason != "" {
		fmt.Printf("Termination Reason: %s\n", finalStats.TermReason)
	} else if finalStats.ExitCode != 0 {
		fmt.Printf("Process exited with code: %d\n", finalStats.ExitCode)
	} else {
		fmt.Println("Process completed successfully")
	}

	fmt.Printf("Loop Iterations: %d\n", finalStats.LoopCount)
	fmt.Printf("Successful Iterations: %d\n", finalStats.SuccessCount)

	// Calculate efficiency
	if finalStats.LoopCount > 0 {
		successRate := float64(finalStats.SuccessCount) / float64(finalStats.LoopCount) * 100
		fmt.Printf("Success Rate: %.1f%%\n", successRate)
	}

	// Calculate resource efficiency
	if duration > 0 {
		cpuEfficiency := finalStats.CpuTimeUsed / duration.Seconds() * 100
		fmt.Printf("CPU Efficiency: %.1f%%\n", cpuEfficiency)
	}

	fmt.Println("===================================================")
}

// ReportProgress reports the current progress of execution
func ReportProgress(stats *monitor.Stats) {
	duration := time.Since(stats.StartTime).Round(time.Second)

	// Clear the current line and update with new information
	fmt.Printf("\r                                                                   ")
	fmt.Printf("\r[Running for %v] CPU: %.2fs | Memory: %d KB | Iterations: %d",
		duration, stats.CpuTimeUsed, stats.MaxMemoryKB, stats.LoopCount)
}
