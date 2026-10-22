package runtime

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

// Context 是保留给运行时扩展的基础数据载体；实际模板求值使用 Scope。
type Context struct {
	Data any
}

// Scope 保存根 ViewContext、循环/include 等局部变量、函数目录以及用于递归保护的模板调用链。
type Scope struct {
	Root      any
	Vars      map[string]any
	Functions map[string]Function
	templates []string
}

// NewScope 基于 ViewContext 创建空局部变量作用域；模板函数由 Engine 随后绑定。
func NewScope(root any) *Scope {
	return &Scope{
		Root: root,
		Vars: make(map[string]any),
	}
}

// WithFunctions 绑定 Engine 初始化完成后的函数目录；派生作用域共享该只读 Map。
func (s *Scope) WithFunctions(functions map[string]Function) *Scope {
	s.Functions = functions
	return s
}

// WithVar 复制当前局部变量并加入一个值，用于循环和数据标签派生互不污染的子作用域。
func (s *Scope) WithVar(name string, value any) *Scope {
	next := &Scope{
		Root:      s.Root,
		Vars:      make(map[string]any, len(s.Vars)+1),
		Functions: s.Functions,
		templates: s.templates,
	}

	for key, item := range s.Vars {
		next.Vars[key] = item
	}

	next.Vars[name] = value
	return next
}

// Clone 复制局部变量 Map，并共享只读的根数据、函数目录和模板调用链。
func (s *Scope) Clone() *Scope {
	next := &Scope{
		Root:      s.Root,
		Vars:      make(map[string]any, len(s.Vars)),
		Functions: s.Functions,
		templates: s.templates,
	}

	for key, item := range s.Vars {
		next.Vars[key] = item
	}

	return next
}

// EnterTemplate 追加不可变模板调用链并拒绝循环或过深组合；Vars 继续共享，以保持 include 的既有作用域语义。
func (s *Scope) EnterTemplate(name string, maxDepth int) (*Scope, error) {
	for _, current := range s.templates {
		if current == name {
			chain := append(append([]string(nil), s.templates...), name)
			return nil, fmt.Errorf("模板组合存在循环调用: %s", strings.Join(chain, " -> "))
		}
	}
	if len(s.templates) >= maxDepth {
		return nil, fmt.Errorf("模板组合超过最大深度 %d", maxDepth)
	}

	return &Scope{
		Root:      s.Root,
		Vars:      s.Vars,
		Functions: s.Functions,
		templates: append(append([]string(nil), s.templates...), name),
	}, nil
}

// WithData 复制当前局部变量后展开附加 map 或 struct，供 include/component 传参使用，不修改父作用域。
func (s *Scope) WithData(data any) *Scope {
	next := s.Clone()
	mergeRootData(next.Vars, data)
	return next
}

// SetVar 写入当前局部变量 Map；include 共享作用域时，该值可被调用模板后续内容读取。
func (s *Scope) SetVar(name string, value any) {
	if s.Vars == nil {
		s.Vars = make(map[string]any)
	}

	s.Vars[name] = value
}

func mergeRootData(target map[string]any, value any) {
	if value == nil {
		return
	}

	reflected := reflect.ValueOf(value)
	for reflected.Kind() == reflect.Pointer || reflected.Kind() == reflect.Interface {
		if reflected.IsNil() {
			return
		}
		reflected = reflected.Elem()
	}

	switch reflected.Kind() {
	case reflect.Map:
		iter := reflected.MapRange()
		for iter.Next() {
			key := iter.Key()
			if key.Kind() != reflect.String {
				continue
			}
			target[key.String()] = iter.Value().Interface()
		}
	case reflect.Struct:
		typeInfo := reflected.Type()
		for index := 0; index < typeInfo.NumField(); index++ {
			field := typeInfo.Field(index)
			if field.PkgPath != "" {
				continue
			}

			value := reflected.Field(index)
			if !value.CanInterface() {
				continue
			}

			target[field.Name] = value.Interface()
			target[lowerFirst(field.Name)] = value.Interface()
		}
	default:
		target["Value"] = reflected.Interface()
	}
}

func lowerFirst(value string) string {
	if value == "" {
		return value
	}

	runes := []rune(value)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
