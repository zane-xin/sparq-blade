// Package blade 提供受控表达式、上下文自动转义、模板组合和进程内扩展能力。
//
// 模板源码、自定义指令和输出型业务标签属于应用的可信代码资产，不应由最终用户动态提交。
// ViewContext 和数据标签返回值可以包含外部数据，普通 {{ }} 输出会交给 html/template 按上下文转义；
// {!! !!} 与 TagFunc 返回的 template.HTML 会绕过转义，只能承载应用生成或已可靠清洗的内容。
package blade
