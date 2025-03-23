<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;

class QuerySearchController extends Controller
{
    protected function getTopImages($query, $limit = 4)
    {
        // Split the query string
        $query = str_replace('+', ' ', $query);
        $words = explode(' ', strtolower($query));

        try {
            $results = DB::connection('mongodb')->table('word_images')
                ->whereIn('_id', $words)
                ->get();

            if ($results->count() <= 0) {
                // Return view for SSR
                return [];
            }

            $wordImagesHashmap = [];
            $urls = [];

            // Rank pages
            foreach ($words as $word) {
                $pages = $results->where('id', $word)->first()->pages;
                foreach ($pages as ['url' => $url, 'weight' => $weight]) {
                    $urls[$url] = true;
                    if (!isset($wordImagesHashmap[$url])) {
                        $wordImagesHashmap[$url] = [
                            'count' => 0,
                            'weight' => 0,
                            'rank' => 1,
                            'alt' => "",
                            'filename' => "",
                            'page_url' => "",
                            'page_title' => "",
                            'page_text' => "",
                        ];
                    }

                    $wordImagesHashmap[$url]['count'] += 1;
                    $wordImagesHashmap[$url]['weight'] += $weight;
                }
            }

            $uniqueUrls = array_keys($urls);

            // Fetch image data
            $images_data = DB::connection('mongodb')->table('image')
                ->whereIn('_id', $uniqueUrls)
                ->get();

            // Calculate the rank
            $unique_page_urls = [];
            foreach ($uniqueUrls as $url) {
                $count = $wordImagesHashmap[$url]['count'];
                $weight = $wordImagesHashmap[$url]['weight'];
                $wordImagesHashmap[$url]['rank'] = 500 * $count + 500 * $weight;

                // Add image data
                $img_data = $images_data->where('id', $url)->first();
                $wordImagesHashmap[$url]['alt'] = $img_data->alt;
                $wordImagesHashmap[$url]['filename'] = $img_data->filename;
                $wordImagesHashmap[$url]['page_url'] = $img_data->page_url;
                $unique_page_urls[$img_data->page_url] = true;
            }

            $unique_page_urls = array_keys($unique_page_urls);
            // Fetch page data to get the title
            $pages_metadata = DB::connection('mongodb')->table('metadata')
                ->whereIn('_id', $unique_page_urls)
                ->get();

            // Add page metadata
            foreach ($uniqueUrls as $url) {
                $page_url = $wordImagesHashmap[$url]['page_url'];
                $current_page_metadata = $pages_metadata->where('id', $page_url)->first();

                if ($current_page_metadata) {
                    $wordImagesHashmap[$url]['page_title'] = $current_page_metadata->title;
                    $summary_text = $current_page_metadata->summary_text;
                    $wordImagesHashmap[$url]['page_text'] = strlen($summary_text) > 30
                        ? substr($summary_text, 0, 30) . '...'
                        : $summary_text;
                }
            }



            $sortedWordImagesResults = collect($wordImagesHashmap)->sortByDesc('rank');

            // Return the top N images
            return $sortedWordImagesResults->take($limit);
        } catch (Exception $e) {
            return collect(); // Return an empty collection in case of error
        }
    }

    public function search(Request $request)
    {
        // $test = 'upload.wikimedia.org/wikipedia/en/thumb/5/5a/Kamen_Rider_Ex-Aid_Trilogy_poster.jpg/220px-Kamen_Rider_Ex-Aid_Trilogy_poster.jpg';
        $query = $request->input('q');
        if (!$query) {
            return view('search-results', [
                'query' => $query,
                'results' => [],
                'total' => 0,
                'topImages' => [],
            ]);
        }

        // Split the query string
        $query = str_replace('+', ' ', $query);
        $words = explode(' ', strtolower($query));


        // Set the number of results per page
        $perPage = 20;
        $page = $request->input('page', 1); // Default page 1

        try {
            $results = DB::connection('mongodb')->table('word')
                ->whereIn('_id', $words)
                ->get();
            $wordHashmap = [];
            $urls = [];

            if ($results->count() <= 0) {
                // Return view for SSR
                return view('search-results', [
                    'query' => $query,
                    'results' => [],
                    'total' => 0,
                    'topImages' => [],
                ]);
            }

            // Rank pages
            foreach ($words as $word) {
                $pages = $results->where('id', $word)->first()->pages;
                foreach ($pages as ['url' => $url, 'weight' => $weight]) {
                    if ($url == "en.wikipedia.org/wiki/Kamen_Rider") {
                        error_log("TESINTg\n\n");
                        error_log("URL: $url");
                    }
                    $urls[$url] = true;
                    if (!isset($wordHashmap[$url])) {
                        $wordHashmap[$url] = [
                            'count' => 0,
                            'weight' => 0,
                            'backlinks' => 0,
                            'rank' => 1,
                            'description' => "",
                            'last_crawled' => "",
                            'summary_text' => "",
                            'title' => "",
                        ];
                    }

                    $wordHashmap[$url]['count'] += 1;
                    $wordHashmap[$url]['weight'] += $weight;
                }
            }

            // Calculate backlinks weight
            $uniqueUrls = array_keys($urls);
            $results = DB::connection('mongodb')->table('backlinks')
                ->whereIn('_id', $uniqueUrls)
                ->get();

            foreach ($results as $result) {
                $backlinks = count($result->links);
                $url = $result->id;
                $wordHashmap[$url]['backlinks'] = $backlinks;
            }

            // Fetch page metadata
            $pages_metadata = DB::connection('mongodb')->table('metadata')
                ->whereIn('_id', $uniqueUrls)
                ->get();

            /*return $pages_metadata;*/
            foreach ($uniqueUrls as $url) {
                $count = $wordHashmap[$url]['count'];
                $weight = $wordHashmap[$url]['weight'];
                $backlinks = $wordHashmap[$url]['backlinks'];
                $wordHashmap[$url]['rank'] = 500 * $count + 500 * $weight + 10000 * $backlinks;

                // Add page metadata
                $page_metadata = $pages_metadata->where('id', $url)->first();
                $wordHashmap[$url]['description'] = $page_metadata->description;
                $wordHashmap[$url]['last_crawled'] = $page_metadata->last_crawled;
                $wordHashmap[$url]['summary_text'] = $page_metadata->summary_text;
                $wordHashmap[$url]['title'] = $page_metadata->title;
            }

            $sortedWordResults = collect($wordHashmap)->sortByDesc('rank');

            $paginatedResults = $sortedWordResults->forPage($page, $perPage);

            // Get top 4 images if it's the first page
            $topImages = [];
            if ($page == 1) {
                $topImages = $this->getTopImages($query, 4);
            }

            // error_log('CHECK: ' . '' . json_encode($topImages[$test]));

            // Return view for SSR
            return view('search-results', [
                'query' => $query,
                'results' => $paginatedResults,
                'total' => count($sortedWordResults),
                'topImages' => $topImages,
            ]);

        } catch (Exception $e) {
            return response()->json(['server error' => $e . getMessage()], 500);
        }
    }

