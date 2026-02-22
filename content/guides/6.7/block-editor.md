---
title: "Block Editor"
summary: "Block development in WordPress 6.7 — Block Bindings, style variations, and the Interactivity API."
weight: 10
---

## What's New in 6.7

WordPress 6.7 brought major enhancements to the block editor:

- **Block Bindings API** reaches stable status
- **Style variations** for themes with expanded presets
- **Interactivity API** improvements with new directives
- **Zoom-out view** for full-page editing

## Block Bindings API

The Block Bindings API (stable in 6.7) lets you connect block attributes to dynamic data sources — custom fields, site options, or any registered source.

### Registering a Binding Source

```php
add_action('init', function () {
    register_block_bindings_source('myplugin/user-data', [
        'label'              => 'User Data',
        'get_value_callback' => function ($source_args, $block_instance) {
            $field = $source_args['field'] ?? '';
            $user  = wp_get_current_user();

            return match ($field) {
                'name'  => $user->display_name,
                'email' => $user->user_email,
                'role'  => implode(', ', $user->roles),
                default => '',
            };
        },
    ]);
});
```

### Using Bindings in Templates

In the block editor or theme templates, bind a paragraph's content to a custom field:

```html
<!-- wp:paragraph {"metadata":{"bindings":{"content":{"source":"core/post-meta","args":{"key":"subtitle"}}}}} -->
<p></p>
<!-- /wp:paragraph -->
```

Or bind an image to a custom field URL:

```html
<!-- wp:image {"metadata":{"bindings":{"url":{"source":"core/post-meta","args":{"key":"hero_image"}}}}} -->
<figure class="wp-block-image"><img src="" alt=""/></figure>
<!-- /wp:image -->
```

### Supported Core Sources

| Source | Description |
|--------|-------------|
| `core/post-meta` | Post custom fields |
| `core/pattern-overrides` | Pattern override values |

## Style Variations

WordPress 6.7 expands theme style variations with more preset control:

```json
// styles/ocean.json
{
  "version": 3,
  "title": "Ocean",
  "settings": {
    "color": {
      "palette": [
        { "slug": "primary", "color": "#0077b6", "name": "Primary" },
        { "slug": "secondary", "color": "#90e0ef", "name": "Secondary" }
      ]
    },
    "typography": {
      "fontFamilies": [
        {
          "fontFamily": "'Merriweather', serif",
          "slug": "heading",
          "name": "Heading"
        }
      ]
    }
  },
  "styles": {
    "color": {
      "background": "#caf0f8",
      "text": "#023e8a"
    }
  }
}
```

Place style variation JSON files in your theme's `styles/` directory.

## Interactivity API

The Interactivity API (introduced 6.5, improved in 6.7) enables reactive client-side interactions for blocks without custom JavaScript build steps.

### Store Definition

```js
// view.js
import { store, getContext } from '@wordpress/interactivity';

store('myplugin/counter', {
    state: {
        get doubleCount() {
            const ctx = getContext();
            return ctx.count * 2;
        },
    },
    actions: {
        increment() {
            const ctx = getContext();
            ctx.count++;
        },
        decrement() {
            const ctx = getContext();
            ctx.count--;
        },
    },
});
```

### Block Markup

```php
<?php
// render.php
$context = json_encode(['count' => 0]);
?>
<div
    <?php echo get_block_wrapper_attributes(); ?>
    data-wp-interactive="myplugin/counter"
    data-wp-context='<?php echo esc_attr($context); ?>'
>
    <button data-wp-on--click="actions.increment">+</button>
    <span data-wp-text="context.count"></span>
    <button data-wp-on--click="actions.decrement">-</button>
    <p>Double: <span data-wp-text="state.doubleCount"></span></p>
</div>
```

### Directives Reference

| Directive | Purpose |
|-----------|---------|
| `data-wp-interactive` | Declare the interactive namespace |
| `data-wp-context` | Set local reactive context |
| `data-wp-on--{event}` | Bind event handlers |
| `data-wp-bind--{attr}` | Bind HTML attributes reactively |
| `data-wp-text` | Set element text content reactively |
| `data-wp-class--{name}` | Toggle CSS classes reactively |
| `data-wp-watch` | Run side effects on state changes |

## Creating a Custom Block

```bash
npx @wordpress/create-block my-block --variant dynamic
cd my-block
npm start
```

### block.json (6.7)

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
  "render": "file:./render.php",
  "viewScriptModule": "file:./view.js"
}
```

Note: `viewScriptModule` (not `viewScript`) loads the view script as an ES module, required for the Interactivity API.

> Browse the [Functions reference](/{{version}}/functions/) for `register_block_bindings_source`, `register_block_type`, and related functions.
