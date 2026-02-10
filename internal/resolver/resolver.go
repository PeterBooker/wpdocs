package resolver

import (
	"strings"

	"github.com/peter/wpdocs/internal/model"
)

// Stats tracks cross-reference resolution metrics.
type Stats struct {
	Resolved     int
	Unresolved   int
	Inheritance  int
	HookBindings int
}

// Resolver connects symbols via cross-references, inheritance, and hook bindings.
type Resolver struct {
	registry *model.Registry
	stats    Stats
}

func New(reg *model.Registry) *Resolver {
	return &Resolver{registry: reg}
}

func (r *Resolver) Stats() Stats { return r.stats }

// ResolveAll performs all cross-reference resolution passes.
func (r *Resolver) ResolveAll() {
	r.resolveInheritance()
	r.resolveHookBindings()
	r.resolveSeeReferences()
	r.resolveMethodOverrides()
}

// resolveInheritance connects extends/implements to actual symbol IDs.
func (r *Resolver) resolveInheritance() {
	for _, sym := range r.registry.All() {
		if sym.Kind != model.KindClass && sym.Kind != model.KindInterface {
			continue
		}

		for i, ext := range sym.Extends {
			if resolved := r.findSymbol(ext); resolved != nil {
				sym.Extends[i] = resolved.ID
				r.stats.Inheritance++
				r.stats.Resolved++
			}
		}
		for i, impl := range sym.Implements {
			if resolved := r.findSymbol(impl); resolved != nil {
				sym.Implements[i] = resolved.ID
				r.stats.Inheritance++
				r.stats.Resolved++
			}
		}
	}
}

// resolveHookBindings links add_action/add_filter calls to hook definitions.
func (r *Resolver) resolveHookBindings() {
	hooks := r.registry.ByKind(model.KindHook)
	hooksByTag := make(map[string]*model.Symbol)
	for _, h := range hooks {
		hooksByTag[h.HookTag] = h
	}

	for _, sym := range r.registry.All() {
		if sym.Kind != model.KindFunction && sym.Kind != model.KindMethod {
			continue
		}
		for _, hookID := range sym.Uses {
			if hook, ok := hooksByTag[hookID]; ok {
				hook.UsedBy = appendUnique(hook.UsedBy, sym.ID)
				r.stats.HookBindings++
				r.stats.Resolved++
			}
		}
	}
}

// resolveSeeReferences resolves @see tags to symbol IDs.
func (r *Resolver) resolveSeeReferences() {
	for _, sym := range r.registry.All() {
		for i, ref := range sym.Doc.SeeAlso {
			// Try to resolve the reference to an actual symbol
			ref = strings.TrimSpace(ref)
			if ref == "" {
				continue
			}
			// Strip trailing () for function references
			cleanRef := strings.TrimSuffix(ref, "()")
			if resolved := r.findSymbol(cleanRef); resolved != nil {
				sym.Doc.SeeAlso[i] = resolved.ID
				r.stats.Resolved++
			} else {
				r.stats.Unresolved++
			}
		}
	}
}

// resolveMethodOverrides finds parent methods that child methods override.
func (r *Resolver) resolveMethodOverrides() {
	for _, sym := range r.registry.All() {
		if sym.Kind != model.KindMethod || sym.ParentID == "" {
			continue
		}

		parent := r.registry.Get(sym.ParentID)
		if parent == nil {
			continue
		}

		// Walk up the inheritance chain
		for _, extID := range parent.Extends {
			extSym := r.registry.Get(extID)
			if extSym == nil {
				continue
			}
			// Look for a method with the same name in the parent class
			parentMethodID := extID + "::" + sym.Name
			if parentMethod := r.registry.Get(parentMethodID); parentMethod != nil {
				sym.Overrides = parentMethod.ID
				parentMethod.UsedBy = appendUnique(parentMethod.UsedBy, sym.ID)
				r.stats.Resolved++
				break
			}
		}
	}
}

// findSymbol attempts to locate a symbol by name, trying various qualification strategies.
func (r *Resolver) findSymbol(name string) *model.Symbol {
	// Direct lookup
	if s := r.registry.Get(name); s != nil {
		return s
	}

	// Try with backslash-separated namespace
	if s := r.registry.Get(strings.ReplaceAll(name, "/", "\\")); s != nil {
		return s
	}

	// Try matching just the short name against all symbols
	shortName := name
	if idx := strings.LastIndexAny(name, "\\/:"); idx >= 0 {
		shortName = name[idx+1:]
	}

	// Search all symbols for a match by short name
	var candidates []*model.Symbol
	for _, sym := range r.registry.All() {
		if sym.Name == shortName {
			candidates = append(candidates, sym)
		}
	}

	if len(candidates) == 1 {
		return candidates[0]
	}

	// If multiple candidates, prefer same language
	// (this is a heuristic; could be improved with namespace context)

	return nil
}

func appendUnique(slice []string, val string) []string {
	for _, s := range slice {
		if s == val {
			return slice
		}
	}
	return append(slice, val)
}
