package main

import (
	"fmt"
	"os"

	"gotty-piko-client/src"

	"github.com/spf13/cobra"
)

// main 程序入口点
func main() {
	rootCmd := createMainCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// createMainCommand 创建主命令
func createMainCommand() *cobra.Command {
	var (
		name       string
		remote     string
		serverPort int
		terminal   string
		autoExit   bool
		pass       string
	)

	cmd := &cobra.Command{
		Use:   "gottyp",
		Short: "Gotty-Piko Client - Terminal-based remote assistance tool",
		Long: `Gotty-Piko is an efficient terminal-based remote assistance tool that integrates gotty and piko services.
Designed for remote assistance in complex network environments, avoiding the high bandwidth dependency of traditional remote desktop.

Features:
- Terminal-based remote access
- HTTP authentication support
- Automatic port allocation
- Cross-platform compatibility
- 24-hour auto-exit option

Examples:
  gottyp --name=my-server --remote=192.168.1.100:8088 --pass=mypassword
  gottyp --name=client1 --remote=piko.example.com:8022 --pass=secret123
  gottyp --name=local --remote=192.168.1.100:8088 --terminal=zsh --pass=localpass
  gottyp --name=server --remote=192.168.1.100:8088 --auto-exit=false --pass=serverpass`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGottypService(name, remote, serverPort, terminal, autoExit, pass)
		},
	}

	// Add command line flags
	addCommandFlags(cmd, &name, &remote, &serverPort, &terminal, &autoExit, &pass)

	// Set required flags
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("remote")

	return cmd
}

// addCommandFlags 添加命令行标志
func addCommandFlags(cmd *cobra.Command, name, remote *string, serverPort *int, terminal *string, autoExit *bool, pass *string) {
	cmd.Flags().StringVar(name, "name", "", "Piko client identifier name")
	cmd.Flags().StringVar(remote, "remote", "", "Remote piko server address (format: host:port)")
	cmd.Flags().IntVar(serverPort, "server-port", 8022, "Piko server port")
	cmd.Flags().StringVar(terminal, "terminal", "", "Specify terminal type (zsh, bash, sh, powershell, etc.)")
	cmd.Flags().BoolVar(autoExit, "auto-exit", true, "Enable 24-hour auto-exit (default: true)")
	cmd.Flags().StringVar(pass, "pass", "", "HTTP authentication password")
}

// runGottypService 运行gottyp服务
func runGottypService(name, remote string, serverPort int, terminal string, autoExit bool, pass string) error {
	// Create configuration
	config := &src.Config{
		Name:       name,
		Remote:     remote,
		ServerPort: serverPort,
		Terminal:   terminal,
		AutoExit:   autoExit,
		Pass:       pass,
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %v", err)
	}

	// Create service manager
	manager := src.NewServiceManager(config)

	// Start service (blocks until service stops)
	if err := manager.Start(); err != nil {
		return fmt.Errorf("failed to start service: %v", err)
	}

	return nil
}
