package runtime

import (
	"fmt"
	"html/template"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/zane-xin/sparq-blade/internal/literal"
)

// Function 表示模板表达式可以调用的纯值转换函数。
type Function func(args ...any) (any, error)

var trustedTemplateContentTypes = map[reflect.Type]struct{}{
	reflect.TypeOf(template.CSS("")):      {},
	reflect.TypeOf(template.HTML("")):     {},
	reflect.TypeOf(template.HTMLAttr("")): {},
	reflect.TypeOf(template.JS("")):       {},
	reflect.TypeOf(template.JSStr("")):    {},
	reflect.TypeOf(template.Srcset("")):   {},
	reflect.TypeOf(template.URL("")):      {},
}

/*
ResolveExpression
根据当前作用域解析模板表达式。
当前版本支持字符串、数字、布尔值、nil、变量路径以及常用逻辑比较运算。
*/
func ResolveExpression(scope *Scope, expression string) (any, error) {
	tokens, err := tokenizeExpression(expression)
	if err != nil {
		return nil, err
	}

	parser := &expressionParser{
		scope:    scope,
		tokens:   tokens,
		evaluate: true,
	}

	value, err := parser.parseExpression()
	if err != nil {
		return nil, err
	}

	if parser.current().kind != exprTokenEOF {
		return nil, fmt.Errorf("表达式包含未消费的内容: %s", parser.current().literal)
	}

	return value, nil
}

/*
ResolveString
把表达式解析为字符串。
*/
func ResolveString(scope *Scope, expression string) (string, error) {
	value, err := ResolveExpression(scope, expression)
	if err != nil {
		return "", err
	}

	if value == nil {
		return "", nil
	}

	return fmt.Sprint(value), nil
}

/*
Truthy
判断一个值在模板条件中的真值语义。
*/
func Truthy(value any) bool {
	if value == nil {
		return false
	}

	switch item := value.(type) {
	case bool:
		return item
	case string:
		return item != ""
	case int:
		return item != 0
	case int8:
		return item != 0
	case int16:
		return item != 0
	case int32:
		return item != 0
	case int64:
		return item != 0
	case uint:
		return item != 0
	case uint8:
		return item != 0
	case uint16:
		return item != 0
	case uint32:
		return item != 0
	case uint64:
		return item != 0
	case float32:
		return item != 0
	case float64:
		return item != 0
	}

	reflected := reflect.ValueOf(value)
	for reflected.Kind() == reflect.Pointer || reflected.Kind() == reflect.Interface {
		if reflected.IsNil() {
			return false
		}
		reflected = reflected.Elem()
	}

	switch reflected.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return reflected.Len() > 0
	}

	return true
}

type exprTokenKind string

const (
	exprTokenEOF        exprTokenKind = "eof"
	exprTokenIdentifier exprTokenKind = "identifier"
	exprTokenString     exprTokenKind = "string"
	exprTokenNumber     exprTokenKind = "number"
	exprTokenOperator   exprTokenKind = "operator"
	exprTokenLParen     exprTokenKind = "lparen"
	exprTokenComma      exprTokenKind = "comma"
	exprTokenRParen     exprTokenKind = "rparen"
)

type exprToken struct {
	kind    exprTokenKind
	literal string
}

type expressionParser struct {
	scope    *Scope
	tokens   []exprToken
	index    int
	evaluate bool
}

type includeSpec struct {
	ViewExpression string
	DataExpression string
}

type componentSpec struct {
	ViewExpression string
	DataExpression string
}

type tagSpec struct {
	TargetExpression string
	Params           map[string]string
}

func (p *expressionParser) parseExpression() (any, error) {
	return p.parseOr()
}

func (p *expressionParser) parseWithoutEvaluation(parse func() (any, error)) (any, error) {
	evaluate := p.evaluate
	p.evaluate = false
	value, err := parse()
	p.evaluate = evaluate
	return value, err
}

func (p *expressionParser) parseOr() (any, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.matchOperator("||") {
		if p.evaluate && Truthy(left) {
			if _, err := p.parseWithoutEvaluation(p.parseAnd); err != nil {
				return nil, err
			}
			left = true
			continue
		}

		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		if p.evaluate {
			left = Truthy(left) || Truthy(right)
		}
	}

	return left, nil
}

