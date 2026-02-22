## WordPress 6.7 Notes

### Block Bindings Support

In WordPress 6.7, blocks registered with `register_block_type()` can declare binding sources in their `block.json`. The Block Bindings API is now stable.

To make your block's attributes bindable, no changes to `register_block_type()` are needed — bindings are configured in the block markup, not the registration.

However, if you're creating a **binding source**, register it alongside your block:

```php
add_action('init', function () {
    register_block_type(__DIR__ . '/build');

    register_block_bindings_source('myplugin/dynamic', [
        'label'              => 'Dynamic Data',
        'get_value_callback' => function ($args) {
            return get_option($args['key'] ?? '', '');
        },
    ]);
});
```

### viewScriptModule

WordPress 6.7 recommends using `viewScriptModule` instead of `viewScript` in `block.json` for Interactivity API blocks. This loads the view script as a native ES module:

```json
{
  "viewScriptModule": "file:./view.js"
}
```

The `register_block_type()` function automatically handles both `viewScript` (classic) and `viewScriptModule` (ES module) loading.
