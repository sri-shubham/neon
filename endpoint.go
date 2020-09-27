package neon

import (
	"fmt"
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
	_, ok := st.MethodByName(handlerName)
	if !ok {
		return nil, false
	}

	handlerMethod := sv.MethodByName(handlerName)

	handler := func(w http.ResponseWriter, r *http.Request) {
		rValue := reflect.ValueOf(r)
		wValue := reflect.ValueOf(w)
		fmt.Fprintln(w, "adsfgnh g")
		fmt.Println(handlerMethod.Call([]reflect.Value{wValue, rValue}))
	}
	return &handler, true
}
