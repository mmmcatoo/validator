package rules

import (
	"fmt"
	"net/mail"
	"reflect"
	"strconv"
	"strings"
)

func Email(values reflect.Value, object interface{}, params, name, chains, message string) (bool, bool, string) {
	_, err := mail.ParseAddress(values.String())
	if err != nil {
		// 按照失败
		if len(message) == 0 {
			message = fmt.Sprintf("%s 不是有效的邮件地址", chains)
		} else {
			message = strings.Replace(message, "{field}", chains, -1)
		}
		return false, true, message
	}
	return true, true, ""
}

func Digits(values reflect.Value, object interface{}, params, name, chains, message string) (bool, bool, string) {
	// 先读取字符串类型的数字
	typeValues := values.String()
	// 然后按照小数点切割
	chunk := strings.Split(typeValues, ".")
	// 先验证整数位长度 最小必须为1
	paramsSet := strings.Split(params, ",")
	var (
		integerLength    int64 = 0
		fractionLength   int64 = 0
		typeValuesLength int64 = int64(len(strings.Replace(typeValues, chunk[0], "", 1)))
	)
	integerLength, _ = strconv.ParseInt(paramsSet[0], 10, 64)
	if 1 <= len(chunk[0]) && int64(len(chunk[0])) <= integerLength {
		// 证数验证通过
		if len(paramsSet) >= 2 {
			fractionLength, _ = strconv.ParseInt(paramsSet[1], 10, 64)
			if fractionLength >= typeValuesLength {
				// 验证通过
				return true, true, ""
			}
		}
	}

	if len(message) == 0 {
		message = fmt.Sprintf("%s 需要整数小于等于{integer}个字符并且小数小于等于{fraction}个字符", chains)
	}
	message = strings.Replace(strings.Replace(strings.Replace(message, "{field}", chains, -1), "{integer}", strconv.FormatInt(integerLength, 10), -1), "{fraction}", strconv.FormatInt(fractionLength, 10), -1)

	return false, true, message
}

func Enum(values reflect.Value, object interface{}, params, name, chains, message string) (bool, bool, string) {
	// 读取成字符串
	actualValues := values.String()
	// 解析成数组
	paramsSet := strings.Split(params, ",")
	for _, set := range paramsSet {
		if actualValues == strings.TrimLeft(set, "\r\n\b\t ") {
			// 匹配成功
			return true, true, ""
		}
	}
	// 匹配失败
	if len(message) == 0 {
		message = fmt.Sprintf("%s 需要在如下的字符串中%s", chains, params)
	}
	message = strings.Replace(strings.Replace(message, "{field}", chains, -1), "{enum}", params, -1)

	return false, true, message
}
