---
title: "Plugin Development"
summary: "Build WordPress plugins from activation to distribution."
weight: 6
---

## Overview

Plugins extend WordPress functionality without modifying core files. A plugin can be as simple as a single PHP file or as complex as a full application.

## Plugin Structure

### Minimal Plugin

```php
<?php
/**
 * Plugin Name: My Plugin
 * Description: A short description of what this plugin does.
 * Version:     1.0.0
 * Author:      Your Name
 * License:     GPL-2.0-or-later
 */

// Prevent direct access
if (!defined('ABSPATH')) {
    exit;
}
```

### Recommended File Layout

```
my-plugin/
├── my-plugin.php           # Main plugin file (header + bootstrap)
├── includes/
│   ├── class-plugin.php    # Main plugin class
│   ├── class-admin.php     # Admin functionality
│   └── class-api.php       # REST API endpoints
├── assets/
│   ├── css/
│   └── js/
├── templates/              # Template files
├── languages/              # Translation files
├── tests/                  # PHPUnit tests
├── readme.txt              # WordPress.org readme
└── uninstall.php           # Cleanup on uninstall
```

## Activation and Deactivation

```php
register_activation_hook(__FILE__, function () {
    // Create database tables, set default options, flush rewrite rules
    add_option('myplugin_version', '1.0.0');
    flush_rewrite_rules();
});

register_deactivation_hook(__FILE__, function () {
    // Clean up temporary data, flush rewrite rules
    flush_rewrite_rules();
});
```

For complete cleanup when the plugin is **deleted**, use `uninstall.php`:

```php
<?php
// uninstall.php
if (!defined('WP_UNINSTALL_PLUGIN')) {
    exit;
}

delete_option('myplugin_version');
delete_option('myplugin_settings');

global $wpdb;
$wpdb->query("DROP TABLE IF EXISTS {$wpdb->prefix}myplugin_data");
```

## Settings API

Register settings pages the WordPress way:

```php
add_action('admin_menu', function () {
    add_options_page(
        'My Plugin Settings',       // Page title
        'My Plugin',                // Menu title
        'manage_options',           // Capability
        'myplugin-settings',        // Slug
        'myplugin_settings_page'    // Callback
    );
});

add_action('admin_init', function () {
    register_setting('myplugin_options', 'myplugin_api_key', [
        'type'              => 'string',
        'sanitize_callback' => 'sanitize_text_field',
        'default'           => '',
    ]);

    add_settings_section('myplugin_main', 'API Settings', null, 'myplugin-settings');

    add_settings_field('myplugin_api_key', 'API Key', function () {
        $value = get_option('myplugin_api_key');
        echo '<input type="text" name="myplugin_api_key" value="' . esc_attr($value) . '" class="regular-text">';
    }, 'myplugin-settings', 'myplugin_main');
});

function myplugin_settings_page() {
    echo '<div class="wrap">';
    echo '<h1>My Plugin Settings</h1>';
    echo '<form method="post" action="options.php">';
    settings_fields('myplugin_options');
    do_settings_sections('myplugin-settings');
    submit_button();
    echo '</form></div>';
}
```

## Custom Post Types

```php
add_action('init', function () {
    register_post_type('product', [
        'labels' => [
            'name'          => 'Products',
            'singular_name' => 'Product',
            'add_new_item'  => 'Add New Product',
            'edit_item'     => 'Edit Product',
        ],
        'public'       => true,
        'has_archive'  => true,
        'show_in_rest' => true,  // Enable block editor + REST API
        'supports'     => ['title', 'editor', 'thumbnail', 'excerpt'],
        'rewrite'      => ['slug' => 'products'],
        'menu_icon'    => 'dashicons-cart',
    ]);
});
```

## Custom Taxonomies

```php
add_action('init', function () {
    register_taxonomy('genre', ['product'], [
        'labels' => [
            'name'          => 'Genres',
            'singular_name' => 'Genre',
        ],
        'public'       => true,
        'hierarchical' => true,  // true = categories, false = tags
        'show_in_rest' => true,
        'rewrite'      => ['slug' => 'genre'],
    ]);
});
```

## Enqueueing Scripts and Styles

```php
add_action('wp_enqueue_scripts', function () {
    // Only on specific pages
    if (!is_singular('product')) {
        return;
    }

    wp_enqueue_style(
        'myplugin-style',
        plugin_dir_url(__FILE__) . 'assets/css/frontend.css',
        [],
        '1.0.0'
    );

    wp_enqueue_script(
        'myplugin-script',
        plugin_dir_url(__FILE__) . 'assets/js/frontend.js',
        ['jquery'],
        '1.0.0',
        true  // Load in footer
    );

    // Pass data to JavaScript
    wp_localize_script('myplugin-script', 'myPlugin', [
        'ajaxUrl' => admin_url('admin-ajax.php'),
        'nonce'   => wp_create_nonce('myplugin_nonce'),
    ]);
});
```

## Security Best Practices

1. **Sanitize input**: `sanitize_text_field()`, `absint()`, `wp_kses_post()`
2. **Escape output**: `esc_html()`, `esc_attr()`, `esc_url()`, `wp_kses()`
3. **Verify nonces**: `wp_verify_nonce()`, `check_admin_referer()`
4. **Check capabilities**: `current_user_can('edit_posts')`
5. **Use prepared queries**: `$wpdb->prepare()`

```php
// Example: safe AJAX handler
add_action('wp_ajax_myplugin_save', function () {
    check_ajax_referer('myplugin_nonce', 'nonce');

    if (!current_user_can('manage_options')) {
        wp_send_json_error('Unauthorized', 403);
    }

    $value = sanitize_text_field($_POST['value'] ?? '');
    update_option('myplugin_value', $value);

    wp_send_json_success(['saved' => true]);
});
```

> See the [Hooks guide](/{{version}}/guides/hooks/) for action and filter patterns used in plugin development.
