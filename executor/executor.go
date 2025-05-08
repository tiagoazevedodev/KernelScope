package executor

import (
	"fmt"
	"kernelscope/cli"
	"os"
	"os/exec"
	"syscall"
)

// Process represents a running process
type Process struct {
	Cmd    *exec.Cmd
	Pid    int
	Config *cli.Config
}

// Executor handles process execution
type Executor struct {
	Config *cli.Config
}

// NewExecutor creates a new Executor with the given configuration
func NewExecutor(config *cli.Config) *Executor {
	return &Executor{
		Config: config,
	}
}

// StartProcess starts the binary with the specified arguments
func (e *Executor) StartProcess() (*Process, error) {
	fmt.Printf("Starting process: %s\n", e.Config.BinaryPath)
	
	cmd := exec.Command(e.Config.BinaryPath)
	
	// Configure output redirection
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	// Configure process attributes for resource control
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// We'll use Setpgid to create a new process group
		// This helps with killing child processes later
		Setpgid: true,
	}
	
	// Start the process
	err := cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start process: %v", err)
	}
	
	fmt.Printf("Process started with PID: %d\n", cmd.Process.Pid)
	
	return &Process{
		Cmd:    cmd,
		Pid:    cmd.Process.Pid,
		Config: e.Config,
	}, nil
}

// KillProcess kills the specified process
func (e *Executor) KillProcess(process *Process) error {
	fmt.Printf("Killing process with PID: %d\n", process.Pid)
	
	// Kill the process group to also terminate any child processes
	pgid, err := syscall.Getpgid(process.Pid)
	if err == nil {
		// First try a graceful termination
		_ = syscall.Kill(-pgid, syscall.SIGTERM)
		
		// Then force kill after a small delay if still running
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
	}
	
	return process.Cmd.Process.Kill()
}

// WaitForProcess waits for the process to complete and returns exit code
func (e *Executor) WaitForProcess(process *Process) (int, error) {
	err := process.Cmd.Wait()
	
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Process exited with non-zero status
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus(), nil
			}
		}
		return -1, err
	}
	
	return 0, nil // Process exited successfully
} 