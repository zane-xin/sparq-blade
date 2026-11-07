package compiler

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
	"unicode"

	"github.com/zane-xin/sparq-blade/internal/literal"
)

/*
Generator
负责把 AST 编译为 Go 模板文本。
*/
type Generator struct {
	directiveCompiler DirectiveCompiler
}

/*
DirectiveCompiler
表示自定义指令的编译器协议。
*/
type DirectiveCompiler interface {
	Compile(name string, expression string) (string, error)
	CompileBlock(name string, expression string, body string, emptyBody string) (string, error)
}

/*
GeneratorOption
表示生成器构造选项。
*/
type GeneratorOption func(*Generator)

/*
NewGenerator
创建模板生成器实例。
*/
func NewGenerator(options ...GeneratorOption) *Generator {
	generator := &Generator{}
	for _, option := range options {
		option(generator)
	}

	return generator
}

/*
WithDirectiveCompiler
向生成器注入自定义指令编译器。
*/
func WithDirectiveCompiler(compiler DirectiveCompiler) GeneratorOption {
	return func(generator *Generator) {
		generator.directiveCompiler = compiler
	}
}

/*
Generate
把文档 AST 编译为可执行的 Go 模板文本。
*/
func (i *Generator) Generate(document *Document) (string, error) {
	if document == nil {
		return "", nil
	}

	var builder strings.Builder
	if err := i.writeNodes(&builder, document.Nodes); err != nil {
		return "", err
	}

	return builder.String(), nil
}

func (i *Generator) writeNodes(builder *strings.Builder, nodes []Node) error {
	for _, node := range nodes {
		if err := i.writeNode(builder, node); err != nil {
			return err
		}
	}

	return nil
}

func (i *Generator) writeNode(builder *strings.Builder, node Node) error {
	switch value := node.(type) {
	case *TextNode:
		builder.WriteString(value.Text)
	case *EchoNode:
		if value.Raw {
			builder.WriteString(fmt.Sprintf("{{ bladeRaw . %q }}", value.Expression))
			return nil
		}
		builder.WriteString(fmt.Sprintf("{{ bladeEcho . %q }}", value.Expression))
	case *CommentNode:
		return nil
	case *ExtendsNode:
		return nil
	case *IncludeNode:
		builder.WriteString(fmt.Sprintf("{{ bladeInclude . %q }}", value.Target))
	case *EachNode:
		builder.WriteString(fmt.Sprintf("{{ bladeEach . %q }}", value.Expression))
	case *ComponentNode:
		bodyText, err := i.generateNodes(value.Body)
		if err != nil {
			return err
		}

		slotTexts := make(map[string]string, len(value.Slots))
		for _, slot := range value.Slots {
			slotName, ok := literal.ParseString(slot.Name)
			if !ok {
				slotName = slot.Name
			}

			text, err := i.generateNodes(slot.Body)
			if err != nil {
				return err
			}

			slotTexts[slotName] = text
		}

		slotJSON, err := json.Marshal(slotTexts)
		if err != nil {
			return err
		}

		builder.WriteString(fmt.Sprintf("{{ bladeComponent . %q %q %q }}", value.Target, bodyText, string(slotJSON)))
	case *CustomDirectiveNode:
		if i.directiveCompiler == nil {
			return &Error{
				Code:    ErrCodeUnsupportedDirective,
				Message: "未配置自定义指令编译器: @" + value.Name,
				Span:    value.Span,
			}
		}

		if value.EndName != "" {
			bodyText, err := i.generateNodes(value.Body)
			if err != nil {
				return err
			}

			emptyBodyText, err := i.generateNodes(value.EmptyBody)
			if err != nil {
				return err
			}

			text, err := i.directiveCompiler.CompileBlock(value.Name, value.Expression, bodyText, emptyBodyText)
			if err != nil {
				return err
			}
			builder.WriteString(text)
			return nil
		}

		text, err := i.directiveCompiler.Compile(value.Name, value.Expression)
		if err != nil {
			return err
		}
		builder.WriteString(text)
	case *YieldNode:
		builder.WriteString(fmt.Sprintf("{{ bladeYield . %q }}", value.Target))
	case *SectionNode:
		return i.writeNodes(builder, value.Body)
	case *ConditionalNode:
		return i.writeConditional(builder, value)
	case *LoopNode:
		rangeFunc, err := i.loopRangeFunction(value)
		if err != nil {
			return err
		}
		builder.WriteString(fmt.Sprintf("{{ range %s . %q }}", rangeFunc, value.Expression))
		if err := i.writeNodes(builder, value.Body); err != nil {
			return err
		}
		builder.WriteString("{{ end }}")
	case *ListNode:
		builder.WriteString(fmt.Sprintf("{{ range bladeList . %q }}", value.Expression))
		builder.WriteString(`{{ if bladeTruthy . "Empty" }}`)
		if err := i.writeNodes(builder, value.EmptyBody); err != nil {
			return err
		}
		builder.WriteString("{{ else }}")
		if err := i.writeNodes(builder, value.Body); err != nil {
			return err
		}
		builder.WriteString("{{ end }}{{ end }}")
	case *ChildrenNode:
		spec, err := parseChildrenCompileSpec(value.Expression)
		if err != nil {
			return err
		}

		builder.WriteString(fmt.Sprintf("{{ range bladeChildren . %q }}", value.Expression))
		builder.WriteString(`{{ if bladeTruthy . "Empty" }}`)
		if err := i.writeNodes(builder, value.EmptyBody); err != nil {
			return err
		}
		builder.WriteString("{{ else }}")
		if spec.Wrap != "" {
			builder.WriteString(fmt.Sprintf(`{{ if bladeTruthy . "loop.first" }}%s{{ end }}`, spec.openTag()))
		}
		if err := i.writeNodes(builder, value.Body); err != nil {
			return err
		}
		if spec.Wrap != "" {
			builder.WriteString(fmt.Sprintf(`{{ if bladeTruthy . "loop.last" }}</%s>{{ end }}`, spec.Wrap))
		}
		builder.WriteString("{{ end }}{{ end }}")
	default:
		return &Error{
			Code:    ErrCodeUnexpectedToken,
			Message: fmt.Sprintf("无法生成未知节点类型: %T", node),
			Span:    node.NodeSpan(),
		}
	}

	return nil
}

