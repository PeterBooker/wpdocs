## WordPress 6.6 Notes

### Per-Block CSS Loading

WordPress 6.6 introduced per-block CSS loading, which makes `wp_enqueue_block_style()` more important than ever. Block stylesheets registered this way are only loaded when the block appears on the page.

### Registering Block Styles

```php
add_action('init', function () {
    // Style that loads only when core/quote is on the page
    wp_enqueue_block_style('core/quote', [
        'handle' => 'mytheme-quote-style',
        'src'    => get_theme_file_uri('assets/css/blocks/quote.css'),
        'ver'    => '1.0.0',
        'path'   => get_theme_file_path('assets/css/blocks/quote.css'), // Enables inline loading
    ]);
});
```

### Inline vs External Loading

When you provide the `path` argument, WordPress can inline the CSS in the `<head>` for small files rather than making a separate HTTP request. This is the recommended approach in 6.6:

```php
wp_enqueue_block_style('core/group', [
    'handle' => 'mytheme-group-style',
    'src'    => get_theme_file_uri('assets/css/blocks/group.css'),
    'path'   => get_theme_file_path('assets/css/blocks/group.css'),
]);
```

### Theme Block Styles

For block themes, place CSS files in `assets/css/blocks/{block-name}.css` and register them in `functions.php`. This pattern replaces the older approach of bundling all block styles into a single `style.css`.
