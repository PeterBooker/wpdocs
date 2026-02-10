package parser

import sitter "github.com/smacker/go-tree-sitter"

// nodeText returns the source text of a tree-sitter node.
func nodeText(node *sitter.Node, src []byte) string {
	if node == nil {
		return ""
	}
	return node.Content(src)
}

// startLine returns the 1-based start line of a node.
func startLine(node *sitter.Node) int {
	if node == nil {
		return 0
	}
	return int(node.StartPoint().Row) + 1
}

// endLine returns the 1-based end line of a node.
func endLine(node *sitter.Node) int {
	if node == nil {
		return 0
	}
	return int(node.EndPoint().Row) + 1
}

// childByType finds the first named child with the given node type.
func childByType(node *sitter.Node, typeName string) *sitter.Node {
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == typeName {
			return child
		}
	}
	return nil
}

// childrenByType finds all named children with the given node type.
func childrenByType(node *sitter.Node, typeName string) []*sitter.Node {
	var result []*sitter.Node
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == typeName {
			result = append(result, child)
		}
	}
	return result
}

// walkTree performs a depth-first traversal, calling fn on every named node.
func walkTree(node *sitter.Node, fn func(*sitter.Node)) {
	if node == nil {
		return
	}
	fn(node)
	for i := 0; i < int(node.NamedChildCount()); i++ {
		walkTree(node.NamedChild(i), fn)
	}
}
