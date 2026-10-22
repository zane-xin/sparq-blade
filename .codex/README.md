# 使用go实现的参考 laravel 的 blade 模版解析库
> 参考 blade 使用语法实现, 解析器可依赖go的解析库

## 一、项目概述
### 1.1 目标
> 实现一个类似 Laravel Blade 的 Go 模板引擎，支持以下核心特性：

特性	优先级
模板继承 (@extends, @section, @yield)	P0
组件/插槽 (@component, @slot)	P1
条件指令 (@if, @elseif, @else, @unless)	P0
循环指令 (@for, @foreach, @while)	P0
变量输出与转义 ({{ var }}, {!! var !!})	P0
内置指令 (@include, @each, @go)	P1
自定义指令注册	P2

### 1.2 技术选型
基础: html/template 或 text/template 作为底层
扩展: 通过 FuncMap 注入自定义函数
编译方式: 代码生成或运行时解析


## 二、架构设计

### 2.1 模块划分
```
blade/
├── compiler/           # 编译器核心
│   ├── lexer.go       # 词法分析
│   ├── parser.go      # 语法分析
│   ├── ast.go         # AST 节点定义
│   └── generator.go   # Go 模板生成
├── engine/             # 引擎封装
│   ├── engine.go      # 主引擎
│   ├── cache.go       # 编译缓存
│   └── fs.go          # 文件系统支持
├── directives/         # 内置指令
│   ├── if.go
│   ├── foreach.go
│   ├── include.go
│   └── extends.go
├── runtime/            # 运行时
│   ├── context.go     # 渲染上下文
│   └── helpers.go     # 辅助函数
└── blade.go           # 公开 API

```

### 2.2 核心流程
```
sequenceDiagram
    participant App as 应用代码
    participant Engine as Blade引擎
    participant Cache as 编译缓存
    participant Compiler as 编译器
    participant GoTemplate as Go模板
    
    App->>Engine: Render("home", data)
    Engine->>Cache: 检查缓存
    alt 已缓存
        Cache-->>Engine: 返回Go模板
    else 未缓存
        Engine->>Compiler: Compile("home.blade")
        Compiler->>Compiler: 解析Blade语法
        Compiler->>Compiler: 生成Go模板
        Compiler-->>Engine: 返回*Template
        Engine->>Cache: 存入缓存
    end
    Engine->>GoTemplate: Execute(data)
    GoTemplate-->>App: HTML输出
```

## 三、语法规范

Blade 语法	说明	转义
{{ variable }}	自动 HTML 转义	✅
{!! variable !!}	原始输出（不转义）	❌
{{-- 注释 --}}	模板注释	-

> 其中 $ 可以考虑是否支持自定义或者 go的 . 语法
>

