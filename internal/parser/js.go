package parser

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/peter/wpdocs/internal/model"
)

// extractJS walks the tree-sitter AST and extracts JS/TS symbols.
func extractJS(root *sitter.Node, src []byte, file string, reg *model.Registry) {
	ctx := &jsContext{
		src:  src,
		file: file,
		reg:  reg,
	}
	ctx.processChildren(root, nil)
}

type jsContext struct {
	src  []byte
	file string
	reg  *model.Registry
}

func (ctx *jsContext) processChildren(node *sitter.Node, classStack []string) {
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		ctx.processNode(child, classStack)
	}
}

func (ctx *jsContext) processNode(node *sitter.Node, classStack []string) {
	switch node.Type() {
	case "function_declaration":
		ctx.handleFunction(node)
	case "class_declaration":
		ctx.handleClass(node, classStack)
	case "interface_declaration":
		ctx.handleInterface(node)
	case "export_statement":
		// Recurse into exported declarations
		ctx.processChildren(node, classStack)
	case "lexical_declaration", "variable_declaration":
		ctx.handleVarDecl(node)
	}
}

func (ctx *jsContext) handleFunction(node *sitter.Node) {
	nameNode := node.ChildByFieldName("name")
	name := nodeText(nameNode, ctx.src)
	if name == "" {
		return
	}

	doc := findDocComment(node, ctx.src)

	sym := &model.Symbol{
		ID:       name,
		Name:     name,
		Kind:     model.KindFunction,
		Language: "js",
		Doc:      doc,
		Params:   extractJSParams(node.ChildByFieldName("parameters"), ctx.src, doc),
		Returns:  jsReturn(node, ctx.src, doc),
		Location: model.SourceLocation{
			File:      ctx.file,
			StartLine: startLine(node),
			EndLine:   endLine(node),
		},
	}
	ctx.reg.Add(sym)
}

func (ctx *jsContext) handleClass(node *sitter.Node, classStack []string) {
	nameNode := node.ChildByFieldName("name")
	name := nodeText(nameNode, ctx.src)
	if name == "" {
		return
	}

	doc := findDocComment(node, ctx.src)
	sym := &model.Symbol{
		ID:       name,
		Name:     name,
		Kind:     model.KindClass,
		Language: "js",
		Doc:      doc,
		Location: model.SourceLocation{
			File:      ctx.file,
			StartLine: startLine(node),
			EndLine:   endLine(node),
		},
	}

	// Heritage: extends/implements
	if heritage := childByType(node, "class_heritage"); heritage != nil {
		for i := 0; i < int(heritage.NamedChildCount()); i++ {
			clause := heritage.NamedChild(i)
			switch clause.Type() {
			case "extends_clause":
				for j := 0; j < int(clause.NamedChildCount()); j++ {
					sym.Extends = append(sym.Extends, nodeText(clause.NamedChild(j), ctx.src))
				}
			case "implements_clause":
				for j := 0; j < int(clause.NamedChildCount()); j++ {
					sym.Implements = append(sym.Implements, nodeText(clause.NamedChild(j), ctx.src))
				}
			}
		}
	}

	ctx.reg.Add(sym)

	// Process class body
	if body := node.ChildByFieldName("body"); body != nil {
		newStack := append(append([]string{}, classStack...), name)
		ctx.processClassBody(body, newStack)
	}
}

func (ctx *jsContext) processClassBody(body *sitter.Node, classStack []string) {
	for i := 0; i < int(body.NamedChildCount()); i++ {
		child := body.NamedChild(i)
		if child.Type() == "method_definition" {
			ctx.handleMethod(child, classStack)
		}
	}
}

func (ctx *jsContext) handleMethod(node *sitter.Node, classStack []string) {
	nameNode := node.ChildByFieldName("name")
	name := nodeText(nameNode, ctx.src)
	if name == "" || len(classStack) == 0 {
		return
	}

	classFQN := classStack[len(classStack)-1]
	methodID := classFQN + "." + name

	doc := findDocComment(node, ctx.src)
	sym := &model.Symbol{
		ID:       methodID,
		Name:     name,
		Kind:     model.KindMethod,
		Language: "js",
		Doc:      doc,
		Params:   extractJSParams(node.ChildByFieldName("parameters"), ctx.src, doc),
		Returns:  jsReturn(node, ctx.src, doc),
		ParentID: classFQN,
		Location: model.SourceLocation{
			File:      ctx.file,
			StartLine: startLine(node),
			EndLine:   endLine(node),
		},
	}
	ctx.reg.Add(sym)

	if parent := ctx.reg.Get(classFQN); parent != nil {
		parent.Members = append(parent.Members, methodID)
	}
}

func (ctx *jsContext) handleInterface(node *sitter.Node) {
	nameNode := node.ChildByFieldName("name")
	name := nodeText(nameNode, ctx.src)
	if name == "" {
		return
	}

	doc := findDocComment(node, ctx.src)
	sym := &model.Symbol{
		ID:       name,
		Name:     name,
		Kind:     model.KindInterface,
		Language: "js",
		Doc:      doc,
		Location: model.SourceLocation{
			File:      ctx.file,
			StartLine: startLine(node),
			EndLine:   endLine(node),
		},
	}

	// Interface extends
	if heritage := childByType(node, "extends_type_clause"); heritage != nil {
		for i := 0; i < int(heritage.NamedChildCount()); i++ {
			sym.Extends = append(sym.Extends, nodeText(heritage.NamedChild(i), ctx.src))
		}
	}

	ctx.reg.Add(sym)
}

