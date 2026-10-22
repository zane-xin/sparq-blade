package engine

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"sync"

	"github.com/zane-xin/sparq-blade/internal/compiler"
	"github.com/zane-xin/sparq-blade/internal/directives"
	bladeRuntime "github.com/zane-xin/sparq-blade/internal/runtime"
)

const (
	maxTemplateCompositionDepth     = 64
	maxCompiledTemplateCacheEntries = 512
	maxFileTemplateCacheEntries     = 512
)

// Option 表示引擎初始化选项。
type Option func(*Engine)

// Engine 表示模板引擎主入口。
type Engine struct {
	fs          TemplateFS
	cache       Cache
	directives  *directives.Registry
	generator   *compiler.Generator
	syntax      compiler.Syntax
	tags        map[string]TagFunc
	dataTags    map[string]DataTagFunc
	functions   map[string]bladeRuntime.Function
	assetPrefix string

	templateMu         sync.RWMutex
	templateCache      map[[32]byte]compiledTemplate
	templateCacheOrder [][32]byte
	fileMu             sync.RWMutex
	fileCache          map[string]compiledFileTemplate
	fileCacheOrder     []string

	funcsOnce sync.Once
	funcs     template.FuncMap
}

type compiledTemplate struct {
	text     string
	template *template.Template
}

type compiledFileTemplate struct {
	version  string
	template *template.Template
}

// New 创建模板引擎实例。
func New(options ...Option) *Engine {
	engine := &Engine{
		cache:         NewMemoryCache(),
		directives:    directives.NewRegistry(),
		syntax:        compiler.DefaultSyntax(),
		tags:          make(map[string]TagFunc),
		dataTags:      make(map[string]DataTagFunc),
		functions:     make(map[string]bladeRuntime.Function),
		assetPrefix:   "/assets",
		templateCache: make(map[[32]byte]compiledTemplate),
		fileCache:     make(map[string]compiledFileTemplate),
	}

	for _, option := range options {
		option(engine)
	}
	if engine.directives == nil {
		panic("自定义指令注册表不能为空")
	}

	engine.installBuiltinFunctions()
	engine.generator = compiler.NewGenerator(compiler.WithDirectiveCompiler(engine.directives))

	return engine
}

// WithFS 为引擎注入模板文件系统。
func WithFS(fs TemplateFS) Option {
	return func(engine *Engine) {
		engine.fs = fs
	}
}

// WithCache 为引擎注入模板缓存实现。
func WithCache(cache Cache) Option {
	return func(engine *Engine) {
		engine.cache = cache
	}
}

// WithDirectives 复制业务指令注册表，并将其中的局部名称统一映射到固定品牌入口。
func WithDirectives(registry *directives.Registry) Option {
	return func(engine *Engine) {
		if registry == nil {
			panic("自定义指令注册表不能为空")
		}
		engine.directives = registry.CloneWithNames(normalizeBusinessDirectiveName, normalizeBusinessDirectiveEndName)
		mustValidateBusinessDirectiveRegistry(engine.directives)
		engine.generator = nil
	}
}

// WithDirective 按业务局部名注册单指令，模板侧固定使用 sq: 品牌入口。
func WithDirective(name string, compile directives.CompileFunc) Option {
	return func(engine *Engine) {
		directiveName := normalizeBusinessDirectiveName(name)
		if engine.directives == nil {
			engine.directives = directives.NewRegistry()
		}
		mustAllowDirectiveNames(engine.directives, "业务", []string{directiveName})
		engine.directives.Register(directiveName, compile)
		engine.generator = nil
	}
}

// WithSyntax 覆盖词法器使用的非空分隔符，未配置字段继续沿用默认 Blade 兼容语法。
func WithSyntax(syntax compiler.Syntax) Option {
	return func(engine *Engine) {
		engine.syntax = syntax
	}
}

// WithBlockDirective 按业务局部名注册块指令，模板侧固定使用 sq: 品牌入口。
// endName 使用局部名称；空值生成品牌关联结束名，所有块同时接受内建 @end。
func WithBlockDirective(name string, endName string, compile directives.BlockCompileFunc) Option {
	return func(engine *Engine) {
		directiveName := normalizeBusinessDirectiveName(name)
		directiveEndName := normalizeBusinessDirectiveEndName(name, endName)
		if engine.directives == nil {
			engine.directives = directives.NewRegistry()
		}
		mustAllowDirectiveNames(engine.directives, "业务块", directiveNamesForBlock([]string{directiveName}, directiveEndName))
		engine.directives.RegisterBlock(directiveName, directiveEndName, compile)
		engine.generator = nil
	}
}

// ParseString 直接解析一段模板源码并返回 AST。
func (e *Engine) ParseString(source string) (*compiler.Document, error) {
	return e.parseString(source, "")
}

