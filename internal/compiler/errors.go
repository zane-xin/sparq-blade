package compiler

import (
	"errors"
	"fmt"
	"strings"
)

// Position 表示模板源码中的一个位置。
// Offset 为字节偏移量，Line 和 Column 为从 1 开始的行列号。
type Position struct {
	Offset int
	Line   int
	Column int
}

// Span 表示模板源码中的一个连续范围。
type Span struct {
	Start Position
	End   Position
}

// ErrorCode 表示编译阶段使用的错误码。
type ErrorCode string

const (
	/*
		ErrCodeUnexpectedEOF
		表示源码在尚未闭合的结构中提前结束。
	*/
	ErrCodeUnexpectedEOF ErrorCode = "unexpected_eof"
	/*
		ErrCodeUnexpectedToken
		表示解析过程中遇到了不符合预期的 token。
	*/
	ErrCodeUnexpectedToken ErrorCode = "unexpected_token"
	/*
		ErrCodeUnsupportedDirective
		表示当前版本尚未支持某个指令。
	*/
	ErrCodeUnsupportedDirective ErrorCode = "unsupported_directive"
	/*
		ErrCodeInvalidDirective
		表示指令本身格式非法或参数缺失。
	*/
	ErrCodeInvalidDirective ErrorCode = "invalid_directive"
)

// Error 表示模板编译阶段的结构化错误。
type Error struct {
	Code       ErrorCode
	Message    string
	Span       Span
	SourceName string
	SourceLine string
	Pointer    string
}

// String 返回位置的行列表示。
func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// String 返回范围的起止位置表示。
func (s Span) String() string {
	return fmt.Sprintf("%s-%s", s.Start.String(), s.End.String())
}

// Error 返回结构化错误的文本表示。
func (e *Error) Error() string {
	message := fmt.Sprintf("%s at %s: %s", e.Code, e.Span.Start.String(), e.Message)
	if e.SourceLine == "" {
		return message
	}

	var builder strings.Builder
	builder.WriteString(message)
	builder.WriteByte('\n')
	if e.SourceName != "" {
		builder.WriteString("--> ")
		builder.WriteString(e.SourceName)
		builder.WriteByte(':')
		builder.WriteString(e.Span.Start.String())
		builder.WriteByte('\n')
	}
	builder.WriteString(e.SourceLine)
	builder.WriteByte('\n')
	builder.WriteString(e.Pointer)

	return builder.String()
}

// WithSource 为编译错误附加源码上下文。
func WithSource(err error, source string) error {
	return WithNamedSource(err, "", source)
}

// WithNamedSource 为编译错误附加带名称的源码上下文。
func WithNamedSource(err error, name string, source string) error {
	if err == nil {
		return nil
	}

	var compileErr *Error
	if !errors.As(err, &compileErr) {
		return err
	}
	if compileErr.SourceLine != "" {
		return err
	}

	enriched := *compileErr
	enriched.SourceName = name
	enriched.SourceLine, enriched.Pointer = sourceContext(source, compileErr.Span.Start)

	return &enriched
}

func sourceContext(source string, position Position) (string, string) {
	if position.Line <= 0 {
		return "", ""
	}

	lines := strings.Split(source, "\n")
	if position.Line > len(lines) {
		return "", ""
	}

	line := lines[position.Line-1]
	column := position.Column
	if column < 1 {
		column = 1
	}

	runeCount := len([]rune(line))
	if column > runeCount+1 {
		column = runeCount + 1
	}

	return line, strings.Repeat(" ", column-1) + "^"
}
