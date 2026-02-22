---
title: "Scheduled Events (WP-Cron)"
summary: "Schedule recurring tasks, one-time events, and background processing with WP-Cron."
weight: 15
---

## Introduction

WP-Cron is WordPress's built-in task scheduler. Unlike a system cron that runs on a fixed schedule, WP-Cron is triggered by page visits — when someone loads your site, WordPress checks if any scheduled events are due and runs them.

## Scheduling a Recurring Event

Register your event on plugin activation and define the callback that runs on each occurrence:

```php
// Schedule on activation
register_activation_hook(__FILE__, function () {
    if (!wp_next_scheduled('myplugin_daily_cleanup')) {
        wp_schedule_event(time(), 'daily', 'myplugin_daily_cleanup');
    }
});

// Define the callback
add_action('myplugin_daily_cleanup', function () {
    global $wpdb;

    // Delete expired tokens older than 30 days
    $wpdb->query($wpdb->prepare(
        "DELETE FROM {$wpdb->prefix}myplugin_tokens WHERE expires_at < %s",
        gmdate('Y-m-d H:i:s', strtotime('-30 days'))
    ));
});

// Unschedule on deactivation
register_deactivation_hook(__FILE__, function () {
    $timestamp = wp_next_scheduled('myplugin_daily_cleanup');
    if ($timestamp) {
        wp_unschedule_event($timestamp, 'myplugin_daily_cleanup');
    }
});
```

### Built-in Schedules

| Schedule | Interval |
|----------|----------|
| `hourly` | Once every hour |
| `twicedaily` | Once every 12 hours |
| `daily` | Once every 24 hours |
| `weekly` | Once every 7 days |

### Custom Schedules

Register custom intervals with the `cron_schedules` filter:

```php
add_filter('cron_schedules', function ($schedules) {
    $schedules['every_fifteen_minutes'] = [
        'interval' => 15 * MINUTE_IN_SECONDS,
        'display'  => 'Every 15 Minutes',
    ];

    $schedules['every_five_minutes'] = [
        'interval' => 5 * MINUTE_IN_SECONDS,
        'display'  => 'Every 5 Minutes',
    ];

    return $schedules;
});

// Now you can use it
wp_schedule_event(time(), 'every_fifteen_minutes', 'myplugin_check_inventory');
```

## One-Time Events

Schedule a task to run once at a specific time:

```php
// Run a task 1 hour from now
wp_schedule_single_event(time() + HOUR_IN_SECONDS, 'myplugin_send_welcome_email', [$user_id]);

// The callback receives the arguments
add_action('myplugin_send_welcome_email', function ($user_id) {
    $user = get_userdata($user_id);
    wp_mail($user->user_email, 'Welcome!', 'Thanks for joining.');
});
```

### Passing Arguments

Both `wp_schedule_event()` and `wp_schedule_single_event()` accept an array of arguments that are passed to the callback:

```php
wp_schedule_single_event(
    time() + MINUTE_IN_SECONDS * 5,
    'myplugin_process_order',
    [$order_id, 'express']    // Arguments array
);

add_action('myplugin_process_order', function ($order_id, $shipping_method) {
    // Process the order
}, 10, 2);
```

## Checking Scheduled Events

```php
// Get the next scheduled timestamp for an event
$next = wp_next_scheduled('myplugin_daily_cleanup');

if ($next) {
    echo 'Next run: ' . gmdate('Y-m-d H:i:s', $next);
} else {
    echo 'Event is not scheduled.';
}
```

## Unscheduling Events

```php
// Unschedule a specific occurrence
$timestamp = wp_next_scheduled('myplugin_daily_cleanup');
wp_unschedule_event($timestamp, 'myplugin_daily_cleanup');

// Unschedule all occurrences of an event
wp_unschedule_hook('myplugin_daily_cleanup');

// Clear all scheduled events (use with caution)
wp_clear_scheduled_hook('myplugin_daily_cleanup');
```

## WP-Cron Limitations

WP-Cron is triggered by page visits, which means:

1. **Low-traffic sites** may not run events on time — if nobody visits the site for 6 hours, a hourly event won't fire during that period
2. **Heavy tasks block the page load** that triggered them unless handled carefully
3. **No guaranteed execution time** — events run on the _next_ page load after they're due, not at the exact scheduled time

### Using System Cron Instead

For production sites that need reliable timing, disable WP-Cron's page-visit trigger and use a real system cron:

```php
// In wp-config.php
define('DISABLE_WP_CRON', true);
```

Then set up a system cron to hit the WP-Cron endpoint:

```bash
# Run WP-Cron every minute via system cron
* * * * * curl -s https://example.com/wp-cron.php?doing_wp_cron > /dev/null 2>&1

# Or with WP-CLI (preferred)
* * * * * cd /var/www/html && wp cron event run --due-now > /dev/null 2>&1
```

## Background Processing Pattern

For long-running tasks that shouldn't block a page load, split the work into small batches:

```php
// Schedule the batch processor
add_action('myplugin_process_batch', function () {
    $items = get_option('myplugin_pending_items', []);

    if (empty($items)) {
        return;
    }

    // Process 10 items per batch
    $batch = array_splice($items, 0, 10);
    update_option('myplugin_pending_items', $items);

    foreach ($batch as $item_id) {
        process_item($item_id);
    }

    // If there are more items, schedule another batch
    if (!empty($items)) {
        wp_schedule_single_event(time() + 30, 'myplugin_process_batch');
    }
});
```

## Debugging with WP-CLI

WP-CLI provides commands for inspecting and managing scheduled events:

```bash
# List all scheduled events
wp cron event list

# Run a specific event immediately
wp cron event run myplugin_daily_cleanup

# Run all events that are due
wp cron event run --due-now

# Delete a scheduled event
wp cron event delete myplugin_daily_cleanup

# List registered cron schedules
wp cron schedule list

# Test if WP-Cron is working
wp cron test
```
