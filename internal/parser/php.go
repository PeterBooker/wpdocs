package parser

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/peter/wpdocs/internal/model"
)

// extractPHP walks the tree-sitter AST and extracts PHP symbols.
func extractPHP(root *sitter.Node, src []byte, file string, reg *model.Registry) {
	ctx := &phpContext{
		src:  src,
		file: file,
		reg:  reg,
	}
	ctx.processChildren(root, "", nil)
}

type phpContext struct {
	src  []byte
	file string
	reg  *model.Registry
}

// processChildren iterates named children, tracking namespace changes across siblings.
func (ctx *phpContext) processChildren(node *sitter.Node, namespace string, classStack []string) {
	currentNS := namespace
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "namespace_definition" {
			nameNode := child.ChildByFieldName("name")
			ns := nodeText(nameNode, ctx.src)
			if body := child.ChildByFieldName("body"); body != nil {
				// Braced namespace: process children within
				ctx.processChildren(body, ns, classStack)
			} else {
				// Semicolon namespace: update current namespace for subsequent siblings
				currentNS = ns
			}
			continue
		}
		ctx.processNode(child, currentNS, classStack)
	}
}

func (ctx *phpContext) processNode(node *sitter.Node, namespace string, classStack []string) {
	switch node.Type() {
	case "function_definition":
		ctx.handleFunction(node, namespace)
	case "class_declaration":
		ctx.handleClass(node, namespace, classStack)
	case "interface_declaration":
		ctx.handleInterface(node, namespace, classStack)
	case "trait_declaration":
		ctx.handleTrait(node, namespace, classStack)
	}
}

func (ctx *phpContext) handleFunction(node *sitter.Node, namespace string) {
	nameNode := node.ChildByFieldName("name")
	name := nodeText(nameNode, ctx.src)
	if name == "" {
		return
	}
	fqn := qualifyPHP(namespace, name)

	doc := findDocComment(node, ctx.src)

	sym := &model.Symbol{
		ID:       fqn,
		Name:     name,
		Kind:     model.KindFunction,
		Language: "php",
		Doc:      doc,
		Params:   extractPHPParams(node.ChildByFieldName("parameters"), ctx.src, doc),
		Returns:  ParseReturn(doc),
		Location: model.SourceLocation{
			File:      ctx.file,
			StartLine: startLine(node),
			EndLine:   endLine(node),
		},
	}
	ctx.reg.Add(sym)

	// Scan function body for hooks
	if body := node.ChildByFieldName("body"); body != nil {
		scanForHooks(body, ctx.src, ctx.file, fqn, ctx.reg)
	}
}

func (ctx *phpContext) handleClass(node *sitter.Node, namespace string, classStack []string) {
	nameNode := node.ChildByFieldName("name")
	name := nodeText(nameNode, ctx.src)
	if name == "" {
		return
	}
	fqn := qualifyPHP(namespace, name)

	doc := findDocComment(node, ctx.src)
	sym := &model.Symbol{
		ID:       fqn,
		Name:     name,
		Kind:     model.KindClass,
		Language: "php",
		Doc:      doc,
		Location: model.SourceLocation{
			File:      ctx.file,
			StartLine: startLine(node),
			EndLine:   endLine(node),
		},
	}

	// Extends (single parent class)
	if bc := childByType(node, "base_clause"); bc != nil {
		for i := 0; i < int(bc.NamedChildCount()); i++ {
			sym.Extends = append(sym.Extends, nodeText(bc.NamedChild(i), ctx.src))
		}
	}

	// Implements (multiple interfaces)
	if ic := childByType(node, "class_interface_clause"); ic != nil {
		for i := 0; i < int(ic.NamedChildCount()); i++ {
			sym.Implements = append(sym.Implements, nodeText(ic.NamedChild(i), ctx.src))
		}
	}

	ctx.reg.Add(sym)

	// Process class body members
	if body := childByType(node, "declaration_list"); body != nil {
		newStack := append(append([]string{}, classStack...), fqn)
		ctx.processClassBody(body, namespace, newStack)
	}
}

func (ctx *phpContext) handleInterface(node *sitter.Node, namespace string, classStack []string) {
	nameNode := node.ChildByFieldName("name")
	name := nodeText(nameNode, ctx.src)
	if name == "" {
		return
	}
	fqn := qualifyPHP(namespace, name)

	doc := findDocComment(node, ctx.src)
	sym := &model.Symbol{
		ID:       fqn,
		Name:     name,
		Kind:     model.KindInterface,
		Language: "php",
		Doc:      doc,
		Location: model.SourceLocation{
			File:      ctx.file,
			StartLine: startLine(node),
			EndLine:   endLine(node),
		},
	}

	// Interface extends
	if bc := childByType(node, "base_clause"); bc != nil {
		for i := 0; i < int(bc.NamedChildCount()); i++ {
			sym.Extends = append(sym.Extends, nodeText(bc.NamedChild(i), ctx.src))
		}
	}

	ctx.reg.Add(sym)

	if body := childByType(node, "declaration_list"); body != nil {
		newStack := append(append([]string{}, classStack...), fqn)
		ctx.processClassBody(body, namespace, newStack)
	}
}

