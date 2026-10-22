package compiler

import "strings"

/*
Parser
负责把 token 序列解析为 AST。
*/
type Parser struct {
	tokens              []Token
	index               int
	customDirectives    map[string]CustomDirectiveSpec
	customDirectiveEnds map[string]struct{}
}

// CustomDirectiveSpec 表示一个自定义指令的结构定义。
// AllowSingle 表示支持单指令，EndName 不为空表示支持块指令。
type CustomDirectiveSpec struct {
	Name         string
	AllowSingle  bool
	AllowEmpty   bool
	DefaultBlock bool
	EndName      string
}

/*
ParserOption
表示语法分析器的构造选项。
*/
type ParserOption func(*Parser)

/*
NewParser
创建语法分析器实例。
*/
func NewParser(tokens []Token, options ...ParserOption) *Parser {
	parser := &Parser{
		tokens:              tokens,
		customDirectives:    make(map[string]CustomDirectiveSpec),
		customDirectiveEnds: make(map[string]struct{}),
	}

	for _, option := range options {
		option(parser)
	}

	return parser
}

/*
WithCustomDirectives
向语法分析器声明一组可识别的自定义指令名。
*/
func WithCustomDirectives(names ...string) ParserOption {
	return func(parser *Parser) {
		for _, name := range names {
			parser.customDirectives[name] = CustomDirectiveSpec{Name: name, AllowSingle: true}
		}
	}
}

/*
WithCustomDirectiveSpecs
向语法分析器声明一组自定义指令结构定义。
*/
func WithCustomDirectiveSpecs(specs ...CustomDirectiveSpec) ParserOption {
	return func(parser *Parser) {
		for _, spec := range specs {
			parser.customDirectives[spec.Name] = spec
			if spec.EndName != "" {
				parser.customDirectiveEnds[spec.EndName] = struct{}{}
			}
		}
	}
}

/*
Parse
执行完整的语法分析流程。
*/
func (p *Parser) Parse() (*Document, error) {
	nodes, stop, err := p.parseNodes(nil)
	if err != nil {
		return nil, err
	}
	if stop != nil {
		return nil, p.unexpectedDirective(*stop)
	}

	span := Span{}
	if len(nodes) > 0 {
		span.Start = nodes[0].NodeSpan().Start
		span.End = nodes[len(nodes)-1].NodeSpan().End
	}

	return &Document{
		Nodes: nodes,
		Span:  span,
	}, nil
}

func (p *Parser) parseNodes(stops map[string]struct{}) ([]Node, *Token, error) {
	nodes := make([]Node, 0, 8)

	for {
		token := p.current()

		switch token.Kind {
		case TokenKindEOF:
			if stops != nil {
				return nil, nil, &Error{
					Code:    ErrCodeUnexpectedEOF,
					Message: "块结构未正常闭合",
					Span:    token.Span,
				}
			}
			return nodes, nil, nil
		case TokenKindText:
			nodes = append(nodes, &TextNode{
				Text: token.Literal,
				Span: token.Span,
			})
			p.advance()
		case TokenKindEchoEscaped:
			nodes = append(nodes, &EchoNode{
				Expression: token.Literal,
				Raw:        false,
				Span:       token.Span,
			})
			p.advance()
		case TokenKindEchoRaw:
			nodes = append(nodes, &EchoNode{
				Expression: token.Literal,
				Raw:        true,
				Span:       token.Span,
			})
			p.advance()
		case TokenKindComment:
			nodes = append(nodes, &CommentNode{
				Content: token.Literal,
				Span:    token.Span,
			})
			p.advance()
		case TokenKindDirective:
			if stops != nil {
				if _, ok := stops[token.Name]; ok {
					return nodes, &token, nil
				}
			}

			node, err := p.parseDirective(stops)
			if err != nil {
				return nil, nil, err
			}
			nodes = append(nodes, node)
		default:
			return nil, nil, &Error{
				Code:    ErrCodeUnexpectedToken,
				Message: "遇到了未知 token 类型",
				Span:    token.Span,
			}
		}
	}
}

