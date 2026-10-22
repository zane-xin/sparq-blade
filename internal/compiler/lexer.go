package compiler

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Syntax 定义词法器识别的输出、原样输出、注释和 @ 转义分隔符；空字段在合并时沿用默认值。
type Syntax struct {
	EchoOpen     string
	EchoClose    string
	RawOpen      string
	RawClose     string
	CommentOpen  string
	CommentClose string
	EscapeAt     string
}

// DefaultSyntax 返回 Blade 兼容的默认分隔符集合。
func DefaultSyntax() Syntax {
	return Syntax{
		EchoOpen:     "{{",
		EchoClose:    "}}",
		RawOpen:      "{!!",
		RawClose:     "!!}",
		CommentOpen:  "{{--",
		CommentClose: "--}}",
		EscapeAt:     "@@",
	}
}

// Merge 只覆盖 override 中的非空字段，便于调用方局部定制语法而不重复声明整套分隔符。
func (s Syntax) Merge(override Syntax) Syntax {
	base := s.normalized()

	if override.EchoOpen != "" {
		base.EchoOpen = override.EchoOpen
	}
	if override.EchoClose != "" {
		base.EchoClose = override.EchoClose
	}
	if override.RawOpen != "" {
		base.RawOpen = override.RawOpen
	}
	if override.RawClose != "" {
		base.RawClose = override.RawClose
	}
	if override.CommentOpen != "" {
		base.CommentOpen = override.CommentOpen
	}
	if override.CommentClose != "" {
		base.CommentClose = override.CommentClose
	}
	if override.EscapeAt != "" {
		base.EscapeAt = override.EscapeAt
	}

	return base
}

func (s Syntax) normalized() Syntax {
	base := DefaultSyntax()
	if s.EchoOpen != "" {
		base.EchoOpen = s.EchoOpen
	}
	if s.EchoClose != "" {
		base.EchoClose = s.EchoClose
	}
	if s.RawOpen != "" {
		base.RawOpen = s.RawOpen
	}
	if s.RawClose != "" {
		base.RawClose = s.RawClose
	}
	if s.CommentOpen != "" {
		base.CommentOpen = s.CommentOpen
	}
	if s.CommentClose != "" {
		base.CommentClose = s.CommentClose
	}
	if s.EscapeAt != "" {
		base.EscapeAt = s.EscapeAt
	}

	return base
}

// LexerOption 在词法分析开始前调整 Lexer 配置。
type LexerOption func(*Lexer)

// Lexer 按 UTF-8 位置扫描模板源码，并为每个 token 保留字节偏移和从 1 开始的行列范围。
type Lexer struct {
	source string
	offset int
	line   int
	column int
	syntax Syntax
}

// NewLexer 创建使用默认语法的词法器，再按顺序应用 options。
func NewLexer(source string, options ...LexerOption) *Lexer {
	lexer := &Lexer{
		source: source,
		line:   1,
		column: 1,
		syntax: DefaultSyntax(),
	}

	for _, option := range options {
		option(lexer)
	}

	lexer.syntax = lexer.syntax.normalized()
	return lexer
}

// WithSyntax 把非空分隔符合并到当前词法器配置中。
func WithSyntax(syntax Syntax) LexerOption {
	return func(l *Lexer) {
		l.syntax = l.syntax.Merge(syntax)
	}
}

// Tokenize 扫描完整源码并在末尾追加 EOF token；未闭合表达式会返回带源码位置的编译错误。
func (l *Lexer) Tokenize() ([]Token, error) {
	tokens := make([]Token, 0, 16)

	for !l.eof() {
		switch {
		case l.hasPrefix(l.syntax.CommentOpen):
			token, err := l.readComment()
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
		case l.hasPrefix(l.syntax.RawOpen):
			token, err := l.readRawEcho()
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
		case l.hasPrefix(l.syntax.EchoOpen):
			token, err := l.readEscapedEcho()
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
		case l.syntax.EscapeAt != "" && l.hasPrefix(l.syntax.EscapeAt):
			tokens = append(tokens, l.readEscapedAt())
		case l.isDirectiveStart():
			token, err := l.readDirective()
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
		default:
			token := l.readText()
			if token.Literal != "" {
				tokens = append(tokens, token)
			}
		}
	}

	pos := l.position()
	tokens = append(tokens, Token{
		Kind: TokenKindEOF,
		Span: Span{Start: pos, End: pos},
	})

	return tokens, nil
}

func (l *Lexer) readComment() (Token, error) {
	return l.readDelimited(TokenKindComment, l.syntax.CommentOpen, l.syntax.CommentClose)
}

func (l *Lexer) readRawEcho() (Token, error) {
	return l.readDelimited(TokenKindEchoRaw, l.syntax.RawOpen, l.syntax.RawClose)
}

func (l *Lexer) readEscapedEcho() (Token, error) {
	return l.readDelimited(TokenKindEchoEscaped, l.syntax.EchoOpen, l.syntax.EchoClose)
}

func (l *Lexer) readEscapedAt() Token {
	start := l.position()
	l.consume(l.syntax.EscapeAt)
	end := l.position()

	return Token{
		Kind:    TokenKindText,
		Literal: "@",
		Span:    Span{Start: start, End: end},
	}
}

