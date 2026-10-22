package blade

import (
	"github.com/zane-xin/sparq-blade/internal/compiler"
	"github.com/zane-xin/sparq-blade/internal/directives"
	"github.com/zane-xin/sparq-blade/internal/engine"
	"github.com/zane-xin/sparq-blade/internal/runtime"
)

// Engine 是可复用的模板解析与渲染实例。所有 Option 只在 New 阶段生效，创建后应作为只读配置并发复用。
type Engine = engine.Engine

// Option 描述 Engine 的初始化配置，用于组合模板文件系统、扩展能力和静态资源前缀。
type Option = engine.Option

// TemplateFS 定义按模板逻辑名称读取源码的协议；实现必须返回稳定的规范名称和可用于缓存失效的版本。
type TemplateFS = engine.TemplateFS

// TemplateSource 是 TemplateFS 返回的不可变模板源码快照。
type TemplateSource = engine.TemplateSource

// Cache 保存按模板名称和版本解析后的 AST。自定义实现需要支持多个渲染请求并发访问。
type Cache = engine.Cache

// Syntax 定义转义输出、原样输出、注释和 @ 转义使用的分隔符；原样输出仍只允许可信内容。
type Syntax = compiler.Syntax

// OSFS 从受限根目录读取模板，并阻止绝对路径、上级目录和符号链接逃逸。
type OSFS = engine.OSFS

// LayeredFS 按配置顺序组合多个模板来源，适合实现应用模板覆盖插件默认模板。
type LayeredFS = engine.LayeredFS

// DirectiveRegistry 保存业务指令编译规则；注入 Engine 后会复制并统一映射到固定的 sq: 模板入口。
type DirectiveRegistry = directives.Registry

// DirectiveCompileFunc 把单指令表达式编译为 Go 模板片段；返回内容属于可信编译产物。
type DirectiveCompileFunc = directives.CompileFunc

// BlockDirectiveCompileFunc 把块指令表达式和已编译的块内容组合为 Go 模板片段。
type BlockDirectiveCompileFunc = directives.BlockCompileFunc

// TagCall 提供输出型业务标签的参数、当前作用域和子模板渲染能力。
type TagCall = engine.TagCall

// TagFunc 生成可直接嵌入页面的可信 HTML；实现必须自行转义来自请求、数据库或外部服务的数据。
type TagFunc = engine.TagFunc

// DataTagCall 提供数据标签的参数、目标名称和当前渲染上下文，不负责生成 HTML。
type DataTagCall = engine.DataTagCall

// DataTagFunc 为模板提供单值或列表数据，输出仍由普通模板表达式执行上下文转义。
type DataTagFunc = engine.DataTagFunc

// DataTagResult 表示数据标签返回的单值或列表结果；同一次调用不能同时设置两种结果。
type DataTagResult = engine.DataTagResult

// DataFields 是构造数据标签对象结果时使用的字段集合。
type DataFields = engine.DataFields

// Function 是表达式中的纯值转换函数，不应执行文件、网络、数据库访问或返回可信模板类型。
type Function = runtime.Function

// Lex 把模板源码拆分为 token，并在词法错误中附加源码行列和指针上下文。
func Lex(source string) ([]compiler.Token, error) {
	tokens, err := compiler.NewLexer(source).Tokenize()
	if err != nil {
		return nil, compiler.WithSource(err, source)
	}

	return tokens, nil
}

// Parse 把模板源码解析为 AST，适合模板检查、调试和开发工具，不执行模板。
func Parse(source string) (*compiler.Document, error) {
	tokens, err := Lex(source)
	if err != nil {
		return nil, err
	}

	document, err := compiler.NewParser(tokens).Parse()
	if err != nil {
		return nil, compiler.WithSource(err, source)
	}

	return document, nil
}

// New 创建带默认语法、内建函数和有界缓存的 Engine；应用应在启动阶段完成配置后复用返回实例。
func New(options ...Option) *Engine {
	return engine.New(options...)
}