func (p *Parser) parseDirective(stops map[string]struct{}) (Node, error) {
	token := p.current()

	switch token.Name {
	case "extends":
		p.advance()
		return &ExtendsNode{
			Target: token.Expression,
			Span:   token.Span,
		}, nil
	case "include":
		p.advance()
		return &IncludeNode{
			Target: token.Expression,
			Span:   token.Span,
		}, nil
	case "each":
		p.advance()
		return &EachNode{
			Expression: token.Expression,
			Span:       token.Span,
		}, nil
	case "component":
		return p.parseComponent()
	case "yield":
		p.advance()
		return &YieldNode{
			Target: token.Expression,
			Span:   token.Span,
		}, nil
	case "section":
		return p.parseSection()
	case "if", "unless":
		return p.parseConditional()
	case "for", "foreach":
		return p.parseLoop()
	case "list":
		return p.parseList()
	case "children":
		return p.parseChildren()
	case "slot", "elseif", "else", "empty", "end", "endif", "endunless", "endfor", "endforeach", "endlist", "endchildren", "endsection", "endcomponent", "endslot":
		return nil, p.unexpectedDirective(token)
	default:
		if p.isCustomDirectiveEnd(token.Name) {
			return nil, p.unexpectedDirective(token)
		}

		if spec, ok := p.customDirective(token.Name); ok {
			if spec.EndName != "" && !spec.AllowSingle {
				return p.parseCustomDirectiveBlock(spec)
			}
			if spec.EndName != "" && spec.DefaultBlock && hasTagParam(token.Expression, "out") {
				p.advance()
				return &CustomDirectiveNode{
					Name:       token.Name,
					Expression: token.Expression,
					Span:       token.Span,
				}, nil
			}
			if spec.EndName != "" && hasCustomDirectiveSpecificEndAhead(p.tokens, spec, p.index, stops) {
				return p.parseCustomDirectiveBlock(spec)
			}
			if spec.EndName != "" && hasTagParam(token.Expression, "out") {
				p.advance()
				return &CustomDirectiveNode{
					Name:       token.Name,
					Expression: token.Expression,
					Span:       token.Span,
				}, nil
			}
			if spec.EndName != "" && spec.DefaultBlock {
				return p.parseCustomDirectiveBlock(spec)
			}
			p.advance()
			return &CustomDirectiveNode{
				Name:       token.Name,
				Expression: token.Expression,
				Span:       token.Span,
			}, nil
		}

		return nil, &Error{
			Code:    ErrCodeUnsupportedDirective,
			Message: "当前版本尚未支持该指令: @" + token.Name,
			Span:    token.Span,
		}
	}
}

func (p *Parser) parseSection() (Node, error) {
	start := p.current()
	p.advance()

	body, stop, err := p.parseNodes(map[string]struct{}{
		"endsection": {},
		"end":        {},
	})
	if err != nil {
		return nil, p.wrapUnexpectedEOF(err, "section 未正常闭合", start.Span)
	}
	if stop == nil {
		return nil, &Error{
			Code:    ErrCodeUnexpectedEOF,
			Message: "section 未正常闭合",
			Span:    start.Span,
		}
	}

	end := p.current()
	p.advance()

	return &SectionNode{
		Name: start.Expression,
		Body: body,
		Span: Span{
			Start: start.Span.Start,
			End:   end.Span.End,
		},
	}, nil
}

func (p *Parser) parseComponent() (Node, error) {
	start := p.current()
	p.advance()

	defaultBody := make([]Node, 0, 4)
	slots := make([]ComponentSlot, 0, 2)

	for {
		body, stop, err := p.parseNodes(map[string]struct{}{
			"slot":         {},
			"endcomponent": {},
			"end":          {},
		})
		if err != nil {
			return nil, p.wrapUnexpectedEOF(err, "component 未正常闭合", start.Span)
		}

		defaultBody = append(defaultBody, body...)

		if stop == nil {
			return nil, &Error{
				Code:    ErrCodeUnexpectedEOF,
				Message: "component 未正常闭合",
				Span:    start.Span,
			}
		}

		switch stop.Name {
		case "slot":
			slot, err := p.parseComponentSlot()
			if err != nil {
				return nil, err
			}
			slots = append(slots, slot)
		case "endcomponent", "end":
			end := p.current()
			p.advance()

			return &ComponentNode{
				Target: start.Expression,
				Body:   defaultBody,
				Slots:  slots,
				Span: Span{
					Start: start.Span.Start,
					End:   end.Span.End,
				},
			}, nil
		}
	}
}

func (p *Parser) parseComponentSlot() (ComponentSlot, error) {
	start := p.current()
	p.advance()

	body, stop, err := p.parseNodes(map[string]struct{}{
		"endslot":      {},
		"endcomponent": {},
		"end":          {},
	})
	if err != nil {
		return ComponentSlot{}, p.wrapUnexpectedEOF(err, "slot 未正常闭合", start.Span)
	}
	if stop == nil || (stop.Name != "endslot" && stop.Name != "end") {
		return ComponentSlot{}, &Error{
			Code:    ErrCodeUnexpectedEOF,
			Message: "slot 未正常闭合",
			Span:    start.Span,
		}
	}

	end := p.current()
	p.advance()

	return ComponentSlot{
		Name: start.Expression,
		Body: body,
		Span: Span{
			Start: start.Span.Start,
			End:   end.Span.End,
		},
	}, nil
}

