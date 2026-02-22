---
title: "Hooks"
summary: "Understanding WordPress actions and filters."
weight: 2
---

## Overview

Hooks are the backbone of WordPress extensibility. They allow plugins and themes to modify WordPress behavior without editing core files.

## Actions

Actions let you **execute code** at specific points in the WordPress lifecycle.

```php
// Register the hook
add_action('wp_head', 'my_custom_meta');

function my_custom_meta() {
    echo '<meta name="author" content="Developer">';
}
```

### Common Actions

| Hook | When it fires |
|------|---------------|
| `init` | After WordPress loads, before headers are sent |
| `wp_enqueue_scripts` | When scripts and styles should be enqueued |
| `save_post` | After a post is saved to the database |
| `wp_head` | Inside the `<head>` element |
| `wp_footer` | Before the closing `</body>` tag |

## Filters

Filters let you **modify data** as it passes through WordPress.

```php
// Modify the excerpt length
add_filter('excerpt_length', function ($length) {
    return 30; // 30 words instead of default 55
});
```

### Common Filters

| Hook | What it filters |
|------|-----------------|
| `the_content` | Post content before display |
| `the_title` | Post title before display |
| `body_class` | CSS classes on the `<body>` element |
| `query_vars` | Allowed public query variables |
| `upload_mimes` | Allowed upload MIME types |

## Priority and Arguments

Both `add_action` and `add_filter` accept priority (default 10) and argument count:

```php
// Run late (priority 99), accept 2 arguments
add_filter('the_content', 'my_filter', 99, 2);

function my_filter($content, $post_id) {
    return $content . '<p>Post ID: ' . $post_id . '</p>';
}
```

Lower priority numbers run first. Use this to control execution order when multiple callbacks are registered on the same hook.

## Removing Hooks

```php
// Remove a specific callback
remove_action('wp_head', 'wp_generator');

// Remove all callbacks at a specific priority
remove_all_actions('wp_head', 10);
```

## Creating Custom Hooks

Define your own hooks so other plugins can extend your code:

```php
// In your plugin: fire a custom action
do_action('my_plugin_after_save', $post_id, $data);

// In your plugin: apply a custom filter
$output = apply_filters('my_plugin_output', $default_output, $context);
```

> Browse all WordPress hooks in the [Hooks reference](/{{version}}/hooks/).
