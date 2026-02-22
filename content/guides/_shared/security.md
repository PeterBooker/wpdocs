---
title: "Security"
summary: "Data validation, sanitization, escaping, nonces, and secure coding practices."
weight: 7
---

## Introduction

WordPress processes untrusted input from users, browsers, and external APIs on every request. Secure code follows a simple principle: **validate on input, escape on output**. This guide covers the core security APIs that make this practical.

## Data Validation

Validation checks whether data meets expected criteria _before_ you use it. WordPress provides helpers for common checks, but you should always validate at the boundary where data enters your code.

```php
// Check types
$page = absint($_GET['page'] ?? 1);              // Positive integer
$email = is_email($input) ? $input : '';          // Valid email or empty

// Check against allowed values
$status = in_array($input, ['draft', 'publish', 'private'], true)
    ? $input
    : 'draft';

// Validate URLs
$url = wp_http_validate_url($input);
if (!$url) {
    wp_die('Invalid URL provided.');
}
```

> Always use strict comparisons (`===`, `in_array(..., true)`) when validating. Loose comparisons in PHP can produce unexpected results.

## Sanitization

Sanitization cleans input data, removing or encoding anything dangerous. Use it when _storing_ user input.

### Common Sanitization Functions

| Function | Purpose |
|----------|---------|
| `sanitize_text_field()` | Strip tags, remove extra whitespace, encode entities |
| `sanitize_textarea_field()` | Like `sanitize_text_field` but preserves line breaks |
| `sanitize_email()` | Remove characters not allowed in email addresses |
| `sanitize_file_name()` | Remove special characters from filenames |
| `sanitize_html_class()` | Ensure a string is a valid CSS class name |
| `sanitize_key()` | Lowercase alphanumeric with dashes and underscores |
| `sanitize_title()` | Create a URL-safe slug |
| `sanitize_url()` | Clean a URL for safe storage |
| `wp_kses()` | Allow only specific HTML tags and attributes |
| `wp_kses_post()` | Allow HTML appropriate for post content |

```php
// Saving user input
update_post_meta($post_id, 'subtitle', sanitize_text_field($_POST['subtitle']));
update_post_meta($post_id, 'bio', sanitize_textarea_field($_POST['bio']));
update_post_meta($post_id, 'website', sanitize_url($_POST['website']));
```

### Allowing Specific HTML

When you need to accept HTML but control which tags are allowed, use `wp_kses()`:

```php
$allowed_tags = [
    'a'      => ['href' => [], 'title' => [], 'target' => []],
    'strong' => [],
    'em'     => [],
    'br'     => [],
];

$clean_html = wp_kses($user_html, $allowed_tags);
```

## Escaping (Output)

Escaping converts special characters for safe output in a given context. Always escape immediately before output — never escape then store.

### Escaping Functions

| Function | Context | Example |
|----------|---------|---------|
| `esc_html()` | Inside HTML elements | `<p><?php echo esc_html($text); ?></p>` |
| `esc_attr()` | Inside HTML attributes | `<input value="<?php echo esc_attr($val); ?>">` |
| `esc_url()` | In `href`, `src`, or URL contexts | `<a href="<?php echo esc_url($url); ?>">` |
| `esc_js()` | Inside inline JavaScript | Avoid inline JS when possible |
| `esc_textarea()` | Inside `<textarea>` elements | `<textarea><?php echo esc_textarea($text); ?></textarea>` |
| `wp_kses_post()` | Render post-like HTML | `<?php echo wp_kses_post($content); ?>` |

```php
// Always escape at the point of output
<h2><?php echo esc_html($title); ?></h2>
<a href="<?php echo esc_url($link); ?>"><?php echo esc_html($label); ?></a>
<div class="<?php echo esc_attr($class); ?>">
    <?php echo wp_kses_post($content); ?>
</div>
```

> **Never** trust data just because it came from the database. Database values may have been stored before your sanitization code existed, or another plugin may have written to the same row.

## Nonces

Nonces (number used once) protect against **Cross-Site Request Forgery (CSRF)**. They verify that a request originated from your site and was intentional.

### Creating Nonces

```php
// In a form
<form method="post">
    <?php wp_nonce_field('save_settings', 'my_nonce'); ?>
    <input type="text" name="option_value" />
    <button type="submit">Save</button>
</form>

// In a URL
$url = wp_nonce_url(admin_url('admin-post.php?action=delete_item&id=42'), 'delete_item_42');
```

### Verifying Nonces

```php
// Verify a form nonce
if (!isset($_POST['my_nonce']) || !wp_verify_nonce($_POST['my_nonce'], 'save_settings')) {
    wp_die('Security check failed.');
}

// Verify a URL nonce
check_admin_referer('delete_item_42');
```

### AJAX Nonces

```php
// Enqueue with nonce
wp_localize_script('my-script', 'myAjax', [
    'url'   => admin_url('admin-ajax.php'),
    'nonce' => wp_create_nonce('my_ajax_action'),
]);

// Verify in the handler
add_action('wp_ajax_my_action', function () {
    check_ajax_referer('my_ajax_action', 'nonce');
    // Handle the request
});
```

## Capability Checks

Always verify the current user has permission to perform an action. Combine with nonces for complete protection:

```php
add_action('admin_post_save_settings', function () {
    // 1. Verify nonce (CSRF protection)
    check_admin_referer('save_settings', 'my_nonce');

    // 2. Verify capabilities (authorization)
    if (!current_user_can('manage_options')) {
        wp_die('You do not have permission to do this.');
    }

    // 3. Sanitize and save
    $value = sanitize_text_field($_POST['option_value']);
    update_option('my_option', $value);

    wp_safe_redirect(admin_url('options-general.php?page=my-settings&saved=1'));
    exit;
});
```

## Database Queries

When writing direct SQL queries, always use `$wpdb->prepare()` to prevent **SQL injection**:

```php
global $wpdb;

// Parameterized query — safe
$results = $wpdb->get_results(
    $wpdb->prepare(
        "SELECT * FROM {$wpdb->posts} WHERE post_type = %s AND post_status = %s",
        'product',
        'publish'
    )
);

// NEVER interpolate user input into SQL
// $wpdb->query("DELETE FROM {$wpdb->posts} WHERE ID = {$_GET['id']}"); // DANGEROUS
```

> Prefer `WP_Query`, `get_posts()`, or other high-level APIs when possible. Use `$wpdb` directly only when the query API doesn't cover your needs.

## File Uploads

Handle file uploads with WordPress's built-in functions rather than moving files manually:

```php
// Check file type
$filetype = wp_check_filetype($filename, null);
if (!$filetype['ext']) {
    wp_die('File type not allowed.');
}

// Use WordPress's upload handler
$upload = wp_handle_upload($_FILES['my_file'], ['test_form' => false]);
if (isset($upload['error'])) {
    wp_die($upload['error']);
}
```

## Security Checklist

When reviewing plugin or theme code, verify each point:

- [ ] All form submissions include and verify a nonce
- [ ] All user input is sanitized before storage
- [ ] All output is escaped with the context-appropriate function
- [ ] Capability checks guard every privileged action
- [ ] Direct database queries use `$wpdb->prepare()`
- [ ] File operations use WordPress upload functions
- [ ] AJAX handlers verify nonces and capabilities
- [ ] No sensitive data is exposed in JavaScript variables or HTML source
