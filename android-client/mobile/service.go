package mobile

import (
	"context"
	"time"

	"gotty-piko-client/src"
)

// ServiceWrapper wraps the Go service for mobile use
type ServiceWrapper struct {
	serviceManager *src.ServiceManager
	ctx            context.Context
	cancel         context.CancelFunc
}

// NewServiceWrapper creates a new service wrapper
func NewServiceWrapper() *ServiceWrapper {
	return &ServiceWrapper{}
}

// Start starts the gotty-piko service
func (sw *ServiceWrapper) Start(remoteAddr, name string) error {
	config := &src.Config{
		Remote:   remoteAddr,
		Name:     name,
		AutoExit: false,
	}

	sw.ctx, sw.cancel = context.WithCancel(context.Background())
	sw.serviceManager = src.NewServiceManager(config)

	return sw.serviceManager.Start()
}

// Stop stops the gotty-piko service
func (sw *ServiceWrapper) Stop() error {
	if sw.cancel != nil {
		sw.cancel()
	}

	// Wait a bit for graceful shutdown
	time.Sleep(1 * time.Second)

	return nil
}

// GetStatus returns the service status
func (sw *ServiceWrapper) GetStatus() string {
	if sw.serviceManager != nil {
		return "Running"
	}
	return "Stopped"
}

// IsRunning returns whether the service is running
func (sw *ServiceWrapper) IsRunning() bool {
	return sw.serviceManager != nil
}
