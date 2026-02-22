## WordPress 6.8 Changes

### Performance Improvements

WordPress 6.8 includes internal optimizations to `WP_Query` that reduce memory usage for large result sets. The `fields => 'ids'` mode is now faster due to reduced object allocation.

### Improved Cache Invalidation

Query caching in 6.8 is more granular. Saving a post now only invalidates queries that could have included that post, rather than flushing all query caches. This benefits sites with object caching enabled.

### New: `update_menu_item_cache` Parameter

```php
// When querying nav menu items, enable this for better performance
$query = new WP_Query([
    'post_type'              => 'nav_menu_item',
    'posts_per_page'         => -1,
    'update_menu_item_cache' => true,
]);
```
