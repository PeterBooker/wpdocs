package parser

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/peter/wpdocs/internal/model"
)

// WordPress hook functions we detect.
var hookFunctions = map[string]model.HookType{
	"do_action":               model.HookAction,
	"do_action_ref_array":     model.HookAction,
	"apply_filters":           model.HookFilter,
	"apply_filters_ref_array": model.HookFilter,
}

// scanForHooks walks the AST subtree looking for WordPress hook calls.
func scanForHooks(bodyNode *sitter.Node, src []byte, file string, callerID string, reg *model.Registry) {
	walkTree(bodyNode, func(node *sitter.Node) {
		if node.Type() != "function_call_expression" {
			return
		}
		fnNode := node.ChildByFieldName("function")
		if fnNode == nil {
			return
		}
		fnName := nodeText(fnNode, src)
		hookType, isHook := hookFunctions[fnName]
		if !isHook {
			return
		}
		registerHook(node, hookType, callerID, src, file, reg)
	})
}

// registerHook extracts the hook tag and creates a hook symbol.
func registerHook(call *sitter.Node, hookType model.HookType, callerID string, src []byte, file string, reg *model.Registry) {
	args := call.ChildByFieldName("arguments")
	if args == nil || args.NamedChildCount() == 0 {
		return
	}

	// First argument is the hook tag
	firstArg := args.NamedChild(0)
	// If wrapped in an argument node, unwrap it
	if firstArg.Type() == "argument" {
		if firstArg.NamedChildCount() > 0 {
			firstArg = firstArg.NamedChild(0)
		}
	}

	tag := extractHookTag(firstArg, src)
	if tag == "" {
		return
	}

	hookID := "hook:" + tag

	// Check if hook already registered (hooks can be fired from multiple places)
	existing := reg.Get(hookID)
	if existing != nil {
		existing.CallSites = append(existing.CallSites, callerID)
		return
	}

	// Extract doc block from the hook call site
	doc := findDocComment(call, src)

	// Build params from the doc's @param tags
	var params []model.Param
	for _, raw := range doc.Tags["param"] {
		if m := paramRegex.FindStringSubmatch("@param " + raw); m != nil {
			params = append(params, model.Param{
				Type:        m[1],
				Name:        strings.TrimPrefix(m[2], "$"),
				Description: strings.TrimSpace(m[3]),
			})
		}
	}

	sym := &model.Symbol{
		ID:        hookID,
		Name:      tag,
		Kind:      model.KindHook,
		Language:  "php",
		HookType:  hookType,
		HookTag:   tag,
		Doc:       doc,
		Params:    params,
		CallSites: []string{callerID},
		Location: model.SourceLocation{
			File:      file,
			StartLine: startLine(call),
			EndLine:   endLine(call),
		},
	}
	reg.Add(sym)
}

// extractHookTag resolves the hook tag string from the AST node.
// Handles simple strings, concatenation, and variable interpolation.
func extractHookTag(node *sitter.Node, src []byte) string {
	if node == nil {
		return ""
	}
	switch node.Type() {
	case "string":
		// Simple string: 'init', "save_post"
		s := nodeText(node, src)
		s = strings.Trim(s, "'\"")
		return s

	case "encapsed_string":
		// Interpolated string: "save_post_{$post->post_type}"
		var parts []string
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			switch child.Type() {
			case "string_content", "string_value":
				parts = append(parts, nodeText(child, src))
			default:
				parts = append(parts, "{$var}")
			}
		}
		if len(parts) == 0 {
			// Fallback: extract from the raw text
			s := nodeText(node, src)
			s = strings.Trim(s, "\"")
			return s
		}
		return strings.Join(parts, "")

	case "binary_expression":
		// Concatenation: 'save_post_' . $post->post_type
		left := node.ChildByFieldName("left")
		right := node.ChildByFieldName("right")
		leftStr := extractHookTag(left, src)
		rightStr := extractHookTag(right, src)
		if leftStr != "" || rightStr != "" {
			if leftStr == "" {
				leftStr = "{$var}"
			}
			if rightStr == "" {
				rightStr = "{$var}"
			}
			return leftStr + rightStr
		}
	}

	return ""
}
