---
title: "What's New"
summary: "WordPress 6.6 \"Dorsey\" — Automatic plugin rollbacks, pattern overrides, and section styles."
weight: 0
---

## WordPress 6.6 "Dorsey"

WordPress 6.6, released on July 16, 2024, makes plugin updates safer with automatic rollbacks, gives synced patterns per-instance content overrides, and introduces section-level styling. It also bumps the minimum PHP version to 7.2.24 and ships theme.json version 3.

### Highlights

- [Automatic Plugin Rollbacks](#automatic-plugin-rollbacks) revert failed auto-updates
- [Pattern Overrides](#pattern-overrides) let synced patterns have per-instance content
- [Section Styles](#section-styles) apply unique styling to individual page sections
- [theme.json Version 3](#themejson-version-3) with CSS specificity changes
- [Data Views API](#data-views-api) for building custom admin interfaces

## Automatic Plugin Rollbacks

When a plugin auto-update causes a **PHP fatal error**, WordPress 6.6 automatically deactivates the updated plugin, restores the previous version, and sends a notification email to the site administrator. This safety net removes the biggest risk of enabling auto-updates.

The rollback system:

1. Creates a backup of the plugin before updating
2. Activates the updated version in a sandboxed check
3. If a fatal error is detected, restores the backup
4. Sends an email notification explaining what happened

No developer action is required — the feature works automatically for all plugins with auto-updates enabled.

## Pattern Overrides

Synced patterns (formerly Reusable Blocks) can now have **content that varies per instance** while maintaining the same structure and styling everywhere. This is powered by the Block Bindings API.

To make a block's content overridable within a synced pattern:

1. Create or edit a synced pattern
2. Select a Paragraph, Heading, Button, or Image block
3. Open the block's **Advanced** settings
4. Enable **"Allow overrides"**

When the pattern is used on a page, editors can customize the content of overridable blocks without affecting other instances.

For developers, pattern overrides use the `core/pattern-overrides` binding source:

```html
<!-- wp:paragraph {
    "metadata": {
        "bindings": {
            "content": {
                "source": "core/pattern-overrides"
            }
        },
        "name": "testimonial-text"
    }
} -->
<p>Default testimonial text</p>
<!-- /wp:paragraph -->
```

## Section Styles

Theme developers can define **section-specific style variations** in `theme.json`. A section style applies a complete visual treatment — colors, typography, spacing — to a Group block and all its inner blocks:

```json
{
    "version": 3,
    "styles": {
        "blocks": {
            "core/group": {
                "variations": {
                    "dark-section": {
                        "color": {
                            "background": "#1a1a2e",
                            "text": "#ffffff"
                        },
                        "elements": {
                            "link": {
                                "color": { "text": "#e94560" }
                            },
                            "heading": {
                                "color": { "text": "#f5f5f5" }
                            }
                        }
                    }
                }
            }
        }
    }
}
```

Section styles cascade to all inner blocks, making it easy to create visually distinct page regions without writing custom CSS.

## theme.json Version 3

WordPress 6.6 introduces **version 3** of the `theme.json` schema. The main change is how CSS specificity is handled for theme styles, which may require updates to existing block themes:

```json
{
    "version": 3,
    "settings": { },
    "styles": { }
}
```

Key changes:

- **CSS specificity** for block styles is adjusted to prevent theme styles from being overridden by lower-specificity core styles
- Theme developers should test their existing `theme.json` styles after upgrading to version 3
- Version 2 continues to work but receives no new features

> If you maintain a block theme, test thoroughly after bumping to `"version": 3`. Some styles may need specificity adjustments.

## Data Views API

A new Data Views API provides a standardized way to build **tabular admin interfaces** for managing content. WordPress uses it internally for the Templates and Patterns screens in the Site Editor.

Plugin developers can register custom actions for post types when building Data Views:

```js
import { registerPostTypeActions } from '@wordpress/editor';

registerPostTypeActions('product', [
    {
        id: 'export-csv',
        label: 'Export as CSV',
        callback: (items) => {
            // Handle export
        },
    },
]);
```

## Interactivity API: Async Directives

A new `data-wp-on-async` directive provides **non-blocking event handling** for the Interactivity API. Unlike `data-wp-on`, async directives yield to the main thread before executing, improving Interaction to Next Paint (INP) scores:

```html
<button data-wp-on-async--click="actions.handleClick">
    Click me
</button>
```

Use `data-wp-on-async` for event handlers that don't need to call `event.preventDefault()` or read synchronous event properties.

## Grid Layout

The Group block gains a new **Grid** layout variation, enabling CSS Grid layouts directly in the editor:

```html
<!-- wp:group {"layout":{"type":"grid","columnCount":3}} -->
<div class="wp-block-group">
    <!-- Child blocks are placed in a 3-column grid -->
</div>
<!-- /wp:group -->
```

## Additional Changes

| Change | Details |
|--------|---------|
| Minimum PHP | Raised to 7.2.24; PHP 7.0 and 7.1 are no longer supported |
| Negative margins | Now supported natively in the block editor |
| Summary panel | Redesigned inspector with featured image, excerpt, and commenting toggles |
| Color/typography mixing | Style variations can have color and typography mixed independently |
| Performance | ~35–40% reduction in template loading time in the editor |

## Upgrading

1. **PHP version** — Ensure your hosting runs PHP 7.2.24 or later (PHP 8.x recommended)
2. **theme.json** — If you maintain a block theme, test with version 3 for CSS specificity changes
3. **Interactivity API** — Migrate non-critical event handlers to `data-wp-on-async` for better INP
4. **Synced patterns** — Review pattern overrides if you use the Block Bindings API
