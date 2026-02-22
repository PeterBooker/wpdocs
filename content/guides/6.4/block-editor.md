---
title: "Block Editor"
summary: "Block development in WordPress 6.4 — Block Hooks, lightbox, and HTML API foundations."
weight: 10
---

## What's New in 6.4

WordPress 6.4 introduces the Block Hooks API as the major block development feature, along with image lightbox support and the first iteration of the HTML API Processor.

- **Block Hooks API** for automatic block insertion into templates
- **Image lightbox** via the "Expand on Click" setting
- **HTML API** with tag processing and CSS class helpers
- **Command Palette** enhancements for faster editor navigation

## Block Hooks API

Block Hooks let plugins automatically insert blocks at specific positions relative to an anchor block — before, after, as the first child, or as the last child. This is the recommended way to inject blocks into templates without requiring users to manually edit their layouts.

### Declaring Hooks in block.json

```json
{
    "apiVersion": 3,
    "name": "myplugin/share-buttons",
    "title": "Share Buttons",
    "category": "widgets",
    "blockHooks": {
        "core/post-content": "after"
    },
    "editorScript": "file:./index.js",
    "render": "file:./render.php"
}
```

### Controlling Insertion via PHP

```php
add_filter('hooked_block_types', function ($hooked_blocks, $position, $anchor_block, $context) {
    if ($position === 'after' && $anchor_block === 'core/post-content') {
        // Only on single post templates
        if ($context instanceof WP_Block_Template && $context->slug === 'single') {
            $hooked_blocks[] = 'myplugin/share-buttons';
        }
    }
    return $hooked_blocks;
}, 10, 4);
```

> Block Hooks only work with block themes and the Site Editor. Classic themes that use PHP template files are not affected.

## Creating a Custom Block (6.4)

The block development workflow in 6.4 uses `@wordpress/create-block` and `apiVersion` 3:

```bash
npx @wordpress/create-block my-notice-block
cd my-notice-block
npm start
```

### block.json

```json
{
    "$schema": "https://schemas.wp.org/trunk/block.json",
    "apiVersion": 3,
    "name": "myplugin/notice",
    "title": "Notice",
    "category": "design",
    "icon": "info",
    "description": "Display a styled notice.",
    "supports": {
        "html": false,
        "color": {
            "background": true,
            "text": true
        },
        "typography": {
            "fontSize": true
        }
    },
    "textdomain": "myplugin",
    "editorScript": "file:./index.js",
    "editorStyle": "file:./index.css",
    "style": "file:./style-index.css",
    "render": "file:./render.php"
}
```

### Server-Side Rendering

```php
<?php
// render.php
$classes = 'wp-block-myplugin-notice';
if (!empty($attributes['className'])) {
    $classes .= ' ' . esc_attr($attributes['className']);
}
?>
<div <?php echo get_block_wrapper_attributes(['class' => $classes]); ?>>
    <?php echo wp_kses_post($content); ?>
</div>
```

## Image Lightbox

WordPress 6.4 adds a native lightbox for Image blocks. Enable it per-image in the block settings with the "Expand on Click" toggle, or enable it globally via `theme.json`:

```json
{
    "version": 2,
    "settings": {
        "blocks": {
            "core/image": {
                "lightbox": {
                    "enabled": true
                }
            }
        }
    }
}
```

The lightbox implementation:
- Uses semantic HTML and ARIA attributes for accessibility
- Supports keyboard navigation (Escape to close, Tab to navigate)
- Traps focus within the overlay
- Animates open/close transitions

## HTML API

WordPress 6.4 introduces the HTML Tag Processor, a safe way to modify HTML without regex:

```php
$html = '<div class="card"><h2>Title</h2><p>Content</p></div>';
$processor = new WP_HTML_Tag_Processor($html);

// Find and modify a tag
if ($processor->next_tag('div')) {
    $processor->add_class('featured');
    $processor->set_attribute('data-id', '42');
}

echo $processor->get_updated_html();
// <div class="card featured" data-id="42"><h2>Title</h2><p>Content</p></div>
```

### CSS Class Helpers

```php
$processor = new WP_HTML_Tag_Processor('<button class="btn">Click</button>');
$processor->next_tag('button');

$processor->has_class('btn');      // true
$processor->add_class('btn-lg');   // Add a class
$processor->remove_class('btn');   // Remove a class

echo $processor->get_updated_html();
// <button class="btn-lg">Click</button>
```

## Script Loading Strategies

WordPress 6.4 moves core frontend scripts from the footer to the `<head>` using the `defer` loading strategy, which was introduced in 6.3. This allows browsers to discover scripts earlier while still executing them after the DOM is ready:

```php
wp_enqueue_script(
    'my-script',
    plugin_dir_url(__FILE__) . 'js/frontend.js',
    [],
    '1.0.0',
    [
        'strategy'  => 'defer',     // 'defer' or 'async'
        'in_footer' => false,       // Can now be in <head> with defer
    ]
);
```

> The `defer` strategy is preferred over `async` for scripts that depend on DOM elements, as it guarantees execution order and waits for the DOM to be parsed.
