#!/bin/bash
set -e

# Fix permissions
mkdir -p /var/www/storage/logs
mkdir -p /var/www/storage/framework/sessions
mkdir -p /var/www/storage/framework/views
mkdir -p /var/www/storage/framework/cache
mkdir -p /var/www/bootstrap/cache
chown -R www-data:www-data /var/www/storage
chown -R www-data:www-data /var/www/bootstrap/cache
chmod -R 775 /var/www/storage
chmod -R 775 /var/www/bootstrap/cache

# Clear Laravel caches if artisan exists
if [ -f /var/www/artisan ]; then
    php artisan config:clear
    php artisan cache:clear
    php artisan view:clear
    php artisan route:clear
    php artisan optimize:clear
fi

# Start PHP-FPM in background
php-fpm -D

# Start Caddy in foreground
caddy run --config /etc/caddy/Caddyfile
