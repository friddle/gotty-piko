package gottyp

import (
	"context"
	"fmt"
	"gotty-piko-client/src"
	"sync"
	"time"
)

// ServiceWrapper 包装ServiceManager，提供gomobile友好的接口
type ServiceWrapper struct {
	manager *src.ServiceManager
	ctx     context.Context
	cancel  context.CancelFunc
	mutex   sync.Mutex
	running bool
}

// NewServiceWrapper 创建新的服务包装器
func NewServiceWrapper() *ServiceWrapper {
	return &ServiceWrapper{
		running: false,
	}
}

// StartService 启动gottyp服务
// 参数: name - 客户端名称, remote - 远程服务器地址, terminal - 终端类型, pass - 密码
// 返回: 错误信息字符串，成功时返回空字符串
func (sw *ServiceWrapper) StartService(name, remote, terminal, pass string) string {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()

	if sw.running {
		return "服务已在运行中"
	}

	// 验证输入参数
	if name == "" {
		return "客户端名称不能为空"
	}
	if remote == "" {
		return "远程服务器地址不能为空"
	}

	// 创建配置
	config := &src.Config{
		Name:       name,
		Remote:     remote,
		ServerPort: 8022, // 默认端口
		Terminal:   terminal,
		AutoExit:   false, // Android模式下不自动退出
		Pass:       pass,
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return fmt.Sprintf("配置验证失败: %v", err)
	}

	// 创建上下文
	sw.ctx, sw.cancel = context.WithCancel(context.Background())

	// 创建服务管理器
	sw.manager = src.NewServiceManager(config)

	// 在goroutine中启动服务
	go func() {
		err := sw.manager.Start()
		if err != nil && err != context.Canceled {
			fmt.Printf("服务启动错误: %v\n", err)
			// 如果启动失败，重置运行状态
			sw.mutex.Lock()
			sw.running = false
			sw.mutex.Unlock()
		}
	}()

	// 等待一小段时间确保服务启动
	time.Sleep(500 * time.Millisecond)

	sw.running = true
	return ""
}

// StopService 停止gottyp服务
// 返回: 错误信息字符串，成功时返回空字符串
func (sw *ServiceWrapper) StopService() string {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()

	if !sw.running {
		return "服务未在运行"
	}

	if sw.cancel != nil {
		sw.cancel()
	}

	// 等待服务停止
	time.Sleep(500 * time.Millisecond)

	sw.running = false
	return ""
}

// IsRunning 检查服务是否正在运行
// 返回: true表示正在运行，false表示未运行
func (sw *ServiceWrapper) IsRunning() bool {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()
	return sw.running
}

// GetStatus 获取服务状态信息
// 返回: 状态信息字符串
func (sw *ServiceWrapper) GetStatus() string {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()

	if !sw.running {
		return "服务未运行"
	}

	if sw.manager == nil {
		return "服务管理器未初始化"
	}

	return "服务正在运行"
}

// GetDetailedStatus 获取详细的服务状态信息
// 返回: 详细状态信息字符串
func (sw *ServiceWrapper) GetDetailedStatus() string {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()

	if !sw.running {
		return "服务未运行"
	}

	if sw.manager == nil {
		return "服务管理器未初始化"
	}

	// 获取配置信息
	config := sw.manager.Config
	if config == nil {
		return "配置信息不可用"
	}

	return fmt.Sprintf("服务正在运行\n客户端名称: %s\n远程服务器: %s\n本地端口: %d\n终端类型: %s\n认证: %s",
		config.Name, config.Remote, config.GottyPort, config.Terminal,
		func() string {
			if config.Pass != "" {
				return "已启用"
			}
			return "未启用"
		}())
}

// GetLocalPort 获取本地监听端口
// 返回: 端口号，如果服务未运行返回0
func (sw *ServiceWrapper) GetLocalPort() int {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()

	if !sw.running || sw.manager == nil || sw.manager.Config == nil {
		return 0
	}

	// 从ServiceManager的配置中获取端口信息
	return sw.manager.Config.GottyPort
}

// GetVersion 获取版本信息
// 返回: 版本字符串
func (sw *ServiceWrapper) GetVersion() string {
	return "1.0.0"
}

// 全局服务实例
var globalService *ServiceWrapper

// GetService 获取全局服务实例
func GetService() *ServiceWrapper {
	if globalService == nil {
		globalService = NewServiceWrapper()
	}
	return globalService
}

// 初始化函数
func init() {
	globalService = NewServiceWrapper()
}

