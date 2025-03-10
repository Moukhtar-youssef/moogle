<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
use Illuminate\Support\Facades\Redis;

class WordController extends Controller
{
    public function getWord($key)
    {
        error_log("word:$key");
        $value = Redis::zrevrange("word:$key", 0, -1, true);
        if ($value) {
            return response()->json(['value' => $value], 200);
        }

        return response()->json(['error' => 'Key not found'], 404);
    }

    public function search()
    {
        $query = request()->query('q');
        error_log("Query: $query");
        if (!$query) {
            return response()->json(['error' => 'No query provided'], 400);
        }

        /*error_log('Searching for words:');*/

        // Initialize empty hashmap
        // hashmap[url] = [
        //      "count" => int,
        //      "score" => int,
        // ]
        $wordHashmap = [];

        // Split the query string
        $words = explode(' ', strtolower($query));

        // Create a batch redis calls
        $pipeline = Redis::pipeline();
        foreach ($words as $word) {
            $pipeline->zrevrange("word:$word", 0, -1, true);
        }

        // Execute batch call
        $wordResults = $pipeline->exec();
        if (!$wordResults) {
            return response()->json(['error' => 'No word results'], 404);
        }

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

                // Update count and score
                $wordHashmap[$url]['count'] += 1;
                $wordHashmap[$url]['score'] += $score;
            }
        }

        // Sort the results to rank them
        $sortedWordResults = collect($wordHashmap)->sortByDesc(function ($value, $key) {
            return $value['count'] * 1000 + $value['score'];
        });

        // Initialize empty hashmap
        // urlMetadataHashmap[url] = [
        //      "title" => str,
        //      "description" => str,
        //      "text" => str,
        // ]
        $urlMetadataHashmap = [];

        // Fetch metadata to render the pages in the frontend
        /*$counter = 0;*/

        // FIXME: In order to implement the pipeline,
        // I need to modify the metadata entry in the database
        // to also include the url.
        // Create another pipeline to fetch metadata in batches
        /*$pipeline = Redis::pipeline();*/
        /*foreach($sortedWordResults as $url => $data) {*/
        /*    if ($counter >= 50) break;*/
        /*    $counter += 1;*/
        /*    $pipeline->hgetall("url_metadata:$url");*/
        /*}*/
        /**/
        /*$urlMetadataResults = $pipeline->exec();*/
        /*foreach($urlMetadataResults as $urlMetadata) {*/
        /*    $title = $urlMetadata['title'];*/
        /*    $description = $urlMetadata['description'];*/
        /*    $text = $urlMetadata['text'];*/
        /**/
        /*    // Create urlMetadataHashmap entry*/
        /*}*/

        $counter = 0;
        foreach ($sortedWordResults as $url => $data) {
            // TODO: Configure how to check for max results
            if ($counter >= 50) {
                break;
            }
            $counter += 1;

            // Fetch metadata
            $resp = Redis::hgetall("url_metadata:$url");
            if (!$resp) {
                /*error_log("url_metadata:$url not found");*/
                continue;
            }

            $title = $resp['title'];
            $description = $resp['description'];
            $text = $resp['summary_text'];

            // Create metadata entry
            $urlMetadataHashmap[$url] = [
                'title' => $title,
                'description' => $description,
                'text' => $text,
            ];
        }

        // FIXME: Return a view for SSR
        // return response()->json(['response' => $urlMetadataHashmap], 200);
        return view('search-results', [
            'query' => $query,
            'results' => $urlMetadataHashmap,
        ]);
    }
}
