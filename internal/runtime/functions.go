package runtime

import (
	"fmt"
	"net/url"
	"path"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"
)

// BuiltinFunctions 返回当前引擎使用的内建函数集合。
func BuiltinFunctions(assetPrefix string) map[string]Function {
	return map[string]Function{
		"asset":      assetFunction(assetPrefix),
		"coalesce":   coalesceFunction,
		"contains":   containsFunction,
		"default":    defaultFunction,
		"formatDate": formatDateFunction,
		"len":        lengthFunction,
		"lower":      lowerFunction,
		"upper":      upperFunction,
	}
}

func lengthFunction(args ...any) (any, error) {
	if err := requireArgCount("len", args, 1); err != nil {
		return nil, err
	}
	if args[0] == nil {
		return int64(0), nil
	}
	if value, ok := args[0].(string); ok {
		return int64(utf8.RuneCountInString(value)), nil
	}
	value := indirectValue(args[0])
	if !value.IsValid() {
		return int64(0), nil
	}
	switch value.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice:
		return int64(value.Len()), nil
	default:
		return nil, fmt.Errorf("len 不支持 %T", args[0])
	}
}

func upperFunction(args ...any) (any, error) {
	value, err := stringArg("upper", args)
	if err != nil {
		return nil, err
	}
	return strings.ToUpper(value), nil
}

func lowerFunction(args ...any) (any, error) {
	value, err := stringArg("lower", args)
	if err != nil {
		return nil, err
	}
	return strings.ToLower(value), nil
}

func containsFunction(args ...any) (any, error) {
	if err := requireArgCount("contains", args, 2); err != nil {
		return nil, err
	}
	if value, ok := args[0].(string); ok {
		target, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("contains 的字符串目标必须是 string，实际为 %T", args[1])
		}
		return strings.Contains(value, target), nil
	}
	value := indirectValue(args[0])
	if !value.IsValid() {
		return false, nil
	}
	switch value.Kind() {
	case reflect.Array, reflect.Slice:
		for index := 0; index < value.Len(); index++ {
			if equalValues(value.Index(index).Interface(), args[1]) {
				return true, nil
			}
		}
		return false, nil
	case reflect.Map:
		iter := value.MapRange()
		for iter.Next() {
			if equalValues(iter.Key().Interface(), args[1]) {
				return true, nil
			}
		}
		return false, nil
	default:
		return nil, fmt.Errorf("contains 不支持 %T", args[0])
	}
}

func coalesceFunction(args ...any) (any, error) {
	for _, value := range args {
		if value != nil {
			return value, nil
		}
	}
	return nil, nil
}

func defaultFunction(args ...any) (any, error) {
	if err := requireArgCount("default", args, 2); err != nil {
		return nil, err
	}
	if Truthy(args[0]) {
		return args[0], nil
	}
	return args[1], nil
}

func formatDateFunction(args ...any) (any, error) {
	if err := requireArgCount("formatDate", args, 2); err != nil {
		return nil, err
	}
	layout, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("formatDate 的格式参数必须是 string，实际为 %T", args[1])
	}
	if args[0] == nil {
		return "", nil
	}
	var value time.Time
	switch date := args[0].(type) {
	case time.Time:
		value = date
	case *time.Time:
		if date == nil {
			return "", nil
		}
		value = *date
	default:
		return nil, fmt.Errorf("formatDate 仅支持 time.Time 或 *time.Time，实际为 %T", args[0])
	}
	return value.Format(layout), nil
}

func assetFunction(prefix string) Function {
	return func(args ...any) (any, error) {
		value, err := stringArg("asset", args)
		if err != nil {
			return nil, err
		}
		parsed, err := url.Parse(value)
		if err != nil {
			return nil, fmt.Errorf("asset 路径非法: %w", err)
		}
		if parsed.IsAbs() || parsed.Host != "" {
			return nil, fmt.Errorf("asset 仅支持站内相对路径: %s", value)
		}
		for _, segment := range strings.Split(parsed.Path, "/") {
			if segment == ".." {
				return nil, fmt.Errorf("asset 路径不能包含上级目录: %s", value)
			}
		}
		parsed.Path = path.Join(prefix, strings.TrimPrefix(parsed.Path, "/"))
		if parsed.Path == "." {
			parsed.Path = "/"
		}
		return parsed.String(), nil
	}
}

func stringArg(name string, args []any) (string, error) {
	if err := requireArgCount(name, args, 1); err != nil {
		return "", err
	}
	value, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("%s 的参数必须是 string，实际为 %T", name, args[0])
	}
	return value, nil
}

func requireArgCount(name string, args []any, count int) error {
	if len(args) != count {
		return fmt.Errorf("%s 需要 %d 个参数，实际收到 %d 个", name, count, len(args))
	}
	return nil
}

func indirectValue(value any) reflect.Value {
	reflected := reflect.ValueOf(value)
	for reflected.IsValid() && (reflected.Kind() == reflect.Pointer || reflected.Kind() == reflect.Interface) {
		if reflected.IsNil() {
			return reflect.Value{}
		}
		reflected = reflected.Elem()
	}
	return reflected
}
