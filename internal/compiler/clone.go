package compiler

// CloneDocument 深复制模板 AST，使缓存、渲染流程和公开检查结果各自持有独立节点。
func CloneDocument(document *Document) *Document {
	if document == nil {
		return nil
	}

	return &Document{
		Nodes: cloneNodes(document.Nodes),
		Span:  document.Span,
	}
}

func cloneNodes(nodes []Node) []Node {
	if nodes == nil {
		return nil
	}

	cloned := make([]Node, len(nodes))
	for index, node := range nodes {
		cloned[index] = cloneNode(node)
	}
	return cloned
}

func cloneNode(node Node) Node {
	switch item := node.(type) {
	case *TextNode:
		cloned := *item
		return &cloned
	case *EchoNode:
		cloned := *item
		return &cloned
	case *CommentNode:
		cloned := *item
		return &cloned
	case *ExtendsNode:
		cloned := *item
		return &cloned
	case *IncludeNode:
		cloned := *item
		return &cloned
	case *EachNode:
		cloned := *item
		return &cloned
	case *ComponentNode:
		slots := make([]ComponentSlot, len(item.Slots))
		for index, slot := range item.Slots {
			slots[index] = ComponentSlot{
				Name: slot.Name,
				Body: cloneNodes(slot.Body),
				Span: slot.Span,
			}
		}
		return &ComponentNode{
			Target: item.Target,
			Body:   cloneNodes(item.Body),
			Slots:  slots,
			Span:   item.Span,
		}
	case *CustomDirectiveNode:
		return &CustomDirectiveNode{
			Name:       item.Name,
			Expression: item.Expression,
			Body:       cloneNodes(item.Body),
			EmptyBody:  cloneNodes(item.EmptyBody),
			EndName:    item.EndName,
			Span:       item.Span,
		}
	case *YieldNode:
		cloned := *item
		return &cloned
	case *SectionNode:
		return &SectionNode{
			Name: item.Name,
			Body: cloneNodes(item.Body),
			Span: item.Span,
		}
	case *ConditionalNode:
		branches := make([]ConditionalBranch, len(item.Branches))
		for index, branch := range item.Branches {
			branches[index] = ConditionalBranch{
				Directive:  branch.Directive,
				Expression: branch.Expression,
				Body:       cloneNodes(branch.Body),
				Span:       branch.Span,
			}
		}
		return &ConditionalNode{
			Branches: branches,
			ElseBody: cloneNodes(item.ElseBody),
			Span:     item.Span,
		}
	case *LoopNode:
		return &LoopNode{
			Directive:  item.Directive,
			Expression: item.Expression,
			Body:       cloneNodes(item.Body),
			Span:       item.Span,
		}
	case *ListNode:
		return &ListNode{
			Expression: item.Expression,
			Body:       cloneNodes(item.Body),
			EmptyBody:  cloneNodes(item.EmptyBody),
			Span:       item.Span,
		}
	case *ChildrenNode:
		return &ChildrenNode{
			Expression: item.Expression,
			Body:       cloneNodes(item.Body),
			EmptyBody:  cloneNodes(item.EmptyBody),
			Span:       item.Span,
		}
	default:
		return node
	}
}
