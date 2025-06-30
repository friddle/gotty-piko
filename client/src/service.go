package src

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/andydunstall/piko/agent/config"
	"github.com/andydunstall/piko/agent/reverseproxy"
	"github.com/andydunstall/piko/client"
	"github.com/andydunstall/piko/pkg/log"
	"github.com/oklog/run"
	"github.com/sorenisanerd/gotty/backend/localcommand"
	"github.com/sorenisanerd/gotty/server"
)

// ServiceManager 服务管理器
type ServiceManager struct {
	config *Config
	ctx    context.Context
	cancel context.CancelFunc
}

// NewServiceManager 创建新的服务管理器
func NewServiceManager(config *Config) *ServiceManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &ServiceManager{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start 启动所有服务
func (sm *ServiceManager) Start() error {
	fmt.Printf("🚀 启动 gotty-piko 客户端\n")
	fmt.Printf("客户端名称: %s\n", sm.config.Name)
	fmt.Printf("远程服务器: %s\n", sm.config.Remote)
	fmt.Printf("使用终端: %s\n", sm.getShell())
	fmt.Printf("自动退出: %t\n", sm.config.AutoExit)

	// 自动分配可用端口
	sm.config.GottyPort = sm.config.FindAvailablePort()
	fmt.Printf("本地监听端口: %d\n", sm.config.GottyPort)

	// 使用 oklog/run 启动服务
	return sm.startServices()
}

// startServices 使用 oklog/run 启动所有服务
func (sm *ServiceManager) startServices() error {
	var g run.Group

	// 启动 piko 服务
	g.Add(func() error {
		err := sm.startPiko()
		if err != nil {
			fmt.Printf("启动piko失败:%v\n", err)
			return err
		}
		// 等待 context 取消
		<-sm.ctx.Done()
		return sm.ctx.Err()
	}, func(error) {
		// piko 服务会在 context 取消时自动停止
	})

	// 启动 gotty 服务
	g.Add(func() error {
		err := sm.startGotty()
		if err != nil {
			fmt.Printf("启动gotty失败:%v\n", err)
			return err
		}
		// 等待 context 取消
		<-sm.ctx.Done()
		return sm.ctx.Err()
	}, func(error) {
		// gotty 服务会在 context 取消时自动停止
	})

	// 信号处理 - 移到主流程中
	g.Add(func() error {
		c := make(chan os.Signal, 1)

		// 根据操作系统设置不同的信号
		if runtime.GOOS == "windows" {
			// Windows 支持 Ctrl+C (SIGINT) 和 Ctrl+Break
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		} else {
			// Unix-like 系统支持更多信号
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		}

		select {
		case sig := <-c:
			fmt.Printf("\n🛑 收到停止信号 %v，正在关闭服务...\n", sig)
			sm.cancel() // 立即取消 context
			return nil
		case <-sm.ctx.Done():
			return sm.ctx.Err()
		}
	}, func(error) {
		sm.cancel()
	})

	// 24小时超时 - 只有当 AutoExit 为 true 时才启用
	if sm.config.AutoExit {
		g.Add(func() error {
			timeoutCtx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
			defer cancel()

			select {
			case <-timeoutCtx.Done():
				fmt.Printf("\n⏰ 服务运行时间达到24小时，正在停止...\n")
				sm.cancel()
				return nil
			case <-sm.ctx.Done():
				return sm.ctx.Err()
			}
		}, func(error) {
			sm.cancel()
		})
	}

	fmt.Printf("✅ 服务启动成功！\n")
	fmt.Printf("🌐 访问地址: http://localhost:%d\n", sm.config.GottyPort)
	if sm.config.Pass != "" {
		fmt.Printf("🔐 HTTP认证: 用户名=%s, 密码=%s\n", sm.config.Name, sm.config.Pass)
	} else {
		fmt.Printf("⚠️  未启用HTTP认证\n")
	}
	fmt.Printf("按 Ctrl+C 停止服务\n")

	// 运行所有服务
	return g.Run()
}

// Wait 等待服务运行（已废弃，使用 Start 方法）
func (sm *ServiceManager) Wait() {
	fmt.Printf("⚠️  Wait 方法已废弃，请使用 Start 方法\n")
}

// Stop 停止所有服务
func (sm *ServiceManager) Stop() {
	fmt.Printf("✅ 服务已停止\n")
}

// startGotty 启动 gotty
func (sm *ServiceManager) startGotty() error {
	// 创建 gotty 服务器选项
	fmt.Print("启动gotty中....")
	options := &server.Options{
		Address:         "127.0.0.1",
		Port:            fmt.Sprintf("%d", sm.config.GottyPort),
		Path:            "/" + sm.config.Name,
		PermitWrite:     true,
		TitleFormat:     "{{ .command }}@{{ .hostname }}",
		WSOrigin:        ".*",                 // 允许所有来源的 WebSocket 连接
		EnableBasicAuth: sm.config.Pass != "", // 只有当密码不为空时才启用HTTP基本认证
	}

	if sm.config.Pass != "" {
		options.Credential = sm.config.Name + ":" + sm.config.Pass // 设置认证凭据：用户名:密码
	}

	// 创建本地命令工厂
	backendOptions := &localcommand.Options{}
	factory, err := localcommand.NewFactory(sm.getShell(), []string{}, backendOptions)
	if err != nil {
		return fmt.Errorf("创建 gotty 工厂失败: %v", err)
	}

	// 创建 gotty 服务器
	srv, err := server.New(factory, options)
	if err != nil {
		return fmt.Errorf("创建 gotty 服务器失败: %v", err)
	}

	// 在独立的 goroutine 中启动 gotty 服务器
	go func() {
		err := srv.Run(sm.ctx)
		if err != nil && err != context.Canceled {
			fmt.Printf("gotty 服务器运行错误: %v\n", err)
		}
	}()

	fmt.Print("启动gotty结束\n")
	return nil
}

// startPiko 启动 piko 客户端
func (sm *ServiceManager) startPiko() error {
	// 创建 piko 配置
	fmt.Printf("启动piko中\n")
	remote := sm.config.Remote
	if strings.HasPrefix(remote, "http") {
		remote = sm.config.Remote
	} else {
		remote = fmt.Sprintf("http://%s", sm.config.Remote)
	}
	conf := &config.Config{
		Connect: config.ConnectConfig{
			URL:     remote,
			Timeout: 30 * time.Second,
		},
		Listeners: []config.ListenerConfig{
			{
				EndpointID: sm.config.Name,
				Protocol:   config.ListenerProtocolHTTP,
				Addr:       fmt.Sprintf("127.0.0.1:%d", sm.config.GottyPort),
				AccessLog:  false,
				Timeout:    30 * time.Second,
				TLS:        config.TLSConfig{},
			},
		},
		Log: log.Config{
			Level:      "info",
			Subsystems: []string{},
		},
		GracePeriod: 30 * time.Second,
	}

	// 创建日志记录器
	logger, err := log.NewLogger("info", []string{})
	if err != nil {
		return fmt.Errorf("创建日志记录器失败: %v", err)
	}

	// 验证配置
	if err := conf.Validate(); err != nil {
		return fmt.Errorf("piko 配置验证失败: %v", err)
	}

	// 解析连接 URL
	connectURL, err := url.Parse(conf.Connect.URL)
	if err != nil {
		return fmt.Errorf("解析连接 URL 失败: %v", err)
	}

	// 创建上游客户端
	upstream := &client.Upstream{
		URL:       connectURL,
		TLSConfig: nil, // 不使用 TLS
		Logger:    logger.WithSubsystem("client"),
	}

	// 为每个监听器创建连接
	for _, listenerConfig := range conf.Listeners {
		fmt.Printf("正在连接到端点: %s\n", listenerConfig.EndpointID)

		ln, err := upstream.Listen(sm.ctx, listenerConfig.EndpointID)
		if err != nil {
			return fmt.Errorf("监听端点失败 %s: %v", listenerConfig.EndpointID, err)
		}

		fmt.Printf("成功连接到端点: %s\n", listenerConfig.EndpointID)

		// 创建 HTTP 代理服务器，传入正确的配置而不是 nil
		metrics := reverseproxy.NewMetrics("proxy")
		server := reverseproxy.NewServer(listenerConfig, metrics, logger)
		if server == nil {
			return fmt.Errorf("创建 HTTP 代理服务器失败")
		}

		// 启动代理服务器
		go func() {
			if err := server.Serve(ln); err != nil && err != context.Canceled {
				fmt.Printf("代理服务器运行错误: %v\n", err)
			}
		}()
	}

	fmt.Printf("启动piko结束\n")
	return nil
}

// getShell 根据操作系统获取对应的shell
func (sm *ServiceManager) getShell() string {
	// 如果配置中指定了 terminal，优先使用配置的
	if sm.config.Terminal != "" {
		// 验证指定的 terminal 是否可用
		if sm.isShellAvailable(sm.config.Terminal) {
			return sm.config.Terminal
		}
		// 如果指定的 terminal 不可用，输出警告并继续使用默认逻辑
		fmt.Printf("⚠️  指定的终端 %s 不可用，将使用默认终端\n", sm.config.Terminal)
	}

	// 使用默认的 shell 选择逻辑
	switch runtime.GOOS {
	case "windows":
		return "powershell"
	case "linux":
		// 在 Linux 上优先使用 zsh，然后是 bash，最后是 sh
		if sm.isShellAvailable("zsh") {
			return "zsh"
		}
		if sm.isShellAvailable("bash") {
			return "bash"
		}
		return "sh"
	case "darwin":
		return "bash"
	default:
		return "sh"
	}
}

// isShellAvailable 检查指定的 shell 是否可用
func (sm *ServiceManager) isShellAvailable(shell string) bool {
	_, err := os.Stat(fmt.Sprintf("/bin/%s", shell))
	if err == nil {
		return true
	}

	// 也检查 /usr/bin 目录
	_, err = os.Stat(fmt.Sprintf("/usr/bin/%s", shell))
	return err == nil
}
