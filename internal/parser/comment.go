package parser

import (
	"strings"

	"github.com/peter/wpdocs/internal/model"
	sitter "github.com/smacker/go-tree-sitter"
)

// findDocComment searches backwards from a node's start position for a /** ... */ comment.
// This preserves the proven byte-offset backward search approach from the original parser.
func findDocComment(node *sitter.Node, src []byte) model.DocBlock {
	startByte := int(node.StartByte())
	if startByte <= 0 || startByte > len(src) {
		return model.DocBlock{}
	}

	chunk := string(src[:startByte])
	idx := strings.LastIndex(chunk, "/**")
	if idx == -1 {
		return model.DocBlock{}
	}

	endIdx := strings.Index(chunk[idx:], "*/")
	if endIdx == -1 {
		return model.DocBlock{}
	}

	raw := chunk[idx : idx+endIdx+2]

	// Make sure there's no code between the docblock and the node
	between := strings.TrimSpace(chunk[idx+endIdx+2:])
	if between != "" && !isOnlyWhitespaceOrModifiers(between) {
		return model.DocBlock{}
	}

	return ParseDocBlock(raw)
}

// isOnlyWhitespaceOrModifiers checks if a string contains only keyword modifiers
// that can appear between a doc comment and a declaration.
func isOnlyWhitespaceOrModifiers(s string) bool {
	modifiers := []string{
		// PHP modifiers
		"public", "private", "protected", "static", "abstract", "final", "readonly",
		// JS/TS modifiers
		"export", "default", "async", "function", "class", "interface", "const", "let", "var",
	}
	words := strings.Fields(s)
	for _, w := range words {
		found := false
		for _, m := range modifiers {
			if w == m {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
