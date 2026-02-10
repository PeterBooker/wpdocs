package model

import (
	"sync"
)

// SymbolKind identifies what type of code symbol this is.
type SymbolKind string

const (
	KindFunction  SymbolKind = "function"
	KindClass     SymbolKind = "class"
	KindMethod    SymbolKind = "method"
	KindProperty  SymbolKind = "property"
	KindConstant  SymbolKind = "constant"
	KindInterface SymbolKind = "interface"
	KindTrait     SymbolKind = "trait"
	KindEnum      SymbolKind = "enum"
	KindHook      SymbolKind = "hook"
	KindComponent SymbolKind = "component" // React components in Gutenberg
)

// HookType distinguishes actions from filters.
type HookType string

const (
	HookAction HookType = "action"
	HookFilter HookType = "filter"
)

// Param represents a function/method parameter.
type Param struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Default     string `json:"default,omitempty"`
	IsVariadic  bool   `json:"is_variadic,omitempty"`
	IsNullable  bool   `json:"is_nullable,omitempty"`
	IsPassByRef bool   `json:"is_pass_by_ref,omitempty"`
}

// ReturnValue represents a function/method return.
type ReturnValue struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// DocBlock represents a parsed documentation comment.
type DocBlock struct {
	Summary     string              `json:"summary"`
	Description string              `json:"description"`
	Tags        map[string][]string `json:"tags,omitempty"`
	Since       string              `json:"since,omitempty"`
	Deprecated  string              `json:"deprecated,omitempty"`
	SeeAlso     []string            `json:"see_also,omitempty"`
	Links       []string            `json:"links,omitempty"`
	Access      string              `json:"access,omitempty"` // public, private, protected
}

// SourceLocation pinpoints where a symbol is defined.
type SourceLocation struct {
	File      string `json:"file"`       // Relative to WP root
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

// Symbol is the unified representation of any documented code entity.
type Symbol struct {
	// Identity
	ID        string     `json:"id"`       // Fully qualified: e.g., "wp_insert_post" or "WP_Query::query"
	Name      string     `json:"name"`     // Short name
	Kind      SymbolKind `json:"kind"`
	Language  string     `json:"language"` // "php" or "js"
	Namespace string     `json:"namespace,omitempty"`

	// Documentation
	Doc DocBlock `json:"doc"`

	// For functions/methods
	Params  []Param      `json:"params,omitempty"`
	Returns *ReturnValue `json:"returns,omitempty"`

	// For classes/interfaces/traits
	Extends    []string `json:"extends,omitempty"`
	Implements []string `json:"implements,omitempty"`
	Members    []string `json:"members,omitempty"`  // IDs of child symbols (methods, properties)
	ParentID   string   `json:"parent_id,omitempty"` // For methods: the owning class ID

	// For hooks
	HookType  HookType `json:"hook_type,omitempty"`
	HookTag   string   `json:"hook_tag,omitempty"`   // The hook name/tag string
	CallSites []string `json:"call_sites,omitempty"` // Where do_action/apply_filters is called

	// Cross-references (populated by resolver)
	UsedBy    []string `json:"used_by,omitempty"`   // Symbols that call this
	Uses      []string `json:"uses,omitempty"`      // Symbols this calls
	Overrides string   `json:"overrides,omitempty"` // Parent method this overrides

	// Source
	Location SourceLocation `json:"location"`
}

// Registry is the central store for all extracted symbols.
type Registry struct {
	mu      sync.RWMutex
	symbols map[string]*Symbol
	byKind  map[SymbolKind][]*Symbol
	byFile  map[string][]*Symbol
}

func NewRegistry() *Registry {
	return &Registry{
		symbols: make(map[string]*Symbol),
		byKind:  make(map[SymbolKind][]*Symbol),
		byFile:  make(map[string][]*Symbol),
	}
}

func (r *Registry) Add(s *Symbol) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.symbols[s.ID] = s
	r.byKind[s.Kind] = append(r.byKind[s.Kind], s)
	r.byFile[s.Location.File] = append(r.byFile[s.Location.File], s)
}

func (r *Registry) Get(id string) *Symbol {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.symbols[id]
}

func (r *Registry) ByKind(k SymbolKind) []*Symbol {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.byKind[k]
}

func (r *Registry) ByFile(path string) []*Symbol {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.byFile[path]
}

func (r *Registry) All() []*Symbol {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*Symbol, 0, len(r.symbols))
	for _, s := range r.symbols {
		result = append(result, s)
	}
	return result
}

func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.symbols)
}

func (r *Registry) CountByLanguage(lang string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, s := range r.symbols {
		if s.Language == lang {
			count++
		}
	}
	return count
}
