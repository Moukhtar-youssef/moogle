FROM php:8.3-fpm

ARG user=query-engine-user
ARG uid=1000

# Install dependencies
RUN apt-get update && apt-get install -y \
    git \
    curl \
    zip \
    unzip \
    libssl-dev \
    libcurl4-openssl-dev \
    pkg-config \
    libonig-dev \
    libxml2-dev \
    libpng-dev \
    nodejs \
    npm \
    debian-keyring \
    debian-archive-keyring \
    apt-transport-https \
    gnupg

# Install Caddy
RUN curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg \
    && curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list \
    && apt-get update \
    && apt-get install -y caddy

# Clear cache
RUN apt-get clean && rm -rf /var/lib/apt/lists/*

# Install PHP extensions
RUN docker-php-ext-install mbstring xml pdo pdo_mysql

# Install MongoDB extension
RUN pecl install mongodb-1.21.0 && docker-php-ext-enable mongodb

# Install composer
COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

# Create system user
RUN useradd -u $uid -ms /bin/bash -g www-data $user

# Copy application
COPY . /var/www
COPY --chown=$user:www-data . /var/www

# Create Caddyfile
RUN echo "{\n\
    auto_https off\n\
    }\n\
    :8080 {\n\
    root * /var/www/public\n\
    php_fastcgi 127.0.0.1:9000\n\
    file_server\n\
    }" > /etc/caddy/Caddyfile

# Make sure entrypoint script is executable
COPY docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

WORKDIR /var/www

# Install frontend dependencies and build assets
RUN npm install && npm run build

# Install PHP dependencies
RUN composer install --no-dev --optimize-autoloader

EXPOSE 8080

ENTRYPOINT ["docker-entrypoint.sh"]
