#!/bin/bash
set -e

# Clear Laravel caches if artisan exists
if [ -f /var/www/artisan ]; then
    php artisan config:clear
    php artisan cache:clear
    php artisan view:clear
    php artisan route:clear
    php artisan optimize:clear
fi

# Execute the main command
exec "$@"
