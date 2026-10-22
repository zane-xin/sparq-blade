package engine

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"unicode"

	"github.com/zane-xin/sparq-blade/internal/directives"
	bladeRuntime "github.com/zane-xin/sparq-blade/internal/runtime"
)

const businessDirectiveBrandPrefix = "sq"

// TagCall 保存输出型业务标签已经完成表达式求值的输入。
type TagCall struct {
	// Target 是标签的可选业务目标，例如 cms:article；未使用目标语法时为空。
	Target string
	// Params 保存已求值的命名参数，值仍保持原始 Go 类型。
	Params map[string]any
	// Body 是块标签已经完成模板渲染的可信 HTML；单指令调用时为空。
	Body   template.HTML
	scope  *bladeRuntime.Scope
	engine *Engine
}

// TagFunc 返回直接嵌入页面的可信 HTML；处理函数必须主动转义所有外部数据，或通过 TagCall.Render 复用模板转义。
type TagFunc func(call TagCall) (template.HTML, error)

// Scope 返回标签调用所在的渲染作用域，供 Render 等需要保留当前局部变量的扩展逻辑使用。
func (c TagCall) Scope() *bladeRuntime.Scope {
	return c.scope
}

// Context 返回本次渲染传入的根 ViewContext，不包含循环和 include 注入的局部变量。
func (c TagCall) Context() any {
	if c.scope == nil {
		return nil
	}

	return c.scope.Root
}

// Has 判断是否显式传入某个参数。
func (c TagCall) Has(key string) bool {
	_, ok := c.Params[key]
	return ok
}

// Get 返回原始参数值。
func (c TagCall) Get(key string) (any, bool) {
	value, ok := c.Params[key]
	return value, ok
}

// GetString 读取字符串参数。
func (c TagCall) GetString(key string) (string, error) {
	value, ok := c.Params[key]
	if !ok {
		return "", fmt.Errorf("缺少参数: %s", key)
	}

	return toString(value, key)
}

// GetStringDefault 读取字符串参数，并在未传入时返回默认值。
func (c TagCall) GetStringDefault(key string, fallback string) (string, error) {
	if !c.Has(key) {
		return fallback, nil
	}

	return c.GetString(key)
}

// GetInt 读取整型参数。
func (c TagCall) GetInt(key string) (int, error) {
	value, ok := c.Params[key]
	if !ok {
		return 0, fmt.Errorf("缺少参数: %s", key)
	}

	return toInt(value, key)
}

// GetIntDefault 读取整型参数，并在未传入时返回默认值。
func (c TagCall) GetIntDefault(key string, fallback int) (int, error) {
	if !c.Has(key) {
		return fallback, nil
	}

	return c.GetInt(key)
}

// GetBool 读取布尔参数。
func (c TagCall) GetBool(key string) (bool, error) {
	value, ok := c.Params[key]
	if !ok {
		return false, fmt.Errorf("缺少参数: %s", key)
	}

	return toBool(value, key)
}

// GetBoolDefault 读取布尔参数，并在未传入时返回默认值。
func (c TagCall) GetBoolDefault(key string, fallback bool) (bool, error) {
	if !c.Has(key) {
		return fallback, nil
	}

	return c.GetBool(key)
}

// Render 在当前作用域基础上合并 data 并渲染子模板；子模板已完成上下文转义，因此结果可作为可信 HTML 嵌入标签输出。
func (c TagCall) Render(view string, data any) (template.HTML, error) {
	if c.engine == nil {
		return "", fmt.Errorf("当前标签调用未绑定模板引擎")
	}
	if strings.TrimSpace(view) == "" {
		return "", fmt.Errorf("模板名不能为空")
	}

	scope := c.scope
	if scope == nil {
		scope = bladeRuntime.NewScope(nil)
	}

	content, err := c.engine.renderFileWithScope(view, scope.WithData(data))
	if err != nil {
		return "", wrapRenderFileError(view, err)
	}

	return template.HTML(content), nil
}

// DataTagResult 表示数据标签的单值或列表结果；Value 与 Items 同时非 nil 会在渲染阶段返回错误。
type DataTagResult struct {
	// Value 用于 out 单指令导出单值；块模式不接受单值结果。
	Value any
	// Items 用于块模式逐项渲染，也可以通过 out 导出完整列表。
	Items []any
}

// DataTagFunc 根据已求值参数获取模板数据；返回值仍由模板普通输出执行上下文转义。
type DataTagFunc func(call DataTagCall) (DataTagResult, error)

// DataTagCall 保存已经完成表达式求值的参数和当前作用域，供数据查询处理函数读取。
type DataTagCall struct {
	target string
	params map[string]any
	scope  *bladeRuntime.Scope
}

