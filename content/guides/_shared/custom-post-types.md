---
title: "Custom Post Types & Taxonomies"
summary: "Register and work with custom post types, taxonomies, and meta fields."
weight: 8
---

## Introduction

WordPress stores most content as "posts" — but not everything is a blog post. Custom post types let you create distinct content types like products, events, or portfolios, each with their own admin screens, templates, and REST API endpoints. Custom taxonomies add structured categorization to any post type.

## Registering a Post Type

Use `register_post_type()` on the `init` hook. The first argument is the post type slug (max 20 characters, no uppercase):

```php
add_action('init', function () {
    register_post_type('product', [
        'labels' => [
            'name'               => 'Products',
            'singular_name'      => 'Product',
            'add_new_item'       => 'Add New Product',
            'edit_item'          => 'Edit Product',
            'view_item'          => 'View Product',
            'search_items'       => 'Search Products',
            'not_found'          => 'No products found.',
            'not_found_in_trash' => 'No products found in Trash.',
        ],
        'public'       => true,
        'has_archive'  => true,
        'menu_icon'    => 'dashicons-cart',
        'supports'     => ['title', 'editor', 'thumbnail', 'excerpt', 'revisions'],
        'show_in_rest' => true,   // Required for the block editor
        'rewrite'      => ['slug' => 'products'],
    ]);
});
```

> After registering a post type with `has_archive` or custom rewrite rules, flush permalinks by visiting **Settings > Permalinks** in the admin, or call `flush_rewrite_rules()` on plugin activation.

### Key Arguments

| Argument | Default | Purpose |
|----------|---------|---------|
| `public` | `false` | Makes the post type visible in the admin and on the frontend |
| `has_archive` | `false` | Creates an archive page at the rewrite slug |
| `show_in_rest` | `false` | Exposes to the REST API and enables the block editor |
| `supports` | `['title', 'editor']` | Editor features to enable |
| `menu_position` | `null` | Position in the admin menu (5 = below Posts, 20 = below Pages) |
| `capability_type` | `'post'` | Base capability type for permissions |
| `taxonomies` | `[]` | Taxonomies to connect at registration time |

### Supported Features

The `supports` argument controls which editor features appear:

```php
'supports' => [
    'title',           // Post title
    'editor',          // Main content editor
    'thumbnail',       // Featured image
    'excerpt',         // Excerpt field
    'revisions',       // Revision history
    'custom-fields',   // Custom fields meta box
    'page-attributes', // Page parent and menu order
    'comments',        // Comments support
],
```

## Registering a Taxonomy

Custom taxonomies add structured categorization. Register them on `init`, then connect them to one or more post types:

```php
add_action('init', function () {
    register_taxonomy('product_category', ['product'], [
        'labels' => [
            'name'          => 'Product Categories',
            'singular_name' => 'Product Category',
            'search_items'  => 'Search Categories',
            'all_items'     => 'All Categories',
            'edit_item'     => 'Edit Category',
            'add_new_item'  => 'Add New Category',
        ],
        'hierarchical' => true,     // true = category-like, false = tag-like
        'show_in_rest' => true,     // Required for block editor
        'rewrite'      => ['slug' => 'product-category'],
    ]);
});
```

### Hierarchical vs. Flat

| Type | `hierarchical` | UI | Example |
|------|---------------|-----|---------|
| Category-like | `true` | Checkbox tree with parent/child | Genres, departments |
| Tag-like | `false` | Comma-separated text input | Skills, ingredients |

## Custom Fields (Post Meta)

Register meta fields for structured data that doesn't belong in the post content:

```php
add_action('init', function () {
    register_post_meta('product', 'price', [
        'type'              => 'number',
        'single'            => true,
        'show_in_rest'      => true,
        'sanitize_callback' => 'absint',
        'default'           => 0,
    ]);

    register_post_meta('product', 'sku', [
        'type'              => 'string',
        'single'            => true,
        'show_in_rest'      => true,
        'sanitize_callback' => 'sanitize_text_field',
    ]);
});
```

