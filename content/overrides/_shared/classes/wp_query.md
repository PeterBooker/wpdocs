## Examples & Best Practices

### Basic Query

```php
$query = new WP_Query([
    'post_type'      => 'post',
    'posts_per_page' => 10,
    'post_status'    => 'publish',
]);

while ($query->have_posts()) {
    $query->the_post();
    printf('<h2>%s</h2>', get_the_title());
}
wp_reset_postdata();
```

### Complex Filtering

Combine taxonomy queries, meta queries, and date queries:

```php
$query = new WP_Query([
    'post_type'  => 'product',
    'tax_query'  => [
        [
            'taxonomy' => 'category',
            'field'    => 'slug',
            'terms'    => 'electronics',
        ],
    ],
    'meta_query' => [
        [
            'key'     => 'price',
            'value'   => [10, 100],
            'type'    => 'NUMERIC',
            'compare' => 'BETWEEN',
        ],
    ],
    'date_query' => [
        ['after' => '2024-01-01'],
    ],
    'orderby'  => 'meta_value_num',
    'meta_key' => 'price',
    'order'    => 'ASC',
]);
```

### Performance Tips

- Use `'fields' => 'ids'` when you only need post IDs
- Use `'no_found_rows' => true` when you don't need pagination
- Use `'update_post_meta_cache' => false` if you won't access post meta
- Use `'update_post_term_cache' => false` if you won't access taxonomy terms
- Avoid `'posts_per_page' => -1` on large sites — always set a reasonable limit

```php
// Optimized query when you only need IDs
$ids = new WP_Query([
    'post_type'      => 'post',
    'posts_per_page' => 100,
    'fields'         => 'ids',
    'no_found_rows'  => true,
]);
```

### Common Gotchas

- Always call `wp_reset_postdata()` after a custom query loop
- `WP_Query` modifies the global `$post` via `the_post()` — this can break the main loop if not reset
- `'post__in' => []` (empty array) returns **all** posts, not zero. Check before querying
- The `offset` parameter overrides `paged` — use one or the other, not both
