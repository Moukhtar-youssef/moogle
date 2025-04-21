<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
use Illuminate\Support\Facades\Redis;
use App\Http\Controllers\QuerySearchController;

class RedisController extends Controller
{
    public function get_top_searches(Request $request)
    {
        // Fetch the top searches from Redis
        $topSearches = Redis::zrevrange('top_searches', 0, -1);

        // If there are no top searches, return an empty array
        if (!$topSearches) {
            return response()->json(['searches' => []]);
        }

        $top10Searches = array_slice($topSearches, 0, 10);

        // Return the results as a JSON response
        return response()->json([
            'searches' => $top10Searches
        ]);
    }

    public function get_search_suggestions(Request $request)
    {
        // Get the search term from the request
        $searchTerm = $request->get('q');

        // Fetch top searches from Redis
        $topSearches = Redis::zrevrange('top_searches', 0, -1);

        // If no top searches, return empty
        if (empty($topSearches)) {
            return response()->json(['searches' => []]);
        }

        // Case-insensitive regex-like matching
        $suggestions = array_filter($topSearches, function ($search) use ($searchTerm) {
            // Convert both to lowercase for case-insensitive matching
            return stripos($search, $searchTerm) === 0;
        });

        // Limit to top 10 suggestions
        $suggestions = array_slice($suggestions, 0, 10);

        return response()->json([
            'searches' => array_values($suggestions)
        ]);
    }

    public function cringe(Request $request)
    {
        // Fetch the top searches from Redis
        $topSearches = Redis::zrevrange('top_searches', 0, -1);

        // Fetch the number of searches performed from Redis
        $totalSearches = Redis::get('total_searches');

        // Call the get_random_page function from QuerySearchController
        $querySearchController = new QuerySearchController();
        $topRankedPage = $querySearchController->get_top_ranked_page($request);

        error_log('Top Ranked Page: ' . json_encode($topRankedPage));

        // Return view
        return view('cringe-results', [
            'topSearches' => $topSearches,
            'totalSearches' => $totalSearches,
            'topRankedPage' => $topRankedPage,
        ]);
    }

}
