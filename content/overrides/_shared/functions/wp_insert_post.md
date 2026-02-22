## Examples & Best Practices

### Basic Usage

```php
$post_id = wp_insert_post([
    'post_title'   => 'My New Post',
    'post_content' => 'Hello, world!',
    'post_status'  => 'publish',
    'post_type'    => 'post',
]);

if (is_wp_error($post_id)) {
    error_log('Failed to create post: ' . $post_id->get_error_message());
}
```

### With Custom Fields

```php
$post_id = wp_insert_post([
    'post_title'  => 'Product Review',
    'post_status' => 'draft',
    'post_type'   => 'review',
    'meta_input'  => [
        'rating'     => 5,
        'product_id' => 42,
    ],
]);
```

### Common Gotchas

- Always check for `WP_Error` — the function returns a `WP_Error` object on failure, not `false`.
- The `post_name` (slug) is auto-generated from `post_title` if not provided.
- Passing `'ID'` in the array will **update** the existing post instead of creating a new one.
- Use `wp_update_post()` explicitly when updating to make intent clear.
