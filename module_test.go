package neon

import (
	"net/http"
	"testing"
)

func TestModulePlaceholder(t *testing.T) {
	module := Module{}

	// Test that placeholder method exists (it returns nothing)
	module.placeholder() // Should not panic
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
