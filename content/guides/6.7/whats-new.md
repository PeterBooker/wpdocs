---
title: "What's New"
summary: "WordPress 6.7 \"Rollins\" — Twenty Twenty-Five, Zoom Out mode, and block metadata collections."
weight: 0
---

## WordPress 6.7 "Rollins"

WordPress 6.7, released on November 12, 2024, introduces the Twenty Twenty-Five default theme, a new Zoom Out editing mode for layout composition, and significant performance improvements for block registration. Over 300 Core Trac tickets were resolved, including 87 enhancements and 200+ bug fixes.

### Highlights

- [Twenty Twenty-Five](#twenty-twenty-five) — a versatile new default theme
- [Zoom Out Mode](#zoom-out-mode) for composing layouts with patterns at a high level
- [Block Metadata Collections](#block-metadata-collections) for faster plugin loading
- [Plugin Template Registration](#plugin-template-registration) via new API functions
- [HTML API Enhancements](#html-api-enhancements) with full SVG and MathML support

## Twenty Twenty-Five

Twenty Twenty-Five is designed for bloggers, businesses, and portfolios alike. It ships with dozens of patterns for landing pages, photo blogs, and news sites, all built on block theme architecture.

Key features for developers:

- Full support for block theme APIs including `theme.json`, template parts, and block patterns
- Multiple style variations for different visual identities
- Extensive pattern library covering common layout needs
- Uses the latest block editor capabilities including section styles and Block Bindings

## Zoom Out Mode

A new editing mode provides a zoomed-out view of the entire page, making it easy to compose layouts by dragging and inserting patterns. Think of it as a high-level layout canvas that lets editors work with sections rather than individual blocks.

Zoom Out mode is available in the Site Editor via the toolbar. It automatically activates when inserting patterns from the inserter panel, providing a natural workflow for assembling pages.

## Block Metadata Collections

WordPress 6.7 introduces `wp_register_block_metadata_collection()`, a function that lets plugins register block metadata from a **single manifest file** instead of loading multiple individual `block.json` files:

```php
// Generate a block-manifest.php with: npx @wordpress/scripts build
// Then register all blocks at once:
wp_register_block_metadata_collection(
    plugin_dir_path(__FILE__) . 'build',
    plugin_dir_path(__FILE__) . 'build/block-manifest.php'
);
```

The manifest file is a PHP array that maps block directories to their metadata. The `@wordpress/scripts` build tool generates it automatically:

```php
<?php
// build/block-manifest.php (auto-generated)
return array(
    'my-block' => array(
        'apiVersion' => 3,
        'name' => 'myplugin/my-block',
        'title' => 'My Block',
        // ... rest of block.json contents
    ),
);
```

This eliminates filesystem lookups for each `block.json` file, delivering measurable performance improvements for plugins that register multiple blocks.

## Plugin Template Registration

Two new functions let plugins register block templates without bundling them as files:

```php
// Register a custom template
register_block_template('myplugin//product', [
    'title'       => 'Product Page',
    'description' => 'A template for displaying products.',
    'content'     => '<!-- wp:template-part {"slug":"header"} /-->
                      <!-- wp:group {"layout":{"type":"constrained"}} -->
                      <div class="wp-block-group">
                          <!-- wp:post-title /-->
                          <!-- wp:post-content /-->
                      </div>
                      <!-- /wp:group -->
                      <!-- wp:template-part {"slug":"footer"} /-->',
]);

// Remove a template (including core or theme templates)
unregister_block_template('myplugin//product');
```

> Template identifiers use the `plugin-slug//template-slug` format, matching the convention used by themes.

## HTML API Enhancements

The HTML API now supports processing of **almost all HTML tags**, including SVG and MathML elements. A new `set_modifiable_text()` method lets you modify text content programmatically:

```php
$processor = new WP_HTML_Tag_Processor('<p>Hello World</p>');
$processor->next_tag('p');
$processor->set_modifiable_text('Updated text');
echo $processor->get_updated_html();
// Output: <p>Updated text</p>
```

The HTML Processor also gained full-parser mode, enabling complete DOM-like traversal of HTML documents.

## Block Bindings UI

WordPress 6.7 adds a user interface in the block editor for managing Block Bindings directly. Editors can connect block attributes to custom fields without writing code:

1. Select a supported block (Paragraph, Heading, Image, Button)
2. Open the block settings sidebar
3. Look for the "Bindings" panel
4. Select a data source and field

For developers, the Block Bindings API introduced in 6.5 continues to grow. Custom binding sources registered via `register_block_bindings_source()` automatically appear in the UI.

## Performance Improvements

| Optimization | Impact |
|-------------|--------|
| `wp_register_block_metadata_collection()` | Eliminates per-block `block.json` file reads |
| `_wp_image_editor_choose` caching | Avoids redundant image editor compatibility checks |
| AVIF encoding speed | `heic:speed` raised to 7, reducing generation time by ~20% |
| `WP_Theme_JSON::compute_style_properties` caching | CSS for block nodes cached in transients, previously 6–11% of server time |

## Additional Changes

- **Editor hooks**: New `editor.preSavePost` async filter for customizing save behavior
- **Human-readable dates**: Post Date and Comment Date blocks support relative time ("3 hours ago")
- **Writing mode**: Added to Button, Verse, Site Title, and Site Tagline blocks
- **23 performance tickets** and **21 accessibility improvements** resolved

## Upgrading

1. **Block plugins** — Adopt `wp_register_block_metadata_collection()` with the `@wordpress/scripts` build tool for faster registration
2. **Theme developers** — Test with Twenty Twenty-Five and review section styles compatibility
3. **Block Bindings** — If you registered custom binding sources, verify they appear correctly in the new UI
