package engine

import (
	"fmt"
	"path"
	"strings"
	"unicode"

	bladeRuntime "github.com/zane-xin/sparq-blade/internal/runtime"
)

var builtinFunctionNames = map[string]struct{}{
	"asset": {}, "coalesce": {}, "contains": {}, "default": {},
	"formatDate": {}, "len": {}, "lower": {}, "upper": {},
}

// WithFunction 为表达式求值器注册一个受控函数。
func WithFunction(name string, function bladeRuntime.Function) Option {
	return func(engine *Engine) {
		if err := ValidateFunction(name, function); err != nil {
			panic(err)
		}
		if _, exists := engine.functions[name]; exists {
			panic(fmt.Sprintf("模板函数重复注册: %s", name))
		}
		engine.functions[name] = function
	}
}

// WithAssetPrefix 设置 asset 函数生成静态资源路径时使用的路由前缀。
func WithAssetPrefix(prefix string) Option {
	return func(engine *Engine) {
		if prefix == "" || !strings.HasPrefix(prefix, "/") {
			panic("asset 静态资源前缀必须是以 / 开头的非空路径")
		}
		engine.assetPrefix = path.Clean(prefix)
	}
}

func (e *Engine) installBuiltinFunctions() {
	for name, function := range bladeRuntime.BuiltinFunctions(e.assetPrefix) {
		e.functions[name] = function
	}
}

func (e *Engine) newScope(data any) *bladeRuntime.Scope {
	return bladeRuntime.NewScope(data).WithFunctions(e.functions)
}

// ValidateFunction 校验表达式函数的名称和处理函数。
func ValidateFunction(name string, function bladeRuntime.Function) error {
	if function == nil {
		return fmt.Errorf("模板函数 %s 的处理函数不能为空", name)
	}
	if !isFunctionName(name) {
		return fmt.Errorf("模板函数名称非法: %s", name)
	}
	if _, builtin := builtinFunctionNames[name]; builtin {
		return fmt.Errorf("内建模板函数不允许覆盖: %s", name)
	}
	return nil
}

func isFunctionName(name string) bool {
	if name == "" {
		return false
	}
	segments := strings.Split(name, ".")
	for _, segment := range segments {
		if segment == "" {
			return false
		}
		for index, current := range segment {
			if index == 0 {
				if current != '_' && !unicode.IsLetter(current) {
					return false
				}
				continue
			}
			if current != '_' && current != '-' && !unicode.IsLetter(current) && !unicode.IsDigit(current) {
				return false
			}
		}
	}
	return true
}
