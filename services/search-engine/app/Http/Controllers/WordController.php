<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
use Illuminate\Support\Facades\Redis;

class WordController extends Controller
{
    public function search()
    {
        $query = request()->query('q');
        error_log("Query: $query");
        if (!$query) {
            return response()->json(['error' => 'No query provided'], 400);
        }

        // Initialize empty hashmap
        // hashmap[url] = [
        //      "count" => int,
        //      "score" => int,
        // ]
        $wordHashmap = [];

        // Split the query string
        $words = explode(' ', strtolower($query));

        // Create a pipline to fetch words from the query
        $pipeline = Redis::pipeline();
        foreach ($words as $word) {
            $pipeline->zrevrange("word:$word", 0, -1, true);
        }

        // Execute batch call
        $wordResults = $pipeline->exec();
        if (!$wordResults) {
            return response()->json(['error' => 'No word results'], 404);
        }

        $pipeline = Redis::pipeline();
        $urlKeys = [];

        // Rank pages
        foreach ($wordResults as $word) {
            foreach ($word as $url => $score) {
                // Initialize hashmap entry if it doesn't exist
                if (!isset($wordHashmap[$url])) {
                    $wordHashmap[$url] = [
                        'count' => 0,
                        'score' => 0,
                    ];
                }

                $pipeline->smembers("backlinks:$url");
                $urlKeys[] = $url;

                // TODO: Implement page rank here
                // Update count and score
                $wordHashmap[$url]['count'] += 1;
                $wordHashmap[$url]['score'] += $score;
                $wordHashmap[$url]['rank'] = 0; // Placeholder for backlinks
            }
        }

        $backlinksResults = $pipeline->exec();
        foreach ($backlinksResults as $index => $backlinks) {
            $url = $urlKeys[$index];
            $total = count($backlinks);
            $wordHashmap[$url]['rank'] = $total;
            /*error_log("Page $url has $total backlinks");*/
        }

        // Sort the results to rank them
        $sortedWordResults = collect($wordHashmap)->sortByDesc(function ($value, $key) {
            return $value['count'] * 1000 + 500 * $value['score'] + 300 * $value['rank'];
        });

        // Initialize empty hashmap
        // urlMetadataHashmap[url] = [
        //      "title" => str,
        //      "description" => str,
        //      "text" => str,
        // ]
        $urlMetadataHashmap = [];

        // Fetch metadata to render the pages in the frontend
        $counter = 0;

        // Create another pipeline to fetch metadata in batches
        $pipeline = Redis::pipeline();
        foreach($sortedWordResults as $url => $data) {
            if ($counter >= 50) break;
            $counter += 1;
            $pipeline->hgetall("url_metadata:$url");
        }

        $urlMetadataResults = $pipeline->exec();
        foreach($urlMetadataResults as $urlMetadata) {
            $title = $urlMetadata['title'];
            $description = $urlMetadata['description'];
            $text = $urlMetadata['summary_text'];
            $url = $urlMetadata['normalized_url'];

            // Create metadata entry
            $urlMetadataHashmap[$url] = [
                'title' => $title,
                'description' => $description,
                'text' => $text,
            ];
        }

        // Return view for SSR
        return view('search-results', [
            'query' => $query,
            'results' => $urlMetadataHashmap,
        ]);
    }
}
