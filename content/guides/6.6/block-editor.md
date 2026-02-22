---
title: "Block Editor"
summary: "Block development in WordPress 6.6 — Data Views, pattern overrides, and per-block CSS loading."
weight: 10
---

## What's New in 6.6

WordPress 6.6 focused on editor polish and performance:

- **Data Views** for managing posts, pages, and patterns
- **Pattern overrides** for reusable blocks with editable placeholders
- **Per-block CSS loading** for better front-end performance
- **Section styles** for applying styles to inner blocks
- **Negative margins** support in block spacing

## Pattern Overrides

Pattern overrides let you create reusable patterns where specific block attributes (content, images, links) can be customized per instance. This bridges the gap between fully locked patterns and fully editable blocks.

### Creating an Overridable Pattern

1. Create a pattern in the editor
2. Select a block within the pattern
3. In the Advanced panel, enable "Allow instance overrides"
4. The block gets a `metadata.bindings` attribute pointing to `core/pattern-overrides`

```html
<!-- wp:group {"metadata":{"name":"Hero Section","patternName":"theme/hero"}} -->
<div class="wp-block-group">
  <!-- wp:heading {"metadata":{"bindings":{"content":{"source":"core/pattern-overrides"}},"name":"Hero Title"}} -->
  <h2 class="wp-block-heading">Default Title</h2>
  <!-- /wp:heading -->

  <!-- wp:paragraph {"metadata":{"bindings":{"content":{"source":"core/pattern-overrides"}},"name":"Hero Description"}} -->
  <p>Default description text.</p>
  <!-- /wp:paragraph -->
</div>
<!-- /wp:group -->
```

When this pattern is inserted, each instance can have its own title and description while maintaining the same layout.

## Data Views

WordPress 6.6 introduces Data Views — a new interface for browsing and managing content in the Site Editor. Data Views provide grid and list layouts with filtering, sorting, and bulk actions.

### Extending Data Views

Data Views are powered by JavaScript and can be extended via the `@wordpress/dataviews` package:

```js
import { DataViews } from '@wordpress/dataviews';

const fields = [
    { id: 'title', header: 'Title', enableSorting: true },
    { id: 'author', header: 'Author', enableSorting: true },
    { id: 'date', header: 'Date', enableSorting: true },
];

function MyDataView({ items }) {
    return (
        <DataViews
            data={items}
            fields={fields}
            view={{ type: 'table' }}
            onChangeView={() => {}}
        />
    );
}
```

## Per-Block CSS Loading

WordPress 6.6 only loads the CSS for blocks actually used on the page. This reduces unused CSS and improves Core Web Vitals.

### What This Means for Block Developers

Block styles registered via `block.json` are automatically handled:

```json
{
  "name": "myplugin/card",
  "style": "file:./style-index.css",
  "editorStyle": "file:./index.css"
}
```

The `style-index.css` file is loaded only when the block appears on the page.

### Opting Out

If your block's styles depend on another block's CSS, declare the dependency:

```php
// Ensure core group styles load whenever your block loads
wp_enqueue_block_style('myplugin/card', [
    'handle' => 'myplugin-card-deps',
    'src'    => false,
    'deps'   => ['wp-block-group'],
]);
```

## Section Styles

Apply styles to a group block and all its inner blocks:

```json
// theme.json
{
  "version": 3,
  "styles": {
    "blocks": {
      "core/group": {
        "variations": {
          "section-dark": {
            "color": {
              "background": "#1a1a2e",
              "text": "#eee"
            },
            "elements": {
              "link": { "color": { "text": "#64b5f6" } },
              "heading": { "color": { "text": "#fff" } }
            }
          }
        }
      }
    }
  }
}
```

## Creating a Custom Block

```bash
npx @wordpress/create-block my-block
cd my-block
npm start
```

### block.json (6.6)

```json
{
  "$schema": "https://schemas.wp.org/trunk/block.json",
  "apiVersion": 3,
  "name": "myplugin/notice",
  "title": "Notice",
  "category": "widgets",
  "icon": "info",
  "supports": {
    "html": false,
    "color": { "background": true, "text": true },
    "spacing": { "margin": true, "padding": true }
  },
  "textdomain": "myplugin",
  "editorScript": "file:./index.js",
  "style": "file:./style-index.css",
  "render": "file:./render.php"
}
```

> Browse the [Functions reference](/{{version}}/functions/) for `register_block_type`, `register_block_pattern`, and `wp_enqueue_block_style`.
