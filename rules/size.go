package rules

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func Length(values reflect.Value, object interface{}, params, name, chains, message string) (bool, bool, string) {
	var targetLength int = 0
	if values.Kind() == reflect.String {
		targetLength = len(values.String())
	} else if values.Kind() == reflect.Array || values.Kind() == reflect.Slice {
		targetLength = values.Len()
	}

	valid, message := compare(float64(targetLength), params, 0, 1, message, chains, "长度不满足约定")

	return valid, true, message
}

func Range(values reflect.Value, object interface{}, params, name, chains, message string) (bool, bool, string) {
	var typeValues float64 = 0
	// 按照Float64读取
	if values.Kind() == reflect.String {
		typeValues, _ = strconv.ParseFloat(values.String(), 10)
	} else {
		typeValues = values.Float()
	}

	valid, message := compare(typeValues, params, 1, 0, message, chains, "大小不满足约定")

	return valid, true, message
}

func compare(actualValue float64, rules string, maxPosition, minPosition int, message, chains, template string) (bool, string) {
	paramsSet := strings.Split(rules, ",")
	var (
		maxLength       float64 = 0
		minLength       float64 = 0
		valid           bool    = true
		paramsSetLength int     = len(paramsSet) - 1
	)

	if paramsSetLength >= maxPosition {
		maxLength, _ = strconv.ParseFloat(paramsSet[maxPosition], 10)
		if actualValue > maxLength {
			valid = valid && false
		}
	}

	if paramsSetLength >= minPosition {
		minLength, _ = strconv.ParseFloat(paramsSet[minPosition], 10)
		if minLength > actualValue {
			valid = valid && false
		}
	}

	if len(message) == 0 {
		message = fmt.Sprintf("%s %s", chains, template)
	} else {
		message = strings.Replace(strings.Replace(strings.Replace(message, "{field}", chains, -1), "{max}", strconv.FormatFloat(maxLength, 'f', -1, 64), -1), "{min}", strconv.FormatFloat(minLength, 'f', -1, 64), -1)
	}

	return valid, message
}
