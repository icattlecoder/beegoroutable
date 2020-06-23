package beegoroutable

import (
	"reflect"
	"runtime"
	"strings"
)

func getFuncName(f interface{}) string {
	name := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	i := strings.LastIndex(name, ".")
	name = name[i+1:]
	i = strings.LastIndex(name, "-")
	name = name[:i]
	return name
}

type HttpMethod func() string

func do(method string, f interface{}) HttpMethod {
	return func() string {
		return method + ":" + getFuncName(f)
	}
}

func GET(f interface{}) HttpMethod {
	return do("get", f)
}

func POST(f interface{}) HttpMethod {
	return do("post", f)
}

func PUT(f interface{}) HttpMethod {
	return do("put", f)
}

func DELETE(f interface{}) HttpMethod {
	return do("delete", f)
}

func HEAD(f interface{}) HttpMethod {
	return do("head", f)
}

func OPTIONS(f interface{}) HttpMethod {
	return do("options", f)
}

func TRACE(f interface{}) HttpMethod {
	return do("trace", f)
}

func PATCH(f interface{}) HttpMethod {
	return do("patch", f)
}

func MappingMethods(methods ...HttpMethod) string {
	mappingMethods := make([]string, 0, len(methods))
	for _, m := range methods {
		mappingMethods = append(mappingMethods, m())
	}
	f := strings.Join(mappingMethods, ";")
	return f
}
