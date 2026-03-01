# Skill: Write WordPress Documentation

Activate this skill when writing or editing guide pages in `content/guides/` or override pages in `content/overrides/`. The goal is documentation that matches the approachability and completeness of the Laravel docs.

## Voice & Tone

**Confident, not hedging.** Say "WordPress stores options in the `wp_options` table" — not "WordPress typically tends to store options in what is usually the `wp_options` table."

**Conversational, uses "you" freely.** Address the reader as a capable developer who simply hasn't learned this particular thing yet. Use phrases like:
- "If you would like to..." (optional features)
- "Of course, you may also..." (alternatives)
- "You are free to..." (developer choice)
- "Let's take a look at..." (transitioning to examples)
- "Before getting started, make sure..." (prerequisites)

**Uses "we" for WordPress's perspective** — sparingly, to convey shared purpose: "We recommend using the Settings API for..."

**Never minimizes complexity.** Don't use "simply," "just," "easy," "straightforward," or "obviously." Present the simplest path first and let the reader judge.

## Page Structure

Every guide page follows this predictable structure:

### 1. Front Matter + Title

```markdown
---
title: "Hooks"
summary: "Understanding WordPress actions and filters."
weight: 2
---
```

Title is a clean noun or noun phrase — "Hooks", not "Working with Hooks."

### 2. Overview Section

One to three paragraphs answering: what is this, why does it exist, when would you use it?

Lead with practical purpose, not a dictionary definition:

> Hooks are WordPress's mechanism for allowing plugins and themes to modify or extend core behavior without editing core files.

Not:

> A hook is a programming concept used in WordPress development.

### 3. Body Sections

H2 for major concepts, H3 for subsections. Rarely go deeper than H4 — if you need to, the page may need splitting.

Within each section, follow this progression:
1. Brief explanation (1-2 paragraphs)
2. Simplest code example
3. Walk through what the code does
4. Progressive complexity — additional parameters, edge cases, real-world patterns
5. Cross-references to related pages

### 4. Callouts

Use sparingly, only for things that will genuinely trip people up:

```markdown
> **Warning**
> Calling `remove_action` must happen at the same priority or later than the `add_action` call it targets.

> **Note**
> This feature was introduced in WordPress 6.0. Check `function_exists()` if supporting older versions.
```

## Code Examples

**Every concept gets one.** Don't describe in prose what code would make clearer.

**Show the simplest version first**, then build to complexity in subsequent examples.

**Use realistic names** — `book`, `event`, `product` for post types; `genre`, `event_type` for taxonomies. Never `foo`, `bar`, `my_function`.

**Include full context** — hook registration, function signatures, everything needed to drop the code into a project.

**Follow WordPress PHP coding standards:**
- Spaces inside parentheses: `array( 'key' => 'value' )`
- Yoda conditions: `if ( true === $value )`
- Snake_case functions: `register_book_post_type`
- Tab indentation (4 spaces in docs)
- Single quotes unless interpolating

**Use fenced code blocks** with language hints: `php`, `js`, `bash`, `json`, `html`.

## Cross-Referencing

Link every mention of a concept documented elsewhere — weave links naturally into prose:

> Custom post types are registered during the [`init` hook](/docs/hooks#actions). Each post type can have its own [taxonomies](/docs/taxonomies).

Don't create "See Also" sections at the bottom.

## Version Awareness

Note when features were introduced or changed:

> The `register_post_type_args` filter, introduced in WordPress 4.4, allows you to modify the arguments of any registered post type.

## Page Length

Comprehensive but focused. One page per concept — cover all its facets, but cross-reference rather than duplicate. If a page exceeds ~3,500 words, consider splitting.

## Markdown Rules

- ATX-style headers (`#`, `##`, `###`)
- Standard Markdown links: `[text](/docs/page-slug#anchor)`
- Backtick inline code for function names, file paths, hook names, class names, parameters, table names
- No HTML unless absolutely necessary

## Content Inventory

### Existing shared guides (`content/guides/_shared/`)
getting-started, hooks, theme-development, database, rest-api, plugins, security, custom-post-types, caching, http-api, internationalization, users-roles, cron, wp-cli

### Version-specific guides (`content/guides/{version}/`)
block-editor, whats-new

### Planned (not yet written)
- Installation (local dev: Local, DDEV, wp-env)
- Configuration (wp-config.php)
- Directory Structure
- Request Lifecycle
- The Loop
- Template Hierarchy
- Routing and Rewrites
- Options API
- Shortcodes
- Widgets / Menus
- Enqueuing Scripts and Styles
- Error Handling and Debugging
- Transients
- Localization (separate from internationalization?)
- Email (wp_mail)
- Admin Notices
- File Handling (WP_Filesystem)
- Nonces and CSRF
- Data Sanitization and Escaping
- Custom Tables
- Pagination
- dbDelta / Schema Migrations
- Object Cache
- Block Patterns / InnerBlocks / Full Site Editing
- Testing (WP_UnitTestCase)
- Multisite
