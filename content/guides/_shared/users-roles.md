---
title: "Users & Roles"
summary: "User management, roles, capabilities, and the authentication system."
weight: 14
---

## Introduction

WordPress has a built-in user system with role-based access control. Every user is assigned a role, and each role has a set of capabilities that determine what the user can do. Understanding this system is essential for building secure plugins that enforce proper authorization.

## Default Roles

WordPress ships with five roles, each with progressively more capabilities:

| Role | Key Capabilities |
|------|-----------------|
| **Subscriber** | Read content, manage own profile |
| **Contributor** | Write and edit own posts (cannot publish) |
| **Author** | Publish and manage own posts |
| **Editor** | Manage and publish all posts, moderate comments |
| **Administrator** | Full access to all administration features |

In multisite installations, a sixth role exists:

| Role | Scope |
|------|-------|
| **Super Admin** | Network-wide administration across all sites |

## Checking Capabilities

Always check capabilities rather than roles. A site owner may customize which capabilities each role has:

```php
// Check if the current user can edit posts
if (current_user_can('edit_posts')) {
    // Show the editing interface
}

// Check against a specific post
if (current_user_can('edit_post', $post_id)) {
    // Allow editing this specific post
}

// Check for admin-level access
if (current_user_can('manage_options')) {
    // Show settings page
}
```

### Common Capabilities

| Capability | Who has it (default) |
|-----------|---------------------|
| `read` | All roles |
| `edit_posts` | Contributor and above |
| `publish_posts` | Author and above |
| `edit_others_posts` | Editor and above |
| `manage_options` | Administrator only |
| `install_plugins` | Administrator only |
| `edit_theme_options` | Administrator only |
| `manage_categories` | Editor and above |
| `upload_files` | Author and above |

## Creating Custom Roles

```php
// Add a role on plugin activation
register_activation_hook(__FILE__, function () {
    add_role('shop_manager', 'Shop Manager', [
        'read'              => true,
        'edit_posts'        => true,
        'edit_others_posts' => true,
        'publish_posts'     => true,
        'manage_categories' => true,
        'upload_files'      => true,
        // Custom capabilities
        'manage_products'   => true,
        'manage_orders'     => true,
    ]);
});

// Remove the role on plugin deactivation
register_deactivation_hook(__FILE__, function () {
    remove_role('shop_manager');
});
```

> Roles are stored in the database. Only call `add_role()` once (e.g., on plugin activation), not on every request.

## Adding Custom Capabilities

Add capabilities to existing roles to integrate with your plugin's features:

```php
register_activation_hook(__FILE__, function () {
    $editor = get_role('editor');
    $editor->add_cap('manage_products');
    $editor->add_cap('view_reports');

    $admin = get_role('administrator');
    $admin->add_cap('manage_products');
    $admin->add_cap('view_reports');
    $admin->add_cap('manage_shop_settings');
});
```

Then check these capabilities in your code:

```php
add_action('admin_menu', function () {
    add_menu_page(
        'Products',
        'Products',
        'manage_products',     // Required capability
        'products',
        'render_products_page',
        'dashicons-cart',
        30
    );
});
```

## User Meta

Store additional data for users with the user meta API:

```php
// Set user meta
update_user_meta($user_id, 'favorite_color', 'blue');

// Get user meta
$color = get_user_meta($user_id, 'favorite_color', true);

// Delete user meta
delete_user_meta($user_id, 'favorite_color');
```

### Registering Meta for the REST API

```php
register_meta('user', 'favorite_color', [
    'type'              => 'string',
    'single'            => true,
    'show_in_rest'      => true,
    'sanitize_callback' => 'sanitize_text_field',
]);
```

## Working with User Objects

```php
// Get the current user
$user = wp_get_current_user();
echo $user->display_name;
echo $user->user_email;

// Get a specific user
$user = get_userdata(42);

// Get user by email
$user = get_user_by('email', 'jane@example.com');

// Get user by login
$user = get_user_by('login', 'jdoe');
```

## Querying Users

```php
$authors = new WP_User_Query([
    'role'     => 'author',
    'orderby'  => 'registered',
    'order'    => 'DESC',
    'number'   => 10,
]);

foreach ($authors->get_results() as $user) {
    echo $user->display_name . ' (' . $user->user_email . ')';
}
```

### Query by Meta

```php
$premium_users = new WP_User_Query([
    'meta_query' => [
        [
            'key'   => 'subscription_level',
            'value' => 'premium',
        ],
    ],
]);
```

## Creating Users Programmatically

```php
$user_id = wp_insert_user([
    'user_login'   => 'jdoe',
    'user_pass'    => wp_generate_password(),
    'user_email'   => 'jdoe@example.com',
    'display_name' => 'Jane Doe',
    'role'         => 'author',
]);

if (is_wp_error($user_id)) {
    error_log($user_id->get_error_message());
} else {
    // Optionally send a password reset email
    wp_new_user_notification($user_id, null, 'user');
}
```

## Hooks

### User Lifecycle

| Hook | When it fires |
|------|--------------|
| `user_register` | After a new user is created |
| `profile_update` | After a user profile is updated |
| `delete_user` | Before a user is deleted |
| `deleted_user` | After a user is deleted |

### Authentication

| Hook | When it fires |
|------|--------------|
| `wp_login` | After successful login |
| `wp_logout` | After logout |
| `wp_login_failed` | After a failed login attempt |
| `authenticate` | During authentication (filter) |

```php
// Log failed login attempts
add_action('wp_login_failed', function ($username) {
    error_log("Failed login attempt for user: {$username} from IP: {$_SERVER['REMOTE_ADDR']}");
});

// Restrict login to verified email addresses
add_filter('authenticate', function ($user, $username, $password) {
    if ($user instanceof WP_User) {
        $verified = get_user_meta($user->ID, 'email_verified', true);
        if (!$verified) {
            return new WP_Error('unverified', 'Please verify your email address.');
        }
    }
    return $user;
}, 30, 3);
```