func (i *Generator) writeConditional(builder *strings.Builder, node *ConditionalNode) error {
	for index, branch := range node.Branches {
		check := "bladeTruthy"
		if branch.Directive == "unless" {
			check = "bladeUnless"
		}

		if index == 0 {
			builder.WriteString(fmt.Sprintf("{{ if %s . %q }}", check, branch.Expression))
		} else {
			builder.WriteString(fmt.Sprintf("{{ else if %s . %q }}", check, branch.Expression))
		}

		if err := i.writeNodes(builder, branch.Body); err != nil {
			return err
		}
	}

	if len(node.ElseBody) > 0 {
		builder.WriteString("{{ else }}")
		if err := i.writeNodes(builder, node.ElseBody); err != nil {
			return err
		}
	}

	builder.WriteString("{{ end }}")
	return nil
}

type childrenCompileSpec struct {
	Wrap  string
	Class string
}

func parseChildrenCompileSpec(expression string) (childrenCompileSpec, error) {
	parts, err := splitTopLevelCompileArgs(expression)
	if err != nil {
		return childrenCompileSpec{}, err
	}

	spec := childrenCompileSpec{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		name, value, ok := strings.Cut(part, "=")
		if !ok {
			return childrenCompileSpec{}, fmt.Errorf("children 参数必须使用 key=value 形式: %s", part)
		}

		key := strings.TrimSpace(name)
		switch key {
		case "wrap":
			wrap, ok := literal.ParseString(value)
			if !ok {
				return childrenCompileSpec{}, fmt.Errorf("children wrap 必须是字符串字面量: %s", value)
			}
			if !isSafeWrapTag(wrap) {
				return childrenCompileSpec{}, fmt.Errorf("children wrap 标签名非法: %s", wrap)
			}
			spec.Wrap = wrap
		case "class":
			className, ok := literal.ParseString(value)
			if !ok {
				return childrenCompileSpec{}, fmt.Errorf("children class 必须是字符串字面量: %s", value)
			}
			spec.Class = className
		case "as", "field":
			continue
		default:
			return childrenCompileSpec{}, fmt.Errorf("children 不支持参数: %s", key)
		}
	}

	return spec, nil
}

func (s childrenCompileSpec) openTag() string {
	if s.Class == "" {
		return "<" + s.Wrap + ">"
	}

	return fmt.Sprintf(`<%s class="%s">`, s.Wrap, html.EscapeString(s.Class))
}

func splitTopLevelCompileArgs(expression string) ([]string, error) {
	if strings.TrimSpace(expression) == "" {
		return nil, nil
	}

	parts := make([]string, 0, 4)
	start := 0
	var quote rune
	escaped := false

	for index, item := range expression {
		if quote != 0 {
			if escaped {
				escaped = false
				continue
			}
			if item == '\\' {
				escaped = true
				continue
			}
			if item == quote {
				quote = 0
			}
			continue
		}

		switch item {
		case '\'', '"':
			quote = item
		case ',':
			parts = append(parts, expression[start:index])
			start = index + 1
		}
	}

	if quote != 0 {
		return nil, fmt.Errorf("children 参数字符串未正常闭合: %s", expression)
	}

	parts = append(parts, expression[start:])
	return parts, nil
}

func isSafeWrapTag(value string) bool {
	if value == "" {
		return false
	}

	for index, item := range value {
		if index == 0 {
			if !unicode.IsLetter(item) {
				return false
			}
			continue
		}

		if !(unicode.IsLetter(item) || unicode.IsDigit(item) || item == '-') {
			return false
		}
	}

	return true
}

func (i *Generator) generateNodes(nodes []Node) (string, error) {
	var builder strings.Builder
	if err := i.writeNodes(&builder, nodes); err != nil {
		return "", err
	}

	return builder.String(), nil
}

func (i *Generator) loopRangeFunction(node *LoopNode) (string, error) {
	switch node.Directive {
	case "foreach":
		return "bladeRange", nil
	case "for":
		return "bladeForRange", nil
	default:
		return "", &Error{
			Code:    ErrCodeUnsupportedDirective,
			Message: "当前版本尚未支持该循环指令: @" + node.Directive,
			Span:    node.Span,
		}
	}
}
