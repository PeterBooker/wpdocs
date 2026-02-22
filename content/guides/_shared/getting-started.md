---
title: "Getting Started"
summary: "Learn the fundamentals of WordPress development."
weight: 1
---

## Introduction

WordPress is one of the most widely used content management systems in the world. This guide walks you through the essential concepts every WordPress developer should know.

## Plugin Structure

A minimal WordPress plugin consists of a single PHP file with a plugin header:

```php
<?php
/**
 * Plugin Name: My Plugin
 * Description: A simple example plugin.
 * Version: 1.0.0
 */

add_action('init', function () {
    // Your code here
});
```

## Theme Structure

A WordPress theme requires at minimum two files: `style.css` and `index.php`.

```
my-theme/
├── style.css          # Theme metadata and styles
├── index.php          # Main template
├── functions.php      # Theme setup and hooks
├── header.php         # Header partial
├── footer.php         # Footer partial
└── single.php         # Single post template
```

## Hooks: Actions and Filters

WordPress uses an event-driven architecture built on **hooks**. There are two types:

- **Actions** — run code at specific points (`do_action` / `add_action`)
- **Filters** — modify data as it passes through (`apply_filters` / `add_filter`)

```php
// Action: run code when WordPress initializes
add_action('init', function () {
    register_post_type('book', ['public' => true, 'label' => 'Books']);
});

// Filter: modify the page title
add_filter('the_title', function ($title) {
    return strtoupper($title);
});
```

> See the [Hooks guide](/{{version}}/guides/hooks/) for a deep dive into the hook system.

## Next Steps

- Read the [Hooks guide](/{{version}}/guides/hooks/) to understand actions and filters
- Browse the [Functions reference](/{{version}}/functions/) for the full API
- Explore [Classes](/{{version}}/classes/) like `WP_Query` and `WP_Post`