func (p *expressionParser) parseAnd() (any, error) {
	left, err := p.parseEquality()
	if err != nil {
		return nil, err
	}

	for p.matchOperator("&&") {
		if p.evaluate && !Truthy(left) {
			if _, err := p.parseWithoutEvaluation(p.parseEquality); err != nil {
				return nil, err
			}
			left = false
			continue
		}

		right, err := p.parseEquality()
		if err != nil {
			return nil, err
		}
		if p.evaluate {
			left = Truthy(left) && Truthy(right)
		}
	}

	return left, nil
}

func (p *expressionParser) parseEquality() (any, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	for {
		switch {
		case p.matchOperator("=="):
			right, err := p.parseComparison()
			if err != nil {
				return nil, err
			}
			if p.evaluate {
				left = equalValues(left, right)
			}
		case p.matchOperator("!="):
			right, err := p.parseComparison()
			if err != nil {
				return nil, err
			}
			if p.evaluate {
				left = !equalValues(left, right)
			}
		default:
			return left, nil
		}
	}
}

func (p *expressionParser) parseComparison() (any, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for {
		var operator string
		switch {
		case p.matchOperator(">"):
			operator = ">"
		case p.matchOperator(">="):
			operator = ">="
		case p.matchOperator("<"):
			operator = "<"
		case p.matchOperator("<="):
			operator = "<="
		default:
			return left, nil
		}

		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		if !p.evaluate {
			continue
		}

		left, err = compareValues(left, right, operator)
		if err != nil {
			return nil, err
		}
	}
}

func (p *expressionParser) parseUnary() (any, error) {
	if p.matchOperator("!") {
		value, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		if !p.evaluate {
			return nil, nil
		}
		return !Truthy(value), nil
	}

	return p.parsePrimary()
}

func (p *expressionParser) parsePrimary() (any, error) {
	token := p.current()

	switch token.kind {
	case exprTokenLParen:
		p.advance()
		value, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if p.current().kind != exprTokenRParen {
			return nil, fmt.Errorf("表达式缺少右括号")
		}
		p.advance()
		return value, nil
	case exprTokenString:
		p.advance()
		value, _ := literal.ParseString(token.literal)
		return value, nil
	case exprTokenNumber:
		p.advance()
		if strings.Contains(token.literal, ".") {
			return strconv.ParseFloat(token.literal, 64)
		}
		return strconv.ParseInt(token.literal, 10, 64)
	case exprTokenIdentifier:
		p.advance()
		if p.current().kind == exprTokenLParen {
			return p.callFunction(token.literal)
		}
		return p.resolveIdentifier(token.literal)
	default:
		return nil, fmt.Errorf("无法解析表达式片段: %s", token.literal)
	}
}

func (p *expressionParser) callFunction(name string) (any, error) {
	p.advance()
	args := make([]any, 0, 2)
	if p.current().kind != exprTokenRParen {
		for {
			value, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			args = append(args, value)
			if p.current().kind != exprTokenComma {
				break
			}
			p.advance()
		}
	}
	if p.current().kind != exprTokenRParen {
		return nil, fmt.Errorf("模板函数 %s 缺少右括号", name)
	}
	p.advance()

	if !p.evaluate {
		return nil, nil
	}

	function, ok := p.scope.Functions[name]
	if !ok {
		return nil, fmt.Errorf("未注册模板函数: %s", name)
	}
	value, err := function(args...)
	if err != nil {
		return nil, fmt.Errorf("调用模板函数 %s 失败: %w", name, err)
	}
	if isTrustedTemplateContent(value) {
		return nil, fmt.Errorf("模板函数 %s 不允许返回可信模板内容", name)
	}
	return value, nil
}

