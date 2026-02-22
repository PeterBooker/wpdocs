## WordPress 6.6 Changes

### Caching Improvements

WordPress 6.6 extended the query caching system introduced in 6.1. Queries using `'cache_results' => true` (the default) now integrate with the object cache more efficiently, reducing database load on repeated queries.

### Per-Block CSS and WP_Query

When building custom blocks that use `WP_Query` on the front end (e.g., a "Related Posts" block), ensure your block's `render.php` returns quickly. WordPress 6.6's per-block CSS loading means your render callback runs during the initial content parse:

```php
// render.php — keep queries lean
$related = new WP_Query([
    'post_type'              => 'post',
    'posts_per_page'         => 3,
    'post__not_in'           => [get_the_ID()],
    'no_found_rows'          => true,
    'update_post_meta_cache' => false,
    'update_post_term_cache' => false,
    'fields'                 => 'ids',
]);
```

### `post__in` Empty Array Behavior

A reminder that `'post__in' => []` still returns all posts (not zero). This was a common source of bugs. Always guard:

```php
$ids = get_saved_post_ids(); // might return []
if (!empty($ids)) {
    $query = new WP_Query(['post__in' => $ids, 'orderby' => 'post__in']);
}
```
