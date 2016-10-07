package libsandbox

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

// workaround use this wait will not get exit status 3 error.
func wait(pid int, rusage *syscall.Rusage) (int, *syscall.WaitStatus, error) {
	var status syscall.WaitStatus
	var siginfo [128]byte
	// If we can block until Wait4 will succeed immediately, do so.
	psig := &siginfo[0]
	_, _, e := syscall.Syscall6(syscall.SYS_WAITID, 1, uintptr(pid), uintptr(unsafe.Pointer(psig)), syscall.WEXITED|syscall.WNOWAIT, 0, 0)
	// psig may be garbage collected before
	// syscall, KeepAlive make it alive util
	// sysacll return.
	runtime.KeepAlive(psig)
	if e != 0 {
		if e != syscall.ENOSYS {
			return 0, nil, os.NewSyscallError("waitid", e)
		}
	}
	wpid, err := syscall.Wait4(pid, &status, 0, rusage) // for rusage collect
	if err != nil {
		return 0, nil, err
	}
	return wpid, &status, err
}

// TODO: Do not use stdout to communicate

// sandbox config
type Config struct {
	Args   []string
	Input  io.Reader
	Memory int64
	Time   int64
}

func (conf Config) Validate() error {
	if len(conf.Args) == 0 {
		return errors.New("process or arguments not set")
	}
	if conf.Memory == 0 {
		return errors.New("memory limit not set")
	}
	if conf.Time == 0 {
		return errors.New("time limit not set")
	}
	return nil
}

var (
	OutOfTimeError   = errors.New("out of time")
	OutOfMemoryError = errors.New("out of memory")
)

func RuntimeError(signal os.Signal) error {
	return fmt.Errorf("receive signal %v", signal)
}

type StdSandbox struct {
	Bin         string    // binary path
	Args        []string  // arguments
	Input       io.Reader // standard input
	TimeLimit   int64     // time limit in ms
	MemoryLimit int64     // memory limit in kb

	time   int64 // time used
	memory int64 // memory used
}

func NewStdSandbox(conf Config) (Sandbox, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}
	var args []string
	if len(conf.Args) > 1 {
		args = conf.Args[1:]

	}
	return &StdSandbox{
		Bin:         conf.Args[0],
		Args:        args,
		Input:       conf.Input,
		MemoryLimit: conf.Memory,
		TimeLimit:   conf.Time,
	}, nil
}

func (s *StdSandbox) Time() int64 {
	return s.time
}
func (s *StdSandbox) Memory() int64 {
	return s.memory
}

func (s *StdSandbox) Run() ([]byte, error) {
	cmd := exec.Command(s.Bin, s.Args...)
	if cmd.Stdin != nil {
		return nil, errors.New("stdin is not nil")
	}
	if cmd.Stderr != nil {
		return nil, errors.New("stdout is not nil")
	}
	if cmd.Stdout != nil {
		return nil, errors.New("stdout is not nil")
	}
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	cmd.Stdout = buf
	cmd.Stdin = s.Input
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	defer cmd.Process.Kill()
	var errCh chan error = make(chan error)
	go func() {
		// workaround 不知道为什么标准的wait不行
		var rusage syscall.Rusage
		_, status, err := wait(cmd.Process.Pid, &rusage)
		if err != nil {
			fmt.Println("wait error", err)
			errCh <- err
		}
		if status.Signaled() {
			errCh <- fmt.Errorf("get signal %s", status.Signal())
		}
		close(errCh)
	}()

	// Send signal SIGSTOP to the process every tick.
	ticker := time.NewTicker(TICK)
	for range ticker.C {
		ok, vm, rss, runningTime, cpuTime := GetResourceUsage(cmd.Process.Pid)
		if !ok {
			break
		}

		// Like sleep, some process consumes no cpu usage, but does
		// consume runnig time, so here limit real runnig time to
		// 150% cpu usage time.
		if cpuTime > s.TimeLimit ||
			runningTime > 3*s.TimeLimit/2 {
			s.time = runningTime
			err = OutOfTimeError
			break

		}

		// RSS size dosen't include swap out memory,
		// virtual memory dosen't include memory demand-loaded int.
		// So set limit: memory < 150% * rss and vm > memory*1000%
		if rss*3 > s.MemoryLimit*2 ||
			vm > s.MemoryLimit*10 {
			err = OutOfMemoryError
			s.memory = rss * 3 / 2
			err = OutOfMemoryError
			break

		}
	}
	if err != nil {
		return nil, err
	}
	err = <-errCh
	return buf.Bytes(), err
}

type Sandbox interface {
	Run() (output []byte, err error)
	Time() int64
	Memory() int64
}
