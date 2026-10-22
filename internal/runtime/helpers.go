package runtime

import (
	"fmt"
	"html/template"
	"reflect"
)

// MaxLoopIterations 是单个模板循环允许处理的最大项目数。
const MaxLoopIterations = 10000
const currentItemVar = "__blade_current"

// SetCurrentItem 记录当前迭代项，供嵌套的 @children 指令读取子级字段；该内部变量不作为业务数据公开。
func SetCurrentItem(scope *Scope, item any) {
	scope.SetVar(currentItemVar, item)
}

/*
String
把任意值转换为字符串表示。
*/
func String(value any) string {
	return fmt.Sprint(value)
}

/*
Echo
解析表达式并返回普通输出值。
返回值交给 html/template 自动转义。
*/
func Echo(scope *Scope, expression string) (any, error) {
	return ResolveExpression(scope, expression)
}

/*
Raw
解析表达式并返回原样 HTML 输出。该函数不会清洗内容，并会绕过 html/template
的上下文转义，因此只允许用于应用明确判定为可信的 HTML。
*/
func Raw(scope *Scope, expression string) (template.HTML, error) {
	value, err := ResolveExpression(scope, expression)
	if err != nil {
		return "", err
	}

	if html, ok := value.(template.HTML); ok {
		return html, nil
	}

	return template.HTML(fmt.Sprint(value)), nil
}

/*
Check
判断表达式在当前作用域中的真值。
*/
func Check(scope *Scope, expression string) (bool, error) {
	value, err := ResolveExpression(scope, expression)
	if err != nil {
		return false, err
	}

	return Truthy(value), nil
}

/*
Unless
判断表达式在当前作用域中的非真值。
*/
func Unless(scope *Scope, expression string) (bool, error) {
	result, err := Check(scope, expression)
	if err != nil {
		return false, err
	}

	return !result, nil
}

/*
Range
把 foreach 表达式展开为一组子作用域。展开前会校验集合类型，并限制最多生成
MaxLoopIterations 个作用域，避免模板数据导致无界内存占用。
*/
func Range(scope *Scope, expression string) ([]*Scope, error) {
	spec, err := parseForeach(expression)
	if err != nil {
		return nil, err
	}

	value, err := ResolveExpression(scope, spec.Collection)
	if err != nil {
		return nil, err
	}

	return rangeValue(scope, value, spec.KeyAlias, spec.ItemAlias)
}

func rangeCollection(scope *Scope, collection string, keyAlias string, itemAlias string) ([]*Scope, error) {
	value, err := ResolveExpression(scope, collection)
	if err != nil {
		return nil, err
	}

	return rangeValue(scope, value, keyAlias, itemAlias)
}

func rangeValue(scope *Scope, value any, keyAlias string, itemAlias string) ([]*Scope, error) {
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
	case reflect.Array, reflect.Slice:
		total := reflected.Len()
		if total > MaxLoopIterations {
			return nil, fmt.Errorf("集合循环超过最大迭代次数 %d", MaxLoopIterations)
		}
		results := make([]*Scope, 0, total)
		for index := 0; index < total; index++ {
			item := reflected.Index(index).Interface()
			child := scope.WithVar(itemAlias, item)
			if keyAlias != "" {
				child = child.WithVar(keyAlias, index)
			}
			child.SetVar(currentItemVar, item)
			child.SetVar("loop", LoopState(index, total))
			results = append(results, child)
		}
		return results, nil
	case reflect.Map:
		total := reflected.Len()
		if total > MaxLoopIterations {
			return nil, fmt.Errorf("集合循环超过最大迭代次数 %d", MaxLoopIterations)
		}
		results := make([]*Scope, 0, total)
		index := 0
		iter := reflected.MapRange()
		for iter.Next() {
			item := iter.Value().Interface()
			child := scope.WithVar(itemAlias, item)
			if keyAlias != "" {
				child = child.WithVar(keyAlias, iter.Key().Interface())
			}
			child.SetVar(currentItemVar, item)
			child.SetVar("loop", LoopState(index, total))
			results = append(results, child)
			index++
		}
		return results, nil
	default:
		return nil, fmt.Errorf("foreach 不支持遍历 %T", value)
	}
}

