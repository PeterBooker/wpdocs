---
title: "Block Editor"
summary: "Block development in WordPress 6.8 — Speculative Loading, Data Views, and the Style Engine."
weight: 10
---

## What's New in 6.8

WordPress 6.8 brings significant improvements to the block editor and surrounding APIs:

- **Speculative Loading** for near-instant page navigations
- **Style Engine** improvements with expanded CSS custom property support
- **Block Hooks auto-insertion** enhancements
- **Improved PHP support** with minimum PHP 7.4 enforcement

## Speculative Loading

WordPress 6.8 introduces speculative loading via the Speculation Rules API. This prefetches pages before the user clicks, making navigation feel instant.

```php
// The feature is enabled by default. To customize:
add_filter('wp_speculation_rules_configuration', function ($config) {
    return [
        'mode'      => 'prerender',    // 'prefetch' or 'prerender'
        'eagerness' => 'moderate',     // 'conservative', 'moderate', or 'eager'
    ];
});

// To disable entirely:
add_filter('wp_speculation_rules_enabled', '__return_false');
```

The speculation rules are automatically output in the page `<head>`:

```json
{
  "prerender": [{
    "source": "document",
    "where": { "href_matches": "/*" },
    "eagerness": "moderate"
  }]
}
```

> Speculative loading works with server-rendered pages. Single-page application patterns in the block editor are unaffected.

## Block Hooks Auto-Insertion

Block Hooks, introduced in 6.5, receive further refinement in 6.8. You can now control auto-insertion positioning with greater precision:

```json
{
  "name": "myplugin/newsletter-signup",
  "title": "Newsletter Signup",
  "blockHooks": {
    "core/post-content": "after",
    "core/template-part/footer": "before"
  }
}
```

### Customizing Auto-Insertion

```php
add_filter('hooked_block_types', function ($hooked_blocks, $position, $anchor_block, $context) {
    // Only auto-insert on single posts
    if ($context instanceof WP_Block_Template && $context->slug === 'single') {
        $hooked_blocks[] = 'myplugin/related-posts';
    }
    return $hooked_blocks;
}, 10, 4);
```

## Style Engine

The Style Engine in 6.8 expands support for CSS custom properties in block styles:

```php
// Register a block style with CSS custom properties
register_block_style('core/group', [
    'name'       => 'card',
    'label'      => 'Card',
    'style_data' => [
        'border'  => ['radius' => 'var(--wp--custom--border-radius)'],
        'shadow'  => 'var(--wp--custom--shadow-card)',
        'spacing' => ['padding' => 'var(--wp--preset--spacing--40)'],
    ],
]);
```

## Creating a Custom Block

The fundamentals remain the same. Scaffold a block with:

```bash
npx @wordpress/create-block my-block
cd my-block
npm start
```

### block.json (6.8)

```json
{
  "$schema": "https://schemas.wp.org/trunk/block.json",
  "apiVersion": 3,
  "name": "myplugin/notice",
  "title": "Notice",
  "category": "widgets",
  "icon": "info",
  "description": "Display an informational notice.",
  "supports": {
    "html": false,
    "color": {
      "background": true,
      "text": true
    },
    "spacing": {
      "margin": true,
      "padding": true
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

Use `render.php` for dynamic blocks:

```php
<?php
// render.php — $attributes, $content, and $block are available
$classes = 'wp-block-myplugin-notice';
if (!empty($attributes['className'])) {
    $classes .= ' ' . esc_attr($attributes['className']);
}
?>
<div <?php echo get_block_wrapper_attributes(['class' => $classes]); ?>>
    <?php echo wp_kses_post($content); ?>
</div>
```

> Browse the [Functions reference](/{{version}}/functions/) for block-related functions like `register_block_type` and `register_block_style`.
