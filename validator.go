package validator

import (
	"fmt"
	"reflect"
	"strings"
	"validator/rules"
)

type ValidateError struct {
	// 字段名称
	Field string `json:"field"`
	// 错误信息
	Message string `json:"message"`
}

// Callback 验证回调方法
type Callback func(values reflect.Value, object interface{}, params, name, chains, message string) (bool, bool, string)

// 验证器对象
type validator struct {
	// 注册过的验证器
	registerRules map[string]Callback
	// 当前的场景
	groupName string
	// 规则名称
	ruleTag string
	// 分组名称
	groupTag string
	// 信息名称
	messageTag string
	// 原始数据结构
	rawObject interface{}
	// 需要过滤的内容
	filterVars string
	// 验证结果
	validateResult []*ValidateError
}

// NewValidator 生成验证器实例
func NewValidator() *validator {
	v := &validator{
		registerRules:  make(map[string]Callback, 0),
		validateResult: make([]*ValidateError, 0),
		ruleTag:        "validator",
		messageTag:     "message",
		groupName:      "groups",
	}
	v.registerRules["required"] = rules.Required
	v.registerRules["optional"] = rules.Optional
	v.registerRules["length"] = rules.Length
	v.registerRules["email"] = rules.Email
	v.registerRules["range"] = rules.Range
	v.registerRules["digits"] = rules.Digits
	v.registerRules["enum"] = rules.Enum
	return v
}

// AddValidator 添加自定义验证器实现
func (receiver *validator) AddValidator(tags string, callback Callback) *validator {
	receiver.registerRules[tags] = callback
	return receiver
}

// AddGroup 添加验证场景
func (receiver *validator) AddGroup(group string) *validator {
	receiver.groupName = strings.Trim(group, receiver.filterVars)
	return receiver
}

// SetTagNames 设置验证规则的相关字段
func (receiver *validator) SetTagNames(rule, group, message string) *validator {
	receiver.ruleTag = rule
	receiver.groupTag = group
	receiver.messageTag = message
	return receiver
}

// Check 验证传递的数据
func (receiver *validator) Check(object interface{}) (bool, []*ValidateError) {
	receiver.rawObject = object
	receiver.parseObject(object, "", -1)
	if len(receiver.validateResult) > 0 {
		return false, receiver.validateResult
	}
	return true, nil
}

// 分析数据
func (receiver *validator) parseObject(object interface{}, prefix string, index int) {
	reflectTypes := reflect.TypeOf(object)
	reflectValues := reflect.ValueOf(object)
	// 如果是个Zero对象 直接返回
	if reflectValues.IsZero() {
		return
	}
	if reflectValues.Kind() == reflect.Ptr {
		reflectValues = reflectValues.Elem()
		reflectTypes = reflectTypes.Elem()
	}
	// 读取属性值
	reflectProperties := reflectTypes.NumField()
	for i := 0; i < reflectProperties; i++ {
		// 是否验证器全部通过
		propertyValid := true
		// 读取属性和相应的值
		reflectProperty := reflectTypes.Field(i)
		reflectPropertyValues := reflectValues.Field(i)
		// 读取分组Tags
		if expectGroups, ok := reflectProperty.Tag.Lookup(receiver.groupTag); ok {
			// 检查分组是否匹配
			if !receiver.matchGroup(expectGroups) {
				// 不匹配 继续下一个属性
				continue
			}
		}
		// 读取信息Tags
		messageSet := make(map[string]string, 0)
		if messages, ok := reflectProperty.Tag.Lookup(receiver.messageTag); ok {
			messageSet = receiver.parseMessage(messages)
		}
		// 读取验证Tags
		if callbackRule, ok := reflectProperty.Tag.Lookup(receiver.ruleTag); ok {
			callbackRules := strings.Split(callbackRule, "|")
			// 先生成属性链
			fieldName := strings.TrimLeft(fmt.Sprintf("%s.%s", prefix, reflectProperty.Name), ".")
			if index > -1 {
				fieldName = strings.TrimLeft(fmt.Sprintf("%s.%d.%s", prefix, index, reflectProperty.Name), ".")
			}

			// 分析规则
			for _, cbRule := range callbackRules {
				cbRuleSet := strings.Split(cbRule, ":")
				if len(cbRuleSet) == 0 {
					// 奇怪的数据 跳过
					continue
				}
				cbName := cbRuleSet[0]
				cbParams := ""
				if len(cbRuleSet) >= 2 {
					cbParams = cbRuleSet[1]
				}
				// 判断是否注册了对于的函数
				if cbExecutor, ok := receiver.registerRules[cbName]; ok {

					// 如果是指针需要二次读取
					if reflectPropertyValues.Kind() == reflect.Ptr {
						reflectPropertyValues = reflectPropertyValues.Elem()
					}

					var cbMessage string = ""
					// 读取自定义错误信息
					if msg, ok := messageSet[cbName]; ok {
						cbMessage = msg
					}

					// 执行验证
					cbResult, continueFlag, cbMessage := cbExecutor(reflectPropertyValues, receiver.rawObject, cbParams, reflectProperty.Name, fieldName, cbMessage)
					if !cbResult {

						// 验证失败了
						receiver.validateResult = append(receiver.validateResult, &ValidateError{
							Field:   fieldName,
							Message: cbMessage,
						})
					}

					// 是否影响了其他的验证器
					if !continueFlag {
						// 继续下一个验证器
						continue
					}

					propertyValid = propertyValid && cbResult
				}
			}

			// 如果全部通过 需要判断类型
			if propertyValid {
				if reflectPropertyValues.Kind() == reflect.Struct {
					// 递归查询
					receiver.parseObject(reflectPropertyValues.Interface(), fieldName, -1)
				} else if reflectPropertyValues.Kind() == reflect.Array || reflectPropertyValues.Kind() == reflect.Slice {
					// 递归处理
					for key := 0; key < reflectPropertyValues.Len(); key++ {
						receiver.parseObject(reflectPropertyValues.Index(key).Interface(), fieldName, key)
					}
				}
			}
		}
	}
}

// 检查验证的场景是否存在
func (receiver *validator) matchGroup(groups string) bool {
	groupList := strings.Split(strings.Trim(groups, receiver.filterVars), ",")
	if len(receiver.groupName) == 0 && len(groupList) == 0 {
		return true
	}
	for _, group := range groupList {
		if strings.Trim(group, receiver.filterVars) == receiver.groupName {
			return true
		}
	}

	return false
}

func (receiver *validator) parseMessage(messages string) map[string]string {
	res := make(map[string]string, 0)
	messageSet := strings.Split(messages, "|")
	for _, rule := range messageSet {
		rules := strings.Split(rule, ":")
		if len(rules) >= 2 {
			// 需要大于2个才能读取
			res[rules[0]] = rules[1]
		}
	}

	return res
}