func (e *Engine) parseString(source string, sourceName string) (*compiler.Document, error) {
	tokens, err := compiler.NewLexer(source, compiler.WithSyntax(e.syntax)).Tokenize()
	if err != nil {
		return nil, compiler.WithNamedSource(err, sourceName, source)
	}

	document, err := compiler.NewParser(tokens, compiler.WithCustomDirectiveSpecs(e.directiveSpecs()...)).Parse()
	if err != nil {
		return nil, compiler.WithNamedSource(err, sourceName, source)
	}

	return document, nil
}

// ParseFile 从文件系统中读取模板并解析，返回 AST。
func (e *Engine) ParseFile(name string) (*compiler.Document, error) {
	if e.fs == nil {
		return nil, &compiler.Error{
			Code:    compiler.ErrCodeInvalidDirective,
			Message: "模板文件系统未配置",
		}
	}

	_, document, err := e.loadParsedFile(name)
	if err != nil {
		return nil, err
	}

	return compiler.CloneDocument(document), nil
}

func (e *Engine) loadParsedFile(name string) (*TemplateSource, *compiler.Document, error) {
	source, err := e.fs.Load(name)
	if err != nil {
		return nil, nil, err
	}

	if e.cache != nil {
		if document, ok := e.cache.Get(source.Name, source.Version); ok {
			return source, document, nil
		}
	}

	document, err := e.parseString(string(source.Content), source.Name)
	if err != nil {
		return nil, nil, err
	}

	if e.cache != nil {
		e.cache.Set(source.Name, source.Version, document)
	}

	return source, document, nil
}

// RenderString 直接渲染一段模板源码。
func (e *Engine) RenderString(source string, data any) (string, error) {
	document, err := e.ParseString(source)
	if err != nil {
		return "", err
	}

	return e.renderDocument(document, e.newScope(data))
}

// RenderFile 读取模板文件并完成渲染。
func (e *Engine) RenderFile(name string, data any) (string, error) {
	scope := e.newScope(data)
	content, err := e.renderFileWithScope(name, scope)
	if err != nil {
		return "", wrapRenderFileError(name, err)
	}

	return content, nil
}

func (e *Engine) renderFileWithScope(name string, scope *bladeRuntime.Scope) (string, error) {
	document, normalizedName, version, err := e.resolveFileDocument(name)
	if err != nil {
		return "", err
	}

	renderScope, err := scope.EnterTemplate(normalizedName, maxTemplateCompositionDepth)
	if err != nil {
		return "", err
	}
	tmpl, err := e.compiledFileTemplate(normalizedName, version, document)
	if err != nil {
		return "", err
	}

	return e.executeTemplate(tmpl, renderScope)
}

func (e *Engine) renderDocument(document *compiler.Document, scope *bladeRuntime.Scope) (string, error) {
	text, err := e.templateGenerator().Generate(document)
	if err != nil {
		return "", err
	}

	return e.renderTemplateText(text, scope)
}

func (e *Engine) directiveSpecs() []compiler.CustomDirectiveSpec {
	if e.directives == nil {
		return nil
	}

	definitions := e.directives.Definitions()
	specs := make([]compiler.CustomDirectiveSpec, 0, len(definitions))
	for _, definition := range definitions {
		specs = append(specs, compiler.CustomDirectiveSpec{
			Name:         definition.Name,
			AllowSingle:  definition.AllowSingle,
			AllowEmpty:   definition.AllowEmpty,
			DefaultBlock: definition.DefaultBlock,
			EndName:      definition.EndName,
		})
	}

	return specs
}

func (e *Engine) renderTemplateText(text string, scope *bladeRuntime.Scope) (string, error) {
	tmpl, err := e.compiledTemplate(text)
	if err != nil {
		return "", err
	}

	return e.executeTemplate(tmpl, scope)
}

func (e *Engine) executeTemplate(tmpl *template.Template, scope *bladeRuntime.Scope) (string, error) {
	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, scope); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func (e *Engine) compiledTemplate(text string) (*template.Template, error) {
	key := sha256.Sum256([]byte(text))

	e.templateMu.RLock()
	cached, ok := e.templateCache[key]
	e.templateMu.RUnlock()
	if ok && cached.text == text {
		return cached.template, nil
	}

	tmpl, err := template.New("blade").Funcs(e.templateFuncs()).Parse(text)
	if err != nil {
		return nil, err
	}

	e.templateMu.Lock()
	defer e.templateMu.Unlock()
	if cached, ok := e.templateCache[key]; ok && cached.text == text {
		return cached.template, nil
	}
	if _, exists := e.templateCache[key]; !exists {
		if len(e.templateCacheOrder) >= maxCompiledTemplateCacheEntries {
			oldest := e.templateCacheOrder[0]
			delete(e.templateCache, oldest)
			e.templateCacheOrder = e.templateCacheOrder[1:]
		}
		e.templateCacheOrder = append(e.templateCacheOrder, key)
	}
	e.templateCache[key] = compiledTemplate{
		text:     text,
		template: tmpl,
	}

	return tmpl, nil
}