func isTrustedTemplateContent(value any) bool {
	// html/template 会持续解引用接口和指针；这里采用相同规则，避免插件通过
	// *template.HTML 等包装类型绕过可信内容返回值限制。
	reflected := reflect.ValueOf(value)
	for reflected.IsValid() && (reflected.Kind() == reflect.Pointer || reflected.Kind() == reflect.Interface) {
		if reflected.IsNil() {
			return false
		}
		reflected = reflected.Elem()
	}
	if !reflected.IsValid() {
		return false
	}
	_, trusted := trustedTemplateContentTypes[reflected.Type()]
	return trusted
}

func (p *expressionParser) resolveIdentifier(literal string) (any, error) {
	if !p.evaluate {
		return nil, nil
	}

	switch literal {
	case "true":
		return true, nil
	case "false":
		return false, nil
	case "nil", "null":
		return nil, nil
	}

	return resolveVariable(p.scope, literal)
}

func (p *expressionParser) current() exprToken {
	if p.index >= len(p.tokens) {
		return exprToken{kind: exprTokenEOF}
	}
	return p.tokens[p.index]
}

func (p *expressionParser) advance() {
	if p.index < len(p.tokens) {
		p.index++
	}
}

func (p *expressionParser) matchOperator(operator string) bool {
	token := p.current()
	if token.kind != exprTokenOperator || token.literal != operator {
		return false
	}

	p.advance()
	return true
}

func tokenizeExpression(expression string) ([]exprToken, error) {
	tokens := make([]exprToken, 0, 8)

	for len(expression) > 0 {
		r, size := utf8.DecodeRuneInString(expression)
		switch {
		case unicode.IsSpace(r):
			expression = expression[size:]
		case r == '(':
			tokens = append(tokens, exprToken{kind: exprTokenLParen, literal: "("})
			expression = expression[size:]
		case r == ')':
			tokens = append(tokens, exprToken{kind: exprTokenRParen, literal: ")"})
			expression = expression[size:]
		case isOperatorPrefix(expression):
			operator, width := readOperator(expression)
			tokens = append(tokens, exprToken{kind: exprTokenOperator, literal: operator})
			expression = expression[width:]
		case r == ',':
			tokens = append(tokens, exprToken{kind: exprTokenComma, literal: ","})
			expression = expression[size:]
		case r == '\'' || r == '"':
			literal, width, err := readQuotedToken(expression)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, exprToken{kind: exprTokenString, literal: literal})
			expression = expression[width:]
		case r == '$':
			return nil, fmt.Errorf("模板变量不再支持 $ 前缀，请直接使用变量名")
		case unicode.IsLetter(r) || r == '_' || unicode.IsDigit(r):
			literal, kind, width := readScalarToken(expression)
			tokens = append(tokens, exprToken{kind: kind, literal: literal})
			expression = expression[width:]
		default:
			return nil, fmt.Errorf("不支持的表达式字符: %q", r)
		}
	}

	tokens = append(tokens, exprToken{kind: exprTokenEOF})
	return tokens, nil
}

func isOperatorPrefix(expression string) bool {
	for _, item := range []string{"&&", "||", "==", "!=", ">=", "<=", "!", ">", "<"} {
		if strings.HasPrefix(expression, item) {
			return true
		}
	}
	return false
}

func readOperator(expression string) (string, int) {
	for _, item := range []string{"&&", "||", "==", "!=", ">=", "<=", "!", ">", "<"} {
		if strings.HasPrefix(expression, item) {
			return item, len(item)
		}
	}
	return "", 0
}

func readQuotedToken(expression string) (string, int, error) {
	quote, size := utf8.DecodeRuneInString(expression)
	offset := size
	escaped := false

	for offset < len(expression) {
		r, width := utf8.DecodeRuneInString(expression[offset:])
		offset += width

		if escaped {
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if r == quote {
			return expression[:offset], offset, nil
		}
	}

	return "", 0, fmt.Errorf("字符串字面量未正常闭合")
}

func readScalarToken(expression string) (string, exprTokenKind, int) {
	offset := 0
	hasDot := false

	for offset < len(expression) {
		r, width := utf8.DecodeRuneInString(expression[offset:])
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			offset += width
			continue
		}

		if r == '.' {
			hasDot = true
			offset += width
			continue
		}

		break
	}

	literal := expression[:offset]
	if isNumericLiteral(literal, hasDot) {
		return literal, exprTokenNumber, offset
	}

	return literal, exprTokenIdentifier, offset
}