// Target 返回数据标签目标，例如 cms:arc。
func (c DataTagCall) Target() string {
	return c.target
}

// Scope 返回数据标签所在的渲染作用域，包含当前循环项和 include 注入的局部变量。
func (c DataTagCall) Scope() *bladeRuntime.Scope {
	return c.scope
}

// Context 返回本次渲染传入的根 ViewContext；需要标签参数时应使用 Get 系列方法。
func (c DataTagCall) Context() any {
	if c.scope == nil {
		return nil
	}

	return c.scope.Root
}

// Has 判断是否显式传入某个参数。
func (c DataTagCall) Has(key string) bool {
	_, ok := c.params[key]
	return ok
}

// Get 返回原始参数值。
func (c DataTagCall) Get(key string) (any, bool) {
	value, ok := c.params[key]
	return value, ok
}

// Parameters 返回已经完成模板表达式求值的参数副本，供声明式数据指令转发到外部 Provider。
func (c DataTagCall) Parameters() map[string]any {
	parameters := make(map[string]any, len(c.params))
	for name, value := range c.params {
		parameters[name] = value
	}
	return parameters
}

// GetString 读取字符串参数。
func (c DataTagCall) GetString(key string) (string, error) {
	value, ok := c.params[key]
	if !ok {
		return "", fmt.Errorf("缺少参数: %s", key)
	}

	return toString(value, key)
}

// GetStringDefault 读取字符串参数，并在未传入时返回默认值。
func (c DataTagCall) GetStringDefault(key string, fallback string) (string, error) {
	if !c.Has(key) {
		return fallback, nil
	}

	return c.GetString(key)
}

// GetInt 读取整型参数。
func (c DataTagCall) GetInt(key string) (int, error) {
	value, ok := c.params[key]
	if !ok {
		return 0, fmt.Errorf("缺少参数: %s", key)
	}

	return toInt(value, key)
}

// GetIntDefault 读取整型参数，并在未传入时返回默认值。
func (c DataTagCall) GetIntDefault(key string, fallback int) (int, error) {
	if !c.Has(key) {
		return fallback, nil
	}

	return c.GetInt(key)
}

// GetBool 读取布尔参数。
func (c DataTagCall) GetBool(key string) (bool, error) {
	value, ok := c.params[key]
	if !ok {
		return false, fmt.Errorf("缺少参数: %s", key)
	}

	return toBool(value, key)
}

// GetBoolDefault 读取布尔参数，并在未传入时返回默认值。
func (c DataTagCall) GetBoolDefault(key string, fallback bool) (bool, error) {
	if !c.Has(key) {
		return fallback, nil
	}

	return c.GetBool(key)
}

// Alias 返回数据标签块模式下的当前项变量名；优先读取 as，其次兼容 id，均未传入时使用 field。
func (c DataTagCall) Alias() (string, error) {
	key := "id"
	value := "field"
	var err error
	if c.Has("as") {
		key = "as"
		value, err = c.GetString("as")
	} else {
		value, err = c.GetStringDefault("id", "field")
	}
	if err != nil {
		return "", err
	}

	return normalizeDataTagVarName(key, value)
}

// Out 解析单指令导出变量；out 等价于 out='field'，out=false 会作为无效配置返回错误。
func (c DataTagCall) Out() (string, bool, error) {
	if !c.Has("out") {
		return "", false, nil
	}

	value, ok := c.Get("out")
	if ok {
		if flag, ok := value.(bool); ok {
			if !flag {
				return "", false, fmt.Errorf("参数 out=false 无效，请使用 out 或 out='name'")
			}

			name, err := defaultDataTagOutName()
			if err != nil {
				return "", false, err
			}
			return name, true, nil
		}
	}

	text, err := c.GetString("out")
	if err != nil {
		return "", false, err
	}

	name, err := normalizeDataTagVarName("out", text)
	if err != nil {
		return "", false, err
	}

	return name, true, nil
}

// WithTag 注册只通过品牌化名称访问的输出型业务指令，handler 返回值按可信 HTML 输出。
func WithTag(name string, handler TagFunc) Option {
	return func(engine *Engine) {
		directiveName := normalizeBusinessDirectiveName(name)
		if handler == nil {
			panic(fmt.Sprintf("业务标签处理函数不能为空: @%s", directiveName))
		}
		if engine.directives == nil {
			engine.directives = directives.NewRegistry()
		}
		mustAllowDirectiveNames(engine.directives, "业务标签", directiveNamesForBlock([]string{directiveName}, ""))
		if engine.tags == nil {
			engine.tags = make(map[string]TagFunc)
		}

		engine.tags[directiveName] = handler
		engine.directives.Register(directiveName, func(expression string) (string, error) {
			return fmt.Sprintf(`{{ bladeTag . %q %q }}`, directiveName, expression), nil
		})
		engine.directives.RegisterBlock(directiveName, normalizeBusinessDirectiveEndName(name, ""), func(expression string, body string) (string, error) {
			return fmt.Sprintf(`{{ bladeTagBlock . %q %q %q }}`, directiveName, expression, body), nil
		})
		engine.generator = nil
	}
}

