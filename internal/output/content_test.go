package output

import (
	"testing"

	"github.com/peter/wpdocs/internal/model"
)

func TestBuildSignature_Function(t *testing.T) {
	sym := &model.Symbol{
		Name: "wp_insert_post",
		Kind: model.KindFunction,
		Params: []model.Param{
			{Name: "postarr", Type: "array"},
			{Name: "wp_error", Type: "bool", Default: "false"},
		},
		Returns: &model.ReturnValue{Type: "int|WP_Error"},
	}
	got := buildSignature(sym)
	want := "wp_insert_post( array $postarr, bool $wp_error = false ): int|WP_Error"
	if got != want {
		t.Errorf("buildSignature(function) =\n  %q\nwant\n  %q", got, want)
	}
}

func TestBuildSignature_FunctionPassByRef(t *testing.T) {
	sym := &model.Symbol{
		Name: "sort_array",
		Kind: model.KindFunction,
		Params: []model.Param{
			{Name: "arr", Type: "array", IsPassByRef: true},
		},
	}
	got := buildSignature(sym)
	want := "sort_array( array &$arr )"
	if got != want {
		t.Errorf("buildSignature(pass-by-ref) =\n  %q\nwant\n  %q", got, want)
	}
}

func TestBuildSignature_HookAction(t *testing.T) {
	sym := &model.Symbol{
		Name:     "init",
		Kind:     model.KindHook,
		HookType: model.HookAction,
		HookTag:  "init",
	}
	got := buildSignature(sym)
	want := "do_action( 'init' )"
	if got != want {
		t.Errorf("buildSignature(action) = %q, want %q", got, want)
	}
}

func TestBuildSignature_HookFilter(t *testing.T) {
	sym := &model.Symbol{
		Name:     "the_content",
		Kind:     model.KindHook,
		HookType: model.HookFilter,
		HookTag:  "the_content",
		Params: []model.Param{
			{Name: "content", Type: "string"},
		},
	}
	got := buildSignature(sym)
	want := "apply_filters( 'the_content', string $content )"
	if got != want {
		t.Errorf("buildSignature(filter) = %q, want %q", got, want)
	}
}

func TestBuildSignature_Class(t *testing.T) {
	sym := &model.Symbol{
		Name:       "WP_Query",
		Kind:       model.KindClass,
		Extends:    []string{"WP_Object"},
		Implements: []string{"Countable", "Iterator"},
	}
	got := buildSignature(sym)
	want := "class WP_Query extends WP_Object implements Countable, Iterator"
	if got != want {
		t.Errorf("buildSignature(class) = %q, want %q", got, want)
	}
}

func TestParseChangelog_MultipleSince(t *testing.T) {
	sym := &model.Symbol{
		Doc: model.DocBlock{
			Tags: map[string][]string{
				"since": {"2.1.0 Introduced.", "4.7.0 Added $post_type parameter."},
			},
		},
	}
	entries := parseChangelog(sym)
	if len(entries) != 2 {
		t.Fatalf("parseChangelog: got %d entries, want 2", len(entries))
	}
	if entries[0].Version != "2.1.0" || entries[0].Description != "Introduced." {
		t.Errorf("entry[0] = %+v", entries[0])
	}
	if entries[1].Version != "4.7.0" || entries[1].Description != "Added $post_type parameter." {
		t.Errorf("entry[1] = %+v", entries[1])
	}
}

func TestParseChangelog_SinceFieldOnly(t *testing.T) {
	sym := &model.Symbol{
		Doc: model.DocBlock{Since: "3.0.0"},
	}
	entries := parseChangelog(sym)
	if len(entries) != 1 {
		t.Fatalf("parseChangelog: got %d entries, want 1", len(entries))
	}
	if entries[0].Version != "3.0.0" || entries[0].Description != "Introduced." {
		t.Errorf("entry = %+v", entries[0])
	}
}

func TestParseChangelog_Empty(t *testing.T) {
	sym := &model.Symbol{
		Doc: model.DocBlock{},
	}
	entries := parseChangelog(sym)
	if len(entries) != 0 {
		t.Fatalf("parseChangelog: got %d entries, want 0", len(entries))
	}
}

func TestParseChangelog_VersionOnly(t *testing.T) {
	sym := &model.Symbol{
		Doc: model.DocBlock{
			Tags: map[string][]string{
				"since": {"5.0.0"},
			},
		},
	}
	entries := parseChangelog(sym)
	if len(entries) != 1 {
		t.Fatalf("parseChangelog: got %d entries, want 1", len(entries))
	}
	if entries[0].Version != "5.0.0" || entries[0].Description != "Introduced." {
		t.Errorf("entry = %+v", entries[0])
	}
}
