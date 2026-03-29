package neon

import (
	"context"
	"net/http"
	"testing"
)

func TestModuleLifecycle(t *testing.T) {
	module := Module{}

	if err := module.Init(); err != nil {
		t.Errorf("Expected Init to return nil, got %v", err)
	}

	if err := module.Close(context.Background()); err != nil {
		t.Errorf("Expected Close to return nil, got %v", err)
	}

	if err := module.HealthCheck(); err != nil {
		t.Errorf("Expected HealthCheck to return nil, got %v", err)
	}
}

func TestModuleInterface(t *testing.T) {
	// Test that Module implements Moduler interface
	var _ Moduler = &Module{}

	// Test that services with embedded Module implement Moduler
	service := &TestModuleService{}
	var _ Moduler = service
} // Test service that embeds Module
type TestModuleService struct {
	Module      `base:"/api" v:"2" middleware:"test"`
	getEndpoint Get `url:"/test"`
}

func (s TestModuleService) GetEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("module test"))
}