### Reading and Writing Meta

```php
// Set a value
update_post_meta($post_id, 'price', 2999);

// Get a value
$price = get_post_meta($post_id, 'price', true);

// Delete a value
delete_post_meta($post_id, 'price');
```

### Meta in WP_Query

Query posts by custom field values using `meta_query`:

```php
$expensive = new WP_Query([
    'post_type'  => 'product',
    'meta_query' => [
        [
            'key'     => 'price',
            'value'   => 5000,
            'compare' => '>=',
            'type'    => 'NUMERIC',
        ],
    ],
    'orderby'  => 'meta_value_num',
    'meta_key' => 'price',
    'order'    => 'ASC',
]);
```

> For frequently queried meta fields, consider adding a database index on the `meta_key`/`meta_value` columns, or using a custom table for high-volume data.

## Querying Custom Post Types

```php
// Get all published products
$products = new WP_Query([
    'post_type'      => 'product',
    'posts_per_page' => 12,
    'post_status'    => 'publish',
]);

// Loop
if ($products->have_posts()) {
    while ($products->have_posts()) {
        $products->the_post();
        echo '<h2>' . get_the_title() . '</h2>';
        echo '<p>Price: $' . get_post_meta(get_the_ID(), 'price', true) . '</p>';
    }
    wp_reset_postdata();
}
```

### Taxonomy Queries

```php
$electronics = new WP_Query([
    'post_type' => 'product',
    'tax_query' => [
        [
            'taxonomy' => 'product_category',
            'field'    => 'slug',
            'terms'    => 'electronics',
        ],
    ],
]);
```

## Template Hierarchy

WordPress automatically looks for templates matching your custom post type:

| Template | Purpose |
|----------|---------|
| `single-product.php` | Single product page |
| `archive-product.php` | Product archive page |
| `taxonomy-product_category.php` | Product category archive |
| `taxonomy-product_category-electronics.php` | Specific term archive |

In block themes, create corresponding templates in the Site Editor or as HTML files:

```
templates/
├── single-product.html
├── archive-product.html
└── taxonomy-product_category.html
```

## REST API

Post types registered with `'show_in_rest' => true` are automatically available via the REST API:

```
GET /wp-json/wp/v2/product
GET /wp-json/wp/v2/product/42
GET /wp-json/wp/v2/product?product_category=5
```

Custom meta fields registered with `'show_in_rest' => true` appear in the response's `meta` object and are writable via `POST`/`PUT` requests.

## Practical Example

Here's a complete plugin that registers a "Book" post type with a genre taxonomy and custom fields:

```php
<?php
/**
 * Plugin Name: Books
 * Description: A custom post type for managing a book collection.
 */

add_action('init', function () {
    register_post_type('book', [
        'labels'       => ['name' => 'Books', 'singular_name' => 'Book'],
        'public'       => true,
        'has_archive'  => true,
        'menu_icon'    => 'dashicons-book',
        'supports'     => ['title', 'editor', 'thumbnail', 'excerpt'],
        'show_in_rest' => true,
        'rewrite'      => ['slug' => 'books'],
        'taxonomies'   => ['genre'],
    ]);

    register_taxonomy('genre', ['book'], [
        'labels'       => ['name' => 'Genres', 'singular_name' => 'Genre'],
        'hierarchical' => true,
        'show_in_rest' => true,
        'rewrite'      => ['slug' => 'genre'],
    ]);

    register_post_meta('book', 'isbn', [
        'type'              => 'string',
        'single'            => true,
        'show_in_rest'      => true,
        'sanitize_callback' => 'sanitize_text_field',
    ]);

    register_post_meta('book', 'page_count', [
        'type'              => 'integer',
        'single'            => true,
        'show_in_rest'      => true,
        'sanitize_callback' => 'absint',
    ]);
});
```
