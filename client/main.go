package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"gotty-piko-client/src"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := MakeMainCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func MakeMainCmd() *cobra.Command {
	var (
		session       string
		authName      string
		remote        string
		terminal      string
		autoExit      bool
		pass          string
		tmux          bool
		auth          bool
		enableNotify  bool
		notifyWebhook string
		staticIndex   string
		attachPort    string
		daemon        bool
		pidFile       string
	)

	cmd := &cobra.Command{
		Use:   "gottyp",
		Short: "Share your terminal as a web application via piko",
		Long: `gottyp is a one-shot tool that integrates gotty and piko.
It starts a local gotty terminal session and registers it with a remote piko server,
making your terminal accessible via a web browser.

Examples:
  gottyp --remote=piko.example.com:8088
  gottyp --remote=piko.example.com:8088 --session myterm --tmux=true
  gottyp --remote=piko.example.com:8088 --auth=false
  gottyp --remote=piko.example.com:8088 --notify-webhook=https://open.feishu.cn/...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := &src.Config{
				Session:       session,
				AuthName:      authName,
				Remote:        remote,
				Terminal:      terminal,
				AutoExit:      autoExit,
				Pass:          pass,
				Tmux:          tmux,
				Auth:          auth,
				EnableNotify:  enableNotify,
				NotifyWebhook: notifyWebhook,
				StaticIndex:   staticIndex,
				AttachPort:    attachPort,
				Daemon:        daemon,
				PidFile:       pidFile,
			}

			if err := config.Validate(); err != nil {
				return err
			}

			manager := src.NewServiceManager(config)

			if config.Daemon {
				staticIndex := manager.PrintInfo()
				if err := src.Daemonize(staticIndex, config.PidFile); err != nil {
					return fmt.Errorf("failed to daemonize: %v", err)
				}
			}

			if src.IsDaemonized() {
				f, err := os.OpenFile("/tmp/gottyp.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
				if err == nil {
					syscall.Dup2(int(f.Fd()), int(os.Stdout.Fd()))
					syscall.Dup2(int(f.Fd()), int(os.Stderr.Fd()))
				}
			}

			if err := manager.Start(); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&session, "session", "", "Session ID for endpoint path (default: user_dir_random)")
	cmd.Flags().StringVar(&authName, "auth-name", "", "Auth username for Basic Auth (auto-generated if not set)")
	cmd.Flags().StringVar(&remote, "remote", "https://clauded.friddle.me", "Remote piko server address")
	cmd.Flags().StringVar(&terminal, "terminal", "", "Terminal type (zsh, bash, sh, powershell, etc.)")
	cmd.Flags().BoolVar(&autoExit, "auto-exit", true, "Enable 24-hour auto exit")
	cmd.Flags().BoolVar(&tmux, "tmux", true, "Use tmux for persistent sessions")
	cmd.Flags().StringVar(&pass, "pass", "", "Auth password (auto-generated if not set)")
	cmd.Flags().BoolVar(&auth, "auth", true, "Enable Basic Authentication")
	cmd.Flags().BoolVar(&enableNotify, "enable-notify", true, "Enable notify-send interception")
	cmd.Flags().StringVar(&notifyWebhook, "notify-webhook", "", "Webhook URL to forward notifications to (Feishu compatible)")
	cmd.Flags().StringVar(&staticIndex, "static-index", ".", "Local directory to serve as static files at /files/")
	cmd.Flags().StringVar(&attachPort, "attach-port", "", "Map a local port to /port/ path (e.g. 3000)")
	cmd.Flags().BoolVar(&daemon, "daemon", true, "Run as daemon (background process)")
	cmd.Flags().StringVar(&pidFile, "pid-file", "/tmp/gottyp.pid", "PID file path for daemon mode")

	cmd.AddCommand(tmuxCmd())

	return cmd
}

func tmuxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tmux",
		Short: "Manage tmux sessions",
	}

	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Short:   "List tmux sessions",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTmux("list-sessions")
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "kill-all",
		Short:   "Kill all tmux sessions and gottyp daemons",
		Aliases: []string{"kill"},
		RunE: func(cmd *cobra.Command, args []string) error {
			src.KillAllDaemons()
			return runTmux("kill-server")
		},
	})

	return cmd
}

func runTmux(args ...string) error {
	bin, err := exec.LookPath("tmux")
	if err != nil {
		return fmt.Errorf("tmux not found")
	}
	cmd := exec.Command(bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
