package output

import "testing"

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"6.7.1", "6.7"},
		{"6.7", "6.7"},
		{"6", "6"},
		{"", ""},
		{"10.0.3", "10.0"},
	}
	for _, tt := range tests {
		if got := normalizeVersion(tt.input); got != tt.want {
			t.Errorf("normalizeVersion(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int // >0, <0, or 0
	}{
		{"6.7", "6.7", 0},
		{"6.8", "6.7", 1},
		{"6.7", "6.8", -1},
		{"10.0", "9.9", 1},
		{"6.7.1", "6.7", 1},
		{"6.7", "6.7.1", -1},
		{"1.0", "2.0", -1},
	}
	for _, tt := range tests {
		got := compareVersions(tt.a, tt.b)
		switch {
		case tt.want > 0 && got <= 0:
			t.Errorf("compareVersions(%q, %q) = %d, want >0", tt.a, tt.b, got)
		case tt.want < 0 && got >= 0:
			t.Errorf("compareVersions(%q, %q) = %d, want <0", tt.a, tt.b, got)
		case tt.want == 0 && got != 0:
			t.Errorf("compareVersions(%q, %q) = %d, want 0", tt.a, tt.b, got)
		}
	}
}

func TestYamlEscape(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", `""`},
		{"hello", `"hello"`},
		{`has "quotes"`, `"has \"quotes\""`},
		{"has\nnewline", "\"has\nnewline\""},
		{`has\backslash`, `"has\\backslash"`},
		{"colon: value", `"colon: value"`},
		{"hash # comment", `"hash # comment"`},
	}
	for _, tt := range tests {
		if got := yamlEscape(tt.input); got != tt.want {
			t.Errorf("yamlEscape(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestYamlMultiline(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", `""`},
		{"single line", `"single line"`},
		{"line1\nline2", `"line1\nline2"`},
		{`has "quotes"`, `"has \"quotes\""`},
		{"has\ttab", `"has\ttab"`},
		{`back\slash`, `"back\\slash"`},
	}
	for _, tt := range tests {
		if got := yamlMultiline(tt.input); got != tt.want {
			t.Errorf("yamlMultiline(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSafeContent(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"normal text", "normal text"},
		{"<script>alert(1)</script>", "&lt;script>alert(1)&lt;/script>"},
		{"<SCRIPT>bad</SCRIPT>", "&lt;SCRIPT>bad&lt;/SCRIPT>"},
		{"<Script>mixed</Script>", "&lt;Script>mixed&lt;/Script>"},
		{"{{ .Title }}", "&#123;&#123; .Title &#125;&#125;"},
		{"no braces", "no braces"},
	}
	for _, tt := range tests {
		if got := safeContent(tt.input); got != tt.want {
			t.Errorf("safeContent(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSymbolSlug(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"wp_insert_post", "wp_insert_post"},
		{"WP_Query::query", "wp_query.query"},
		{`Namespace\Class`, "namespace.class"},
		{"has spaces", "has-spaces"},
		{"remove$parens()", "removeparens"},
		{"", "unnamed"},
		{"MixedCase_Func", "mixedcase_func"},
	}
	for _, tt := range tests {
		if got := symbolSlug(tt.input); got != tt.want {
			t.Errorf("symbolSlug(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
