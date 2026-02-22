## Examples & Best Practices

### Basic Usage

```php
add_action('init', 'my_custom_init');

function my_custom_init() {
    register_post_type('book', ['public' => true, 'label' => 'Books']);
}
```

### With Priority and Arguments

```php
// Run late (priority 99) and accept 2 arguments
add_action('save_post', 'my_save_handler', 99, 2);

function my_save_handler($post_id, $post) {
    if ($post->post_type !== 'product') {
        return;
    }
    update_post_meta($post_id, '_last_modified_by', get_current_user_id());
}
```

### Anonymous Functions

```php
add_action('wp_enqueue_scripts', function () {
    wp_enqueue_style('my-theme', get_stylesheet_uri());
});
```

> **Note:** Anonymous functions cannot be removed with `remove_action()`. Use named functions or class methods when removal may be needed.

### Class Methods

```php
class MyPlugin {
    public function __construct() {
        add_action('init', [$this, 'register_post_types']);
        add_action('admin_menu', [$this, 'add_admin_page']);
    }

    public function register_post_types() {
        register_post_type('event', ['public' => true]);
    }

    public function add_admin_page() {
        add_menu_page('Events', 'Events', 'manage_options', 'events', [$this, 'render_page']);
    }
}
new MyPlugin();
```

### Common Hook Priorities

| Priority | Use case |
|----------|----------|
| `1-9` | Run before most callbacks |
| `10` | Default — most hooks |
| `20-50` | Run after standard callbacks |
| `99+` | Run last — cleanup, final modifications |
| `PHP_INT_MAX` | Absolute last |
