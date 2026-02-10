package parser

import (
	"regexp"
	"strings"

	"github.com/peter/wpdocs/internal/model"
)

var (
	tagRegex    = regexp.MustCompile(`^@(\w+)\s*(.*)$`)
	paramRegex  = regexp.MustCompile(`^@param\s+(\S+)\s+(\$\w+)\s*(.*)$`)
	returnRegex = regexp.MustCompile(`^@return\s+(\S+)\s*(.*)$`)
	sinceRegex  = regexp.MustCompile(`^@since\s+(.+)$`)
)

// ParseDocBlock parses a PHPDoc comment block into a structured DocBlock.
func ParseDocBlock(raw string) model.DocBlock {
	doc := model.DocBlock{
		Tags: make(map[string][]string),
	}

	// Strip comment delimiters
	raw = strings.TrimPrefix(raw, "/**")
	raw = strings.TrimSuffix(raw, "*/")

	lines := strings.Split(raw, "\n")
	var cleaned []string
	for _, line := range lines {
		// Strip leading whitespace to find the * prefix, but preserve
		// indentation AFTER "* " so code blocks remain intact.
		trimmed := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmed, "* ") {
			line = trimmed[2:] // remove "* ", keep remaining indent
		} else if strings.HasPrefix(trimmed, "*") {
			line = trimmed[1:] // lone "*" (blank comment line)
		} else {
			line = trimmed
		}
		cleaned = append(cleaned, line)
	}

	// Separate summary, description, and tags
	var (
		summary     []string
		description []string
		inTags      bool
		currentTag  string
		tagLines    []string
	)

	flushTag := func() {
		if currentTag != "" {
			// Trim each continuation line to collapse aligned whitespace
			var trimmed []string
			for _, tl := range tagLines {
				trimmed = append(trimmed, strings.TrimSpace(tl))
			}
			doc.Tags[currentTag] = append(doc.Tags[currentTag], strings.Join(trimmed, " "))
		}
		currentTag = ""
		tagLines = nil
	}

	for _, line := range cleaned {
		if strings.HasPrefix(line, "@") {
			inTags = true
			flushTag()

			// Parse specific tag types
			if m := sinceRegex.FindStringSubmatch(line); m != nil {
				if doc.Since == "" {
					doc.Since = strings.TrimSpace(m[1])
				}
				currentTag = "since"
				tagLines = []string{m[1]}
			} else if strings.HasPrefix(line, "@deprecated") {
				doc.Deprecated = strings.TrimPrefix(line, "@deprecated")
				doc.Deprecated = strings.TrimSpace(doc.Deprecated)
				currentTag = "deprecated"
				tagLines = []string{doc.Deprecated}
			} else if strings.HasPrefix(line, "@see") {
				ref := strings.TrimSpace(strings.TrimPrefix(line, "@see"))
				doc.SeeAlso = append(doc.SeeAlso, ref)
				currentTag = "see"
				tagLines = []string{ref}
			} else if strings.HasPrefix(line, "@link") {
				link := strings.TrimSpace(strings.TrimPrefix(line, "@link"))
				doc.Links = append(doc.Links, link)
				currentTag = "link"
				tagLines = []string{link}
			} else if strings.HasPrefix(line, "@access") {
				doc.Access = strings.TrimSpace(strings.TrimPrefix(line, "@access"))
				currentTag = "access"
				tagLines = []string{doc.Access}
			} else if m := tagRegex.FindStringSubmatch(line); m != nil {
				currentTag = m[1]
				tagLines = []string{m[2]}
			}
			continue
		}

		if inTags {
			// Continuation of a tag
			if line != "" {
				tagLines = append(tagLines, line)
			}
			continue
		}

		// Summary/description
		if len(summary) == 0 && line == "" {
			continue // skip leading blank lines
		}
		if len(summary) > 0 && line == "" && len(description) == 0 {
			// Blank line after summary = start of description
			description = append(description, "")
			continue
		}

		if len(description) > 0 || (len(summary) > 0 && line == "") {
			description = append(description, line)
		} else {
			summary = append(summary, line)
		}
	}
	flushTag()

	doc.Summary = strings.TrimSpace(strings.Join(summary, " "))
	doc.Description = strings.TrimSpace(strings.Join(description, "\n"))

	return doc
}

// ParseParams extracts @param tags into structured Param entries.
func ParseParams(doc model.DocBlock) []model.Param {
	var params []model.Param
	for _, raw := range doc.Tags["param"] {
		if m := paramRegex.FindStringSubmatch("@param " + raw); m != nil {
			p := model.Param{
				Type:        m[1],
				Name:        strings.TrimPrefix(m[2], "$"),
				Description: strings.TrimSpace(m[3]),
			}
			if strings.HasPrefix(p.Type, "?") {
				p.IsNullable = true
				p.Type = strings.TrimPrefix(p.Type, "?")
			}
			params = append(params, p)
		}
	}
	return params
}

// ParseReturn extracts @return tag into a ReturnValue.
func ParseReturn(doc model.DocBlock) *model.ReturnValue {
	returns := doc.Tags["return"]
	if len(returns) == 0 {
		return nil
	}
	raw := returns[0]
	if m := returnRegex.FindStringSubmatch("@return " + raw); m != nil {
		return &model.ReturnValue{
			Type:        m[1],
			Description: strings.TrimSpace(m[2]),
		}
	}
	return &model.ReturnValue{Type: raw}
}
