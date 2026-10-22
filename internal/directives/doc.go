/*
Package directives
用于放置内置指令实现与指令注册表。

本包位于 internal 下，仅供根包和内部渲染引擎使用，不作为稳定公开 API。
*/
package directives

import (
	"fmt"
	"slices"
)

// CompileFunc 表示自定义指令的编译函数。
// 返回值应为可插入 Go 模板中的模板片段。
type CompileFunc func(expression string) (string, error)

// BlockCompileFunc 表示块级自定义指令的编译函数。
// body 为已经编译好的子模板文本。
type BlockCompileFunc func(expression string, body string) (string, error)

// EmptyBlockCompileFunc 表示带 @empty 分支的块级业务指令编译函数。
// body 和 emptyBody 均为已经编译好的子模板文本。
type EmptyBlockCompileFunc func(expression string, body string, emptyBody string) (string, error)

// Definition 表示一条自定义指令的结构定义。
// AllowSingle 表示支持单指令，EndName 不为空表示支持块指令。
type Definition struct {
	Name         string
	AllowSingle  bool
	AllowEmpty   bool
	DefaultBlock bool
	EndName      string
}

type blockDirective struct {
	endName      string
	compile      BlockCompileFunc
	compileEmpty EmptyBlockCompileFunc
	defaultBlock bool
}

// Registry 表示指令注册表。
type Registry struct {
	directives      map[string]CompileFunc
	blockDirectives map[string]blockDirective
}

// NewRegistry 创建指令注册表。
func NewRegistry() *Registry {
	return &Registry{
		directives:      make(map[string]CompileFunc),
		blockDirectives: make(map[string]blockDirective),
	}
}

// CloneWithNames 创建注册表快照，并通过映射函数统一开放指令及其结束指令的名称。
func (r *Registry) CloneWithNames(mapName func(string) string, mapEndName func(string, string) string) *Registry {
	if r == nil {
		return nil
	}

	cloned := NewRegistry()
	for name, compile := range r.directives {
		cloned.directives[mapName(name)] = compile
	}
	for name, directive := range r.blockDirectives {
		mappedName := mapName(name)
		directive.endName = mapEndName(name, directive.endName)
		cloned.blockDirectives[mappedName] = directive
	}
	return cloned
}

// Register 注册一个自定义指令编译函数。
func (r *Registry) Register(name string, compile CompileFunc) {
	if compile == nil {
		panic(fmt.Sprintf("自定义指令处理函数不能为空: @%s", name))
	}
	if r.directives == nil {
		r.directives = make(map[string]CompileFunc)
	}
	r.directives[name] = compile
}

// RegisterBlock 注册一个块级自定义指令编译函数。
// 当 endName 为空时默认使用 end+name。
func (r *Registry) RegisterBlock(name string, endName string, compile BlockCompileFunc) {
	if compile == nil {
		panic(fmt.Sprintf("自定义块处理函数不能为空: @%s", name))
	}
	if endName == "" {
		endName = "end" + name
	}
	if r.blockDirectives == nil {
		r.blockDirectives = make(map[string]blockDirective)
	}

	r.blockDirectives[name] = blockDirective{
		endName: endName,
		compile: compile,
	}
}

// RegisterBlockWithEmpty 注册支持 @empty 分支的内部业务块指令。
func (r *Registry) RegisterBlockWithEmpty(name string, endName string, compile EmptyBlockCompileFunc) {
	if compile == nil {
		panic(fmt.Sprintf("带空分支的自定义块处理函数不能为空: @%s", name))
	}
	if endName == "" {
		endName = "end" + name
	}
	if r.blockDirectives == nil {
		r.blockDirectives = make(map[string]blockDirective)
	}

	r.blockDirectives[name] = blockDirective{
		endName:      endName,
		compileEmpty: compile,
		defaultBlock: true,
	}
}

// Exists 判断某个指令是否已经注册。
func (r *Registry) Exists(name string) bool {
	_, ok := r.directives[name]
	if ok {
		return true
	}

	if _, ok = r.blockDirectives[name]; ok {
		return true
	}

	for _, item := range r.blockDirectives {
		if item.endName == name {
			return true
		}
	}

	return false
}

// Compile 执行某个自定义指令的编译逻辑。
func (r *Registry) Compile(name string, expression string) (string, error) {
	compile, ok := r.directives[name]
	if !ok {
		return "", fmt.Errorf("未注册的自定义指令: @%s", name)
	}

	return compile(expression)
}

// CompileBlock 执行某个块级自定义指令的编译逻辑，并按注册能力传递可选空分支。
func (r *Registry) CompileBlock(name string, expression string, body string, emptyBody string) (string, error) {
	directive, ok := r.blockDirectives[name]
	if !ok {
		return "", fmt.Errorf("未注册的块级自定义指令: @%s", name)
	}
	if directive.compileEmpty != nil {
		return directive.compileEmpty(expression, body, emptyBody)
	}
	if emptyBody != "" {
		return "", fmt.Errorf("块级自定义指令不支持 @empty: @%s", name)
	}

	return directive.compile(expression, body)
}

// Names 返回当前已注册的自定义指令名。
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.Definitions()))
	for _, definition := range r.Definitions() {
		names = append(names, definition.Name)
	}

	slices.Sort(names)
	return names
}

// Definitions 返回当前所有自定义指令的定义信息。
func (r *Registry) Definitions() []Definition {
	items := make(map[string]Definition, len(r.directives)+len(r.blockDirectives))
	for name := range r.directives {
		item := items[name]
		item.Name = name
		item.AllowSingle = true
		items[name] = item
	}
	for name, item := range r.blockDirectives {
		definition := items[name]
		definition.Name = name
		definition.AllowEmpty = item.compileEmpty != nil
		definition.DefaultBlock = item.defaultBlock
		definition.EndName = item.endName
		items[name] = definition
	}

	definitions := make([]Definition, 0, len(items))
	for _, definition := range items {
		definitions = append(definitions, definition)
	}

	slices.SortFunc(definitions, func(left, right Definition) int {
		return compareStrings(left.Name, right.Name)
	})

	return definitions
}

func compareStrings(left, right string) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}
