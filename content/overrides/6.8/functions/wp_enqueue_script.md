## WordPress 6.8 Notes

### Speculative Loading Compatibility

In WordPress 6.8, scripts enqueued with `wp_enqueue_script()` work seamlessly with the new Speculative Loading feature. Prerendered pages will execute your scripts as expected.

If your script has side effects that should not run during prerender, check the document state:

```php
wp_add_inline_script('my-script', "
    if (document.prerendering) {
        document.addEventListener('prerenderingchange', initMyScript);
    } else {
        initMyScript();
    }
");
```

### Strategy Parameter

The `$args` parameter (added in 6.3) supports loading strategies. In 6.8, `defer` is recommended over `async` for most use cases:

```php
wp_enqueue_script(
    'my-script',
    plugin_dir_url(__FILE__) . 'js/app.js',
    ['wp-dom-ready'],
    '1.0.0',
    ['strategy' => 'defer', 'in_footer' => true]
);
```