// LoopState 构造循环体内可读取的 loop 元数据，index 从 0 开始，iteration 从 1 开始。
func LoopState(index int, total int) map[string]any {
	iteration := index + 1
	return map[string]any{
		"index":     index,
		"iteration": iteration,
		"count":     total,
		"remaining": total - iteration,
		"remainini": total - iteration,
		"first":     index == 0,
		"last":      index == total-1,
		"odd":       iteration%2 == 1,
		"even":      iteration%2 == 0,
	}
}

/*
ListRange
把 list 表达式展开为一组子作用域；空列表时返回一个带 Empty 标记的作用域。
*/
func ListRange(scope *Scope, expression string) ([]*Scope, error) {
	spec, err := parseList(expression)
	if err != nil {
		return nil, err
	}

	scopes, err := rangeCollection(scope, spec.Collection, "", spec.ItemAlias)
	if err != nil {
		return nil, err
	}
	if len(scopes) > 0 {
		for _, item := range scopes {
			item.SetVar("Empty", false)
			item.SetVar("empty", false)
		}
		return scopes, nil
	}

	empty := scope.Clone()
	empty.SetVar("Empty", true)
	empty.SetVar("empty", true)
	return []*Scope{empty}, nil
}

/*
ChildrenRange
把当前项的子级列表展开为一组子作用域；空子级时返回一个带 Empty 标记的作用域。
*/
func ChildrenRange(scope *Scope, expression string) ([]*Scope, error) {
	spec, err := parseChildren(expression)
	if err != nil {
		return nil, err
	}

	current, ok := scope.Vars[currentItemVar]
	if !ok {
		return nil, fmt.Errorf("@children 只能在列表项上下文中使用")
	}

	children, err := resolvePath(current, []string{spec.Field})
	if err != nil {
		return nil, err
	}

	scopes, err := rangeValue(scope, children, "", spec.ItemAlias)
	if err != nil {
		return nil, fmt.Errorf("@children 字段 %s 不是可遍历列表: %w", spec.Field, err)
	}
	if len(scopes) > 0 {
		for _, item := range scopes {
			item.SetVar("Empty", false)
			item.SetVar("empty", false)
		}
		return scopes, nil
	}

	empty := scope.Clone()
	empty.SetVar("Empty", true)
	empty.SetVar("empty", true)
	return []*Scope{empty}, nil
}

/*
EachRange
把 each 表达式展开为视图名、子作用域和空视图信息。
*/
func EachRange(scope *Scope, expression string) (string, []*Scope, string, error) {
	spec, err := parseEach(expression)
	if err != nil {
		return "", nil, "", err
	}

	scopes, err := Range(scope, spec.Collection+" as "+spec.ItemAlias)
	if err != nil {
		return "", nil, "", err
	}

	return spec.View, scopes, spec.EmptyView, nil
}

/*
IncludeScope
把 include 表达式展开为视图名和渲染作用域。
当 include 带第二个参数时，会把该数据合并到当前作用域的局部变量中。
*/
func IncludeScope(scope *Scope, expression string) (string, *Scope, error) {
	spec, err := parseInclude(expression)
	if err != nil {
		return "", nil, err
	}

	view, err := ResolveString(scope, spec.ViewExpression)
	if err != nil {
		return "", nil, err
	}

	if spec.DataExpression == "" {
		return view, scope, nil
	}

	data, err := ResolveExpression(scope, spec.DataExpression)
	if err != nil {
		return "", nil, err
	}

	return view, scope.WithData(data), nil
}

/*
ComponentScope
把 component 表达式展开为视图名和组件渲染作用域。
当 component 带第二个参数时，会把该数据合并到组件模板作用域中。
*/
func ComponentScope(scope *Scope, expression string) (string, *Scope, error) {
	spec, err := parseComponent(expression)
	if err != nil {
		return "", nil, err
	}

	view, err := ResolveString(scope, spec.ViewExpression)
	if err != nil {
		return "", nil, err
	}

	if spec.DataExpression == "" {
		return view, scope.Clone(), nil
	}

	data, err := ResolveExpression(scope, spec.DataExpression)
	if err != nil {
		return "", nil, err
	}

	return view, scope.WithData(data), nil
}

