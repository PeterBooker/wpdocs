# Skill: Write Symbol Overrides

Activate this skill when creating or editing files in `content/overrides/`. Overrides provide hand-curated documentation that supplements auto-generated symbol reference pages.

## How Overrides Work

The generator (`internal/output/hugo.go:readOverride`) looks for a markdown file matching the symbol's section and slug. If found, its content is appended to the auto-generated page inside a `<div class="override-content">` block, after the auto-generated description.

**Resolution order:** version-specific file wins over `_shared/`.

```
content/overrides/
  _shared/
    functions/add_action.md        ← applies to all versions
    classes/wp_query.md
  6.8/
    functions/wp_enqueue_script.md  ← only for 6.8, overrides _shared/ if both exist
```

## File Path Convention

```
content/overrides/{_shared|version}/{section}/{slug}.md
```

- **section** matches the symbol kind directory: `functions`, `classes`, `methods`, `hooks`, `interfaces`, `traits`, `enums`, `components`
- **slug** is the symbol ID lowercased with `::` → `.`, `\` → `.`, `$` stripped (see `symbolSlug` in hugo.go)
  - `wp_insert_post` → `wp_insert_post.md`
  - `WP_Query` → `wp_query.md`
  - `WP_Query::query` → `wp_query.query.md`

## File Format

Override files are **plain markdown with no front matter**. They get injected into the page body directly. The auto-generated page already provides the title, signature, parameters, return type, source links, and changelog — the override adds what the auto-extraction can't: context, examples, and gotchas.

## Content Structure

Follow this pattern (all sections optional — include what's useful):

```markdown
## Examples & Best Practices

### Basic Usage

\`\`\`php
// Simplest working example
\`\`\`

### [Scenario-specific heading]

\`\`\`php
// More complex example
\`\`\`

### Common Gotchas

- Bullet points covering non-obvious behavior
- Version-specific quirks
- Related functions that are often confused
```

## Writing Guidelines

- **Don't repeat what's auto-generated.** The page already has the function signature, parameter list, return type, source location, and changelog. Focus on what machines can't extract: practical examples, edge cases, and "why."
- **Start with an H2.** The page already has an H1 (the symbol name) and an auto-generated Description section. Override content appears after those.
- **Use realistic examples.** Same rules as guides: `book`, `product`, `event` — not `foo`, `bar`.
- **Follow WordPress PHP coding standards** in code blocks (spaces inside parentheses, Yoda conditions, snake_case).
- **Keep it focused.** An override for `add_action` covers `add_action`. Link to related symbols rather than documenting them inline.
- **Prefer `_shared/`** unless the content is genuinely version-specific (e.g., a parameter added in 6.7, a behavior change in 6.8).

## Existing Overrides

Shared: `add_action`, `register_post_type`, `wp_insert_post`, `wp_query` (class)
Version-specific: `wp_enqueue_script` (6.7, 6.8), `register_block_type` (6.7), `wp_enqueue_block_style` (6.6), `wp_query` (6.6, 6.8)
