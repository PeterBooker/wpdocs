---
title: "HTTP API"
summary: "Making HTTP requests with wp_remote_get, wp_remote_post, and the HTTP API."
weight: 12
---

## Introduction

WordPress provides a unified HTTP API for making outbound requests to external services. It abstracts away the differences between cURL, PHP streams, and other transports, giving you a consistent interface with built-in timeout handling, SSL verification, and WordPress-specific hooks.

## GET Requests

Use `wp_remote_get()` to fetch data from an external URL:

```php
$response = wp_remote_get('https://api.example.com/posts');

// Always check for errors first
if (is_wp_error($response)) {
    error_log('API request failed: ' . $response->get_error_message());
    return;
}

$status_code = wp_remote_retrieve_response_code($response);
$body        = wp_remote_retrieve_body($response);

if (200 !== $status_code) {
    error_log("API returned status {$status_code}");
    return;
}

$data = json_decode($body, true);
```

### Request Arguments

```php
$response = wp_remote_get('https://api.example.com/data', [
    'timeout'     => 15,                          // Seconds (default: 5)
    'headers'     => [
        'Authorization' => 'Bearer ' . $api_key,
        'Accept'        => 'application/json',
    ],
    'sslverify'   => true,                        // Verify SSL certificate (default: true)
    'user-agent'  => 'MyPlugin/1.0',
]);
```

## POST Requests

Use `wp_remote_post()` to send data:

```php
$response = wp_remote_post('https://api.example.com/submit', [
    'body' => [
        'name'  => 'Jane Doe',
        'email' => 'jane@example.com',
    ],
    'timeout' => 15,
]);

if (is_wp_error($response)) {
    wp_die('Request failed: ' . $response->get_error_message());
}

$result = json_decode(wp_remote_retrieve_body($response), true);
```

### Sending JSON

```php
$response = wp_remote_post('https://api.example.com/webhook', [
    'headers' => [
        'Content-Type' => 'application/json',
    ],
    'body'    => wp_json_encode([
        'event' => 'order.created',
        'data'  => ['order_id' => 123],
    ]),
    'timeout' => 15,
]);
```

## Other Methods

For PUT, DELETE, PATCH, and HEAD requests, use `wp_remote_request()`:

```php
// PUT request
$response = wp_remote_request('https://api.example.com/items/42', [
    'method'  => 'PUT',
    'headers' => ['Content-Type' => 'application/json'],
    'body'    => wp_json_encode(['name' => 'Updated Item']),
]);

// DELETE request
$response = wp_remote_request('https://api.example.com/items/42', [
    'method' => 'DELETE',
]);

// HEAD request (retrieve headers only)
$response = wp_remote_head('https://api.example.com/status');
$headers  = wp_remote_retrieve_headers($response);
```

## Response Helpers

WordPress provides functions to safely extract parts of the response:

```php
$response = wp_remote_get($url);

// Status code (int)
$code = wp_remote_retrieve_response_code($response);    // 200

// Status message (string)
$message = wp_remote_retrieve_response_message($response); // "OK"

// Response body (string)
$body = wp_remote_retrieve_body($response);

// All headers (Requests_Utility_CaseInsensitiveDictionary)
$headers = wp_remote_retrieve_headers($response);

// Single header
$content_type = wp_remote_retrieve_header($response, 'content-type');

// Cookies
$cookies = wp_remote_retrieve_cookies($response);
```

## Caching API Responses

External API calls are slow and may be rate-limited. Always cache responses with transients:

```php
function get_github_repos($username) {
    $cache_key = 'github_repos_' . sanitize_key($username);
    $repos = get_transient($cache_key);

    if (false !== $repos) {
        return $repos;
    }

    $response = wp_remote_get("https://api.github.com/users/{$username}/repos", [
        'headers' => ['Accept' => 'application/vnd.github.v3+json'],
        'timeout' => 10,
    ]);

    if (is_wp_error($response) || 200 !== wp_remote_retrieve_response_code($response)) {
        return [];
    }

    $repos = json_decode(wp_remote_retrieve_body($response), true);
    set_transient($cache_key, $repos, HOUR_IN_SECONDS);

    return $repos;
}
```

## Error Handling

`wp_remote_get()` and related functions return a `WP_Error` object on failure. Always check before accessing the response:

```php
$response = wp_remote_get($url, ['timeout' => 10]);

if (is_wp_error($response)) {
    // Network error, timeout, DNS failure, etc.
    $error_message = $response->get_error_message();
    $error_code    = $response->get_error_code();    // e.g., 'http_request_failed'

    error_log("HTTP request to {$url} failed: [{$error_code}] {$error_message}");
    return new WP_Error('api_error', 'Unable to reach the service.');
}

// HTTP-level errors (4xx, 5xx)
$code = wp_remote_retrieve_response_code($response);
if ($code >= 400) {
    error_log("API returned HTTP {$code} for {$url}");
    return new WP_Error('api_http_error', "API returned status {$code}");
}
```

## Filtering Requests

WordPress fires hooks that let you modify or intercept all outbound HTTP requests:

```php
// Modify request arguments globally
add_filter('http_request_args', function ($args, $url) {
    // Add a custom header to all requests to your API
    if (str_contains($url, 'api.example.com')) {
        $args['headers']['X-Plugin-Version'] = '2.0';
    }
    return $args;
}, 10, 2);

// Block requests to specific domains
add_filter('pre_http_request', function ($preempt, $args, $url) {
    if (str_contains($url, 'analytics.tracker.com')) {
        return new WP_Error('blocked', 'Request blocked by policy.');
    }
    return $preempt;
}, 10, 3);
```

## Timeouts and Limits

| Setting | Default | Notes |
|---------|---------|-------|
| `timeout` | 5 seconds | Increase for slow APIs, but keep it reasonable |
| `redirection` | 5 | Maximum number of redirects to follow |
| `blocking` | `true` | Set to `false` for fire-and-forget requests |
| `sslverify` | `true` | Never disable in production |

For non-blocking requests (logging, analytics):

```php
wp_remote_post('https://api.example.com/log', [
    'blocking' => false,    // Don't wait for the response
    'body'     => ['event' => 'page_view'],
]);
```

> Non-blocking requests still consume server resources to initiate the connection. For high-volume events, consider using WP-Cron to batch and send them asynchronously.