func isNumericLiteral(value string, hasDot bool) bool {
	if value == "" {
		return false
	}

	dotCount := 0
	for _, item := range value {
		if item == '.' {
			dotCount++
			continue
		}
		if !unicode.IsDigit(item) {
			return false
		}
	}

	if hasDot {
		return dotCount == 1
	}

	return dotCount == 0
}

type foreachSpec struct {
	Collection string
	KeyAlias   string
	ItemAlias  string
}

type listSpec struct {
	Collection string
	ItemAlias  string
}

type childrenSpec struct {
	Field     string
	ItemAlias string
}

type eachSpec struct {
	View       string
	Collection string
	ItemAlias  string
	EmptyView  string
}

type forSpec struct {
	Alias string
	Start string
	End   string
}

func parseForeach(expression string) (*foreachSpec, error) {
	parts := strings.SplitN(expression, " as ", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("foreach 表达式格式非法: %s", expression)
	}

	spec := &foreachSpec{
		Collection: strings.TrimSpace(parts[0]),
	}

	aliases := strings.SplitN(strings.TrimSpace(parts[1]), "=>", 2)
	if len(aliases) == 2 {
		keyAlias, err := normalizeAlias(aliases[0])
		if err != nil {
			return nil, err
		}
		itemAlias, err := normalizeAlias(aliases[1])
		if err != nil {
			return nil, err
		}
		spec.KeyAlias = keyAlias
		spec.ItemAlias = itemAlias
	} else {
		itemAlias, err := normalizeAlias(parts[1])
		if err != nil {
			return nil, err
		}
		spec.ItemAlias = itemAlias
	}

	if spec.ItemAlias == "" {
		return nil, fmt.Errorf("foreach 缺少迭代变量: %s", expression)
	}

	return spec, nil
}

func parseList(expression string) (*listSpec, error) {
	parts, err := splitTopLevel(expression, ',')
	if err != nil {
		return nil, err
	}
	if len(parts) != 2 {
		return nil, fmt.Errorf("list 表达式格式非法: %s", expression)
	}

	spec := &listSpec{
		Collection: strings.TrimSpace(parts[0]),
	}
	if spec.Collection == "" {
		return nil, fmt.Errorf("list 缺少数据表达式: %s", expression)
	}

	name, value, ok := splitTopLevelFirst(strings.TrimSpace(parts[1]), '=')
	if !ok || strings.TrimSpace(name) != "as" {
		return nil, fmt.Errorf("list 必须使用 as=item 指定列表项变量: %s", expression)
	}

	alias, err := normalizeAlias(value)
	if err != nil {
		return nil, err
	}
	if alias == "" {
		return nil, fmt.Errorf("list 缺少列表项变量: %s", expression)
	}
	spec.ItemAlias = alias

	return spec, nil
}

func parseChildren(expression string) (*childrenSpec, error) {
	spec := &childrenSpec{
		Field:     "children",
		ItemAlias: "child",
	}

	parts, err := splitTopLevel(expression, ',')
	if err != nil {
		return nil, err
	}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		name, value, ok := splitTopLevelFirst(part, '=')
		if !ok {
			return nil, fmt.Errorf("children 参数必须使用 key=value 形式: %s", part)
		}

		key := strings.TrimSpace(name)
		value = strings.TrimSpace(value)
		switch key {
		case "as":
			alias, err := normalizeAlias(value)
			if err != nil {
				return nil, err
			}
			if alias == "" {
				return nil, fmt.Errorf("children as 不能为空")
			}
			spec.ItemAlias = alias
		case "field":
			field, ok := literal.ParseString(value)
			if !ok {
				field = value
			}
			field = strings.TrimSpace(field)
			if field == "" {
				return nil, fmt.Errorf("children field 不能为空")
			}
			if strings.Contains(field, ".") {
				return nil, fmt.Errorf("children field 仅支持当前项的直接字段: %s", field)
			}
			spec.Field = field
		case "wrap", "class":
			continue
		default:
			return nil, fmt.Errorf("children 不支持参数: %s", key)
		}
	}

	return spec, nil
}