func (l *Lexer) readDelimited(kind TokenKind, open, close string) (Token, error) {
	start := l.position()
	l.consume(open)
	endIndex := strings.Index(l.source[l.offset:], close)
	if endIndex < 0 {
		return Token{}, &Error{
			Code:    ErrCodeUnexpectedEOF,
			Message: "表达式未正常闭合",
			Span:    Span{Start: start, End: start},
		}
	}

	content := l.source[l.offset : l.offset+endIndex]
	l.consume(content)
	l.consume(close)
	end := l.position()

	return Token{
		Kind:    kind,
		Literal: strings.TrimSpace(content),
		Span:    Span{Start: start, End: end},
	}, nil
}

func (l *Lexer) readDirective() (Token, error) {
	start := l.position()
	l.consume("@")

	name := l.readIdentifier()
	if name == "" {
		return Token{}, &Error{
			Code:    ErrCodeInvalidDirective,
			Message: "无效的指令名称",
			Span:    Span{Start: start, End: l.position()},
		}
	}

	l.skipInlineSpaces()

	expression := ""
	if l.peek() == '(' {
		value, err := l.readParenthesized()
		if err != nil {
			return Token{}, err
		}
		expression = strings.TrimSpace(value)
	}

	end := l.position()

	return Token{
		Kind:       TokenKindDirective,
		Literal:    l.source[start.Offset:end.Offset],
		Name:       name,
		Expression: expression,
		Span:       Span{Start: start, End: end},
	}, nil
}

func (l *Lexer) readText() Token {
	start := l.position()
	begin := l.offset

	for !l.eof() {
		if l.hasPrefix(l.syntax.CommentOpen) || l.hasPrefix(l.syntax.RawOpen) || l.hasPrefix(l.syntax.EchoOpen) || (l.syntax.EscapeAt != "" && l.hasPrefix(l.syntax.EscapeAt)) || l.isDirectiveStart() {
			break
		}

		l.consumeRune()
	}

	end := l.position()

	return Token{
		Kind:    TokenKindText,
		Literal: l.source[begin:l.offset],
		Span:    Span{Start: start, End: end},
	}
}

func (l *Lexer) readIdentifier() string {
	begin := l.offset
	allowSegmentStart := false

	for !l.eof() {
		r := l.peek()
		if r == ':' {
			if l.offset == begin || allowSegmentStart {
				break
			}
			allowSegmentStart = true
			l.consumeRune()
			continue
		}
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_') {
			break
		}
		allowSegmentStart = false
		l.consumeRune()
	}

	value := l.source[begin:l.offset]
	if strings.HasSuffix(value, ":") {
		l.offset--
		l.column--
		return strings.TrimSuffix(value, ":")
	}

	return value
}

func (l *Lexer) readParenthesized() (string, error) {
	start := l.position()
	if l.peek() != '(' {
		return "", nil
	}

	l.consumeRune()
	begin := l.offset
	depth := 1
	var quote rune
	escaped := false

	for !l.eof() {
		r := l.peek()
		l.consumeRune()

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
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return l.source[begin : l.offset-1], nil
			}
		}
	}

	return "", &Error{
		Code:    ErrCodeUnexpectedEOF,
		Message: "指令表达式未正常闭合",
		Span:    Span{Start: start, End: l.position()},
	}
}

func (l *Lexer) skipInlineSpaces() {
	for !l.eof() {
		r := l.peek()
		if r != ' ' && r != '\t' {
			return
		}
		l.consumeRune()
	}
}

func (l *Lexer) isDirectiveStart() bool {
	if !l.hasPrefix("@") {
		return false
	}

	next := l.peekAt(1)
	return unicode.IsLetter(next) || next == '_'
}

func (l *Lexer) hasPrefix(prefix string) bool {
	return prefix != "" && strings.HasPrefix(l.source[l.offset:], prefix)
}

func (l *Lexer) consume(text string) {
	for len(text) > 0 {
		_, size := utf8.DecodeRuneInString(text)
		l.consumeRune()
		text = text[size:]
	}
}

func (l *Lexer) consumeRune() {
	r, size := utf8.DecodeRuneInString(l.source[l.offset:])
	l.offset += size
	if r == '\n' {
		l.line++
		l.column = 1
		return
	}
	l.column++
}

func (l *Lexer) position() Position {
	return Position{
		Offset: l.offset,
		Line:   l.line,
		Column: l.column,
	}
}

func (l *Lexer) peek() rune {
	return l.peekAt(0)
}

func (l *Lexer) peekAt(distance int) rune {
	if l.eof() {
		return 0
	}

	cursor := l.offset
	for i := 0; i < distance; i++ {
		if cursor >= len(l.source) {
			return 0
		}
		_, size := utf8.DecodeRuneInString(l.source[cursor:])
		cursor += size
	}

	if cursor >= len(l.source) {
		return 0
	}

	r, _ := utf8.DecodeRuneInString(l.source[cursor:])
	return r
}

func (l *Lexer) eof() bool {
	return l.offset >= len(l.source)
}
