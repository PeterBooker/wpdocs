---
title: "Block Editor"
summary: "Block development in WordPress 6.5 — Block Bindings, Interactivity API, and Font Library."
weight: 10
---

## What's New in 6.5

WordPress 6.5 is a landmark release for block development. Three major APIs reach stable: Block Bindings, the Interactivity API, and the Font Library. Together, they transform blocks from static content containers into dynamic, interactive components connected to live data.

- **Block Bindings API** connects block attributes to dynamic data sources
- **Interactivity API** provides a standard framework for frontend interactions
- **Font Library** enables in-editor font management
- **HTML API** now scans all tokens, not just tags

## Block Bindings API

The Block Bindings API lets you bind block attributes to **external data sources** — custom fields, site options, or any programmatic value. The bound value is resolved at render time, keeping the editor experience clean while the frontend shows live data.

### Registering a Source

```php
register_block_bindings_source('myplugin/product', [
    'label'              => 'Product Data',
    'get_value_callback' => function ($source_args, $block_instance, $attribute_name) {
        $post_id = $block_instance->context['postId'];
        return get_post_meta($post_id, $source_args['field'], true);
    },
    'uses_context' => ['postId'],
]);
```

### Binding in Block Markup

```html
<!-- wp:heading {
    "metadata": {
        "bindings": {
            "content": {
                "source": "myplugin/product",
                "args": { "field": "product_name" }
            }
        }
    }
} -->
<h2>Product Name</h2>
<!-- /wp:heading -->
```

The heading displays "Product Name" in the editor but renders the actual custom field value on the frontend.

### Built-in: Post Meta Source

WordPress ships with `core/post-meta` for binding to custom fields:

```html
<!-- wp:paragraph {
    "metadata": {
        "bindings": {
            "content": {
                "source": "core/post-meta",
                "args": { "key": "subtitle" }
            }
        }
    }
} -->
<p></p>
<!-- /wp:paragraph -->
```

> Custom fields must be registered with `register_post_meta()` and have `'show_in_rest' => true` to work with Block Bindings.

## Interactivity API

The Interactivity API replaces ad-hoc JavaScript in blocks with a **declarative, reactive system**. It uses HTML directives (similar to Alpine.js or Vue) and a store pattern for state management.

### Setting Up

In your `block.json`, declare interactivity support and a view script module:

```json
{
    "supports": {
        "interactivity": true
    },
    "viewScriptModule": "file:./view.js"
}
```

### Defining the Store

```js
// view.js
import { store, getContext, getElement } from '@wordpress/interactivity';

store('myPlugin', {
    state: {
        get buttonLabel() {
            const { isOpen } = getContext();
            return isOpen ? 'Close' : 'Open';
        },
    },
    actions: {
        toggle() {
            const ctx = getContext();
            ctx.isOpen = !ctx.isOpen;
        },
    },
    callbacks: {
        logState() {
            const { isOpen } = getContext();
            console.log('Panel is', isOpen ? 'open' : 'closed');
        },
    },
});
```

### Server-Side Markup

```php
<?php
// render.php
$context = wp_json_encode(['isOpen' => false]);
?>
<div
    <?php echo get_block_wrapper_attributes(); ?>
    data-wp-interactive="myPlugin"
    data-wp-context='<?php echo esc_attr($context); ?>'
>
    <button
        data-wp-on--click="actions.toggle"
        data-wp-text="state.buttonLabel"
    >
        Open
    </button>

    <div
        data-wp-bind--hidden="!context.isOpen"
        data-wp-watch="callbacks.logState"
    >
        <p>This content toggles on click.</p>
    </div>
</div>
```

### Directive Reference

| Directive | Purpose | Example |
|-----------|---------|---------|
| `data-wp-interactive` | Mark a region as interactive | `data-wp-interactive="myPlugin"` |
| `data-wp-context` | Provide reactive context | `data-wp-context='{"count":0}'` |
| `data-wp-on--{event}` | Bind event handlers | `data-wp-on--click="actions.increment"` |
| `data-wp-bind--{attr}` | Bind HTML attributes | `data-wp-bind--hidden="!state.visible"` |
| `data-wp-text` | Set text content reactively | `data-wp-text="state.label"` |
| `data-wp-class--{name}` | Toggle CSS classes | `data-wp-class--active="state.isActive"` |
| `data-wp-style--{prop}` | Set inline styles | `data-wp-style--color="state.color"` |
| `data-wp-each` | Render a list from an array | `data-wp-each="state.items"` |
| `data-wp-watch` | Run side effects on state changes | `data-wp-watch="callbacks.onUpdate"` |

### How It Differs from React

The Interactivity API is **not** a client-side rendering framework. Key differences:

1. **Server-rendered first** — HTML is generated on the server; the API adds interactivity on top
2. **Progressive enhancement** — Content works without JavaScript; directives add behavior
3. **No virtual DOM** — Uses signals for fine-grained reactivity without diffing
4. **Scoped by namespace** — Each block's store is isolated by its namespace string

## Font Library

The Font Library lets users manage typography directly in the editor. For theme developers, fonts declared in `theme.json` integrate automatically:

```json
{
    "version": 2,
    "settings": {
        "typography": {
            "fontFamilies": [
                {
                    "fontFamily": "\"Source Sans 3\", sans-serif",
                    "name": "Source Sans 3",
                    "slug": "source-sans-3",
                    "fontFace": [
                        {
                            "fontFamily": "Source Sans 3",
                            "fontWeight": "400",
                            "fontStyle": "normal",
                            "src": ["file:./assets/fonts/source-sans-3-regular.woff2"]
                        },
                        {
                            "fontFamily": "Source Sans 3",
                            "fontWeight": "700",
                            "fontStyle": "normal",
                            "src": ["file:./assets/fonts/source-sans-3-bold.woff2"]
                        }
                    ]
                }
            ]
        }
    }
}
```

Users access the Font Library in the Site Editor under **Styles > Typography**, where they can:
- Upload custom font files
- Activate or deactivate installed fonts
- Use any active font in block typography settings

## HTML API: Full Token Scanning

The HTML Tag Processor in 6.5 can now scan **every token** in an HTML document, not just tags:

```php
$processor = new WP_HTML_Tag_Processor('<p>Hello <em>world</em>!</p>');

while ($processor->next_token()) {
    $type = $processor->get_token_type();
    $name = $processor->get_token_name();
    // Visits: <p>, text "Hello ", <em>, text "world", </em>, text "!", </p>
}
```

This enables use cases like extracting text content, counting words, or processing comments and CDATA sections.

## Design Tools

New block design controls in 6.5:

- **Background image** — Size, Repeat, and Position controls for blocks with background images
- **Cover block** — Aspect Ratio setting for consistent responsive behavior
- **Shadow support** — Expanded to more block types

```json
{
    "version": 2,
    "settings": {
        "shadow": {
            "defaultPresets": true,
            "presets": [
                {
                    "name": "Card",
                    "slug": "card",
                    "shadow": "0 2px 8px rgba(0,0,0,0.12)"
                }
            ]
        }
    }
}
```
