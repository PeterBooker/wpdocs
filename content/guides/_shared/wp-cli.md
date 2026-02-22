---
title: "WP-CLI"
summary: "Manage WordPress from the command line — core operations, custom commands, and scripting."
weight: 11
---

## Introduction

[WP-CLI](https://wp-cli.org/) is the official command-line interface for WordPress. It lets you manage plugins, run database operations, generate scaffolding, and automate workflows without opening a browser. For developers, it's as essential as `composer` or `npm`.

## Installation

```bash
# Download the phar
curl -O https://raw.githubusercontent.com/wp-cli/builds/gh-pages/phar/wp-cli.phar

# Make it executable
chmod +x wp-cli.phar

# Move to PATH
sudo mv wp-cli.phar /usr/local/bin/wp

# Verify
wp --info
```

## Essential Commands

### Core Management

```bash
# Download WordPress
wp core download

# Install WordPress
wp core install --url=example.com --title="My Site" \
    --admin_user=admin --admin_email=admin@example.com

# Update WordPress
wp core update

# Check version
wp core version
```

### Plugins

```bash
# Install and activate a plugin
wp plugin install woocommerce --activate

# List all plugins
wp plugin list

# Deactivate a plugin
wp plugin deactivate akismet

# Update all plugins
wp plugin update --all

# Search the plugin directory
wp plugin search "security" --per-page=5
```

### Themes

```bash
# Install and activate a theme
wp theme install twentytwentyfive --activate

# List themes
wp theme list

# Delete an inactive theme
wp theme delete twentytwentythree
```

### Database

```bash
# Export the database
wp db export backup.sql

# Import a database
wp db import backup.sql

# Run a query
wp db query "SELECT COUNT(*) FROM wp_posts WHERE post_type='post'"

# Search and replace (dry run first)
wp search-replace 'http://old-domain.com' 'https://new-domain.com' --dry-run

# Run the replacement
wp search-replace 'http://old-domain.com' 'https://new-domain.com' --precise
```

### Posts and Content

```bash
# Create a post
wp post create --post_type=post --post_title="Hello World" --post_status=publish

# List posts
wp post list --post_type=page --post_status=publish

# Delete a post
wp post delete 42 --force

# Generate test posts
wp post generate --count=50 --post_type=post
```

### Users

```bash
# Create a user
wp user create editor editor@example.com --role=editor

# List users
wp user list --role=administrator

# Update a user's role
wp user set-role 5 author
```

### Options

```bash
# Get an option
wp option get blogname

# Set an option
wp option update blogname "New Site Name"

# Delete an option
wp option delete my_plugin_option

# List autoloaded options sorted by size
wp option list --autoload=on --format=table --orderby=size_bytes --order=desc
```

## Scaffolding

WP-CLI can generate boilerplate code for plugins, themes, blocks, and more:

```bash
# Scaffold a plugin
wp scaffold plugin my-plugin --plugin_name="My Plugin" --plugin_author="Developer"

# Scaffold a custom post type
wp scaffold post-type product --label="Products" --textdomain=my-plugin

# Scaffold a taxonomy
wp scaffold taxonomy genre --post_types=book --label="Genres"

# Scaffold a block
wp scaffold block notice --title="Notice" --plugin=my-plugin

# Scaffold PHPUnit tests for a plugin
wp scaffold plugin-tests my-plugin
```

## Custom Commands

Register your own WP-CLI commands for plugin-specific tasks:

```php
if (defined('WP_CLI') && WP_CLI) {
    WP_CLI::add_command('mystore', 'MyStore_CLI');
}

class MyStore_CLI {
    /**
     * Sync products from the external API.
     *
     * ## OPTIONS
     *
     * [--limit=<number>]
     * : Number of products to sync. Default: all.
     *
     * [--dry-run]
     * : Show what would be synced without making changes.
     *
     * ## EXAMPLES
     *
     *     wp mystore sync --limit=10
     *     wp mystore sync --dry-run
     *
     * @subcommand sync
     */
    public function sync($args, $assoc_args) {
        $limit  = (int) ($assoc_args['limit'] ?? 0);
        $dry_run = isset($assoc_args['dry-run']);

        $products = $this->fetch_products($limit);

        $progress = \WP_CLI\Utils\make_progress_bar('Syncing products', count($products));

        foreach ($products as $product) {
            if (!$dry_run) {
                $this->import_product($product);
            }
            $progress->tick();
        }

        $progress->finish();
        WP_CLI::success(sprintf('Synced %d products.', count($products)));
    }
}
```

### Output Helpers

```php
// Success, warning, and error messages
WP_CLI::success('Operation completed.');
WP_CLI::warning('No records found.');
WP_CLI::error('Failed to connect.'); // Exits with code 1

// Tables
WP_CLI\Utils\format_items('table', $items, ['id', 'name', 'status']);

// Progress bars
$progress = \WP_CLI\Utils\make_progress_bar('Processing', $total);
foreach ($items as $item) {
    // process
    $progress->tick();
}
$progress->finish();

// Confirmation prompts
WP_CLI::confirm('Are you sure you want to delete all transients?');
```

## Cron Management

```bash
# List scheduled events
wp cron event list

# Run a specific event
wp cron event run my_daily_import

# Run all due events
wp cron event run --due-now

# Schedule a new event
wp cron event schedule my_task hourly

# Test if WP-Cron is working
wp cron test
```

## Maintenance and Debugging

```bash
# Flush rewrite rules
wp rewrite flush

# Flush the object cache
wp cache flush

# Delete all transients
wp transient delete --all

# Regenerate thumbnails
wp media regenerate --yes

# Check for WordPress updates
wp core check-update

# Run a PHP expression
wp eval 'echo home_url();'

# Profile a page load
wp profile stage --url=https://example.com/sample-page
```

## Shell Scripting

WP-CLI integrates naturally into shell scripts for automation:

```bash
#!/usr/bin/env bash
set -euo pipefail

SITE_PATH="/var/www/mysite"

# Maintenance mode on
wp maintenance-mode activate --path="$SITE_PATH"

# Backup database
wp db export "$SITE_PATH/backup-$(date +%Y%m%d).sql" --path="$SITE_PATH"

# Update everything
wp core update --path="$SITE_PATH"
wp plugin update --all --path="$SITE_PATH"
wp theme update --all --path="$SITE_PATH"

# Flush caches
wp cache flush --path="$SITE_PATH"
wp rewrite flush --path="$SITE_PATH"

# Maintenance mode off
wp maintenance-mode deactivate --path="$SITE_PATH"

echo "Update complete!"
```
