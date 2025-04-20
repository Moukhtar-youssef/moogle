<?php

namespace App\Http\Middleware;

use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;
use Illuminate\Support\Facades\Redis;

class StoreSearchTerm
{
    /**
     * Handle an incoming request.
     *
     * @param  \Closure(\Illuminate\Http\Request): (\Symfony\Component\HttpFoundation\Response)  $next
     */
    public function handle(Request $request, Closure $next): Response
    {
        // Store the search term in Redis:
        $searchTerm = $request->get('processed_query');

        if (empty($searchTerm)) {
            return $next($request);
        }
        $searchTerm = trim($searchTerm);

        // Increment the search term count in Redis
        Redis::zincrby('top_searches', 1, strtolower($searchTerm));
        // Keep only the top 100 search terms
        Redis::zremrangebyrank('top_searches', 0, -101);

        return $next($request);
    }
}
