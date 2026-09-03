// Package binding 提供了纯 Go 标准库实现的请求参数绑定与基于 Struct Tag 的轻量级数据验证器。
package binding

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// emailRegex 用于匹配标准电子邮箱格式的正则表达式。
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// numericRegex 用于匹配纯数字格式的正则表达式。
var numericRegex = regexp.MustCompile(`^[0-9]+$`)

// ValidationError 包含了所有字段校验失败的具体细节。
type ValidationError struct {
	Errors []FieldError
}

func (ve *ValidationError) Error() string {
	errMsgs := make([]string, len(ve.Errors))
	for i, err := range ve.Errors {
		errMsgs[i] = err.Message
	}
	return strings.Join(errMsgs, "; ")
}

// FieldError 描述单个字段验证失败的错误信息。
type FieldError struct {
	Field   string // 字段名称
	Rule    string // 触发失败的规则 (如 "required", "min", "email")
	Param   string // 规则参数 (如 min=6 中的 "6")
	Message string // 友好的错误描述
}

// Validate 扫描目标结构体中带有 `binding:"..."` 标签的字段，并执行规则校验。
// obj 必须是指向结构体的非空指针。
func Validate(obj any) error {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	valType := val.Type()
	fieldErrors := make([]FieldError, 0)

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := valType.Field(i)

		tag := fieldType.Tag.Get("binding")
		if tag == "" || tag == "-" {
			continue
		}

		rules := strings.Split(tag, ",")
		fieldName := fieldType.Tag.Get("json")
		if fieldName == "" || fieldName == "-" {
			fieldName = fieldType.Tag.Get("form")
		}
		if fieldName == "" || fieldName == "-" {
			fieldName = fieldType.Name
		}

		for _, rule := range rules {
			rule = strings.TrimSpace(rule)
			if rule == "" {
				continue
			}

			ruleName, ruleParam := parseRule(rule)
			if err := validateField(fieldName, field, ruleName, ruleParam); err != nil {
				fieldErrors = append(fieldErrors, *err)
			}
		}
	}

	if len(fieldErrors) > 0 {
		return &ValidationError{Errors: fieldErrors}
	}
	return nil
}

func parseRule(rule string) (string, string) {
	parts := strings.SplitN(rule, "=", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}

func validateField(name string, field reflect.Value, rule, param string) *FieldError {
	switch rule {
	case "required":
		if isZeroValue(field) {
			return &FieldError{
				Field:   name,
				Rule:    rule,
				Param:   param,
				Message: fmt.Sprintf("字段 [%s] 是必填项，不能为空", name),
			}
		}

	case "min":
		minVal, _ := strconv.ParseFloat(param, 64)
		switch field.Kind() {
		case reflect.String:
			if float64(len([]rune(field.String()))) < minVal {
				return &FieldError{
					Field:   name,
					Rule:    rule,
					Param:   param,
					Message: fmt.Sprintf("字段 [%s] 长度不能少于 %s 个字符", name, param),
				}
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if float64(field.Int()) < minVal {
				return &FieldError{
					Field:   name,
					Rule:    rule,
					Param:   param,
					Message: fmt.Sprintf("字段 [%s] 不能小于数值 %s", name, param),
				}
			}
		case reflect.Float32, reflect.Float64:
			if field.Float() < minVal {
				return &FieldError{
					Field:   name,
					Rule:    rule,
					Param:   param,
					Message: fmt.Sprintf("字段 [%s] 不能小于数值 %s", name, param),
				}
			}
		case reflect.Slice, reflect.Map:
			if float64(field.Len()) < minVal {
				return &FieldError{
					Field:   name,
					Rule:    rule,
					Param:   param,
					Message: fmt.Sprintf("字段 [%s] 元素数量不能少于 %s 个", name, param),
				}
			}
		}

	case "max":
		maxVal, _ := strconv.ParseFloat(param, 64)
		switch field.Kind() {
		case reflect.String:
			if float64(len([]rune(field.String()))) > maxVal {
				return &FieldError{
					Field:   name,
					Rule:    rule,
					Param:   param,
					Message: fmt.Sprintf("字段 [%s] 长度不能超过 %s 个字符", name, param),
				}
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if float64(field.Int()) > maxVal {
				return &FieldError{
					Field:   name,
					Rule:    rule,
					Param:   param,
					Message: fmt.Sprintf("字段 [%s] 不能大于数值 %s", name, param),
				}
			}
		case reflect.Float32, reflect.Float64:
			if field.Float() > maxVal {
				return &FieldError{
					Field:   name,
					Rule:    rule,
					Param:   param,
					Message: fmt.Sprintf("字段 [%s] 不能大于数值 %s", name, param),
				}
			}
		case reflect.Slice, reflect.Map:
			if float64(field.Len()) > maxVal {
				return &FieldError{
					Field:   name,
					Rule:    rule,
					Param:   param,
					Message: fmt.Sprintf("字段 [%s] 元素数量不能超过 %s 个", name, param),
				}
			}
		}

	case "email":
		if field.Kind() == reflect.String {
			str := field.String()
			if str != "" && !emailRegex.MatchString(str) {
				return &FieldError{
					Field:   name,
					Rule:    rule,
					Param:   param,
					Message: fmt.Sprintf("字段 [%s] 必须是有效的电子邮箱格式", name),
				}
			}
		}

	case "numeric":
		if field.Kind() == reflect.String {
			str := field.String()
			if str != "" && !numericRegex.MatchString(str) {
				return &FieldError{
					Field:   name,
					Rule:    rule,
					Param:   param,
					Message: fmt.Sprintf("字段 [%s] 必须为纯数字字符串", name),
				}
			}
		}
	}

	return nil
}

// isZeroValue 检查反射值是否为零值。
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return strings.TrimSpace(v.String()) == ""
	case reflect.Array, reflect.Slice, reflect.Map:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	default:
		return false
	}
}
