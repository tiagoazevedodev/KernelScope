# KernelScope

KernelScope is a process execution and monitoring system for Linux, designed to run binaries under controlled resource constraints.

## Features

- Execute binaries with CPU time, memory, and timeout limits
- Monitor process execution in real-time
- Support for process tree monitoring (parent + child processes)
- Interactive loop execution mode
- Prepaid (credit-based) and postpaid execution modes
- Detailed execution reports

## Requirements

- Linux-based system (most features will only work on Linux)
- Go 1.16 or higher

## Installation

```bash
# Clone the repository
git clone https://github.com/username/kernelscope.git
cd kernelscope

# Build the project
go build -o kernelscope
```

## Usage

```bash
# Basic usage
./kernelscope --binary /path/to/executable

# With resource limits
./kernelscope --binary /path/to/executable --cpu 10 --mem 1024 --timeout 30

# Prepaid mode with CPU credits
./kernelscope --binary /path/to/executable --prepaid --credit 5.0

# Postpaid mode
./kernelscope --binary /path/to/executable --prepaid=false
```

### Command-line Options

- `--binary`: Path to the binary to execute (required)
- `--cpu`: CPU time limit in seconds (default: 10)
- `--mem`: Memory limit in KB (default: 1048576)
- `--timeout`: Timeout in seconds (default: 30)
- `--prepaid`: Run in prepaid mode (true) or postpaid mode (false) (default: true)
- `--credit`: CPU credits in seconds for prepaid mode (default: 5.0)

## How It Works

KernelScope operates using the following components:

1. **CLI Module**: Parses command-line arguments and configuration
2. **Executor**: Handles process execution and termination
3. **Resource Manager**: Sets and enforces resource limits
4. **Monitor**: Continuously monitors resource usage
5. **Loop Controller**: Manages the main execution loop
6. **Reporter**: Generates execution reports and statistics

In prepaid mode, KernelScope will only deduct CPU time for successful executions, allowing for a more efficient use of resources. In postpaid mode, all CPU time is counted regardless of success.

## Limitations

- Resource monitoring heavily relies on the Linux `/proc` filesystem
- Some features may not work or provide accurate data on non-Linux systems
- Process resource limits are enforced using Linux-specific system calls

## License

MIT 