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

// ServiceManager manages the gotty and piko services
type ServiceManager struct {
	Config *Config
	ctx    context.Context
	cancel context.CancelFunc
}

// NewServiceManager creates a new service manager
func NewServiceManager(config *Config) *ServiceManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &ServiceManager{
		Config: config,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start starts all services
func (sm *ServiceManager) Start() error {
	fmt.Printf("🚀 Starting gotty-piko client\n")
	fmt.Printf("Client name: %s\n", sm.Config.Name)
	fmt.Printf("Remote server: %s\n", sm.Config.Remote)
	fmt.Printf("Terminal: %s\n", sm.getShell())
	fmt.Printf("Auto-exit: %t\n", sm.Config.AutoExit)

	// Auto-allocate available port
	sm.Config.GottyPort = sm.Config.FindAvailablePort()
	fmt.Printf("Local listening port: %d\n", sm.Config.GottyPort)

	// Use oklog/run to start services
	return sm.startServices()
}

// startServices uses oklog/run to start all services
func (sm *ServiceManager) startServices() error {
	var g run.Group

	// Start piko service
	g.Add(func() error {
		err := sm.startPiko()
		if err != nil {
			fmt.Printf("Failed to start piko: %v\n", err)
			return err
		}
		// Wait for context cancellation
		<-sm.ctx.Done()
		return sm.ctx.Err()
	}, func(error) {
		// Piko service will automatically stop when context is cancelled
	})

	// Start gotty service
	g.Add(func() error {
		err := sm.startGotty()
		if err != nil {
			fmt.Printf("Failed to start gotty: %v\n", err)
			return err
		}
		// Wait for context cancellation
		<-sm.ctx.Done()
		return sm.ctx.Err()
	}, func(error) {
		// Gotty service will automatically stop when context is cancelled
	})

	// Signal handling
	g.Add(func() error {
		c := make(chan os.Signal, 1)

		// Set different signals based on operating system
		if runtime.GOOS == "windows" {
			// Windows supports Ctrl+C (SIGINT) and Ctrl+Break
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		} else {
			// Unix-like systems support more signals
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		}

		select {
		case sig := <-c:
			fmt.Printf("\n🛑 Received stop signal %v, shutting down services...\n", sig)
			sm.cancel() // Immediately cancel context
			return nil
		case <-sm.ctx.Done():
			return sm.ctx.Err()
		}
	}, func(error) {
		sm.cancel()
	})

	// 24-hour timeout - only enabled when AutoExit is true
	if sm.Config.AutoExit {
		g.Add(func() error {
			timeoutCtx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
			defer cancel()

			select {
			case <-timeoutCtx.Done():
				fmt.Printf("\n⏰ Service runtime reached 24 hours, stopping...\n")
				sm.cancel()
				return nil
			case <-sm.ctx.Done():
				return sm.ctx.Err()
			}
		}, func(error) {
			sm.cancel()
		})
	}

	fmt.Printf("✅ Services started successfully!\n")
	fmt.Printf("🌐 Access URL: http://localhost:%d\n", sm.Config.GottyPort)
	if sm.Config.Pass != "" {
		fmt.Printf("🔐 HTTP Authentication: username=%s, password=%s\n", sm.Config.Name, sm.Config.Pass)
	} else {
		fmt.Printf("⚠️  HTTP authentication not enabled\n")
	}
	fmt.Printf("Press Ctrl+C to stop services\n")

	// Run all services
	return g.Run()
}

// Stop stops all services
func (sm *ServiceManager) Stop() {
	fmt.Printf("✅ Services stopped\n")
}

// startGotty starts the gotty service
func (sm *ServiceManager) startGotty() error {
	// Create gotty server options
	fmt.Print("Starting gotty...")
	options := &server.Options{
		Address:         "127.0.0.1",
		Port:            fmt.Sprintf("%d", sm.Config.GottyPort),
		Path:            "/" + sm.Config.Name,
		PermitWrite:     true,
		TitleFormat:     "{{ .command }}@{{ .hostname }}",
		WSOrigin:        ".*",                 // Allow WebSocket connections from all origins
		EnableBasicAuth: sm.Config.Pass != "", // Enable HTTP basic auth only when password is not empty
	}

	if sm.Config.Pass != "" {
		options.Credential = sm.Config.Name + ":" + sm.Config.Pass // Set auth credentials: username:password
	}

	// Create local command factory
	backendOptions := &localcommand.Options{}
	factory, err := localcommand.NewFactory(sm.getShell(), []string{}, backendOptions)
	if err != nil {
		return fmt.Errorf("failed to create gotty factory: %v", err)
	}

	// Create gotty server
	srv, err := server.New(factory, options)
	if err != nil {
		return fmt.Errorf("failed to create gotty server: %v", err)
	}

	// Start gotty server in separate goroutine
	go func() {
		err := srv.Run(sm.ctx)
		if err != nil && err != context.Canceled {
			fmt.Printf("Gotty server runtime error: %v\n", err)
		}
	}()

	fmt.Print("Gotty started\n")
	return nil
}

// startPiko starts the piko client
func (sm *ServiceManager) startPiko() error {
	// Create piko configuration
	fmt.Printf("Starting piko...\n")
	remote := sm.Config.Remote
	if strings.HasPrefix(remote, "http") {
		remote = sm.Config.Remote
	} else {
		remote = fmt.Sprintf("http://%s", sm.Config.Remote)
	}
	conf := &config.Config{
		Connect: config.ConnectConfig{
			URL:     remote,
			Timeout: 30 * time.Second,
		},
		Listeners: []config.ListenerConfig{
			{
				EndpointID: sm.Config.Name,
				Protocol:   config.ListenerProtocolHTTP,
				Addr:       fmt.Sprintf("127.0.0.1:%d", sm.Config.GottyPort),
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

	// Create logger
	logger, err := log.NewLogger("info", []string{})
	if err != nil {
		return fmt.Errorf("failed to create logger: %v", err)
	}

	// Validate configuration
	if err := conf.Validate(); err != nil {
		return fmt.Errorf("piko configuration validation failed: %v", err)
	}

	// Parse connection URL
	connectURL, err := url.Parse(conf.Connect.URL)
	if err != nil {
		return fmt.Errorf("failed to parse connection URL: %v", err)
	}

	// Create upstream client
	upstream := &client.Upstream{
		URL:       connectURL,
		TLSConfig: nil, // No TLS
		Logger:    logger.WithSubsystem("client"),
	}

	// Create connection for each listener
	for _, listenerConfig := range conf.Listeners {
		fmt.Printf("Connecting to endpoint: %s\n", listenerConfig.EndpointID)

		ln, err := upstream.Listen(sm.ctx, listenerConfig.EndpointID)
		if err != nil {
			return fmt.Errorf("failed to listen on endpoint %s: %v", listenerConfig.EndpointID, err)
		}

		fmt.Printf("Successfully connected to endpoint: %s\n", listenerConfig.EndpointID)

		// Create HTTP proxy server
		metrics := reverseproxy.NewMetrics("proxy")
		server := reverseproxy.NewServer(listenerConfig, metrics, logger)
		if server == nil {
			return fmt.Errorf("failed to create HTTP proxy server")
		}

		// Start proxy server
		go func() {
			if err := server.Serve(ln); err != nil && err != context.Canceled {
				fmt.Printf("Proxy server runtime error: %v\n", err)
			}
		}()
	}

	fmt.Printf("Piko started\n")
	return nil
}

// getShell gets the appropriate shell based on operating system
func (sm *ServiceManager) getShell() string {
	// If terminal is specified in config, use it first
	if sm.Config.Terminal != "" {
		// Verify if specified terminal is available
		if sm.isShellAvailable(sm.Config.Terminal) {
			return sm.Config.Terminal
		}
		// If specified terminal is not available, output warning and continue with default logic
		fmt.Printf("⚠️  Specified terminal %s is not available, using default terminal\n", sm.Config.Terminal)
	}

	// Use default shell selection logic
	switch runtime.GOOS {
	case "windows":
		return "powershell"
	case "linux":
		// On Linux, prefer zsh, then bash, finally sh
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

// isShellAvailable checks if specified shell is available
func (sm *ServiceManager) isShellAvailable(shell string) bool {
	_, err := os.Stat(fmt.Sprintf("/bin/%s", shell))
	if err == nil {
		return true
	}

	// Also check /usr/bin directory
	_, err = os.Stat(fmt.Sprintf("/usr/bin/%s", shell))
	return err == nil
}
