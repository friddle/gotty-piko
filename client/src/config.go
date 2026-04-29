package src

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Config 配置结构体
type Config struct {
	AuthName      string
	Session       string
	Remote        string
	ServerPort    int
	GottyPort     int
	Terminal      string
	Pass          string
	AutoExit      bool
	Tmux          bool
	Auth          bool
	EnableNotify  bool
	NotifyWebhook string
	StaticIndex   string
	AttachPort    string
	Daemon        bool
	PidFile       string
}

// NewConfig 创建新的配置实例
func NewConfig() *Config {
	return &Config{
		AuthName:      getEnvOrDefault("AUTH_NAME", ""),
		Session:       getEnvOrDefault("SESSION", ""),
		Remote:        getEnvOrDefault("REMOTE", ""),
		ServerPort:    getEnvIntOrDefault("SERVER_PORT", 8022),
		GottyPort:     0,
		Terminal:      getEnvOrDefault("TERMINAL", ""),
		AutoExit:      getEnvBoolOrDefault("AUTO_EXIT", true),
		Tmux:          getEnvBoolOrDefault("TMUX", true),
		Pass:          getEnvOrDefault("PASS", ""),
		Auth:          getEnvBoolOrDefault("AUTH", true),
		EnableNotify:  getEnvBoolOrDefault("ENABLE_NOTIFY", true),
		NotifyWebhook: getEnvOrDefault("NOTIFY_WEBHOOK", ""),
		StaticIndex:   getEnvOrDefault("STATIC_INDEX", "."),
		AttachPort:    getEnvOrDefault("ATTACH_PORT", ""),
		Daemon:        getEnvBoolOrDefault("DAEMON", true),
		PidFile:       getEnvOrDefault("PID_FILE", "/tmp/gottyp.pid"),
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Session == "" {
		c.Session = generateDefaultSession()
	}
	if c.Remote == "" {
		return fmt.Errorf("remote server address is required")
	}
	if c.Auth {
		if c.AuthName == "" {
			c.AuthName = generateRandomString(8)
		}
		if c.Pass == "" {
			c.Pass = generateRandomString(10)
		}
	}
	return nil
}

// GetRemoteHost 获取远程主机地址（用于显示 URL）
func (c *Config) GetRemoteHost() string {
	host := c.Remote
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	host = strings.Split(host, "/")[0]
	host = strings.Split(host, ":")[0]
	if host == "" {
		host = "clauded.friddle.me"
	}
	return host
}

// GetRemotePort 获取远程端口
func (c *Config) GetRemotePort() int {
	// 解析 remote 参数，格式: host:port
	parts := strings.Split(c.Remote, ":")
	if len(parts) >= 2 {
		if port, err := strconv.Atoi(parts[1]); err == nil {
			return port
		}
	}
	return 8088
}

// FindAvailablePort 查找可用端口，从8080开始
func (c *Config) FindAvailablePort() int {
	startPort := 8080
	for port := startPort; port < startPort+100; port++ {
		if isPortAvailable(port) {
			return port
		}
	}
	return startPort // 如果都不可用，返回默认端口
}

// isPortAvailable 检查端口是否可用
func isPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// getEnvOrDefault 获取环境变量或默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault 获取整数环境变量或默认值
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBoolOrDefault 获取布尔环境变量或默认值
func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func generateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)[:length]
}

func generateDefaultSession() string {
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	if user == "" {
		user = "unknown"
	}
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "app"
	}
	dir := filepath.Base(cwd)
	rand := generateRandomString(4)
	session := fmt.Sprintf("%s-%s-%s", user, dir, rand)
	if runtime.GOOS == "windows" {
		session = strings.ToLower(session)
	}
	return session
}
