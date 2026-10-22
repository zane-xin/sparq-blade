package compiler

// TokenKind 表示词法分析产出的 token 类型。
type TokenKind string

const (
	/*
		TokenKindText
		表示普通文本片段。
	*/
	TokenKindText TokenKind = "text"
	/*
		TokenKindEchoEscaped
		表示会自动转义的输出片段。
	*/
	TokenKindEchoEscaped TokenKind = "echo_escaped"
	/*
		TokenKindEchoRaw
		表示原样输出片段。
	*/
	TokenKindEchoRaw TokenKind = "echo_raw"
	/*
		TokenKindComment
		表示模板注释片段。
	*/
	TokenKindComment TokenKind = "comment"
	/*
		TokenKindDirective
		表示 Blade 指令片段。
	*/
	TokenKindDirective TokenKind = "directive"
	/*
		TokenKindEOF
		表示输入结束。
	*/
	TokenKindEOF TokenKind = "eof"
)

// Token 表示词法分析阶段的一个基础单元。
// Name 和 Expression 主要供指令 token 使用。
type Token struct {
	Kind       TokenKind
	Literal    string
	Name       string
	Expression string
	Span       Span
}