// TagCallData 表示运行时解析后的业务标签调用参数。
type TagCallData struct {
	Target string
	Params map[string]any
}

// TagCallScope 解析业务标签表达式，并返回目标与参数。
func TagCallScope(scope *Scope, expression string) (*TagCallData, error) {
	spec, err := parseTag(expression)
	if err != nil {
		return nil, err
	}

	target, err := ResolveString(scope, spec.TargetExpression)
	if err != nil {
		return nil, err
	}

	params := make(map[string]any, len(spec.Params))
	for key, valueExpression := range spec.Params {
		if (key == "as" || key == "id" || key == "out") && isBareIdentifier(valueExpression) {
			params[key] = valueExpression
			continue
		}

		value, err := ResolveExpression(scope, valueExpression)
		if err != nil {
			return nil, err
		}
		params[key] = value
	}

	return &TagCallData{
		Target: target,
		Params: params,
	}, nil
}

// FlexibleTagCallScope 同时支持带目标名称的标签调用和纯 key=value 参数调用，用于兼容输出型业务标签的两种语法。
func FlexibleTagCallScope(scope *Scope, expression string) (*TagCallData, error) {
	spec, err := parseFlexibleTag(expression)
	if err != nil {
		return nil, err
	}

	target := ""
	if spec.TargetExpression != "" {
		target, err = ResolveString(scope, spec.TargetExpression)
		if err != nil {
			return nil, err
		}
	}

	params, err := resolveTagParams(scope, spec.Params)
	if err != nil {
		return nil, err
	}

	return &TagCallData{
		Target: target,
		Params: params,
	}, nil
}

// NamedTagCallScope 解析数据标签的 key=value 参数；标签目标由指令名称确定，不从表达式中读取。
func NamedTagCallScope(scope *Scope, expression string) (*TagCallData, error) {
	paramsSpec, err := parseNamedTagParams(expression)
	if err != nil {
		return nil, err
	}

	params, err := resolveTagParams(scope, paramsSpec)
	if err != nil {
		return nil, err
	}

	return &TagCallData{
		Params: params,
	}, nil
}

func resolveTagParams(scope *Scope, paramsSpec map[string]string) (map[string]any, error) {
	params := make(map[string]any, len(paramsSpec))
	for key, valueExpression := range paramsSpec {
		if (key == "as" || key == "id" || key == "out") && isBareIdentifier(valueExpression) {
			params[key] = valueExpression
			continue
		}

		value, err := ResolveExpression(scope, valueExpression)
		if err != nil {
			return nil, err
		}
		params[key] = value
	}

	return params, nil
}

/*
ForRange
把 for 表达式展开为一组子作用域。
*/
func ForRange(scope *Scope, expression string) ([]*Scope, error) {
	spec, err := parseFor(expression)
	if err != nil {
		return nil, err
	}

	startValue, err := ResolveExpression(scope, spec.Start)
	if err != nil {
		return nil, err
	}
	endValue, err := ResolveExpression(scope, spec.End)
	if err != nil {
		return nil, err
	}

	start, ok := toInt64(startValue)
	if !ok {
		return nil, fmt.Errorf("for 起始值必须是整数: %s", spec.Start)
	}
	end, ok := toInt64(endValue)
	if !ok {
		return nil, fmt.Errorf("for 结束值必须是整数: %s", spec.End)
	}

	step := int64(1)
	if start > end {
		step = -1
	}
	// 使用无符号差值计算跨度，避免 MinInt64 到 MaxInt64 这类边界范围在
	// int64 减法时先发生溢出，从而绕过循环次数限制。
	var distance uint64
	if start <= end {
		distance = uint64(end) - uint64(start)
	} else {
		distance = uint64(start) - uint64(end)
	}
	if distance >= uint64(MaxLoopIterations) {
		return nil, fmt.Errorf("for 循环超过最大迭代次数 %d", MaxLoopIterations)
	}
	total := int(distance) + 1

	results := make([]*Scope, 0, 8)
	for value, index := start, 0; ; value, index = value+step, index+1 {
		child := scope.WithVar(spec.Alias, value)
		child.SetVar(currentItemVar, value)
		child.SetVar("loop", LoopState(index, total))
		results = append(results, child)

		if value == end {
			break
		}
	}

	return results, nil
}
