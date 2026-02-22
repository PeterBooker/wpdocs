---
title: "What's New"
summary: "WordPress 6.8 \"Cecil\" — Speculative loading, bcrypt passwords, and Style Book for classic themes."
weight: 0
---

## WordPress 6.8 "Cecil"

WordPress 6.8, released on April 15, 2025, focuses on performance and security fundamentals. Speculative loading makes page transitions feel instant, bcrypt replaces legacy password hashing, and the Style Book debuts for classic themes.

### Highlights

- [Speculative Loading](#speculative-loading) prefetches and prerenders pages before the user clicks
- [Bcrypt Password Hashing](#bcrypt-password-hashing) replaces the legacy phpass algorithm
- [Style Book for Classic Themes](#style-book-for-classic-themes) brings visual style management to all themes
- [Query Loop Improvements](#query-loop-improvements) with a redesigned pattern selector
- [Interactivity API Updates](#interactivity-api-updates) with `withSyncEvent()` and expanded `wp-each`

## Speculative Loading

WordPress 6.8 leverages the browser's [Speculation Rules API](https://developer.mozilla.org/en-US/docs/Web/API/Speculation_Rules_API) to prefetch or prerender pages before users navigate to them. The result is near-instant page loads on supporting browsers.

The feature is enabled by default. WordPress outputs speculation rules in the page `<head>` that tell the browser to prerender same-origin links when the user is likely to click them:

```php
// Customize the speculation rules behavior
add_filter('wp_speculation_rules_configuration', function ($config) {
    return [
        'mode'      => 'prerender',   // 'prefetch' or 'prerender'
        'eagerness' => 'moderate',    // 'conservative', 'moderate', or 'eager'
    ];
});
```

You can exclude specific URL patterns from speculative loading, which is useful for pages with side effects like logout URLs:

```php
add_filter('wp_speculation_rules_href_exclude_paths', function ($exclude_paths) {
    $exclude_paths[] = '/checkout/*';
    $exclude_paths[] = '/wp-login.php';
    return $exclude_paths;
});
```

To disable speculative loading entirely:

```php
add_filter('wp_load_speculation_rules', '__return_false');
```

> Speculative loading only affects server-rendered page navigation. It has no impact on the block editor or single-page application patterns.

## Bcrypt Password Hashing

WordPress 6.8 replaces the legacy phpass-based password hashing with **bcrypt**, a modern algorithm that is significantly more resistant to brute-force attacks. Application passwords, password reset keys, personal data request keys, and recovery mode keys now use **BLAKE2b** hashing via the Sodium library.

The migration is transparent. Existing passwords continue to work: when a user logs in with a phpass-hashed password, WordPress automatically rehashes it with bcrypt. No action is required for most developers.

Two new utility functions are available for hashing randomly generated strings:

```php
// Hash a high-entropy string (e.g., an API key)
$hash = wp_fast_hash($random_token);

// Verify against the hash
if (wp_verify_fast_hash($random_token, $hash)) {
    // Token is valid
}
```

> If your plugin directly manipulates password hashes or uses phpass constants, review your code for compatibility. The `wp_hash_password()` and `wp_check_password()` functions continue to work without changes.

## Style Book for Classic Themes

The Style Book — a comprehensive visual overview of your site's colors, typography, and block styles — was previously available only to block themes. WordPress 6.8 extends it to **classic themes**, giving all theme developers and site owners a single place to review how global styles affect every block type.

Access the Style Book from **Appearance > Design** in the admin dashboard.

## Query Loop Improvements

The Query Loop block receives a usability overhaul:

- The "Replace" pattern modal is replaced with a compact **dropdown selector**
- A new **"Change design"** button makes it easy to switch layout patterns
- An **"Ignore sticky posts"** toggle gives editors explicit control

## Interactivity API Updates

### `withSyncEvent()`

A new `withSyncEvent()` utility function prepares your Interactivity API actions for a future release where event handlers will become asynchronous by default:

```php
// In block.json, declare interactivity support
{
  "supports": { "interactivity": true }
}
```

```js
// In your store, wrap synchronous actions
import { store, withSyncEvent } from '@wordpress/interactivity';

store('myPlugin', {
    actions: {
        handleClick: withSyncEvent(function* (event) {
            event.preventDefault();
            const ctx = getContext();
            ctx.isOpen = !ctx.isOpen;
        }),
    },
});
```

### Expanded `wp-each` Directive

The `wp-each` directive now supports **any iterable** — arrays, strings, Maps, Sets, and generators — not just arrays:

```html
<template data-wp-each="state.items">
    <li data-wp-text="context.item"></li>
</template>
```

## Additional Developer Changes

| Change | Details |
|--------|---------|
| `wp_register_block_types_from_metadata_collection()` | Register multiple block types from a single metadata collection |
| `rest_menu_read_access` filter | Control public REST API access to navigation menus |
| Block Editor | 387 enhancements, 525 bug fixes, 70 accessibility improvements |
| Minimum PHP | PHP 7.2.24 remains the minimum, but PHP 8.x is strongly recommended |

## Upgrading

Most sites will upgrade seamlessly. Review these areas if applicable:

1. **Password handling** — If your plugin directly stores or verifies password hashes, test with bcrypt
2. **Speculative loading** — Exclude URLs with side effects using the `wp_speculation_rules_href_exclude_paths` filter
3. **Interactivity API** — Start wrapping synchronous event handlers with `withSyncEvent()` to prepare for the async default
