package main

import (
	"fmt"
	"kernelscope/cli"
	"kernelscope/executor"
	"kernelscope/loopcontrol"
	"kernelscope/monitor"
	"os"
	"runtime"
)

func main() {
	fmt.Println("KernelScope - Process Execution and Monitoring System")
	
	// Check if running on Linux
	if runtime.GOOS != "linux" {
		fmt.Println("Warning: KernelScope is designed for Linux systems. Some features may not work on your platform.")
		fmt.Printf("Current platform: %s\n", runtime.GOOS)
	}
	
	// Parse command line arguments
	config := cli.ParseArgs()
	
	// Display configuration
	cli.DisplayConfig(config)
	
	// Initialize the executor
	exec := executor.NewExecutor(config)
	
	// Initialize the monitor
	mon := monitor.NewMonitor(config)
	
	// Initialize the loop controller
	loopCtrl := loopcontrol.NewLoopController(config, exec, mon)
	
	// Start the main execution loop
	loopCtrl.StartLoop()
	
	fmt.Println("KernelScope execution completed")
	os.Exit(0)
} 