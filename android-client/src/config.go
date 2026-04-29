package src

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// Config represents the configuration for the gotty-piko client
type Config struct {
	Name       string // Piko client name
	Remote     string // Remote piko server address (format: host:port)
	ServerPort int    // Piko server port
	GottyPort  int    // Local gotty port (auto-allocated)
	Terminal   string // Terminal type (zsh, bash, sh, powershell, etc.)
	Pass       string // Remote piko server password
	AutoExit   bool   // Enable 24-hour auto-exit (default: true)
}

// NewConfig creates a new configuration instance with environment variables
func NewConfig() *Config {
	return &Config{
		Name:       getEnvOrDefault("NAME", ""),
		Remote:     getEnvOrDefault("REMOTE", ""),
		ServerPort: getEnvIntOrDefault("SERVER_PORT", 8022),
		GottyPort:  0,                                      // Will be auto-allocated on startup
		Terminal:   getEnvOrDefault("TERMINAL", ""),        // Read terminal type from environment
		AutoExit:   getEnvBoolOrDefault("AUTO_EXIT", true), // Read auto-exit setting from environment, default true
		Pass:       getEnvOrDefault("PASS", ""),
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("client name cannot be empty")
	}
	if c.Remote == "" {
		return fmt.Errorf("remote server address cannot be empty")
	}
	return nil
}

// GetRemoteHost extracts the remote host from the remote address
func (c *Config) GetRemoteHost() string {
	// Parse remote parameter, format: host:port
	parts := strings.Split(c.Remote, ":")
	if len(parts) >= 1 {
		return parts[0]
	}
	return "localhost"
}

// GetRemotePort extracts the remote port from the remote address
func (c *Config) GetRemotePort() int {
	// Parse remote parameter, format: host:port
	parts := strings.Split(c.Remote, ":")
	if len(parts) >= 2 {
		if port, err := strconv.Atoi(parts[1]); err == nil {
			return port
		}
	}
	return 8088
}

// FindAvailablePort finds an available port starting from 8080
func (c *Config) FindAvailablePort() int {
	startPort := 8080
	for port := startPort; port < startPort+100; port++ {
		if isPortAvailable(port) {
			return port
		}
	}
	return startPort // If none available, return default port
}

// isPortAvailable checks if a port is available
func isPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// getEnvOrDefault gets environment variable or default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault gets integer environment variable or default value
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBoolOrDefault gets boolean environment variable or default value
func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
