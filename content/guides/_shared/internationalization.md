---
title: "Internationalization"
summary: "Translating plugins and themes with gettext, i18n functions, and .pot/.po/.mo files."
weight: 13
---

## Introduction

Internationalization (i18n) is the process of writing code so that strings can be translated into other languages without modifying the source. WordPress has built-in support for gettext-based translations. If you're distributing a plugin or theme, internationalizing your strings is essential.

## Text Domain

Every translatable plugin or theme needs a **text domain** — a unique identifier that maps strings to the correct translation files.

Declare it in your plugin header:

```php
<?php
/**
 * Plugin Name: My Plugin
 * Text Domain: my-plugin
 * Domain Path: /languages
 */
```

Or in your theme's `style.css`:

```css
/*
 * Theme Name: My Theme
 * Text Domain: my-theme
 */
```

> The text domain should match your plugin or theme's slug. WordPress uses it to load the correct `.mo` file.

## Translation Functions

### Basic Strings

```php
// Simple translation
echo __('Settings saved.', 'my-plugin');

// Echo shorthand
_e('Settings saved.', 'my-plugin');
```

### Strings with Context

When the same English word has different meanings in different contexts, use `_x()`:

```php
// "Post" as a verb (to publish)
$label = _x('Post', 'verb', 'my-plugin');

// "Post" as a noun (a blog post)
$label = _x('Post', 'noun', 'my-plugin');
```

### Plurals

Use `_n()` for strings that change based on a count:

```php
$count = 5;

$message = sprintf(
    _n(
        '%d item found.',     // Singular
        '%d items found.',    // Plural
        $count,
        'my-plugin'
    ),
    $count
);
// Output: "5 items found."
```

### Strings with Placeholders

Never concatenate translated strings. Use `sprintf()` with placeholders:

```php
// Correct — translators can reorder placeholders
$message = sprintf(
    /* translators: 1: user name, 2: post title */
    __('Hello %1$s, your post "%2$s" has been published.', 'my-plugin'),
    esc_html($user_name),
    esc_html($post_title)
);

// Wrong — translators can't reorder pieces
$message = __('Hello ', 'my-plugin') . $user_name . __(', your post "', 'my-plugin') . $post_title;
```

### Escaped Output

For strings going into HTML, use the escaping variants:

```php
// Escape for HTML context
esc_html__('Settings', 'my-plugin');     // Returns escaped string
esc_html_e('Settings', 'my-plugin');     // Echoes escaped string

// Escape for HTML attributes
esc_attr__('Click here', 'my-plugin');
esc_attr_e('Click here', 'my-plugin');

// With context
esc_html_x('Post', 'verb', 'my-plugin');
```

## Function Reference

| Function | Purpose |
|----------|---------|
| `__()` | Return a translated string |
| `_e()` | Echo a translated string |
| `_x()` | Return a translated string with context |
| `_ex()` | Echo a translated string with context |
| `_n()` | Return a translated plural string |
| `_nx()` | Return a translated plural string with context |
| `esc_html__()` | Return a translated, HTML-escaped string |
| `esc_html_e()` | Echo a translated, HTML-escaped string |
| `esc_attr__()` | Return a translated, attribute-escaped string |
| `esc_attr_e()` | Echo a translated, attribute-escaped string |
| `esc_html_x()` | Return a translated, HTML-escaped string with context |
| `esc_attr_x()` | Return a translated, attribute-escaped string with context |

## Loading Translations

### Plugins

```php
add_action('init', function () {
    load_plugin_textdomain('my-plugin', false, dirname(plugin_basename(__FILE__)) . '/languages');
});
```

WordPress also auto-loads translations from `wp-content/languages/plugins/my-plugin-{locale}.mo` if the plugin is hosted on wordpress.org.

### Themes

```php
add_action('after_setup_theme', function () {
    load_theme_textdomain('my-theme', get_template_directory() . '/languages');
});
```

## Generating Translation Files

### The .pot File

A `.pot` (Portable Object Template) file is the master template containing all translatable strings. Generate it with WP-CLI:

```bash
# Generate .pot from a plugin
wp i18n make-pot /path/to/my-plugin /path/to/my-plugin/languages/my-plugin.pot

# Generate .pot from a theme
wp i18n make-pot /path/to/my-theme /path/to/my-theme/languages/my-theme.pot
```

### The .po and .mo Files

Translators work with `.po` files (human-readable). These are compiled to `.mo` files (machine-readable) that WordPress loads at runtime:

```bash
# Compile a .po file to .mo
wp i18n make-mo /path/to/languages/my-plugin-fr_FR.po
```

The file naming convention is `{text-domain}-{locale}.{po|mo}`:

```
languages/
├── my-plugin.pot            # Template (source strings)
├── my-plugin-fr_FR.po       # French translation (editable)
├── my-plugin-fr_FR.mo       # French translation (compiled)
├── my-plugin-de_DE.po       # German translation (editable)
└── my-plugin-de_DE.mo       # German translation (compiled)
```

## JavaScript Translations

For JavaScript strings in the block editor or frontend, use the `@wordpress/i18n` package:

```js
import { __, _n, sprintf } from '@wordpress/i18n';

const label = __('Settings', 'my-plugin');
const message = sprintf(__('Found %d results.', 'my-plugin'), count);
```

Register your script's translation file in PHP:

```php
add_action('init', function () {
    wp_set_script_translations('my-plugin-editor', 'my-plugin', plugin_dir_path(__FILE__) . 'languages');
});
```

Generate the JavaScript translation file:

```bash
wp i18n make-json /path/to/languages/ --no-purge
```

This creates `.json` files like `my-plugin-fr_FR-{hash}.json` that WordPress loads automatically for the registered script.

## Best Practices

1. **Mark every user-facing string** — Even if you only speak English, other developers may translate your plugin
2. **Never split sentences** — Translators need the complete sentence to produce correct translations
3. **Use translator comments** — Add `/* translators: */` comments before strings with placeholders
4. **Keep HTML out of translated strings** when possible — Use `sprintf()` to wrap translated text in HTML
5. **Don't translate slugs, option names, or technical identifiers** — Only translate user-visible text
6. **Test with a RTL language** — Install a right-to-left language pack to catch layout issues early

```php
// Good — HTML is separate from the translated string
echo '<p>' . esc_html__('No results found.', 'my-plugin') . '</p>';

// Acceptable — minimal HTML that translators might need to see
echo wp_kses(
    __('<strong>Note:</strong> This action cannot be undone.', 'my-plugin'),
    ['strong' => []]
);

// Bad — complex HTML mixed with text
_e('<div class="notice"><p style="color:red">Error occurred!</p></div>', 'my-plugin');
```
