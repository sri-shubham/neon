package neon

import (
	"net/http"
	"reflect"
	"strings"
)

// endpoint : Endpoint
type endpoint struct {
	method  string
	url     string
	handler func(w http.ResponseWriter, r *http.Request)
}

// checkAPIMethodExists : Checks if api field formt is correct and methods exist
// Field name should beign with Lower Caps; corresponding handler should have same name
// but begin with Upper caps
func checkAPIMethodExists(sv reflect.Value, st reflect.Type, ft reflect.StructField) (*func(w http.ResponseWriter, r *http.Request), bool) {
	if string(ft.Name[0]) == strings.ToUpper(string(ft.Name[0])) {
		return nil, false
	}
	handlerName := strings.ToUpper(string(ft.Name[0])) + ft.Name[1:]
	method, ok := st.MethodByName(handlerName)
	if !ok {
		return nil, false
	}

	if method.Type.NumIn() != 3 || method.Type.NumOut() != 0 {
		return nil, false
	}
	if method.Type.In(1) != reflect.TypeOf((*http.ResponseWriter)(nil)).Elem() {
		return nil, false
	}
	if method.Type.In(2) != reflect.TypeOf(&http.Request{}) {
		return nil, false
	}

	handlerMethod := sv.MethodByName(handlerName)

	// Zero-allocation static closure!
	// In Go, extracting `.Interface()` from a bound reflection method returns a perfectly typed closure.
	if typedHandler, ok := handlerMethod.Interface().(func(http.ResponseWriter, *http.Request)); ok {
		return &typedHandler, true
	}

	// Fallback (should never be reached because we verified the signature above)
	handler := func(w http.ResponseWriter, r *http.Request) {
		panic("FALLBACK")
		rValue := reflect.ValueOf(r)
		wValue := reflect.ValueOf(w)
		handlerMethod.Call([]reflect.Value{wValue, rValue})
	}
	return &handler, true
}
