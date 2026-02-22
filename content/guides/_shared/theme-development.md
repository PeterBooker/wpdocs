---
title: "Theme Development"
summary: "Build custom WordPress themes from scratch."
weight: 3
---

## Overview

WordPress themes control the visual presentation and layout of a site. This guide covers the fundamentals of building a theme from scratch.

## Template Hierarchy

WordPress uses a template hierarchy to decide which PHP file renders a given page. The system falls back through increasingly general templates:

```
Single Post:    single-{post_type}-{slug}.php → single-{post_type}.php → single.php → singular.php → index.php
Page:           page-{slug}.php → page-{id}.php → page.php → singular.php → index.php
Archive:        archive-{post_type}.php → archive.php → index.php
Category:       category-{slug}.php → category-{id}.php → category.php → archive.php → index.php
```

## Theme Setup

Register theme features in `functions.php` using the `after_setup_theme` action:

```php
add_action('after_setup_theme', function () {
    // Enable featured images
    add_theme_support('post-thumbnails');

    // Register navigation menus
    register_nav_menus([
        'primary' => 'Primary Navigation',
        'footer'  => 'Footer Navigation',
    ]);

    // Set content width
    $GLOBALS['content_width'] = 1200;
});
```

## Enqueuing Assets

Always use `wp_enqueue_scripts` to load CSS and JavaScript:

```php
add_action('wp_enqueue_scripts', function () {
    wp_enqueue_style('theme-style', get_stylesheet_uri(), [], '1.0.0');
    wp_enqueue_script('theme-script', get_template_directory_uri() . '/js/app.js', [], '1.0.0', true);
});
```

## The Loop

The Loop is WordPress's core mechanism for displaying posts:

```php
<?php if (have_posts()) : ?>
    <?php while (have_posts()) : the_post(); ?>
        <article>
            <h2><a href="<?php the_permalink(); ?>"><?php the_title(); ?></a></h2>
            <?php the_excerpt(); ?>
        </article>
    <?php endwhile; ?>
    <?php the_posts_pagination(); ?>
<?php else : ?>
    <p>No posts found.</p>
<?php endif; ?>
```

## Template Parts

Split templates into reusable parts with `get_template_part()`:

```php
// In index.php
get_template_part('template-parts/content', get_post_type());

// Loads: template-parts/content-{post_type}.php
// Falls back to: template-parts/content.php
```

## Custom Queries

Use `WP_Query` for custom post queries outside the main loop:

```php
$books = new WP_Query([
    'post_type'      => 'book',
    'posts_per_page' => 10,
    'orderby'        => 'title',
    'order'          => 'ASC',
]);

if ($books->have_posts()) {
    while ($books->have_posts()) {
        $books->the_post();
        the_title('<h3>', '</h3>');
    }
    wp_reset_postdata();
}
```

> See the [WP_Query class reference](/{{version}}/classes/wp_query/) for all available parameters.
