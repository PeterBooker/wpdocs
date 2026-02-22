---
title: "What's New"
summary: "WordPress 6.5 \"Regina\" — Font Library, Block Bindings, Interactivity API, and AVIF support."
weight: 0
---

## WordPress 6.5 "Regina"

WordPress 6.5, released on April 2, 2024, delivers four major developer-facing features: the Font Library for managing typography, the Block Bindings API for connecting blocks to dynamic data, the Interactivity API for frontend interactions, and native AVIF image support. It also introduces plugin dependencies.

### Highlights

- [Font Library](#font-library) — install and manage fonts directly in the editor
- [Block Bindings API](#block-bindings-api) — bind block attributes to dynamic data sources
- [Interactivity API](#interactivity-api) — a standard framework for frontend block interactivity
- [Plugin Dependencies](#plugin-dependencies) — declare required plugins in the plugin header
- [AVIF Support](#avif-support) — next-generation image format with better compression

## Font Library

The Font Library is a built-in font management system that lets users install, activate, and manage fonts directly within the Site Editor. No more manually uploading font files or editing CSS.

For block themes, the Font Library is available under **Appearance > Editor > Styles > Typography**. Users can:

- Upload custom font files (`.woff2`, `.ttf`, `.otf`)
- Install fonts from bundled collections
- Activate or deactivate fonts per-site
- Use installed fonts in any block's typography settings

For theme developers, fonts registered via `theme.json` integrate with the Font Library automatically:

```json
{
    "version": 2,
    "settings": {
        "typography": {
            "fontFamilies": [
                {
                    "fontFamily": "\"Inter\", sans-serif",
                    "name": "Inter",
                    "slug": "inter",
                    "fontFace": [
                        {
                            "fontFamily": "Inter",
                            "fontWeight": "400",
                            "fontStyle": "normal",
                            "src": ["file:./assets/fonts/inter-regular.woff2"]
                        }
                    ]
                }
            ]
        }
    }
}
```

> The Font Library is currently available for block themes. Classic theme support is planned for a future release.

## Block Bindings API

The Block Bindings API lets you **connect block attributes to dynamic data sources** like custom fields, site options, or any programmatic value. Instead of hardcoding content, blocks pull their values from a registered source at render time.

### Registering a Custom Source

```php
register_block_bindings_source('myplugin/product-data', [
    'label'              => 'Product Data',
    'get_value_callback' => function ($source_args, $block_instance, $attribute_name) {
        $post_id = $block_instance->context['postId'];
        $field   = $source_args['field'];
        return get_post_meta($post_id, $field, true);
    },
]);
```

### Using Bindings in Block Markup

```html
<!-- wp:paragraph {
    "metadata": {
        "bindings": {
            "content": {
                "source": "myplugin/product-data",
                "args": { "field": "product_description" }
            }
        }
    }
} -->
<p>Fallback content shown in the editor</p>
<!-- /wp:paragraph -->

<!-- wp:image {
    "metadata": {
        "bindings": {
            "url": {
                "source": "myplugin/product-data",
                "args": { "field": "product_image_url" }
            }
        }
    }
} /-->
```

### Built-in Sources

WordPress ships with a `core/post-meta` source that binds blocks to custom fields:

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

> Register your custom fields with `register_post_meta()` and set `show_in_rest` to `true` for them to work with Block Bindings.

## Interactivity API

The Interactivity API provides a **standard, lightweight framework** for adding frontend interactivity to blocks. It replaces ad-hoc JavaScript with a declarative system of directives, stores, and reactive state.

### Declaring Interactivity Support

```json
{
    "supports": {
        "interactivity": true
    },
    "viewScriptModule": "file:./view.js"
}
```

### Defining a Store

```js
import { store, getContext } from '@wordpress/interactivity';

store('myPlugin', {
    state: {
        get isOpen() {
            return getContext().isOpen;
        },
    },
    actions: {
        toggle() {
            const ctx = getContext();
            ctx.isOpen = !ctx.isOpen;
        },
    },
});
```

### Using Directives in Markup

```php
<?php
// render.php
$context = wp_json_encode(['isOpen' => false]);
?>
<div
    <?php echo get_block_wrapper_attributes(); ?>
    data-wp-interactive="myPlugin"
    data-wp-context='<?php echo $context; ?>'
>
    <button data-wp-on--click="actions.toggle">
        Toggle
    </button>
    <div data-wp-bind--hidden="!state.isOpen">
        Content revealed on click.
    </div>
</div>
```

### Available Directives

| Directive | Purpose |
|-----------|---------|
| `data-wp-interactive` | Mark a block as interactive and set its namespace |
| `data-wp-context` | Provide reactive context data |
| `data-wp-on--{event}` | Bind event listeners |
| `data-wp-bind--{attr}` | Reactively bind HTML attributes |
| `data-wp-text` | Reactively set text content |
| `data-wp-class--{name}` | Toggle CSS classes |
| `data-wp-style--{prop}` | Reactively set inline styles |
| `data-wp-each` | Loop over an array to render items |

> The Interactivity API is designed for server-rendered blocks. It provides progressive enhancement — blocks work without JavaScript and become interactive when the script loads.

## Plugin Dependencies

Plugins can now declare dependencies on other plugins using the `Requires Plugins` header. WordPress will prevent activation unless all required plugins are installed and active:

```php
<?php
/**
 * Plugin Name: My WooCommerce Extension
 * Requires Plugins: woocommerce
 * Description: Adds custom features to WooCommerce.
 * Version: 1.0.0
 */
```

Multiple dependencies are separated by commas:

```php
/**
 * Requires Plugins: woocommerce, jetpack
 */
```

The dependency system uses plugin slugs (the directory name on wordpress.org). On the Plugins screen, WordPress shows a clear message listing missing or inactive dependencies.

## AVIF Support

WordPress 6.5 adds native support for the **AVIF image format**. AVIF typically delivers 30–50% smaller file sizes than WebP at equivalent quality.

When a server's image library (Imagick or GD) supports AVIF, WordPress will:

- Accept `.avif` uploads
- Generate AVIF sub-sizes during upload
- Serve AVIF images in `<img>` and `<picture>` elements

To check if your server supports AVIF:

```php
$editor = wp_get_image_editor('/path/to/image.jpg');
if (!is_wp_error($editor)) {
    $formats = $editor->supports_mime_type('image/avif');
    // true if AVIF is supported
}
```

> AVIF support requires Imagick compiled with libheif/libaom, or GD compiled with AVIF support. Check your hosting environment.

## HTML API Improvements

The HTML Tag Processor can now scan **every token** in an HTML document — not just tags. This means you can process text nodes, comments, and processing instructions:

```php
$processor = new WP_HTML_Tag_Processor('<div>Hello <em>World</em></div>');

while ($processor->next_token()) {
    // Visits: <div>, "Hello ", <em>, "World", </em>, </div>
}
```

## Additional Changes

| Change | Details |
|--------|---------|
| Minimum MySQL | Raised from 5.0 to 5.5.5 |
| Design tools | Background image Size, Repeat, and Position controls |
| Cover block | New Aspect Ratio setting for responsive layouts |
| `hooked_block_{$type}` filter | Filter individual hooked blocks by type |
| RichText | Line breaks now always serialize as `<br>` elements |

## Upgrading

1. **MySQL version** — Ensure your database runs MySQL 5.5.5 or later (or MariaDB 10.0+)
2. **Block Bindings** — If you use custom fields with blocks, consider adopting the Block Bindings API
3. **Interactivity API** — Evaluate migrating custom block JavaScript to the Interactivity API for consistency
4. **Plugin headers** — Add `Requires Plugins` if your plugin depends on others
