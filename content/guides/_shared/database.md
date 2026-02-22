---
title: "Database"
summary: "Working with the WordPress database using wpdb and WP_Query."
weight: 4
---

## Overview

WordPress uses MySQL/MariaDB and provides two primary interfaces for database interaction: the global `$wpdb` object for direct queries, and `WP_Query` for post retrieval.

## The $wpdb Object

The `$wpdb` global provides safe, abstracted access to the database.

### Selecting Data

```php
global $wpdb;

// Get a single value
$count = $wpdb->get_var("SELECT COUNT(*) FROM {$wpdb->posts} WHERE post_status = 'publish'");

// Get a single row
$post = $wpdb->get_row("SELECT * FROM {$wpdb->posts} WHERE ID = 1");

// Get multiple rows
$drafts = $wpdb->get_results("SELECT ID, post_title FROM {$wpdb->posts} WHERE post_status = 'draft'");

// Get a single column
$ids = $wpdb->get_col("SELECT ID FROM {$wpdb->posts} WHERE post_type = 'page'");
```

### Prepared Statements

Always use `$wpdb->prepare()` when including user input in queries:

```php
global $wpdb;

// String placeholder: %s, Integer placeholder: %d, Float: %f
$results = $wpdb->get_results(
    $wpdb->prepare(
        "SELECT * FROM {$wpdb->posts} WHERE post_type = %s AND post_status = %s",
        'product',
        'publish'
    )
);
```

> **Never** concatenate user input directly into SQL queries. Always use `$wpdb->prepare()`.

### Insert, Update, Delete

```php
global $wpdb;

// Insert
$wpdb->insert($wpdb->postmeta, [
    'post_id'    => 42,
    'meta_key'   => 'color',
    'meta_value' => 'blue',
], ['%d', '%s', '%s']);

// Update
$wpdb->update($wpdb->posts,
    ['post_title' => 'New Title'],    // data
    ['ID' => 42],                      // where
    ['%s'],                            // data format
    ['%d']                             // where format
);

// Delete
$wpdb->delete($wpdb->postmeta, ['meta_key' => 'deprecated_field'], ['%s']);
```

## WP_Query

`WP_Query` is the standard way to retrieve posts. It accepts dozens of parameters for filtering, sorting, and pagination.

### Basic Usage

```php
$query = new WP_Query([
    'post_type'      => 'post',
    'posts_per_page' => 10,
    'orderby'        => 'date',
    'order'          => 'DESC',
]);

while ($query->have_posts()) {
    $query->the_post();
    the_title('<h2>', '</h2>');
    the_excerpt();
}
wp_reset_postdata();
```

### Common Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `post_type` | Post type(s) to query | `'post'`, `['post', 'page']` |
| `post_status` | Post status | `'publish'`, `'draft'` |
| `posts_per_page` | Number of posts | `10`, `-1` (all) |
| `orderby` | Sort field | `'date'`, `'title'`, `'menu_order'` |
| `tax_query` | Taxonomy conditions | See below |
| `meta_query` | Custom field conditions | See below |

### Tax Query

```php
$query = new WP_Query([
    'post_type' => 'product',
    'tax_query' => [
        'relation' => 'AND',
        [
            'taxonomy' => 'category',
            'field'    => 'slug',
            'terms'    => ['electronics'],
        ],
        [
            'taxonomy' => 'product_tag',
            'field'    => 'slug',
            'terms'    => ['sale'],
        ],
    ],
]);
```

### Meta Query

```php
$query = new WP_Query([
    'post_type'  => 'product',
    'meta_query' => [
        'relation' => 'AND',
        [
            'key'     => 'price',
            'value'   => 100,
            'type'    => 'NUMERIC',
            'compare' => '>=',
        ],
        [
            'key'     => 'in_stock',
            'value'   => '1',
            'compare' => '=',
        ],
    ],
]);
```

## Custom Tables

For plugins that need their own tables:

```php
function myplugin_create_table() {
    global $wpdb;
    $table = $wpdb->prefix . 'myplugin_logs';
    $charset = $wpdb->get_charset_collate();

    $sql = "CREATE TABLE $table (
        id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
        user_id bigint(20) unsigned NOT NULL,
        action varchar(100) NOT NULL,
        created_at datetime DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY  (id),
        KEY user_id (user_id)
    ) $charset;";

    require_once ABSPATH . 'wp-admin/includes/upgrade.php';
    dbDelta($sql);
}
register_activation_hook(__FILE__, 'myplugin_create_table');
```

> Browse the [Functions reference](/{{version}}/functions/) for `wpdb` methods like `get_var`, `get_row`, `get_results`, `insert`, `update`, and `delete`.