func parseEach(expression string) (*eachSpec, error) {
	parts, err := splitTopLevel(expression, ',')
	if err != nil {
		return nil, err
	}
	if len(parts) < 3 || len(parts) > 4 {
		return nil, fmt.Errorf("each 表达式格式非法: %s", expression)
	}

	view, ok := literal.ParseString(strings.TrimSpace(parts[0]))
	if !ok {
		return nil, fmt.Errorf("each 视图名必须是字符串字面量: %s", parts[0])
	}

	aliasPart := strings.TrimSpace(parts[2])
	alias, ok := literal.ParseString(aliasPart)
	if !ok {
		alias, err = normalizeAlias(aliasPart)
		if err != nil {
			return nil, err
		}
	}
	if alias == "" {
		return nil, fmt.Errorf("each 缺少迭代变量名: %s", expression)
	}

	spec := &eachSpec{
		View:       view,
		Collection: strings.TrimSpace(parts[1]),
		ItemAlias:  alias,
	}

	if len(parts) == 4 {
		emptyView, ok := literal.ParseString(strings.TrimSpace(parts[3]))
		if !ok {
			return nil, fmt.Errorf("each empty 视图名必须是字符串字面量: %s", parts[3])
		}
		spec.EmptyView = emptyView
	}

	return spec, nil
}

func parseInclude(expression string) (*includeSpec, error) {
	parts, err := splitTopLevel(expression, ',')
	if err != nil {
		return nil, err
	}
	if len(parts) < 1 || len(parts) > 2 {
		return nil, fmt.Errorf("include 表达式格式非法: %s", expression)
	}

	spec := &includeSpec{
		ViewExpression: strings.TrimSpace(parts[0]),
	}
	if spec.ViewExpression == "" {
		return nil, fmt.Errorf("include 缺少视图名: %s", expression)
	}

	if len(parts) == 2 {
		spec.DataExpression = strings.TrimSpace(parts[1])
		if spec.DataExpression == "" {
			return nil, fmt.Errorf("include 数据表达式不能为空: %s", expression)
		}
	}

	return spec, nil
}

func parseComponent(expression string) (*componentSpec, error) {
	parts, err := splitTopLevel(expression, ',')
	if err != nil {
		return nil, err
	}
	if len(parts) < 1 || len(parts) > 2 {
		return nil, fmt.Errorf("component 表达式格式非法: %s", expression)
	}

	spec := &componentSpec{
		ViewExpression: strings.TrimSpace(parts[0]),
	}
	if spec.ViewExpression == "" {
		return nil, fmt.Errorf("component 缺少视图名: %s", expression)
	}

	if len(parts) == 2 {
		spec.DataExpression = strings.TrimSpace(parts[1])
		if spec.DataExpression == "" {
			return nil, fmt.Errorf("component 数据表达式不能为空: %s", expression)
		}
	}

	return spec, nil
}

func parseTag(expression string) (*tagSpec, error) {
	parts, err := splitTopLevel(expression, ',')
	if err != nil {
		return nil, err
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("tag 表达式不能为空")
	}

	targetExpression := strings.TrimSpace(parts[0])
	if targetExpression == "" {
		return nil, fmt.Errorf("tag 缺少目标标识: %s", expression)
	}
	if _, ok := literal.ParseString(targetExpression); !ok {
		return nil, fmt.Errorf("tag 目标必须是字符串字面量: %s", targetExpression)
	}

	spec := &tagSpec{
		TargetExpression: targetExpression,
		Params:           make(map[string]string, max(0, len(parts)-1)),
	}

	params, err := parseTagParams(parts[1:], expression)
	if err != nil {
		return nil, err
	}
	spec.Params = params

	return spec, nil
}

