## Examples & Best Practices

### Basic Custom Post Type

```php
add_action('init', function () {
    register_post_type('product', [
        'labels' => [
            'name'               => 'Products',
            'singular_name'      => 'Product',
            'add_new_item'       => 'Add New Product',
            'edit_item'          => 'Edit Product',
            'new_item'           => 'New Product',
            'view_item'          => 'View Product',
            'search_items'       => 'Search Products',
            'not_found'          => 'No products found',
            'not_found_in_trash' => 'No products found in trash',
        ],
        'public'       => true,
        'has_archive'  => true,
        'show_in_rest' => true,
        'supports'     => ['title', 'editor', 'thumbnail', 'excerpt', 'custom-fields'],
        'rewrite'      => ['slug' => 'products'],
        'menu_icon'    => 'dashicons-cart',
        'menu_position' => 5,
    ]);
});
```

### With Block Editor Support

`'show_in_rest' => true` is required for the block editor. Without it, the post type uses the classic editor.

For custom fields in the block editor, also include `'custom-fields'` in `supports`:

```php
register_post_type('recipe', [
    'public'       => true,
    'show_in_rest' => true,
    'supports'     => ['title', 'editor', 'thumbnail', 'custom-fields'],
    'template'     => [
        ['core/paragraph', ['placeholder' => 'Add the recipe description...']],
        ['core/heading', ['level' => 3, 'content' => 'Ingredients']],
        ['core/list'],
        ['core/heading', ['level' => 3, 'content' => 'Instructions']],
        ['core/list', ['ordered' => true]],
    ],
    'template_lock' => 'all', // 'all', 'insert', or false
]);
```

### Internal Post Types (No UI)

```php
register_post_type('log_entry', [
    'public'              => false,
    'show_ui'             => false,
    'show_in_rest'        => false,
    'exclude_from_search' => true,
    'supports'            => ['title'],
]);
```

### Common Gotchas

- Post type names must be **max 20 characters**, lowercase, no spaces
- Always call `register_post_type()` in the `init` hook — not earlier
- After changing `rewrite` slugs, you must flush permalinks (visit Settings → Permalinks or call `flush_rewrite_rules()`)
- Setting `'public' => true` implies `'show_ui'`, `'show_in_nav_menus'`, and `'exclude_from_search' => false` — override individually if needed
