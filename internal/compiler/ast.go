package compiler

/*
NodeKind
表示 AST 节点类型。
*/
type NodeKind string

const (
	/*
		NodeKindText
		表示文本节点。
	*/
	NodeKindText NodeKind = "text"
	/*
		NodeKindEcho
		表示变量输出节点。
	*/
	NodeKindEcho NodeKind = "echo"
	/*
		NodeKindComment
		表示注释节点。
	*/
	NodeKindComment NodeKind = "comment"
	/*
		NodeKindExtends
		表示模板继承节点。
	*/
	NodeKindExtends NodeKind = "extends"
	/*
		NodeKindInclude
		表示模板引入节点。
	*/
	NodeKindInclude NodeKind = "include"
	/*
		NodeKindEach
		表示 each 节点。
	*/
	NodeKindEach NodeKind = "each"
	/*
		NodeKindComponent
		表示组件节点。
	*/
	NodeKindComponent NodeKind = "component"
	/*
		NodeKindCustomDirective
		表示自定义指令节点。
	*/
	NodeKindCustomDirective NodeKind = "custom_directive"
	/*
		NodeKindYield
		表示占位输出节点。
	*/
	NodeKindYield NodeKind = "yield"
	/*
		NodeKindSection
		表示 section 块节点。
	*/
	NodeKindSection NodeKind = "section"
	/*
		NodeKindConditional
		表示条件分支节点。
	*/
	NodeKindConditional NodeKind = "conditional"
	/*
		NodeKindLoop
		表示循环节点。
	*/
	NodeKindLoop NodeKind = "loop"
	/*
		NodeKindList
		表示列表块节点。
	*/
	NodeKindList NodeKind = "list"
	/*
		NodeKindChildren
		表示当前项子级列表块节点。
	*/
	NodeKindChildren NodeKind = "children"
)

/*
Node
表示所有 AST 节点的公共协议。
*/
type Node interface {
	NodeKind() NodeKind
	NodeSpan() Span
}

/*
Document
表示完整模板文档的根节点。
*/
type Document struct {
	Nodes []Node
	Span  Span
}

/*
TextNode
表示纯文本节点。
*/
type TextNode struct {
	Text string
	Span Span
}

/*
EchoNode
表示变量输出节点。
Raw 为 true 时表示原样输出。
*/
type EchoNode struct {
	Expression string
	Raw        bool
	Span       Span
}

/*
CommentNode
表示模板注释节点。
*/
type CommentNode struct {
	Content string
	Span    Span
}

/*
ExtendsNode
表示模板继承声明节点。
*/
type ExtendsNode struct {
	Target string
	Span   Span
}

/*
IncludeNode
表示模板引入节点。
*/
type IncludeNode struct {
	Target string
	Span   Span
}

/*
EachNode
表示 each 批量渲染节点。
*/
type EachNode struct {
	Expression string
	Span       Span
}

/*
ComponentSlot
表示组件中的命名插槽。
*/
type ComponentSlot struct {
	Name string
	Body []Node
	Span Span
}

/*
ComponentNode
表示组件渲染节点。
Body 表示默认插槽内容，Slots 表示命名插槽内容。
*/
type ComponentNode struct {
	Target string
	Body   []Node
	Slots  []ComponentSlot
	Span   Span
}

/*
CustomDirectiveNode
表示运行时可扩展的自定义指令节点。
*/
type CustomDirectiveNode struct {
	Name       string
	Expression string
	Body       []Node
	EmptyBody  []Node
	EndName    string
	Span       Span
}

/*
YieldNode
表示模板占位节点。
*/
type YieldNode struct {
	Target string
	Span   Span
}

/*
SectionNode
表示 section 块节点。
*/
type SectionNode struct {
	Name string
	Body []Node
	Span Span
}

/*
ConditionalBranch
表示条件节点中的单个分支。
Directive 可能为 if、elseif 或 unless。
*/
type ConditionalBranch struct {
	Directive  string
	Expression string
	Body       []Node
	Span       Span
}

/*
ConditionalNode
表示条件分支结构节点。
*/
type ConditionalNode struct {
	Branches []ConditionalBranch
	ElseBody []Node
	Span     Span
}

/*
LoopNode
表示循环结构节点。
Directive 可能为 for 或 foreach。
*/
type LoopNode struct {
	Directive  string
	Expression string
	Body       []Node
	Span       Span
}

/*
ListNode
表示列表结构节点。
Body 表示列表项模板，EmptyBody 表示空列表模板。
*/
type ListNode struct {
	Expression string
	Body       []Node
	EmptyBody  []Node
	Span       Span
}

/*
ChildrenNode
表示当前项子级列表结构节点。
Body 表示子级项模板，EmptyBody 表示空子级模板。
*/
type ChildrenNode struct {
	Expression string
	Body       []Node
	EmptyBody  []Node
	Span       Span
}

func (n *TextNode) NodeKind() NodeKind {
	return NodeKindText
}

func (n *TextNode) NodeSpan() Span {
	return n.Span
}

func (n *EchoNode) NodeKind() NodeKind {
	return NodeKindEcho
}

func (n *EchoNode) NodeSpan() Span {
	return n.Span
}

func (n *CommentNode) NodeKind() NodeKind {
	return NodeKindComment
}

func (n *CommentNode) NodeSpan() Span {
	return n.Span
}

func (n *ExtendsNode) NodeKind() NodeKind {
	return NodeKindExtends
}

func (n *ExtendsNode) NodeSpan() Span {
	return n.Span
}

func (n *IncludeNode) NodeKind() NodeKind {
	return NodeKindInclude
}

func (n *IncludeNode) NodeSpan() Span {
	return n.Span
}

func (n *EachNode) NodeKind() NodeKind {
	return NodeKindEach
}

func (n *EachNode) NodeSpan() Span {
	return n.Span
}

func (n *ComponentNode) NodeKind() NodeKind {
	return NodeKindComponent
}

func (n *ComponentNode) NodeSpan() Span {
	return n.Span
}

func (n *CustomDirectiveNode) NodeKind() NodeKind {
	return NodeKindCustomDirective
}

func (n *CustomDirectiveNode) NodeSpan() Span {
	return n.Span
}

func (n *YieldNode) NodeKind() NodeKind {
	return NodeKindYield
}

func (n *YieldNode) NodeSpan() Span {
	return n.Span
}

func (n *SectionNode) NodeKind() NodeKind {
	return NodeKindSection
}

func (n *SectionNode) NodeSpan() Span {
	return n.Span
}

func (n *ConditionalNode) NodeKind() NodeKind {
	return NodeKindConditional
}

func (n *ConditionalNode) NodeSpan() Span {
	return n.Span
}

func (n *LoopNode) NodeKind() NodeKind {
	return NodeKindLoop
}

func (n *LoopNode) NodeSpan() Span {
	return n.Span
}

func (n *ListNode) NodeKind() NodeKind {
	return NodeKindList
}

func (n *ListNode) NodeSpan() Span {
	return n.Span
}

func (n *ChildrenNode) NodeKind() NodeKind {
	return NodeKindChildren
}

func (n *ChildrenNode) NodeSpan() Span {
	return n.Span
}
