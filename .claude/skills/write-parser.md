# Skill: Parser Code

Activate this skill when adding new symbol extraction, modifying AST traversal, or extending the parser in `internal/parser/`.

## Architecture

```
parser.go          Orchestrator: worker pool, file dispatch, language detection
php.go             PHP extraction: functions, classes, interfaces, traits
js.go              JS/TS extraction: functions, classes, interfaces, exports
php_hooks.go       Hook detection: do_action / apply_filters call scanning
docblock.go        Doc comment parsing: @param, @return, @since, @deprecated, @see
comment.go         Finds /** */ comments adjacent to AST nodes
helpers.go         Tree-sitter utilities: nodeText, startLine, endLine, walkTree
```

## Data Flow

```
parser.ParseFiles(files, registry)
  → spawns N workers, each with own sitter.Parser (not thread-safe)
  → each worker calls parseFile()
    → detectLanguage() picks tree-sitter grammar by file extension
    → sp.ParseCtx() produces AST
    → extractPHP() or extractJS() walks AST, creates model.Symbol structs
    → reg.Add(sym) writes to thread-safe Registry
```

## Adding a New Symbol Kind

Follow this pattern (modeled on how classes/interfaces/traits are handled):

### 1. Add the kind constant to `internal/model/model.go`

```go
const KindEnum SymbolKind = "enum"
```

### 2. Add the node type to `processNode` in the relevant language file

```go
// In php.go processNode():
case "enum_declaration":
    ctx.handleEnum(node, namespace, classStack)
```

### 3. Write the handler function

Follow the established pattern — every handler:
- Extracts the name via `node.ChildByFieldName("name")`
- Returns early if name is empty
- Builds a fully qualified ID (PHP uses `qualifyPHP(namespace, name)`)
- Calls `findDocComment(node, ctx.src)` for the docblock
- Creates a `*model.Symbol` with all applicable fields
- Calls `ctx.reg.Add(sym)`
- Recurses into the body for members (classes, interfaces, traits all do this)

```go
func (ctx *phpContext) handleEnum(node *sitter.Node, namespace string, classStack []string) {
    nameNode := node.ChildByFieldName("name")
    name := nodeText(nameNode, ctx.src)
    if name == "" {
        return
    }
    fqn := qualifyPHP(namespace, name)
    doc := findDocComment(node, ctx.src)

    sym := &model.Symbol{
        ID:        fqn,
        Name:      name,
        Kind:      model.KindEnum,
        Language:  "php",
        Namespace: namespace,
        Doc:       doc,
        Location: model.SourceLocation{
            File:      ctx.file,
            StartLine: startLine(node),
            EndLine:   endLine(node),
        },
    }
    ctx.reg.Add(sym)

    // Recurse into body for members
    body := node.ChildByFieldName("body")
    if body == nil {
        return
    }
    newStack := append(classStack, fqn)
    // ... process child nodes
}
```

### 4. Add a section entry in `hugo.go`

In the `kindSections` slice in `Generate()`:

```go
{model.KindEnum, "enums", "Enums"},
```

### 5. Add to the nav partial

In the `partialNav` constant, add a nav entry for the new section (follow the existing pattern of section links with counts).

## Tree-Sitter Patterns

### Navigating the AST

```go
node.ChildByFieldName("name")       // Named field access (grammar-defined)
node.NamedChild(i)                   // Positional named child
node.NamedChildCount()               // Count of named children
node.Type()                          // Node type string (e.g., "function_definition")
node.StartByte() / node.EndByte()    // Byte offsets in source
```

### Helper functions (`helpers.go`)

```go
nodeText(node, src)          // Extract source text of a node
startLine(node) / endLine(node)  // 1-based line numbers
childByType(node, "type")    // First named child matching type
childrenByType(node, "type") // All named children matching type
walkTree(node, fn)           // Depth-first traversal over all named nodes
```

### Finding the right node types

Use the tree-sitter playground or dump the AST to discover node types for a given language grammar. PHP node types differ from JS/TS node types. Common ones:

**PHP:** `function_definition`, `class_declaration`, `interface_declaration`, `trait_declaration`, `method_declaration`, `property_declaration`, `namespace_definition`, `function_call_expression`, `formal_parameters`

**JS/TS:** `function_declaration`, `class_declaration`, `interface_declaration`, `method_definition`, `public_field_definition`, `export_statement`, `lexical_declaration`, `variable_declaration`, `formal_parameters`

## Conventions

- **Each worker gets its own `sitter.Parser`** — tree-sitter parsers are not thread-safe. Never share a parser across goroutines.
- **The Registry is thread-safe.** `reg.Add()` acquires a write lock. Safe to call from multiple workers.
- **PHP uses fully qualified IDs.** `qualifyPHP(namespace, name)` produces e.g., `WP_REST_Server::dispatch`. JS symbols use simple names.
- **Docblock association uses byte-offset search**, not AST parent traversal. `findDocComment()` in `comment.go` searches backwards from the node's start byte for the nearest `/** */` block, then verifies nothing but whitespace/modifiers sits between them.
- **Hook detection scans function bodies.** After extracting a PHP function/method, `scanForHooks()` walks its body AST looking for `do_action`/`apply_filters` calls. Hooks found from multiple call sites get their `CallSites` list appended to.
