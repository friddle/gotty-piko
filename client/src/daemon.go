package src

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func Daemonize(staticIndex string, pidFile string) error {
	if syscall.Getppid() == 1 {
		return nil
	}

	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	args := os.Args[1:]
	for i, a := range args {
		if a == "--daemon" {
			args[i] = "--daemon=false"
		}
	}

	cmd := exec.Command(self, args...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Env = append(os.Environ(),
		"GOTTYP_DAEMONIZED=1",
		"GOTTYP_STATIC_INDEX="+staticIndex,
	)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	pidData := strconv.Itoa(cmd.Process.Pid) + "\n"
	if err := os.WriteFile(pidFile, []byte(pidData), 0644); err != nil {
		return fmt.Errorf("failed to write pid file: %w", err)
	}

	os.Exit(0)
	return nil
}

func IsDaemonized() bool {
	return os.Getenv("GOTTYP_DAEMONIZED") == "1"
}

func KillAllDaemons() {
	out, err := exec.Command("pgrep", "-f", "gottyp").Output()
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pid, err := strconv.Atoi(line)
		if err != nil {
			continue
		}
		if pid == os.Getpid() {
			continue
		}
		syscall.Kill(pid, syscall.SIGTERM)
		fmt.Printf("killed gottyp daemon (pid %d)\n", pid)
	}
	_ = os.Remove("/tmp/gottyp.pid")
}