func (ctx *jsContext) handleVarDecl(node *sitter.Node) {
	// Look for: const foo = () => {} or const foo = function() {}
	for i := 0; i < int(node.NamedChildCount()); i++ {
		declarator := node.NamedChild(i)
		if declarator.Type() != "variable_declarator" {
			continue
		}

		nameNode := declarator.ChildByFieldName("name")
		valueNode := declarator.ChildByFieldName("value")
		if nameNode == nil || valueNode == nil {
			continue
		}

		// Check if the value is a function expression or arrow function
		switch valueNode.Type() {
		case "arrow_function", "function_expression", "function":
			name := nodeText(nameNode, ctx.src)
			if name == "" {
				continue
			}

			// Use the doc comment from the variable declaration, not the inner function
			doc := findDocComment(node, ctx.src)

			sym := &model.Symbol{
				ID:       name,
				Name:     name,
				Kind:     model.KindFunction,
				Language: "js",
				Doc:      doc,
				Params:   extractJSParams(valueNode.ChildByFieldName("parameters"), ctx.src, doc),
				Returns:  jsReturn(valueNode, ctx.src, doc),
				Location: model.SourceLocation{
					File:      ctx.file,
					StartLine: startLine(node),
					EndLine:   endLine(node),
				},
			}
			ctx.reg.Add(sym)
		}
	}
}

// extractJSParams extracts parameters from a formal_parameters node, merging with JSDoc info.
func extractJSParams(paramsNode *sitter.Node, src []byte, doc model.DocBlock) []model.Param {
	if paramsNode == nil {
		return nil
	}

	// Build doc param map from JSDoc
	docMap := make(map[string]model.Param)
	for _, raw := range doc.Tags["param"] {
		p := parseJSDocParam(raw)
		if p.Name != "" {
			docMap[p.Name] = p
		}
	}

	var result []model.Param
	for i := 0; i < int(paramsNode.NamedChildCount()); i++ {
		param := paramsNode.NamedChild(i)

		var name, typeName string
		switch param.Type() {
		case "identifier":
			name = nodeText(param, src)
		case "required_parameter", "optional_parameter":
			if n := param.ChildByFieldName("pattern"); n != nil {
				name = nodeText(n, src)
			} else if n := param.ChildByFieldName("name"); n != nil {
				name = nodeText(n, src)
			}
			if t := param.ChildByFieldName("type"); t != nil {
				typeName = nodeText(t, src)
				typeName = strings.TrimPrefix(typeName, ": ")
			}
		case "formal_parameter":
			if n := param.ChildByFieldName("name"); n != nil {
				name = nodeText(n, src)
			}
		default:
			name = nodeText(param, src)
		}

		if name == "" {
			continue
		}

		mp := model.Param{Name: name, Type: typeName}

		// Merge JSDoc info
		if dp, ok := docMap[name]; ok {
			if mp.Type == "" {
				mp.Type = dp.Type
			}
			mp.Description = dp.Description
		}

		result = append(result, mp)
	}
	return result
}

// parseJSDocParam handles JSDoc format: {type} name description
func parseJSDocParam(raw string) model.Param {
	raw = strings.TrimSpace(raw)
	var p model.Param

	// Check for {type} prefix
	if strings.HasPrefix(raw, "{") {
		endBrace := strings.Index(raw, "}")
		if endBrace != -1 {
			p.Type = raw[1:endBrace]
			raw = strings.TrimSpace(raw[endBrace+1:])
		}
	}

	// Next is the name, then description
	parts := strings.SplitN(raw, " ", 2)
	if len(parts) >= 1 {
		p.Name = strings.TrimPrefix(parts[0], "$")
	}
	if len(parts) >= 2 {
		p.Description = strings.TrimSpace(parts[1])
	}

	return p
}

// jsReturn extracts the return type from a TS annotation or JSDoc.
func jsReturn(node *sitter.Node, src []byte, doc model.DocBlock) *model.ReturnValue {
	// Check explicit return type annotation
	if retType := node.ChildByFieldName("return_type"); retType != nil {
		text := nodeText(retType, src)
		text = strings.TrimPrefix(text, ": ")
		return &model.ReturnValue{Type: text}
	}

	// Fall back to JSDoc @return
	if ret := ParseReturn(doc); ret != nil {
		return ret
	}

	// Check @returns tag (JS convention)
	if returns, ok := doc.Tags["returns"]; ok && len(returns) > 0 {
		raw := returns[0]
		raw = strings.TrimSpace(raw)
		if strings.HasPrefix(raw, "{") {
			endBrace := strings.Index(raw, "}")
			if endBrace != -1 {
				typeName := raw[1:endBrace]
				desc := strings.TrimSpace(raw[endBrace+1:])
				return &model.ReturnValue{Type: typeName, Description: desc}
			}
		}
		return &model.ReturnValue{Type: raw}
	}

	return nil
}