// WithFS 配置 RenderFile 使用的模板来源；未配置时仍可使用 RenderString。
func WithFS(fs TemplateFS) Option {
	return engine.WithFS(fs)
}

// WithCache 替换 AST 缓存；传入 nil 表示关闭这一层缓存，不影响 Engine 内部的有界编译缓存。
func WithCache(cache Cache) Option {
	return engine.WithCache(cache)
}

// WithSyntax 覆盖非空的模板分隔符，未设置字段继续沿用默认 Blade 兼容语法。
func WithSyntax(syntax Syntax) Option {
	return engine.WithSyntax(syntax)
}

// WithDirectives 复制使用局部业务名注册的指令表，并统一映射到固定的 sq: 模板入口。
func WithDirectives(registry *DirectiveRegistry) Option {
	return engine.WithDirectives(registry)
}

// WithDirective 按业务局部名注册单指令编译函数；模板侧固定通过 @sq:name 访问。
func WithDirective(name string, compile DirectiveCompileFunc) Option {
	return engine.WithDirective(name, compile)
}

// WithBlockDirective 按业务局部名注册块指令；模板入口固定为 @sq:name，并可使用内建 @end 结束。
func WithBlockDirective(name string, endName string, compile BlockDirectiveCompileFunc) Option {
	return engine.WithBlockDirective(name, endName, compile)
}

// WithTag 注册输出型业务标签；业务名统一映射为 sq: 品牌入口，处理函数返回值按可信 HTML 嵌入页面。
func WithTag(name string, handler TagFunc) Option {
	return engine.WithTag(name, handler)
}

// WithDataDirective 注册数据业务指令；无 out 时通过 @end 块直接迭代列表，有 out 时把结果写入当前作用域。
func WithDataDirective(name string, handler DataTagFunc) Option {
	return engine.WithDataDirective(name, handler)
}

// WithFunction 注册纯值转换函数；插件函数应使用带点号的命名空间，内建名称不能覆盖，可信模板返回类型会在执行时拒绝。
func WithFunction(name string, function Function) Option {
	return engine.WithFunction(name, function)
}

// ValidateFunction 校验函数名称和处理函数是否满足注册约束，但不会修改 Engine。
func ValidateFunction(name string, function Function) error {
	return engine.ValidateFunction(name, function)
}

// WithAssetPrefix 配置 asset(path) 生成的站内 URL 前缀；该选项不挂载静态目录，也不读取资源文件。
func WithAssetPrefix(prefix string) Option {
	return engine.WithAssetPrefix(prefix)
}

// Fields 创建可链式写入的数据标签字段集合。
func Fields() DataFields {
	return engine.Fields()
}

// Item 把字段集合转换为一条列表数据，不复制底层 map；后续修改 fields 会反映到该数据项。
func Item(fields DataFields) any {
	return engine.Item(fields)
}

// Items 构造数据标签的列表结果，适合元素已经是 map、struct 或 Item 的场景。
func Items(items ...any) DataTagResult {
	return engine.Items(items...)
}

// Rows 在只传入一个 slice 或 array 时展开其元素；其他调用形式把每个参数视为一条数据。
func Rows(rows ...any) DataTagResult {
	return engine.Rows(rows...)
}

// ValueOf 构造数据标签单值结果，值可以是标量或对象。
func ValueOf(value any) DataTagResult {
	return engine.ValueOf(value)
}

// Object 面向 map 或 struct 场景的 ValueOf 语义别名，不额外执行类型校验。
func Object(value any) DataTagResult {
	return engine.Object(value)
}

// EmptyResult 构造不包含单值和列表的空结果。
func EmptyResult() DataTagResult {
	return engine.EmptyResult()
}

// RenderString 使用默认 Engine 渲染一段可信模板源码；需要文件系统、插件函数或自定义指令时应显式创建并复用 Engine。
func RenderString(source string, data any) (string, error) {
	return New().RenderString(source, data)
}