func parseFlexibleTag(expression string) (*tagSpec, error) {
	parts, err := splitTopLevel(expression, ',')
	if err != nil {
		return nil, err
	}
	if len(parts) == 0 {
		return &tagSpec{Params: make(map[string]string)}, nil
	}

	first := strings.TrimSpace(parts[0])
	if first == "" {
		return nil, fmt.Errorf("tag 参数不能为空: %s", expression)
	}

	offset := 0
	targetExpression := ""
	if _, ok := literal.ParseString(first); ok {
		targetExpression = first
		offset = 1
	} else if _, _, ok := splitTopLevelFirst(first, '='); !ok && first != "out" {
		return nil, fmt.Errorf("tag 目标必须是字符串字面量: %s", first)
	}

	params, err := parseTagParams(parts[offset:], expression)
	if err != nil {
		return nil, err
	}

	return &tagSpec{
		TargetExpression: targetExpression,
		Params:           params,
	}, nil
}

func parseNamedTagParams(expression string) (map[string]string, error) {
	parts, err := splitTopLevel(expression, ',')
	if err != nil {
		return nil, err
	}

	return parseTagParams(parts, expression)
}

func parseTagParams(parts []string, expression string) (map[string]string, error) {
	params := make(map[string]string, max(0, len(parts)))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("tag 参数不能为空: %s", expression)
		}

		name, value, ok := splitTopLevelFirst(part, '=')
		if !ok {
			if part == "out" {
				params["out"] = "true"
				continue
			}
			return nil, fmt.Errorf("tag 参数必须使用 key=value 形式，或使用受控裸参数 out: %s", part)
		}

		key := strings.TrimSpace(name)
		if !isTagParamName(key) {
			return nil, fmt.Errorf("tag 参数名非法: %s", key)
		}

		value = strings.TrimSpace(value)
		if value == "" {
			return nil, fmt.Errorf("tag 参数值不能为空: %s", key)
		}

		params[key] = value
	}

	return params, nil
}

func parseFor(expression string) (*forSpec, error) {
	name, rangeExpression, ok := splitKeywordIn(expression)
	if !ok {
		return nil, fmt.Errorf("for 表达式格式非法，请使用 i in 1..5: %s", expression)
	}

	alias, err := normalizeAlias(name)
	if err != nil {
		return nil, err
	}
	if alias == "" {
		return nil, fmt.Errorf("for 缺少循环变量: %s", expression)
	}

	start, end, ok := splitTopLevelRange(rangeExpression)
	if !ok {
		return nil, fmt.Errorf("for 范围格式非法，请使用 1..5: %s", expression)
	}

	return &forSpec{
		Alias: alias,
		Start: strings.TrimSpace(start),
		End:   strings.TrimSpace(end),
	}, nil
}

func normalizeAlias(alias string) (string, error) {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return "", nil
	}
	if strings.HasPrefix(alias, "$") {
		return "", fmt.Errorf("变量别名不再支持 $ 前缀: %s", alias)
	}
	if !isBareIdentifier(alias) {
		return "", fmt.Errorf("变量别名非法: %s", alias)
	}

	return alias, nil
}

func splitTopLevel(expression string, delimiter rune) ([]string, error) {
	parts := make([]string, 0, 4)
	start := 0
	depth := 0
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
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		default:
			if item == delimiter && depth == 0 {
				parts = append(parts, expression[start:index])
				start = index + 1
			}
		}
	}

	if quote != 0 || depth != 0 {
		return nil, fmt.Errorf("表达式结构未正常闭合: %s", expression)
	}

	parts = append(parts, expression[start:])
	return parts, nil
}

func splitTopLevelFirst(expression string, delimiter rune) (string, string, bool) {
	depth := 0
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
		case '(':
			depth++
		case ')':
			depth--
		default:
			if item == delimiter && depth == 0 {
				return expression[:index], expression[index+1:], true
			}
		}
	}

	return "", "", false
}

func splitKeywordIn(expression string) (string, string, bool) {
	depth := 0
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
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		default:
			if item == 'i' && depth == 0 && strings.HasPrefix(expression[index:], "in") {
				beforeOK := index == 0 || unicode.IsSpace(rune(expression[index-1]))
				after := index + len("in")
				afterOK := after >= len(expression) || unicode.IsSpace(rune(expression[after]))
				if beforeOK && afterOK {
					return strings.TrimSpace(expression[:index]), strings.TrimSpace(expression[after:]), true
				}
			}
		}
	}

	return "", "", false
}