func (p *Parser) parseCustomDirectiveBlock(spec CustomDirectiveSpec) (Node, error) {
	start := p.current()
	p.advance()

	stops := map[string]struct{}{
		spec.EndName: {},
		"end":        {},
	}
	if spec.AllowEmpty {
		stops["empty"] = struct{}{}
	}

	body, stop, err := p.parseNodes(stops)
	if err != nil {
		return nil, p.wrapUnexpectedEOF(err, "自定义指令未正常闭合: @"+spec.Name, start.Span)
	}

	var emptyBody []Node
	if stop != nil && stop.Name == "empty" {
		p.advance()
		emptyBody, stop, err = p.parseNodes(map[string]struct{}{
			spec.EndName: {},
			"end":        {},
		})
		if err != nil {
			return nil, p.wrapUnexpectedEOF(err, "自定义指令未正常闭合: @"+spec.Name, start.Span)
		}
	}
	if stop == nil || (stop.Name != spec.EndName && stop.Name != "end") {
		return nil, &Error{
			Code:    ErrCodeUnexpectedEOF,
			Message: "自定义指令未正常闭合: @" + spec.Name,
			Span:    start.Span,
		}
	}

	end := p.current()
	p.advance()

	return &CustomDirectiveNode{
		Name:       spec.Name,
		Expression: start.Expression,
		Body:       body,
		EmptyBody:  emptyBody,
		EndName:    spec.EndName,
		Span: Span{
			Start: start.Span.Start,
			End:   end.Span.End,
		},
	}, nil
}

func (p *Parser) parseConditional() (Node, error) {
	start := p.current()
	endName := "endif"
	if start.Name == "unless" {
		endName = "endunless"
	}

	p.advance()

	branches := make([]ConditionalBranch, 0, 2)
	body, stop, err := p.parseNodes(map[string]struct{}{
		"elseif": {},
		"else":   {},
		endName:  {},
		"end":    {},
	})
	if err != nil {
		return nil, err
	}

	branches = append(branches, ConditionalBranch{
		Directive:  start.Name,
		Expression: start.Expression,
		Body:       body,
		Span:       start.Span,
	})

	for stop != nil && stop.Name == "elseif" {
		elseif := p.current()
		p.advance()

		body, stop, err = p.parseNodes(map[string]struct{}{
			"elseif": {},
			"else":   {},
			endName:  {},
			"end":    {},
		})
		if err != nil {
			return nil, err
		}

		branches = append(branches, ConditionalBranch{
			Directive:  elseif.Name,
			Expression: elseif.Expression,
			Body:       body,
			Span:       elseif.Span,
		})
	}

	var elseBody []Node
	if stop != nil && stop.Name == "else" {
		p.advance()
		elseBody, stop, err = p.parseNodes(map[string]struct{}{
			endName: {},
			"end":   {},
		})
		if err != nil {
			return nil, err
		}
	}

	if stop == nil || (stop.Name != endName && stop.Name != "end") {
		return nil, &Error{
			Code:    ErrCodeUnexpectedEOF,
			Message: "条件结构未正常闭合",
			Span:    start.Span,
		}
	}

	end := p.current()
	p.advance()

	return &ConditionalNode{
		Branches: branches,
		ElseBody: elseBody,
		Span: Span{
			Start: start.Span.Start,
			End:   end.Span.End,
		},
	}, nil
}

func (p *Parser) parseLoop() (Node, error) {
	start := p.current()
	endName := map[string]string{
		"for":     "endfor",
		"foreach": "endforeach",
	}[start.Name]

	p.advance()

	body, stop, err := p.parseNodes(map[string]struct{}{
		endName: {},
		"end":   {},
	})
	if err != nil {
		return nil, err
	}
	if stop == nil || (stop.Name != endName && stop.Name != "end") {
		return nil, &Error{
			Code:    ErrCodeUnexpectedEOF,
			Message: "循环结构未正常闭合",
			Span:    start.Span,
		}
	}

	end := p.current()
	p.advance()

	return &LoopNode{
		Directive:  start.Name,
		Expression: start.Expression,
		Body:       body,
		Span: Span{
			Start: start.Span.Start,
			End:   end.Span.End,
		},
	}, nil
}

