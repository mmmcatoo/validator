package rules

import (
	"fmt"
	"reflect"
	"strings"
)

// Required 必选验证器
func Required(values reflect.Value, object interface{}, params, name, chains, message string) (bool, bool, string) {
	if len(message) == 0 {
		message = fmt.Sprintf("%s 不能为空", chains)
	} else {
		message = strings.Replace(message, "{field}", chains, -1)
	}
	if values.Kind() == reflect.Invalid {
		return false, false, message
	}

	if values.Kind() == reflect.Ptr && values.IsNil() {
		return false, false, message
	}

	if values.IsZero() {
		return false, false, message
	}

	return true, true, ""
}

// Optional 可选验证器
func Optional(values reflect.Value, object interface{}, params, name, chains, message string) (bool, bool, string) {
	res, _, _ := Required(values, object, params, name, chains, "")
	if res {
		return true, true, ""
	}
	return true, false, ""
}