func splitTopLevelRange(expression string) (string, string, bool) {
	depth := 0
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
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case '.':
			if depth == 0 && strings.HasPrefix(expression[index:], "..") {
				left := strings.TrimSpace(expression[:index])
				right := strings.TrimSpace(expression[index+2:])
				return left, right, left != "" && right != ""
			}
		}
	}

	return "", "", false
}

func isTagParamName(name string) bool {
	if name == "" {
		return false
	}

	for index, item := range name {
		if index == 0 {
			if !(item == '_' || unicode.IsLetter(item)) {
				return false
			}
			continue
		}

		if !(item == '_' || unicode.IsLetter(item) || unicode.IsDigit(item)) {
			return false
		}
	}

	return true
}

func isBareIdentifier(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(value, "$") {
		return false
	}

	for index, item := range value {
		if index == 0 {
			if !(item == '_' || unicode.IsLetter(item)) {
				return false
			}
			continue
		}

		if !(item == '_' || unicode.IsLetter(item) || unicode.IsDigit(item)) {
			return false
		}
	}

	switch value {
	case "true", "false", "nil", "null":
		return false
	default:
		return true
	}
}

func resolveVariable(scope *Scope, expression string) (any, error) {
	segments := splitPath(expression)
	if len(segments) == 0 {
		return nil, nil
	}

	baseName := segments[0]
	base, ok := scope.Vars[baseName]
	if !ok {
		base = scope.Root
		segments = splitPath(expression)
	} else {
		segments = segments[1:]
	}

	return resolvePath(base, segments)
}

func splitPath(expression string) []string {
	raw := strings.Split(expression, ".")
	segments := make([]string, 0, len(raw))
	for _, item := range raw {
		item = strings.TrimSpace(item)
		if item != "" {
			segments = append(segments, item)
		}
	}
	return segments
}

func resolvePath(value any, segments []string) (any, error) {
	current := value
	var err error

	for _, segment := range segments {
		current, err = lookupSegment(current, segment)
		if err != nil {
			return nil, err
		}
	}

	return current, nil
}

func lookupSegment(value any, segment string) (any, error) {
	if value == nil {
		return nil, nil
	}

	reflected := reflect.ValueOf(value)
	for reflected.Kind() == reflect.Pointer || reflected.Kind() == reflect.Interface {
		if reflected.IsNil() {
			return nil, nil
		}
		reflected = reflected.Elem()
	}

	switch reflected.Kind() {
	case reflect.Map:
		iter := reflected.MapRange()
		for iter.Next() {
			key := iter.Key()
			if key.Kind() == reflect.String && key.String() == segment {
				return iter.Value().Interface(), nil
			}
		}
		return nil, nil
	case reflect.Struct:
		field := reflected.FieldByName(segment)
		if field.IsValid() && field.CanInterface() {
			return field.Interface(), nil
		}

		typeInfo := reflected.Type()
		for index := 0; index < typeInfo.NumField(); index++ {
			item := typeInfo.Field(index)
			if strings.EqualFold(item.Name, segment) {
				field = reflected.Field(index)
				if field.IsValid() && field.CanInterface() {
					return field.Interface(), nil
				}
			}
		}

		return nil, nil
	case reflect.Slice, reflect.Array:
		if isDigits(segment) {
			index, _ := strconv.Atoi(segment)
			if index >= 0 && index < reflected.Len() {
				return reflected.Index(index).Interface(), nil
			}
			return nil, nil
		}
	}

	return nil, fmt.Errorf("无法从 %T 上读取字段 %s", value, segment)
}

func isDigits(value string) bool {
	if value == "" {
		return false
	}

	for _, item := range value {
		if !unicode.IsDigit(item) {
			return false
		}
	}

	return true
}

type integerValue struct {
	isSigned      bool
	signedValue   int64
	unsignedValue uint64
}

