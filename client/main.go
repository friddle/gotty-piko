package main

import (
	"fmt"
	"os"

	"gotty-piko-client/src"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := MakeMainCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}

func MakeMainCmd() *cobra.Command {
	var (
		name       string
		remote     string
		serverPort int
		terminal   string
	)

	cmd := &cobra.Command{
		Use:   "gottyp",
		Short: "gotty-piko 客户端 - 基于终端的远程协助工具",
		Long: `gotty-piko 是一个基于终端的高效远程协助工具，集成了 gotty 和 piko 服务。
专为复杂网络环境下的远程协助而设计，避免传统远程桌面对高带宽的依赖。

使用示例:
  gottyp --name=my-server --remote=192.168.1.100:8088  # 连接到远程 piko 服务器
  gottyp --name=client1 --remote=piko.example.com:8022 # 连接到远程 piko 服务器
  gottyp --name=local --remote=192.168.1.100:8088 --terminal=zsh  # 指定使用 zsh`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 创建配置
			config := &src.Config{
				Name:       name,
				Remote:     remote,
				ServerPort: serverPort,
				Terminal:   terminal,
			}

			// 验证配置
			if err := config.Validate(); err != nil {
				return fmt.Errorf("配置验证失败: %v", err)
			}

			// 创建服务管理器
			manager := src.NewServiceManager(config)

			// 启动服务（会阻塞直到服务停止）
			if err := manager.Start(); err != nil {
				return fmt.Errorf("启动服务失败: %v", err)
			}

			return nil
		},
	}

	// 添加命令行参数
	cmd.Flags().StringVar(&name, "name", "", "piko 客户端标识名称")
	cmd.Flags().StringVar(&remote, "remote", "", "远程 piko 服务器地址 (格式: host:port)")
	cmd.Flags().IntVar(&serverPort, "server-port", 8022, "piko 服务器端口")
	cmd.Flags().StringVar(&terminal, "terminal", "", "指定要使用的终端类型 (zsh, bash, sh, powershell 等)")

	// 设置必需参数
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("remote")

	return cmd
}
