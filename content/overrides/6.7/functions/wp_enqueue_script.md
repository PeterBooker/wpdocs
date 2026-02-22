## WordPress 6.7 Notes

### Script Modules

WordPress 6.7 improves support for ES modules alongside the classic script system. For Interactivity API blocks, prefer `viewScriptModule` in `block.json` over `wp_enqueue_script()`:

```json
{
  "viewScriptModule": "file:./view.js"
}
```

For standalone ES module scripts outside of blocks, use the Script Modules API:

```php
add_action('wp_enqueue_scripts', function () {
    wp_register_script_module('my-module', plugin_dir_url(__FILE__) . 'js/app.js', [
        ['id' => '@wordpress/interactivity', 'import' => 'static'],
    ]);
    wp_enqueue_script_module('my-module');
});
```

### When to Use What

| Approach | Use case |
|----------|----------|
| `wp_enqueue_script()` | Classic scripts, jQuery plugins, legacy code |
| `wp_register_script_module()` | ES modules, Interactivity API, modern JS |
| `viewScriptModule` in block.json | Block-specific interactive scripts |