func (ctx *phpContext) handleTrait(node *sitter.Node, namespace string, classStack []string) {
	nameNode := node.ChildByFieldName("name")
	name := nodeText(nameNode, ctx.src)
	if name == "" {
		return
	}
	fqn := qualifyPHP(namespace, name)

	doc := findDocComment(node, ctx.src)
	sym := &model.Symbol{
		ID:       fqn,
		Name:     name,
		Kind:     model.KindTrait,
		Language: "php",
		Doc:      doc,
		Location: model.SourceLocation{
			File:      ctx.file,
			StartLine: startLine(node),
			EndLine:   endLine(node),
		},
	}
	ctx.reg.Add(sym)

	if body := childByType(node, "declaration_list"); body != nil {
		newStack := append(append([]string{}, classStack...), fqn)
		ctx.processClassBody(body, namespace, newStack)
	}
}

// processClassBody handles method declarations inside a class/interface/trait body.
func (ctx *phpContext) processClassBody(body *sitter.Node, namespace string, classStack []string) {
	for i := 0; i < int(body.NamedChildCount()); i++ {
		child := body.NamedChild(i)
		if child.Type() == "method_declaration" {
			ctx.handleMethod(child, namespace, classStack)
		}
	}
}

func (ctx *phpContext) handleMethod(node *sitter.Node, namespace string, classStack []string) {
	nameNode := node.ChildByFieldName("name")
	name := nodeText(nameNode, ctx.src)
	if name == "" || len(classStack) == 0 {
		return
	}
	classFQN := classStack[len(classStack)-1]
	methodID := classFQN + "::" + name

	doc := findDocComment(node, ctx.src)
	sym := &model.Symbol{
		ID:       methodID,
		Name:     name,
		Kind:     model.KindMethod,
		Language: "php",
		Doc:      doc,
		Params:   extractPHPParams(node.ChildByFieldName("parameters"), ctx.src, doc),
		Returns:  ParseReturn(doc),
		ParentID: classFQN,
		Location: model.SourceLocation{
			File:      ctx.file,
			StartLine: startLine(node),
			EndLine:   endLine(node),
		},
	}
	ctx.reg.Add(sym)

	// Register method under parent class
	if parent := ctx.reg.Get(classFQN); parent != nil {
		parent.Members = append(parent.Members, methodID)
	}

	// Scan method body for hooks
	if body := node.ChildByFieldName("body"); body != nil {
		scanForHooks(body, ctx.src, ctx.file, methodID, ctx.reg)
	}
}

func qualifyPHP(namespace, name string) string {
	if namespace != "" {
		return namespace + "\\" + name
	}
	return name
}

// extractPHPParams extracts parameters from a formal_parameters node, merging with docblock info.
func extractPHPParams(paramsNode *sitter.Node, src []byte, doc model.DocBlock) []model.Param {
	if paramsNode == nil {
		return nil
	}

	docParams := ParseParams(doc)
	docMap := make(map[string]model.Param)
	for _, dp := range docParams {
		docMap[dp.Name] = dp
	}

	var result []model.Param
	for i := 0; i < int(paramsNode.NamedChildCount()); i++ {
		param := paramsNode.NamedChild(i)
		if param.Type() != "simple_parameter" && param.Type() != "property_promotion_parameter" {
			continue
		}

		nameNode := param.ChildByFieldName("name")
		name := nodeText(nameNode, src)
		name = strings.TrimPrefix(name, "$")
		if name == "" {
			continue
		}

		mp := model.Param{Name: name}

		// Type from AST
		if typeNode := param.ChildByFieldName("type"); typeNode != nil {
			mp.Type = nodeText(typeNode, src)
		}

		// Merge doc info
		if dp, ok := docMap[name]; ok {
			if mp.Type == "" {
				mp.Type = dp.Type
			}
			mp.Description = dp.Description
			mp.IsNullable = dp.IsNullable
		}

		// Check for variadic and reference via parameter text
		paramText := nodeText(param, src)
		if strings.Contains(paramText, "...") {
			mp.IsVariadic = true
		}
		if strings.Contains(paramText, "&$") {
			mp.IsPassByRef = true
		}

		result = append(result, mp)
	}
	return result
}
