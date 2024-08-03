package runtime

import (
	"reflect"
	"runtime"
)

// GetFunctionName return fn name
func GetFunctionName(i any) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
