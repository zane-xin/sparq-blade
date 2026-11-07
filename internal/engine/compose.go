package engine

import (
	"fmt"
	"strings"

	"github.com/zane-xin/sparq-blade/internal/compiler"
	"github.com/zane-xin/sparq-blade/internal/literal"
	bladeRuntime "github.com/zane-xin/sparq-blade/internal/runtime"
)

func (e *Engine) resolveFileDocument(name string) (*compiler.Document, string, string, error) {
	source, document, err := e.loadParsedFile(name)
	if err != nil {
		return nil, "", "", err
	}

	resolved, version, err := e.resolveDocument(document, source.Name, source.Version, nil, nil)
	if err != nil {
		return nil, "", "", err
	}

	return resolved, source.Name, version, nil
}

func (e *Engine) resolveDocument(document *compiler.Document, sourceName string, sourceVersion string, overrides map[string]*compiler.SectionNode, inheritanceChain []string) (*compiler.Document, string, error) {
	for _, current := range inheritanceChain {
		if current == sourceName {
			chain := append(append([]string(nil), inheritanceChain...), sourceName)
			return nil, "", fmt.Errorf("模板继承存在循环调用: %s", strings.Join(chain, " -> "))
		}
	}
	if len(inheritanceChain) >= maxTemplateCompositionDepth {
		return nil, "", fmt.Errorf("模板继承超过最大深度 %d", maxTemplateCompositionDepth)
	}
	inheritanceChain = append(append([]string(nil), inheritanceChain...), sourceName)

	localSections := collectSections(document)
	effectiveSections := mergeSections(localSections, overrides)
	version := fmt.Sprintf("%s:%s", sourceName, sourceVersion)

	if parent := findExtends(document); parent != nil {
		name, err := bladeRuntime.ResolveString(bladeRuntime.NewScope(nil), parent.Target)
		if err != nil {
			return nil, "", err
		}

		parentSource, parentDocument, err := e.loadParsedFile(name)
		if err != nil {
			return nil, "", err
		}

		resolved, parentVersion, err := e.resolveDocument(parentDocument, parentSource.Name, parentSource.Version, effectiveSections, inheritanceChain)
		if err != nil {
			return nil, "", err
		}

		return resolved, version + "|" + parentVersion, nil
	}

	resolvedNodes, err := applySections(document.Nodes, overrides)
	if err != nil {
		return nil, "", err
	}

	return &compiler.Document{
		Nodes: resolvedNodes,
		Span:  document.Span,
	}, version, nil
}

func findExtends(document *compiler.Document) *compiler.ExtendsNode {
	for _, node := range document.Nodes {
		if item, ok := node.(*compiler.ExtendsNode); ok {
			return item
		}
	}

	return nil
}

func collectSections(document *compiler.Document) map[string]*compiler.SectionNode {
	sections := make(map[string]*compiler.SectionNode)
	for _, node := range document.Nodes {
		section, ok := node.(*compiler.SectionNode)
		if !ok {
			continue
		}

		name, ok := literal.ParseString(section.Name)
		if !ok {
			name = section.Name
		}
		sections[name] = section
	}

	return sections
}

func mergeSections(base map[string]*compiler.SectionNode, overrides map[string]*compiler.SectionNode) map[string]*compiler.SectionNode {
	if len(base) == 0 && len(overrides) == 0 {
		return nil
	}

	merged := make(map[string]*compiler.SectionNode, len(base)+len(overrides))
	for name, section := range base {
		merged[name] = section
	}

	for name, section := range overrides {
		merged[name] = section
	}

	return merged
}

func applySections(nodes []compiler.Node, overrides map[string]*compiler.SectionNode) ([]compiler.Node, error) {
	results := make([]compiler.Node, 0, len(nodes))

	for _, node := range nodes {
		resolved, err := resolveNode(node, overrides)
		if err != nil {
			return nil, err
		}

		results = append(results, resolved...)
	}

	return results, nil
}