func (p *Parser) parseList() (Node, error) {
	start := p.current()
	p.advance()

	body, stop, err := p.parseNodes(map[string]struct{}{
		"empty":   {},
		"endlist": {},
		"end":     {},
	})
	if err != nil {
		return nil, err
	}

	var emptyBody []Node
	if stop != nil && stop.Name == "empty" {
		p.advance()
		emptyBody, stop, err = p.parseNodes(map[string]struct{}{
			"endlist": {},
			"end":     {},
		})
		if err != nil {
			return nil, err
		}
	}

	if stop == nil || (stop.Name != "endlist" && stop.Name != "end") {
		return nil, &Error{
			Code:    ErrCodeUnexpectedEOF,
			Message: "list 未正常闭合",
			Span:    start.Span,
		}
	}

	end := p.current()
	p.advance()

	return &ListNode{
		Expression: start.Expression,
		Body:       body,
		EmptyBody:  emptyBody,
		Span: Span{
			Start: start.Span.Start,
			End:   end.Span.End,
		},
	}, nil
}

func (p *Parser) parseChildren() (Node, error) {
	start := p.current()
	p.advance()

	body, stop, err := p.parseNodes(map[string]struct{}{
		"empty":       {},
		"endchildren": {},
		"end":         {},
	})
	if err != nil {
		return nil, err
	}

	var emptyBody []Node
	if stop != nil && stop.Name == "empty" {
		p.advance()
		emptyBody, stop, err = p.parseNodes(map[string]struct{}{
			"endchildren": {},
			"end":         {},
		})
		if err != nil {
			return nil, err
		}
	}

	if stop == nil || (stop.Name != "endchildren" && stop.Name != "end") {
		return nil, &Error{
			Code:    ErrCodeUnexpectedEOF,
			Message: "children 未正常闭合",
			Span:    start.Span,
		}
	}

	end := p.current()
	p.advance()

	return &ChildrenNode{
		Expression: start.Expression,
		Body:       body,
		EmptyBody:  emptyBody,
		Span: Span{
			Start: start.Span.Start,
			End:   end.Span.End,
		},
	}, nil
}

func (p *Parser) current() Token {
	if p.index >= len(p.tokens) {
		return Token{
			Kind: TokenKindEOF,
		}
	}
	return p.tokens[p.index]
}

func (p *Parser) advance() {
	if p.index < len(p.tokens) {
		p.index++
	}
}

func (p *Parser) unexpectedDirective(token Token) error {
	return &Error{
		Code:    ErrCodeUnexpectedToken,
		Message: "出现了不应在当前位置出现的指令: @" + token.Name,
		Span:    token.Span,
	}
}

func (p *Parser) wrapUnexpectedEOF(err error, message string, span Span) error {
	compileErr, ok := err.(*Error)
	if !ok {
		return err
	}

	if compileErr.Code != ErrCodeUnexpectedEOF {
		return err
	}

	return &Error{
		Code:    ErrCodeUnexpectedEOF,
		Message: message,
		Span:    span,
	}
}

func (p *Parser) customDirective(name string) (CustomDirectiveSpec, bool) {
	spec, ok := p.customDirectives[name]
	return spec, ok
}

func (p *Parser) isCustomDirectiveEnd(name string) bool {
	_, ok := p.customDirectiveEnds[name]
	return ok
}

func hasCustomDirectiveSpecificEndAhead(tokens []Token, spec CustomDirectiveSpec, index int, stops map[string]struct{}) bool {
	if spec.EndName == "" {
		return false
	}

	depth := 0
	for current := index + 1; current < len(tokens); current++ {
		token := tokens[current]
		if token.Kind != TokenKindDirective {
			continue
		}

		tokenName := token.Name
		switch tokenName {
		case spec.Name:
			depth++
		case spec.EndName:
			if depth == 0 {
				return true
			}
			depth--
		default:
			if stops != nil {
				if _, ok := stops[tokenName]; ok && depth == 0 {
					return false
				}
			}
		}
	}

	return false
}

func hasTagParam(expression string, name string) bool {
	for _, part := range splitDirectiveArgs(expression) {
		part = strings.TrimSpace(part)
		if part == name {
			return true
		}
		key, _, ok := strings.Cut(part, "=")
		if ok && strings.TrimSpace(key) == name {
			return true
		}
	}

	return false
}

func splitDirectiveArgs(expression string) []string {
	parts := make([]string, 0, 4)
	start := 0
	var quote rune
	escaped := false

	for index, r := range expression {
		if quote != 0 {
			if escaped {
				escaped = false
				continue
			}
			if r == '\\' {
				escaped = true
				continue
			}
			if r == quote {
				quote = 0
			}
			continue
		}

		switch r {
		case '\'', '"':
			quote = r
		case ',':
			parts = append(parts, expression[start:index])
			start = index + len(string(r))
		}
	}

	parts = append(parts, expression[start:])
	return parts
}
