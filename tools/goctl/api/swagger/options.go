package swagger

import (
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	"github.com/zeromicro/go-zero/tools/goctl/util"
)

func rangeValueFromOptions(options []string) (minimum *float64, maximum *float64, exclusiveMinimum bool, exclusiveMaximum bool) {
	if len(options) == 0 {
		return nil, nil, false, false
	}
	for _, option := range options {
		if strings.HasPrefix(option, rangeFlag) {
			val := option[6:]
			start, end := val[0], val[len(val)-1]
			if start != '[' && start != '(' {
				return nil, nil, false, false
			}
			if end != ']' && end != ')' {
				return nil, nil, false, false
			}
			exclusiveMinimum = start == '('
			exclusiveMaximum = end == ')'

			content := val[1 : len(val)-1]
			idxColon := strings.Index(content, ":")
			if idxColon < 0 {
				return nil, nil, false, false
			}
			var (
				minStr, maxStr string
				minVal, maxVal *float64
			)
			minStr = util.TrimWhiteSpace(content[:idxColon])
			if len(val) >= idxColon+1 {
				maxStr = util.TrimWhiteSpace(content[idxColon+1:])
			}

			if len(minStr) > 0 {
				min, err := strconv.ParseFloat(minStr, 64)
				if err != nil {
					return nil, nil, false, false
				}
				minVal = &min
			}

			if len(maxStr) > 0 {
				max, err := strconv.ParseFloat(maxStr, 64)
				if err != nil {
					return nil, nil, false, false
				}
				maxVal = &max
			}

			return minVal, maxVal, exclusiveMinimum, exclusiveMaximum
		}
	}
	return nil, nil, false, false
}

func enumsValueFromOptions(options []string) []any {
	if len(options) == 0 {
		return []any{}
	}
	for _, option := range options {
		if strings.HasPrefix(option, enumFlag) {
			var resp = make([]any, 0)
			val := option[8:]
			fields := util.FieldsAndTrimSpace(val, func(r rune) bool {
				return r == '|'
			})
			for _, field := range fields {
				resp = append(resp, field)
			}
			return resp
		}
	}
	return []any{}
}

func defValueFromOptions(ctx Context, options []string, apiType spec.Type) any {
	tp := sampleTypeFromGoType(ctx, apiType)
	return valueFromOptions(ctx, options, defFlag, tp)
}

func exampleValueFromOptions(ctx Context, options []string, apiType spec.Type) any {
	tp := sampleTypeFromGoType(ctx, apiType)
	val := valueFromOptions(ctx, options, exampleFlag, tp)
	if val != nil {
		return val
	}
	return defValueFromOptions(ctx, options, apiType)
}

func valueFromOptions(_ Context, options []string, key string, tp string) any {
	if len(options) == 0 {
		return nil
	}
	for _, option := range options {
		if strings.HasPrefix(option, key) {
			s := option[len(key):]
			switch tp {
			case swaggerTypeInteger:
				val, _ := strconv.ParseInt(s, 10, 64)
				return val
			case swaggerTypeBoolean:
				val, _ := strconv.ParseBool(s)
				return val
			case swaggerTypeNumber:
				val, _ := strconv.ParseFloat(s, 64)
				return val
			case swaggerTypeArray:
				// 修正：支持example传入对应的数组类型
				// 尝试解析为数组案例：example=1|2，同时如果只有单个元素将被自动包装为数组
				// 处理数组类型的example
				return parseArrayExample(s)
				// if strings.Contains(s,"|") {
				// 	parts := strings.Split(s, "|")
				// 	var result []string
				// 	for _, part := range parts {
				// 		trimmed := strings.TrimSpace(part)
				// 		if len(trimmed) > 0 {
				// 			result = append(result, trimmed)
				// 		}
				// 	}
				// 	return result
				// }
				// // 如果没有用到管道错误就直接返回对应字符串
				// return s
			case swaggerTypeString:
				return s
			default:
				return nil
			}
		}
	}
	return nil
}

func parseArrayExample(s string) any {
	s = strings.TrimSpace(s)

	// 情况1：使用管道符分隔的多个元素
	if strings.Contains(s, "|") {
		parts := strings.Split(s, "|")
		var result []any
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if len(trimmed) > 0 {
				result = append(result, parseValue(trimmed))
			}
		}
		return result
	}

	// 情况2：单个值，直接包装为数组
	return []any{parseValue(s)}
}

func parseValue(value string) any {
	value = strings.TrimSpace(value)

	// 尝试解析为整数
	if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
		return intVal
	}

	// 尝试解析为浮点数
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}

	// 尝试解析为布尔值
	if boolVal, err := strconv.ParseBool(value); err == nil {
		return boolVal
	}

	// 默认作为字符串
	return value
}