func resolveNode(node compiler.Node, overrides map[string]*compiler.SectionNode) ([]compiler.Node, error) {
	switch item := node.(type) {
	case *compiler.ExtendsNode:
		return nil, nil
	case *compiler.SectionNode:
		name, ok := literal.ParseString(item.Name)
		if !ok {
			name = item.Name
		}

		if section, ok := overrides[name]; ok {
			return applySections(section.Body, overrides)
		}

		return applySections(item.Body, overrides)
	case *compiler.YieldNode:
		name, ok := literal.ParseString(item.Target)
		if !ok {
			name = item.Target
		}

		if section, ok := overrides[name]; ok {
			return applySections(section.Body, overrides)
		}

		return nil, nil
	case *compiler.ConditionalNode:
		branches := make([]compiler.ConditionalBranch, 0, len(item.Branches))
		for _, branch := range item.Branches {
			body, err := applySections(branch.Body, overrides)
			if err != nil {
				return nil, err
			}

			branches = append(branches, compiler.ConditionalBranch{
				Directive:  branch.Directive,
				Expression: branch.Expression,
				Body:       body,
				Span:       branch.Span,
			})
		}

		elseBody, err := applySections(item.ElseBody, overrides)
		if err != nil {
			return nil, err
		}

		return []compiler.Node{
			&compiler.ConditionalNode{
				Branches: branches,
				ElseBody: elseBody,
				Span:     item.Span,
			},
		}, nil
	case *compiler.LoopNode:
		body, err := applySections(item.Body, overrides)
		if err != nil {
			return nil, err
		}

		return []compiler.Node{
			&compiler.LoopNode{
				Directive:  item.Directive,
				Expression: item.Expression,
				Body:       body,
				Span:       item.Span,
			},
		}, nil
	case *compiler.ListNode:
		body, err := applySections(item.Body, overrides)
		if err != nil {
			return nil, err
		}

		emptyBody, err := applySections(item.EmptyBody, overrides)
		if err != nil {
			return nil, err
		}

		return []compiler.Node{
			&compiler.ListNode{
				Expression: item.Expression,
				Body:       body,
				EmptyBody:  emptyBody,
				Span:       item.Span,
			},
		}, nil
	case *compiler.ChildrenNode:
		body, err := applySections(item.Body, overrides)
		if err != nil {
			return nil, err
		}

		emptyBody, err := applySections(item.EmptyBody, overrides)
		if err != nil {
			return nil, err
		}

		return []compiler.Node{
			&compiler.ChildrenNode{
				Expression: item.Expression,
				Body:       body,
				EmptyBody:  emptyBody,
				Span:       item.Span,
			},
		}, nil
	case *compiler.CustomDirectiveNode:
		body, err := applySections(item.Body, overrides)
		if err != nil {
			return nil, err
		}

		emptyBody, err := applySections(item.EmptyBody, overrides)
		if err != nil {
			return nil, err
		}

		return []compiler.Node{
			&compiler.CustomDirectiveNode{
				Name:       item.Name,
				Expression: item.Expression,
				Body:       body,
				EmptyBody:  emptyBody,
				EndName:    item.EndName,
				Span:       item.Span,
			},
		}, nil

	case *compiler.ComponentNode:
		body, err := applySections(item.Body, overrides)
		if err != nil {
			return nil, err
		}

		slots := make([]compiler.ComponentSlot, 0, len(item.Slots))
		for _, slot := range item.Slots {
			slotBody, err := applySections(slot.Body, overrides)
			if err != nil {
				return nil, err
			}

			slots = append(slots, compiler.ComponentSlot{
				Name: slot.Name,
				Body: slotBody,
				Span: slot.Span,
			})
		}

		return []compiler.Node{
			&compiler.ComponentNode{
				Target: item.Target,
				Body:   body,
				Slots:  slots,
				Span:   item.Span,
			},
		}, nil
	default:
		return []compiler.Node{node}, nil
	}
}