func mustAllowDirectiveNames(registry *directives.Registry, kind string, names []string) {
	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		mustValidDirectiveName(name, kind)
		mustNotConflictBuiltinDirective(name, kind)
		if _, ok := seen[name]; ok {
			panic(fmt.Sprintf("%s指令名重复: @%s", kind, name))
		}
		seen[name] = struct{}{}
	}
	for _, name := range names {
		if registry != nil && registry.Exists(name) {
			panic(fmt.Sprintf("%s指令名已被注册: @%s", kind, name))
		}
	}
}

func directiveNamesForBlock(names []string, endName string) []string {
	results := make([]string, 0, len(names)*2)
	for _, name := range names {
		results = append(results, name)
		if endName != "" {
			results = append(results, endName)
			continue
		}
		results = append(results, "end"+name)
	}
	return results
}

func mustValidDirectiveName(name string, kind string) {
	if !isValidDirectiveName(name) {
		panic(fmt.Sprintf("%s指令名非法: @%s", kind, name))
	}
}

func isValidDirectiveName(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" || strings.HasSuffix(name, ":") {
		return false
	}

	allowSegmentStart := true
	for _, item := range name {
		if item == ':' {
			if allowSegmentStart {
				return false
			}
			allowSegmentStart = true
			continue
		}
		if allowSegmentStart {
			if item != '_' && !unicode.IsLetter(item) {
				return false
			}
			allowSegmentStart = false
			continue
		}
		if item != '_' && !unicode.IsLetter(item) && !unicode.IsDigit(item) {
			return false
		}
	}

	return !allowSegmentStart
}

func mustNotConflictBuiltinDirective(name string, kind string) {
	if isBuiltinDirectiveName(name) {
		panic(fmt.Sprintf("%s指令名与内置指令冲突: @%s", kind, name))
	}
}

func isBuiltinDirectiveName(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}

	if _, ok := builtinDirectiveNames[name]; ok {
		return true
	}
	if inner, ok := strings.CutPrefix(name, businessDirectiveBrandPrefix+":"); ok {
		_, ok = builtinDirectiveNames[inner]
		return ok
	}
	if inner, ok := strings.CutPrefix(name, businessDirectiveBrandPrefix+":end"); ok {
		_, ok = builtinDirectiveNames["end"+inner]
		return ok
	}
	return false
}

var builtinDirectiveNames = map[string]struct{}{
	"component":    {},
	"each":         {},
	"else":         {},
	"elseif":       {},
	"empty":        {},
	"extends":      {},
	"for":          {},
	"foreach":      {},
	"if":           {},
	"include":      {},
	"list":         {},
	"section":      {},
	"slot":         {},
	"unless":       {},
	"yield":        {},
	"children":     {},
	"end":          {},
	"endcomponent": {},
	"endfor":       {},
	"endforeach":   {},
	"endif":        {},
	"endlist":      {},
	"endsection":   {},
	"endslot":      {},
	"endunless":    {},
	"endchildren":  {},
}

// WithDataDirective 注册只提供数据、不直接输出 HTML 的业务指令。
// name 使用业务局部名时生成 @sq:name；块模式由 @end 结束且只能返回 Items，out 模式可导出 Value 或 Items。
func WithDataDirective(name string, handler DataTagFunc) Option {
	return func(engine *Engine) {
		directiveName := normalizeBusinessDirectiveName(name)
		if handler == nil {
			panic(fmt.Sprintf("数据标签处理函数不能为空: @%s", directiveName))
		}
		if engine.directives == nil {
			engine.directives = directives.NewRegistry()
		}
		mustAllowDirectiveNames(engine.directives, "数据标签", []string{directiveName})
		if engine.dataTags == nil {
			engine.dataTags = make(map[string]DataTagFunc)
		}

		engine.dataTags[directiveName] = handler
		engine.directives.Register(directiveName, func(expression string) (string, error) {
			return fmt.Sprintf(`{{ bladeDataDirective . %q %q }}`, directiveName, expression), nil
		})
		engine.directives.RegisterBlockWithEmpty(directiveName, "end", func(expression string, body string, emptyBody string) (string, error) {
			return fmt.Sprintf(`{{ bladeDataDirectiveBlock . %q %q %q %q }}`, directiveName, expression, body, emptyBody), nil
		})
		engine.generator = nil
	}
}

