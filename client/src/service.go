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
func (sm *ServiceManager) PrintInfo() string {
	sm.config.GottyPort = sm.config.FindAvailablePort()
	if sm.config.StaticIndex == "." {
		if cwd, err := os.Getwd(); err == nil {
			sm.config.StaticIndex = cwd
		}
	}
	sm.printInfo()
	return sm.config.StaticIndex
}

func (sm *ServiceManager) Start() error {
	if sm.config.GottyPort == 0 {
		sm.config.GottyPort = sm.config.FindAvailablePort()
	}
	if env := os.Getenv("GOTTYP_STATIC_INDEX"); env != "" {
		sm.config.StaticIndex = env
	}
	return sm.startServices()
}

func (sm *ServiceManager) printInfo() {
	remoteHost := sm.config.GetRemoteHost()
	sessionPath := "/" + sm.config.Session + "/"

	fmt.Println("========================================")
	fmt.Printf("Remote URL: https://%s%s\n", remoteHost, sessionPath)
	if sm.config.Auth {
		fmt.Printf("Username:   %s\n", sm.config.AuthName)
		fmt.Printf("Password:   %s\n", sm.config.Pass)
	}
	if sm.config.AttachPort != "" {
		fmt.Printf("Port Proxy: https://%s%sport/%s\n", remoteHost, sessionPath, sm.config.AttachPort)
	}
	if sm.config.StaticIndex != "" {
		fmt.Printf("Files:      https://%s%sfiles/\n", remoteHost, sessionPath)
	}
	fmt.Println("========================================")
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
			fmt.Printf("\nReceived signal %v, shutting down...\n", sig)
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

func (sm *ServiceManager) startGotty() error {
	options := &server.Options{
		Address:       "127.0.0.1",
		Port:          fmt.Sprintf("%d", sm.config.GottyPort),
		Path:          "/" + sm.config.Session,
		SessionName:   sm.config.Session,
		PermitWrite:   true,
		TitleFormat:   "{{ .session_name }}",
		WSOrigin:      ".*",
		Auth:          sm.config.Auth,
		EnableNotify:  sm.config.EnableNotify,
		NotifyWebhook: sm.config.NotifyWebhook,
		StaticIndex:   sm.config.StaticIndex,
		AttachPort:    sm.config.AttachPort,
		TitleVariables: map[string]interface{}{
			"command":      sm.getShell(),
			"session_name": sm.config.Session,
		},
	}

	if sm.config.Auth {
		options.AuthName = sm.config.AuthName
		options.Password = sm.config.Pass
		options.EnableBasicAuth = true
	}

	notifier := server.NewNotifier(sm.config.NotifyWebhook)
	notifier.Start(sm.config.EnableNotify, "", sm.config.Session)

	backendOptions := &localcommand.Options{}
	if prefix := notifier.PathPrefix(); prefix != "" {
		backendOptions.EnvExtra = map[string]string{
			"PATH": prefix + os.Getenv("PATH"),
		}
	}

	var factory *localcommand.Factory
	var err error

	if sm.config.Tmux {
		if sm.isTmuxAvailable() {
			sessionName := sm.config.TmuxSession
			if sessionName == "" {
				sessionName = "gotty-" + sm.config.Session
			}
			cmd := "tmux"
			args := []string{"new", "-A", "-s", sessionName, sm.getShell()}
			factory, err = localcommand.NewFactory(cmd, args, backendOptions)
			fmt.Printf("✅ 使用 tmux 保持会话\n")
		} else {
			fmt.Printf("ℹ️  tmux not found, using plain shell. Install tmux for a better persistent session experience.\n")
			factory, err = localcommand.NewFactory(sm.getShell(), []string{}, backendOptions)
		}
	} else {
		factory, err = localcommand.NewFactory(sm.getShell(), []string{}, backendOptions)
	}

	if err != nil {
		return fmt.Errorf("创建 gotty 工厂失败: %v", err)
	}

	srv, err := server.NewWithNotifier(factory, options, notifier)
	if err != nil {
		return fmt.Errorf("创建 gotty 服务器失败: %v", err)
	}

	go func() {
		err := srv.Run(sm.ctx)
		if err != nil && err != context.Canceled {
			fmt.Printf("gotty 服务器运行错误: %v\n", err)
		}
	}()

	return nil
}

func (sm *ServiceManager) startPiko() error {
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
				EndpointID: sm.config.Session,
				Protocol:   config.ListenerProtocolHTTP,
				Addr:       fmt.Sprintf("127.0.0.1:%d", sm.config.GottyPort),
				AccessLog:  false,
				Timeout:    30 * time.Second,
				TLS:        config.TLSConfig{},
			},
		},
		Log: log.Config{
			Level:      "debug",
			Subsystems: []string{"client"},
		},
		GracePeriod: 30 * time.Second,
	}

	// 创建日志记录器
	logger, err := log.NewLogger("info", []string{})
	if err != nil {
		return fmt.Errorf("failed to create logger: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return fmt.Errorf("piko config validation failed: %v", err)
	}

	connectURL, err := url.Parse(conf.Connect.URL)
	if err != nil {
		return fmt.Errorf("failed to parse connect URL: %v", err)
	}

	// 创建上游客户端
	upstream := &client.Upstream{
		URL:       connectURL,
		TLSConfig: nil, // 不使用 TLS
		Logger:    logger.WithSubsystem("client"),
	}

	// 为每个监听器创建连接
	for _, listenerConfig := range conf.Listeners {
		fmt.Printf("[piko] connecting to endpoint: %s, remote: %s\n", listenerConfig.EndpointID, remote)
		ln, err := upstream.Listen(sm.ctx, listenerConfig.EndpointID)
		if err != nil {
			return fmt.Errorf("failed to listen on endpoint %s: %v", listenerConfig.EndpointID, err)
		}
		fmt.Printf("[piko] connected to endpoint: %s\n", listenerConfig.EndpointID)

		metrics := reverseproxy.NewMetrics("proxy")
		proxySrv := reverseproxy.NewServer(listenerConfig, metrics, logger)
		if proxySrv == nil {
			return fmt.Errorf("failed to create HTTP proxy server")
		}

		go func() {
			if err := proxySrv.Serve(ln); err != nil && err != context.Canceled {
				fmt.Printf("proxy server error: %v\n", err)
			}
		}()
	}

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

// isTmuxAvailable 检查 tmux 是否可用
func (sm *ServiceManager) isTmuxAvailable() bool {
	// 首先尝试使用 exec.LookPath 来检查命令是否在 PATH 中
	_, err := os.Stat("/usr/bin/tmux")
	if err == nil {
		return true
	}

	_, err = os.Stat("/bin/tmux")
	if err == nil {
		return true
	}

	_, err = os.Stat("/usr/local/bin/tmux")
	return err == nil
}
