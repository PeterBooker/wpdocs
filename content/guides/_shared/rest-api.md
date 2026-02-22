---
title: "REST API"
summary: "Building and consuming the WordPress REST API."
weight: 5
---

## Overview

The WordPress REST API provides HTTP endpoints for interacting with your site programmatically. It powers the block editor, the WordPress mobile apps, and can serve as a backend for headless frontends.

## Built-in Endpoints

WordPress exposes endpoints for all core data types at `/wp-json/wp/v2/`:

| Endpoint | Description |
|----------|-------------|
| `/wp/v2/posts` | Posts |
| `/wp/v2/pages` | Pages |
| `/wp/v2/categories` | Categories |
| `/wp/v2/tags` | Tags |
| `/wp/v2/users` | Users |
| `/wp/v2/media` | Media attachments |
| `/wp/v2/comments` | Comments |

### Fetching Posts

```bash
# List published posts
curl https://example.com/wp-json/wp/v2/posts

# Single post
curl https://example.com/wp-json/wp/v2/posts/42

# Filter by category
curl https://example.com/wp-json/wp/v2/posts?categories=5

# Search
curl https://example.com/wp-json/wp/v2/posts?search=hello
```

### From JavaScript

```js
// Using the built-in apiFetch (available in the block editor)
import apiFetch from '@wordpress/api-fetch';

const posts = await apiFetch({ path: '/wp/v2/posts?per_page=5' });

// Or with native fetch
const response = await fetch('/wp-json/wp/v2/posts?per_page=5');
const posts = await response.json();
```

## Registering Custom Endpoints

Use `register_rest_route()` to add your own API endpoints:

```php
add_action('rest_api_init', function () {
    register_rest_route('myplugin/v1', '/items', [
        'methods'             => 'GET',
        'callback'            => 'myplugin_get_items',
        'permission_callback' => '__return_true',
    ]);
});

function myplugin_get_items(WP_REST_Request $request) {
    $per_page = $request->get_param('per_page') ?? 10;

    $items = get_posts([
        'post_type'      => 'item',
        'posts_per_page' => $per_page,
    ]);

    return rest_ensure_response($items);
}
```

### Route Parameters

```php
register_rest_route('myplugin/v1', '/items/(?P<id>\d+)', [
    'methods'             => 'GET',
    'callback'            => 'myplugin_get_item',
    'permission_callback' => '__return_true',
    'args'                => [
        'id' => [
            'validate_callback' => function ($param) {
                return is_numeric($param);
            },
        ],
    ],
]);

function myplugin_get_item(WP_REST_Request $request) {
    $post = get_post($request['id']);
    if (!$post) {
        return new WP_Error('not_found', 'Item not found', ['status' => 404]);
    }
    return rest_ensure_response($post);
}
```

## Authentication

Public endpoints work without authentication. For protected operations (create, update, delete), you need authentication:

### Application Passwords

WordPress supports Application Passwords for REST API authentication:

```bash
# Using basic auth with an application password
curl -u "username:xxxx xxxx xxxx xxxx" \
     -X POST https://example.com/wp-json/wp/v2/posts \
     -H "Content-Type: application/json" \
     -d '{"title": "New Post", "status": "draft"}'
```

### Nonce Authentication (same-origin)

For JavaScript running on the same WordPress site, use nonce authentication:

```php
// Enqueue script with nonce
wp_enqueue_script('my-api-script', plugin_dir_url(__FILE__) . 'script.js');
wp_localize_script('my-api-script', 'myApi', [
    'nonce' => wp_create_nonce('wp_rest'),
    'root'  => esc_url_raw(rest_url()),
]);
```

```js
// In script.js
const response = await fetch(myApi.root + 'wp/v2/posts', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
        'X-WP-Nonce': myApi.nonce,
    },
    body: JSON.stringify({ title: 'New Post', status: 'draft' }),
});
```

## Modifying REST Responses

### Adding Fields

```php
add_action('rest_api_init', function () {
    register_rest_field('post', 'reading_time', [
        'get_callback' => function ($post) {
            $content = get_post_field('post_content', $post['id']);
            $words = str_word_count(strip_tags($content));
            return ceil($words / 200); // minutes
        },
        'schema' => [
            'type'        => 'integer',
            'description' => 'Estimated reading time in minutes',
        ],
    ]);
});
```

### Permission Callbacks

Always define a `permission_callback` for your routes:

```php
'permission_callback' => function () {
    return current_user_can('edit_posts');
}
```

> Browse the [Classes reference](/{{version}}/classes/) for `WP_REST_Request`, `WP_REST_Response`, and `WP_REST_Controller`.
