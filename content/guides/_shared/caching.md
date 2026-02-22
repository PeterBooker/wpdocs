---
title: "Caching"
summary: "Transients, object cache, fragment caching, and performance optimization."
weight: 9
---

## Introduction

WordPress provides several caching layers to reduce database queries and expensive computations. Understanding when to use each one is the difference between a site that handles ten visitors per second and one that handles ten thousand.

## Transients

Transients store **temporary data with an expiration time** in the database (or in the object cache, if one is configured). They are the simplest caching mechanism in WordPress and the right choice for most plugins.

```php
// Try to get cached data
$data = get_transient('weather_data');

if (false === $data) {
    // Cache miss — fetch fresh data
    $response = wp_remote_get('https://api.example.com/weather');
    $data = json_decode(wp_remote_retrieve_body($response), true);

    // Cache for 1 hour
    set_transient('weather_data', $data, HOUR_IN_SECONDS);
}

// Use $data
```

### Expiration Constants

WordPress provides human-readable constants for common durations:

| Constant | Value |
|----------|-------|
| `MINUTE_IN_SECONDS` | 60 |
| `HOUR_IN_SECONDS` | 3,600 |
| `DAY_IN_SECONDS` | 86,400 |
| `WEEK_IN_SECONDS` | 604,800 |
| `MONTH_IN_SECONDS` | 2,592,000 |
| `YEAR_IN_SECONDS` | 31,536,000 |

### Deleting Transients

```php
// Delete a specific transient
delete_transient('weather_data');
```

### Site Transients

In multisite installations, use site transients to store data shared across the network:

```php
set_site_transient('network_stats', $stats, DAY_IN_SECONDS);
$stats = get_site_transient('network_stats');
delete_site_transient('network_stats');
```

> Transients with no expiration are autoloaded on every request. Always set an expiration to keep the `options` table lean.

## Object Cache

The WordPress object cache stores data **for the duration of a single request** by default. With a persistent object cache plugin (Redis, Memcached), it persists across requests.

```php
// Store a value (default: current request only)
wp_cache_set('user_42_permissions', $permissions, 'my_plugin', HOUR_IN_SECONDS);

// Retrieve it
$permissions = wp_cache_get('user_42_permissions', 'my_plugin');

if (false === $permissions) {
    // Cache miss
    $permissions = compute_permissions(42);
    wp_cache_set('user_42_permissions', $permissions, 'my_plugin', HOUR_IN_SECONDS);
}

// Delete
wp_cache_delete('user_42_permissions', 'my_plugin');
```

### Cache Groups

Groups namespace your cache keys to avoid collisions with other plugins:

```php
wp_cache_set('key', $value, 'my_plugin_group');
$value = wp_cache_get('key', 'my_plugin_group');
```

### When to Use Object Cache vs. Transients

| Feature | Transients | Object Cache |
|---------|-----------|--------------|
| Persistent by default | Yes (database) | No (request-scoped) |
| Persistent with plugin | Yes (uses object cache) | Yes (Redis, Memcached) |
| Best for | External API responses, expensive queries | Repeated lookups within a request |
| Network overhead | Database query | In-memory or socket to cache server |

> When a persistent object cache is active, transients are stored _in_ the object cache (not the database), making them equivalent in performance.

## WP_Query Caching

`WP_Query` has built-in caching that you should be aware of:

```php
$query = new WP_Query([
    'post_type'      => 'post',
    'posts_per_page' => 10,
    'cache_results'  => true,   // Default: true — caches post objects
    'no_found_rows'  => true,   // Skip SQL_CALC_FOUND_ROWS when you don't need pagination
]);
```

### Optimizing Queries

```php
// Fast: only need IDs
$ids = new WP_Query([
    'post_type'              => 'product',
    'fields'                 => 'ids',
    'no_found_rows'          => true,
    'update_post_meta_cache' => false,
    'update_post_term_cache' => false,
]);

// Fast: only need a count
$count = new WP_Query([
    'post_type'              => 'product',
    'posts_per_page'         => 1,
    'no_found_rows'          => false,   // We need the count
    'fields'                 => 'ids',
    'update_post_meta_cache' => false,
    'update_post_term_cache' => false,
]);
$total = $count->found_posts;
```

## Fragment Caching

For expensive template fragments (like a sidebar widget that queries the database), cache the rendered HTML:

```php
function render_popular_posts_widget() {
    $html = get_transient('popular_posts_widget');

    if (false === $html) {
        ob_start();

        $popular = new WP_Query([
            'post_type'      => 'post',
            'posts_per_page' => 5,
            'meta_key'       => 'view_count',
            'orderby'        => 'meta_value_num',
            'order'          => 'DESC',
        ]);

        if ($popular->have_posts()) {
            echo '<ul class="popular-posts">';
            while ($popular->have_posts()) {
                $popular->the_post();
                echo '<li><a href="' . esc_url(get_permalink()) . '">'
                    . esc_html(get_the_title()) . '</a></li>';
            }
            echo '</ul>';
            wp_reset_postdata();
        }

        $html = ob_get_clean();
        set_transient('popular_posts_widget', $html, 15 * MINUTE_IN_SECONDS);
    }

    echo $html;
}
```

## Cache Invalidation

The hardest part of caching is knowing when to invalidate. Use WordPress hooks to clear caches when underlying data changes:

```php
// Clear cache when a post is saved
add_action('save_post', function ($post_id) {
    delete_transient('popular_posts_widget');
    delete_transient('sidebar_cache');
});

// Clear cache when a term is updated
add_action('edited_term', function ($term_id, $tt_id, $taxonomy) {
    if ($taxonomy === 'category') {
        delete_transient('category_list_cache');
    }
}, 10, 3);
```

## Performance Tips

1. **Set `no_found_rows` to `true`** when you don't need pagination — it eliminates a `SQL_CALC_FOUND_ROWS` query
2. **Disable meta/term caches** when you only need post IDs or titles
3. **Use `fields => 'ids'`** to avoid hydrating full post objects when you don't need them
4. **Cache external API responses** with transients — never call an external API on every page load
5. **Install a persistent object cache** (Redis or Memcached) on production sites
6. **Batch meta lookups** — `update_postmeta_cache()` for a list of IDs is faster than individual `get_post_meta()` calls in a loop