func (e *Engine) compiledFileTemplate(name string, version string, document *compiler.Document) (*template.Template, error) {
	e.fileMu.RLock()
	cached, ok := e.fileCache[name]
	e.fileMu.RUnlock()
	if ok && cached.version == version {
		return cached.template, nil
	}

	text, err := e.templateGenerator().Generate(document)
	if err != nil {
		return nil, err
	}

	tmpl, err := e.compiledTemplate(text)
	if err != nil {
		return nil, err
	}

	e.fileMu.Lock()
	defer e.fileMu.Unlock()
	if cached, ok := e.fileCache[name]; ok && cached.version == version {
		return cached.template, nil
	}
	if _, exists := e.fileCache[name]; !exists {
		if len(e.fileCacheOrder) >= maxFileTemplateCacheEntries {
			oldest := e.fileCacheOrder[0]
			delete(e.fileCache, oldest)
			e.fileCacheOrder = e.fileCacheOrder[1:]
		}
		e.fileCacheOrder = append(e.fileCacheOrder, name)
	}
	e.fileCache[name] = compiledFileTemplate{
		version:  version,
		template: tmpl,
	}

	return tmpl, nil
}

func (e *Engine) templateGenerator() *compiler.Generator {
	if e.generator != nil {
		return e.generator
	}

	e.generator = compiler.NewGenerator(compiler.WithDirectiveCompiler(e.directives))
	return e.generator
}

