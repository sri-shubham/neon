package neon

import "context"

// Moduler : Creates interface for input modules
type Moduler interface {
	Init() error
	Close(context.Context) error
	HealthCheck() error
}

// Module : Module is default implementation for Moduler interface
// modules can embed this structure so can create new module without implementing Moduler
type Module struct {
}

// Default implementations for lifecycle methods
func (m Module) Init() error                     { return nil }
func (m Module) Close(ctx context.Context) error { return nil }
func (m Module) HealthCheck() error              { return nil }