func toInteger(value any) (integerValue, bool) {
	switch item := value.(type) {
	case int:
		return integerValue{isSigned: true, signedValue: int64(item)}, true
	case int8:
		return integerValue{isSigned: true, signedValue: int64(item)}, true
	case int16:
		return integerValue{isSigned: true, signedValue: int64(item)}, true
	case int32:
		return integerValue{isSigned: true, signedValue: int64(item)}, true
	case int64:
		return integerValue{isSigned: true, signedValue: item}, true
	case uint:
		return integerValue{unsignedValue: uint64(item)}, true
	case uint8:
		return integerValue{unsignedValue: uint64(item)}, true
	case uint16:
		return integerValue{unsignedValue: uint64(item)}, true
	case uint32:
		return integerValue{unsignedValue: uint64(item)}, true
	case uint64:
		return integerValue{unsignedValue: item}, true
	default:
		return integerValue{}, false
	}
}

func compareIntegers(left, right integerValue) int {
	if left.isSigned && right.isSigned {
		return compareInt64(left.signedValue, right.signedValue)
	}
	if !left.isSigned && !right.isSigned {
		return compareUint64(left.unsignedValue, right.unsignedValue)
	}
	if left.isSigned {
		if left.signedValue < 0 {
			return -1
		}
		return compareUint64(uint64(left.signedValue), right.unsignedValue)
	}
	if right.signedValue < 0 {
		return 1
	}
	return compareUint64(left.unsignedValue, uint64(right.signedValue))
}

func compareInt64(left, right int64) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}

func compareUint64(left, right uint64) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}

func equalValues(left, right any) bool {
	if left == nil || right == nil {
		return left == right
	}

	if li, lok := toInteger(left); lok {
		if ri, rok := toInteger(right); rok {
			return compareIntegers(li, ri) == 0
		}
	}
	if lf, lok := toFloat64(left); lok {
		if rf, rok := toFloat64(right); rok {
			return lf == rf
		}
	}

	return reflect.DeepEqual(left, right)
}

func compareValues(left, right any, operator string) (bool, error) {
	if li, lok := toInteger(left); lok {
		if ri, rok := toInteger(right); rok {
			return compareOrder(compareIntegers(li, ri), operator)
		}
	}
	if lf, lok := toFloat64(left); lok {
		if rf, rok := toFloat64(right); rok {
			switch operator {
			case ">":
				return lf > rf, nil
			case ">=":
				return lf >= rf, nil
			case "<":
				return lf < rf, nil
			case "<=":
				return lf <= rf, nil
			}
		}
	}

	return compareOrder(strings.Compare(fmt.Sprint(left), fmt.Sprint(right)), operator)
}

func compareOrder(order int, operator string) (bool, error) {
	switch operator {
	case ">":
		return order > 0, nil
	case ">=":
		return order >= 0, nil
	case "<":
		return order < 0, nil
	case "<=":
		return order <= 0, nil
	default:
		return false, fmt.Errorf("不支持的比较运算符: %s", operator)
	}
}

func toFloat64(value any) (float64, bool) {
	switch item := value.(type) {
	case int:
		return float64(item), true
	case int8:
		return float64(item), true
	case int16:
		return float64(item), true
	case int32:
		return float64(item), true
	case int64:
		return float64(item), true
	case uint:
		return float64(item), true
	case uint8:
		return float64(item), true
	case uint16:
		return float64(item), true
	case uint32:
		return float64(item), true
	case uint64:
		return float64(item), true
	case float32:
		return float64(item), true
	case float64:
		return item, true
	}

	return 0, false
}

func toInt64(value any) (int64, bool) {
	switch item := value.(type) {
	case int:
		return int64(item), true
	case int8:
		return int64(item), true
	case int16:
		return int64(item), true
	case int32:
		return int64(item), true
	case int64:
		return item, true
	case uint:
		if uint64(item) > uint64(^uint64(0)>>1) {
			return 0, false
		}
		return int64(item), true
	case uint8:
		return int64(item), true
	case uint16:
		return int64(item), true
	case uint32:
		return int64(item), true
	case uint64:
		if item > uint64(^uint64(0)>>1) {
			return 0, false
		}
		return int64(item), true
	}

	return 0, false
}
