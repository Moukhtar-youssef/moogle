<?php

namespace App\Providers;

use Illuminate\Support\ServiceProvider;
use Illuminate\Support\Facades\URL;

class AppServiceProvider extends ServiceProvider
{
    /**
     * Register any application services.
     */
    public function register(): void
    {
        //
    }

    /**
     * Bootstrap any application services.
     */
    public function boot(): void
    {
        // Force HTTPS globally
        URL::forceScheme('https');

        // Add security headers
        $this->app['router']->middleware(function ($request, $next) {
            $response = $next($request);

            return $response->withHeaders([
                'Strict-Transport-Security' => 'max-age=31536000; includeSubDomains',
                'X-Frame-Options' => 'SAMEORIGIN',
                'X-Content-Type-Options' => 'nosniff',
            ]);
        });
    }
}
