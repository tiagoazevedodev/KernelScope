package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ProcStats holds stats read from /proc filesystem
type ProcStats struct {
	CpuTime   float64 // CPU time in seconds
	MemoryKB  uint64  // Memory usage in KB
	Children  []int   // Child process IDs
}

// ReadProcStats reads stats for a process from /proc filesystem
func ReadProcStats(pid int) (*ProcStats, error) {
	stats := &ProcStats{}
	
	// Check if process exists
	procPath := filepath.Join("/proc", strconv.Itoa(pid))
	_, err := os.Stat(procPath)
	if err != nil {
		return nil, fmt.Errorf("process %d does not exist: %v", pid, err)
	}
	
	// Read CPU stats from /proc/[pid]/stat
	statFile := filepath.Join(procPath, "stat")
	statBytes, err := os.ReadFile(statFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read stat file: %v", err)
	}
	
	statFields := strings.Fields(string(statBytes))
	if len(statFields) < 17 {
		return nil, fmt.Errorf("invalid stat file format")
	}
	
	// Extract CPU time (user + system time)
	// Fields 14 and 15 are utime and stime in clock ticks
	utime, err := strconv.ParseUint(statFields[13], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse utime: %v", err)
	}
	
	stime, err := strconv.ParseUint(statFields[14], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse stime: %v", err)
	}
	
	// Convert from clock ticks to seconds
	// In Linux, the number of clock ticks per second is typically defined by sysconf(_SC_CLK_TCK)
	// Most common value is 100, but we should read it from the system in a real implementation
	const clockTicksPerSecond = 100
	stats.CpuTime = float64(utime+stime) / float64(clockTicksPerSecond)
	
	// Read memory stats from /proc/[pid]/status
	statusFile := filepath.Join(procPath, "status")
	file, err := os.Open(statusFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read status file: %v", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		
		// Look for VmRSS line which gives the physical memory usage
		if strings.HasPrefix(line, "VmRSS:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				memKB, err := strconv.ParseUint(fields[1], 10, 64)
				if err == nil {
					stats.MemoryKB = memKB
				}
			}
		}
	}
	
	// Read child processes from /proc/[pid]/task/[pid]/children
	childrenFile := filepath.Join(procPath, "task", strconv.Itoa(pid), "children")
	childrenBytes, err := os.ReadFile(childrenFile)
	if err == nil { // It's okay if this fails, not all systems support it
		childrenStr := strings.TrimSpace(string(childrenBytes))
		if childrenStr != "" {
			childrenFields := strings.Fields(childrenStr)
			for _, child := range childrenFields {
				childPid, err := strconv.Atoi(child)
				if err == nil {
					stats.Children = append(stats.Children, childPid)
				}
			}
		}
	}
	
	return stats, nil
}

// GetAllChildProcesses recursively gets all child processes
func GetAllChildProcesses(pid int) ([]int, error) {
	var allChildren []int
	
	// First get direct children
	stats, err := ReadProcStats(pid)
	if err != nil {
		return nil, err
	}
	
	// Add direct children to the list
	allChildren = append(allChildren, stats.Children...)
	
	// Recursively get children of children
	for _, childPid := range stats.Children {
		grandchildren, err := GetAllChildProcesses(childPid)
		if err != nil {
			continue // It's ok if a child process disappeared
		}
		allChildren = append(allChildren, grandchildren...)
	}
	
	return allChildren, nil
} 