    public function search_images(Request $request)
    {
        $query = $request->input('q');
        if (!$query) {
            return view('search-results', [
                'query' => $query,
                'results' => [],
                'total' => 0,
                'topImages' => [],
            ]);
        }

        // Split the query string
        $query = str_replace('+', ' ', $query);
        $words = explode(' ', strtolower($query));

        // Set the number of results per page
        $perPage = 20;
        $page = $request->input('page', 1); // Default page 1

        try {
            $results = DB::connection('mongodb')->table('word_images')
                ->whereIn('_id', $words)
                ->get();

            $wordImagesHashmap = [];
            $urls = [];

            // Rank pages
            foreach ($words as $word) {
                $pages = $results->where('id', $word)->first()->pages;
                foreach ($pages as ['url' => $url, 'weight' => $weight]) {
                    $urls[$url] = true;
                    if (!isset($wordImagesHashmap[$url])) {
                        $wordImagesHashmap[$url] = [
                            'count' => 0,
                            'weight' => 0,
                            'rank' => 1,
                            'alt' => "",
                            'filename' => "",
                            'page_url' => "",
                            'page_title' => "",
                            'page_text' => "",
                        ];
                    }

                    $wordImagesHashmap[$url]['count'] += 1;
                    $wordImagesHashmap[$url]['weight'] += $weight;
                }
            }

            $uniqueUrls = array_keys($urls);

            // Fetch image data
            $images_data = DB::connection('mongodb')->table('image')
                ->whereIn('_id', $uniqueUrls)
                ->get();

            // Calculate the rank
            $unique_page_urls = [];
            foreach ($uniqueUrls as $url) {
                $count = $wordImagesHashmap[$url]['count'];
                $weight = $wordImagesHashmap[$url]['weight'];
                $wordImagesHashmap[$url]['rank'] = 500 * $count + 500 * $weight;

                // Add image data
                $img_data = $images_data->where('id', $url)->first();
                $wordImagesHashmap[$url]['alt'] = $img_data->alt;
                $wordImagesHashmap[$url]['filename'] = $img_data->filename;
                $wordImagesHashmap[$url]['page_url'] = $img_data->page_url;
                $unique_page_urls[$img_data->page_url] = true;
            }

            $unique_page_urls = array_keys($unique_page_urls);
            // Fetch page data to get the title
            $pages_metadata = DB::connection('mongodb')->table('metadata')
                ->whereIn('_id', $unique_page_urls)
                ->get();

            // Add page metadata
            foreach ($uniqueUrls as $url) {
                $page_url = $wordImagesHashmap[$url]['page_url'];
                $current_page_metadata = $pages_metadata->where('id', $page_url)->first();

                if ($current_page_metadata) {
                    $wordImagesHashmap[$url]['page_title'] = $current_page_metadata->title;
                    $summary_text = $current_page_metadata->summary_text;
                    $wordImagesHashmap[$url]['page_text'] = strlen($summary_text) > 30
                        ? substr($summary_text, 0, 30) . '...'
                        : $summary_text;
                }
            }

            $sortedWordImagesResults = collect($wordImagesHashmap)->sortByDesc('rank');

            $paginatedResults = $sortedWordImagesResults->forPage($page, $perPage);

            // Return view for SSR
            return view('search-image-results', [
                'query' => $query,
                'results' => $paginatedResults,
                'total' => count($sortedWordImagesResults),
            ]);
            /*return response()->json(['data' => $paginatedResults], 200);*/
        } catch (Exception $e) {
            return response()->json(['server error' => $e . getMessage()], 500);
        }
    }
}
