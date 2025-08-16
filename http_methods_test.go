package neon

import (
	"reflect"
	"testing"
)

func TestHTTPMethodTypes(t *testing.T) {
	tests := []struct {
		name     string
		method   interface{}
		expected string
	}{
		{"Get type", (*Get)(nil), "*neon.Get"},
		{"Post type", (*Post)(nil), "*neon.Post"},
		{"Put type", (*Put)(nil), "*neon.Put"},
		{"Patch type", (*Patch)(nil), "*neon.Patch"},
		{"Delete type", (*Delete)(nil), "*neon.Delete"},
		{"Options type", (*Options)(nil), "*neon.Options"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Use reflection to get the type string
			typeStr := reflect.TypeOf(test.method).String()
			if typeStr != test.expected {
				t.Errorf("Expected type %s, got %s", test.expected, typeStr)
			}
		})
	}
}
