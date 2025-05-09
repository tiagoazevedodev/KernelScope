package executor

import (
	"fmt"
	"kernelscope/cli"
	"os"
	"os/exec"
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

	// Create command with the binary path
	cmd := exec.Command(e.Config.BinaryPath)

	// Configure output redirection
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

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

	// Simply kill the process - this works across platforms
	return process.Cmd.Process.Kill()
}

// WaitForProcess waits for the process to complete and returns exit code
func (e *Executor) WaitForProcess(process *Process) (int, error) {
	err := process.Cmd.Wait()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Process exited with non-zero status
			return exitErr.ExitCode(), nil
		}
		return -1, err
	}

	return 0, nil // Process exited successfully
}

// CheckProcess checks if the process is still running without waiting
func (e *Executor) CheckProcess(process *Process) (int, error) {
	if process == nil || process.Cmd == nil || process.Cmd.Process == nil {
		return -1, fmt.Errorf("invalid process")
	}

	// Try to get process state
	if process.Cmd.ProcessState != nil {
		// Process has exited
		return process.Cmd.ProcessState.ExitCode(), nil
	}

	// Process is likely still running
	// Use a platform-independent way to check
	// Just send a signal 0 with Process.Signal
	err := process.Cmd.Process.Signal(os.Signal(nil))
	if err != nil {
		// Process doesn't exist anymore or can't be signaled
		return -1, fmt.Errorf("process no longer exists or can't be signaled")
	}

	// Process is still running
	return -1, nil
}
