package cli

import (
	"flag"
	"fmt"
	"os"
)

// Config holds all the command-line parameters
type Config struct {
	BinaryPath    string  // Path to the binary to execute
	CpuLimit      int     // CPU time limit in seconds
	MemoryLimit   int     // Memory limit in KB
	Timeout       int     // Timeout in seconds
	PrePaidMode   bool    // Run in prepaid mode (true) or postpaid mode (false)
	CpuCredit     float64 // CPU credits in seconds for prepaid mode
}

// ParseArgs parses command-line arguments and returns a Config
func ParseArgs() *Config {
	config := &Config{}

	flag.StringVar(&config.BinaryPath, "binary", "", "Path to the binary to execute (required)")
	flag.IntVar(&config.CpuLimit, "cpu", 10, "CPU time limit in seconds")
	flag.IntVar(&config.MemoryLimit, "mem", 1024*1024, "Memory limit in KB")
	flag.IntVar(&config.Timeout, "timeout", 30, "Timeout in seconds")
	flag.BoolVar(&config.PrePaidMode, "prepaid", true, "Run in prepaid mode (true) or postpaid mode (false)")
	flag.Float64Var(&config.CpuCredit, "credit", 5.0, "CPU credits in seconds for prepaid mode")

	flag.Parse()

	// Validate that binary path is provided
	if config.BinaryPath == "" {
		fmt.Println("Error: Binary path is required")
		flag.Usage()
		os.Exit(1)
	}

	return config
}

// DisplayConfig prints the current configuration
func DisplayConfig(config *Config) {
	fmt.Println("=== KernelScope Configuration ===")
	fmt.Printf("Binary:       %s\n", config.BinaryPath)
	fmt.Printf("CPU Limit:    %d seconds\n", config.CpuLimit)
	fmt.Printf("Memory Limit: %d KB\n", config.MemoryLimit)
	fmt.Printf("Timeout:      %d seconds\n", config.Timeout)
	if config.PrePaidMode {
		fmt.Printf("Mode:         Prepaid with %.2f CPU credits\n", config.CpuCredit)
	} else {
		fmt.Println("Mode:         Postpaid")
	}
	fmt.Println("===============================")
} 