func (e *Engine) templateFuncs() template.FuncMap {
	e.funcsOnce.Do(func() {
		e.funcs = template.FuncMap{
			"bladeEcho": func(scope *bladeRuntime.Scope, expression string) (any, error) {
				return bladeRuntime.Echo(scope, expression)
			},
			"bladeRaw": func(scope *bladeRuntime.Scope, expression string) (template.HTML, error) {
				return bladeRuntime.Raw(scope, expression)
			},
			"bladeTruthy": func(scope *bladeRuntime.Scope, expression string) (bool, error) {
				return bladeRuntime.Check(scope, expression)
			},
			"bladeUnless": func(scope *bladeRuntime.Scope, expression string) (bool, error) {
				return bladeRuntime.Unless(scope, expression)
			},
			"bladeRange": func(scope *bladeRuntime.Scope, expression string) ([]*bladeRuntime.Scope, error) {
				return bladeRuntime.Range(scope, expression)
			},
			"bladeForRange": func(scope *bladeRuntime.Scope, expression string) ([]*bladeRuntime.Scope, error) {
				return bladeRuntime.ForRange(scope, expression)
			},
			"bladeList": func(scope *bladeRuntime.Scope, expression string) ([]*bladeRuntime.Scope, error) {
				return bladeRuntime.ListRange(scope, expression)
			},
			"bladeChildren": func(scope *bladeRuntime.Scope, expression string) ([]*bladeRuntime.Scope, error) {
				return bladeRuntime.ChildrenRange(scope, expression)
			},
			"bladeInclude": func(scope *bladeRuntime.Scope, expression string) (template.HTML, error) {
				name, includeScope, err := bladeRuntime.IncludeScope(scope, expression)
				if err != nil {
					return "", err
				}

				content, err := e.renderFileWithScope(name, includeScope)
				if err != nil {
					return "", wrapRenderFileError(name, err)
				}

				return template.HTML(content), nil
			},
			"bladeEach": func(scope *bladeRuntime.Scope, expression string) (template.HTML, error) {
				view, scopes, emptyView, err := bladeRuntime.EachRange(scope, expression)
				if err != nil {
					return "", err
				}

				if len(scopes) == 0 {
					if emptyView == "" {
						return "", nil
					}

					content, err := e.renderFileWithScope(emptyView, scope)
					if err != nil {
						return "", wrapRenderFileError(emptyView, err)
					}

					return template.HTML(content), nil
				}

				var builder strings.Builder
				for _, itemScope := range scopes {
					content, err := e.renderFileWithScope(view, itemScope)
					if err != nil {
						return "", wrapRenderFileError(view, err)
					}
					builder.WriteString(content)
				}

				return template.HTML(builder.String()), nil
			},
			"bladeComponent": func(scope *bladeRuntime.Scope, expression string, bodyText string, slotsJSON string) (template.HTML, error) {
				target, componentScope, err := bladeRuntime.ComponentScope(scope, expression)
				if err != nil {
					return "", err
				}

				defaultSlot, err := e.renderTemplateText(bodyText, scope)
				if err != nil {
					return "", err
				}
				componentScope.SetVar("slot", template.HTML(defaultSlot))

				slotTemplates := make(map[string]string)
				if err := json.Unmarshal([]byte(slotsJSON), &slotTemplates); err != nil {
					return "", err
				}

				for name, slotText := range slotTemplates {
					content, err := e.renderTemplateText(slotText, scope)
					if err != nil {
						return "", err
					}
					componentScope.SetVar(name, template.HTML(content))
				}

				content, err := e.renderFileWithScope(target, componentScope)
				if err != nil {
					return "", wrapRenderFileError(target, err)
				}

				return template.HTML(content), nil
			},
			"bladeYield": func(_ *bladeRuntime.Scope, _ string) template.HTML {
				return ""
			},
			"bladeTag": func(scope *bladeRuntime.Scope, name string, expression string) (template.HTML, error) {
				handler, ok := e.tags[name]
				if !ok {
					return "", fmt.Errorf("未注册的业务标签: @%s", name)
				}

				call := TagCall{
					Params: make(map[string]any),
					scope:  scope,
					engine: e,
				}
				if strings.TrimSpace(expression) != "" {
					callData, err := bladeRuntime.FlexibleTagCallScope(scope, expression)
					if err != nil {
						return "", err
					}
					call.Target = callData.Target
					call.Params = callData.Params
				}

				return handler(call)
			},
			"bladeTagBlock": func(scope *bladeRuntime.Scope, name string, expression string, bodyText string) (template.HTML, error) {
				handler, ok := e.tags[name]
				if !ok {
					return "", fmt.Errorf("未注册的业务标签: @%s", name)
				}

				body, err := e.renderTemplateText(bodyText, scope)
				if err != nil {
					return "", err
				}

				call := TagCall{
					Params: make(map[string]any),
					Body:   template.HTML(body),
					scope:  scope,
					engine: e,
				}
				if strings.TrimSpace(expression) != "" {
					callData, err := bladeRuntime.FlexibleTagCallScope(scope, expression)
					if err != nil {
						return "", err
					}
					call.Target = callData.Target
					call.Params = callData.Params
				}

				return handler(call)
			},
			"bladeDataDirective": func(scope *bladeRuntime.Scope, name string, expression string) (template.HTML, error) {
				call, handler, err := e.resolveNamedDataTag(scope, name, expression)
				if err != nil {
					return "", err
				}

				out, ok, err := call.Out()
				if err != nil {
					return "", err
				}
				if !ok {
					return "", fmt.Errorf("@%s 单指令需要 out 参数", name)
				}

				result, err := handler(call)
				if err != nil {
					return "", err
				}

				value, err := result.exportValue()
				if err != nil {
					return "", err
				}

				scope.SetVar(out, value)
				return "", nil
			},
			"bladeDataDirectiveBlock": func(scope *bladeRuntime.Scope, name string, expression string, bodyText string, emptyBodyText string) (template.HTML, error) {
				call, handler, err := e.resolveNamedDataTag(scope, name, expression)
				if err != nil {
					return "", err
				}

				if out, ok, err := call.Out(); err != nil {
					return "", err
				} else if ok {
					return "", fmt.Errorf("@%s 块标签暂不支持 out 参数: %s", name, out)
				}

				alias, err := call.Alias()
				if err != nil {
					return "", err
				}

				result, err := handler(call)
				if err != nil {
					return "", err
				}

				items, err := result.blockItems()
				if err != nil {
					return "", err
				}

				var builder strings.Builder
				total := len(items)
				if total > bladeRuntime.MaxLoopIterations {
					return "", fmt.Errorf("@%s 返回项数超过最大迭代次数 %d", name, bladeRuntime.MaxLoopIterations)
				}
				if total == 0 {
					content, err := e.renderTemplateText(emptyBodyText, scope)
					if err != nil {
						return "", err
					}
					return template.HTML(content), nil
				}
				for index, item := range items {
					child := scope.WithVar(alias, item)
					bladeRuntime.SetCurrentItem(child, item)
					child.SetVar("loop", bladeRuntime.LoopState(index, total))

					content, err := e.renderTemplateText(bodyText, child)
					if err != nil {
						return "", err
					}
					builder.WriteString(content)
				}

				return template.HTML(builder.String()), nil
			},
		}
	})

	return e.funcs
}

func (e *Engine) resolveNamedDataTag(scope *bladeRuntime.Scope, name string, expression string) (DataTagCall, DataTagFunc, error) {
	handler, ok := e.dataTags[name]
	if !ok {
		return DataTagCall{}, nil, fmt.Errorf("未注册的数据标签: @%s", name)
	}

	params := make(map[string]any)
	if strings.TrimSpace(expression) != "" {
		callData, err := bladeRuntime.NamedTagCallScope(scope, expression)
		if err != nil {
			return DataTagCall{}, nil, err
		}
		params = callData.Params
	}

	return newDataTagCall(name, params, scope), handler, nil
}