func normalizeBusinessDirectiveName(name string) string {
	name = strings.TrimSpace(name)
	if strings.Contains(name, ":") {
		panic(fmt.Sprintf("业务指令注册只接受局部名称: %s", name))
	}
	return businessDirectiveBrandPrefix + ":" + name
}

func normalizeBusinessDirectiveEndName(name string, endName string) string {
	name = strings.TrimSpace(name)
	endName = strings.TrimSpace(endName)
	if endName == "end" {
		return endName
	}
	if endName == "" || endName == "end"+name {
		return businessDirectiveBrandPrefix + ":end" + name
	}
	return normalizeBusinessDirectiveName(endName)
}

func mustValidateBusinessDirectiveRegistry(registry *directives.Registry) {
	seen := make(map[string]struct{})
	for _, definition := range registry.Definitions() {
		names := []string{definition.Name}
		if definition.EndName != "" && definition.EndName != "end" {
			names = append(names, definition.EndName)
		}

		for _, name := range names {
			mustValidDirectiveName(name, "业务")
			mustNotConflictBuiltinDirective(name, "业务")
			if _, ok := seen[name]; ok {
				panic(fmt.Sprintf("业务指令名重复: @%s", name))
			}
			seen[name] = struct{}{}
		}
	}
}

func newDataTagCall(target string, params map[string]any, scope *bladeRuntime.Scope) DataTagCall {
	return DataTagCall{
		target: target,
		params: params,
		scope:  scope,
	}
}

func (r DataTagResult) exportValue() (any, error) {
	if r.Value != nil && r.Items != nil {
		return nil, fmt.Errorf("数据标签结果不能同时设置 Value 和 Items")
	}
	if r.Value != nil {
		return r.Value, nil
	}
	if r.Items != nil {
		return r.Items, nil
	}
	return nil, nil
}

func (r DataTagResult) blockItems() ([]any, error) {
	if r.Value != nil {
		return nil, fmt.Errorf("@sq 块标签只支持列表结果，请返回 DataTagResult{Items: ...}")
	}
	return r.Items, nil
}

func defaultDataTagOutName() (string, error) {
	value, err := normalizeDataTagVarName("out", "field")
	if err != nil {
		return "", fmt.Errorf("默认 out 变量名无效: %w", err)
	}

	return value, nil
}

func normalizeDataTagVarName(key string, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("参数 %s 不能为空", key)
	}
	if strings.HasPrefix(value, "$") {
		return "", fmt.Errorf("参数 %s 不再支持 $ 前缀: %s", key, value)
	}
	if !isDataTagVarName(value) {
		return "", fmt.Errorf("参数 %s 不是合法变量名: %s", key, value)
	}
	return value, nil
}

func isDataTagVarName(name string) bool {
	for index, item := range name {
		if index == 0 {
			if item != '_' && !unicode.IsLetter(item) {
				return false
			}
			continue
		}
		if item != '_' && !unicode.IsLetter(item) && !unicode.IsDigit(item) {
			return false
		}
	}

	return name != ""
}

func toString(value any, key string) (string, error) {
	switch item := value.(type) {
	case string:
		return item, nil
	case []byte:
		return string(item), nil
	case fmt.Stringer:
		return item.String(), nil
	case int:
		return strconv.Itoa(item), nil
	case int8, int16, int32, int64:
		return fmt.Sprint(item), nil
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprint(item), nil
	case float32, float64:
		return fmt.Sprint(item), nil
	case bool:
		return strconv.FormatBool(item), nil
	case nil:
		return "", fmt.Errorf("参数 %s 不能为 nil", key)
	default:
		return "", fmt.Errorf("参数 %s 不能转换为 string: %T", key, value)
	}
}

func toInt(value any, key string) (int, error) {
	switch item := value.(type) {
	case int:
		return item, nil
	case int8:
		return int(item), nil
	case int16:
		return int(item), nil
	case int32:
		return int(item), nil
	case int64:
		return int(item), nil
	case uint:
		return int(item), nil
	case uint8:
		return int(item), nil
	case uint16:
		return int(item), nil
	case uint32:
		return int(item), nil
	case uint64:
		return int(item), nil
	case float32:
		return int(item), nil
	case float64:
		return int(item), nil
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(item))
		if err != nil {
			return 0, fmt.Errorf("参数 %s 不能转换为 int: %v", key, err)
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("参数 %s 不能转换为 int: %T", key, value)
	}
}

func toBool(value any, key string) (bool, error) {
	switch item := value.(type) {
	case bool:
		return item, nil
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(item))
		if err != nil {
			return false, fmt.Errorf("参数 %s 不能转换为 bool: %v", key, err)
		}
		return parsed, nil
	default:
		return false, fmt.Errorf("参数 %s 不能转换为 bool: %T", key, value)
	}
}
