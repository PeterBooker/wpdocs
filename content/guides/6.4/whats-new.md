---
title: "What's New"
summary: "WordPress 6.4 \"Shirley\" — Block Hooks, admin notice API, lightbox, and post meta revisions."
weight: 0
---

## WordPress 6.4 "Shirley"

WordPress 6.4, released on November 7, 2023, introduces the Block Hooks API for automatic block insertion, standardized admin notice functions, image lightbox support, and a framework for tracking post meta revisions. It also ships the Twenty Twenty-Four default theme.

### Highlights

- [Block Hooks API](#block-hooks-api) — dynamically inject blocks at defined positions
- [Admin Notice Functions](#admin-notice-functions) — standardized HTML for admin notices
- [Image Lightbox](#image-lightbox) — fullscreen image viewing with "Expand on Click"
- [Post Meta Revisions](#post-meta-revisions) — opt-in revision tracking for post metadata
- [Twenty Twenty-Four](#twenty-twenty-four) — a versatile multi-purpose default theme

## Block Hooks API

The Block Hooks API provides a mechanism for plugins and themes to **automatically insert blocks** at specific positions relative to another block. This replaces the need to manually edit templates to add features like shopping carts, newsletter signups, or social sharing buttons.

### Declaring Block Hooks in block.json

```json
{
    "name": "myplugin/newsletter-form",
    "title": "Newsletter Form",
    "blockHooks": {
        "core/post-content": "after",
        "core/template-part/footer": "before"
    }
}
```

The supported positions are:

| Position | Behavior |
|----------|----------|
| `before` | Insert immediately before the anchor block |
| `after` | Insert immediately after the anchor block |
| `firstChild` | Insert as the first child inside the anchor block |
| `lastChild` | Insert as the last child inside the anchor block |

### Filtering Hooked Blocks

Use the `hooked_block_types` filter to control which blocks are hooked and under what conditions:

```php
add_filter('hooked_block_types', function ($hooked_blocks, $position, $anchor_block, $context) {
    // Only hook the block on single post templates
    if ($context instanceof WP_Block_Template && $context->slug === 'single') {
        if ($position === 'after' && $anchor_block === 'core/post-content') {
            $hooked_blocks[] = 'myplugin/related-posts';
        }
    }
    return $hooked_blocks;
}, 10, 4);
```

> Block Hooks work with block themes and the Site Editor. Classic themes that don't use block templates are unaffected.

## Admin Notice Functions

WordPress 6.4 introduces two new functions for generating **standardized admin notice markup**:

```php
// Output an admin notice directly
wp_admin_notice('Settings saved successfully.', [
    'type'               => 'success',    // 'success', 'error', 'warning', 'info'
    'dismissible'        => true,
    'additional_classes'  => ['my-notice'],
]);

// Get the HTML string without outputting
$html = wp_get_admin_notice('Something went wrong.', [
    'type' => 'error',
]);
```

A new `wp_admin_notice` action fires before each notice is output, letting plugins modify or suppress notices:

```php
add_action('wp_admin_notice', function ($message, $args) {
    // Log all admin notices
    error_log("Admin notice ({$args['type']}): {$message}");
}, 10, 2);
```

> These functions replace the common pattern of manually constructing `<div class="notice notice-success">` markup, ensuring consistency and accessibility across the admin.

## Image Lightbox

Image blocks gain an **"Expand on Click"** setting that displays images in a fullscreen lightbox overlay. When enabled, clicking the image opens it in a modal with a dark backdrop, allowing visitors to view the full-resolution image without leaving the page.

The lightbox uses WordPress's built-in animation and accessibility features — it traps focus, supports keyboard navigation, and closes with Escape.

To enable the lightbox by default for all images in your theme via `theme.json`:

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

## Post Meta Revisions

WordPress 6.4 adds a framework for **storing revisions of post metadata**. This is opt-in: each meta key must explicitly enable revision tracking.

```php
register_post_meta('post', 'custom_subtitle', [
    'type'              => 'string',
    'single'            => true,
    'show_in_rest'      => true,
    'revisions_enabled' => true,  // Enable meta revisions
]);
```

When enabled, WordPress saves the meta value with each post revision. When a revision is restored, the associated meta values are restored as well. This solves a long-standing pain point where restoring a revision lost custom field data.

## Twenty Twenty-Four

The new default theme is a **multi-purpose block theme** suitable for blogs, portfolios, business sites, and personal pages. It ships with a rich library of patterns and demonstrates modern block theme capabilities:

- Full site editing support with customizable templates
- Multiple style variations for different visual identities
- Pattern library covering hero sections, portfolios, testimonials, and more
- Designed to showcase block editor features like font management and design tools

## HTML API Updates

The HTML API receives a significant update with the introduction of a minimal HTML Processor:

```php
$processor = new WP_HTML_Tag_Processor('<div class="wrapper"><p>Hello</p></div>');

// Navigate to elements with breadcrumb awareness
$processor->next_tag('p');
echo $processor->get_tag(); // "P"
```

New CSS class helpers make it easier to work with class attributes:

```php
$processor->next_tag('div');
$processor->add_class('active');
$processor->remove_class('hidden');
$processor->has_class('wrapper'); // true
```

## Performance Improvements

WordPress 6.4 delivers approximately **4% improvement** in server response time over 6.3.2:

| Optimization | Impact |
|-------------|--------|
| Pattern caching | `_wp_get_block_patterns()` eliminates redundant file lookups |
| Theme API | Avoids unnecessary file existence checks when stylesheet matches template directory |
| Script loading | Frontend core scripts use `defer` strategy and move from footer to head |

## Additional Changes

| Change | Details |
|--------|---------|
| Attachment pages | Disabled by default for new installations |
| `TEMPLATEPATH` / `STYLESHEETPATH` | Constants officially deprecated |
| Command Palette | Enhanced search and navigation in the editor |
| Pattern categories | Users can categorize custom patterns (synced and unsynced) |

## Upgrading

1. **Block plugins** — Consider adopting the Block Hooks API to auto-insert blocks instead of requiring manual template editing
2. **Admin notices** — Migrate to `wp_admin_notice()` for consistent, accessible notice markup
3. **Post meta** — Enable `revisions_enabled` for important custom fields
4. **Deprecated constants** — Replace `TEMPLATEPATH` and `STYLESHEETPATH` with `get_template_directory()` and `get_stylesheet_directory()`
