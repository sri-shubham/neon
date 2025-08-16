package neon

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestCheckAPIMethodExists(t *testing.T) {
	service := &TestEndpointService{}
	serviceValue := reflect.ValueOf(service)
	serviceType := reflect.TypeOf(service).Elem()

	// Get the field info for getTest
	field, _ := serviceType.FieldByName("getTest")

	handler, exists := checkAPIMethodExists(serviceValue, serviceType, field)

	if !exists {
		t.Fatal("Expected handler to exist for GetTest method")
	}

	if handler == nil {
		t.Fatal("Expected handler to be non-nil")
	}

	// Test the handler
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	(*handler)(w, req)

	if w.Body.String() != "test endpoint response" {
		t.Errorf("Expected 'test endpoint response', got '%s'", w.Body.String())
	}
}

func TestCheckAPIMethodExists_NonExistent(t *testing.T) {
	service := &TestEndpointService{}
	serviceValue := reflect.ValueOf(service)
	serviceType := reflect.TypeOf(service).Elem()

	// Get the field info for nonExistent (should not have a corresponding handler)
	field, _ := serviceType.FieldByName("nonExistent")

	handler, exists := checkAPIMethodExists(serviceValue, serviceType, field)

	if exists {
		t.Error("Expected handler to not exist for non-existent method")
	}

	if handler != nil {
		t.Error("Expected handler to be nil for non-existent method")
	}
}

func TestCheckAPIMethodExists_LowercaseField(t *testing.T) {
	service := &TestEndpointService{}
	serviceValue := reflect.ValueOf(service)
	serviceType := reflect.TypeOf(service).Elem()

	// Get the field info for uppercaseField (should fail because field starts with uppercase)
	field, _ := serviceType.FieldByName("UppercaseField")

	handler, exists := checkAPIMethodExists(serviceValue, serviceType, field)

	if exists {
		t.Error("Expected handler to not exist for uppercase field")
	}

	if handler != nil {
		t.Error("Expected handler to be nil for uppercase field")
	}
}

// Test service for endpoint testing
type TestEndpointService struct {
	Module         `base:"/test" v:"1"`
	getTest        Get  `url:"/endpoint"`
	nonExistent    Post `url:"/missing"`   // No corresponding handler method
	UppercaseField Get  `url:"/uppercase"` // Should fail - starts with uppercase
}

func (s TestEndpointService) GetTest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("test endpoint response"))
}

// Note: No handler for nonExistent, and no handler for UppercaseField
