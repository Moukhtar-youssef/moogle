<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;

ini_set('memory_limit', '256M'); // Increase to 256MB or higher

class QuerySearchController extends Controller
{
    protected function getTopImages($query, $limit = 4)
    {
        error_log('Memory usage: ' . memory_get_usage());
        error_log('Peak memory usage: ' . memory_get_peak_usage());

        // Split the query string
        $query = str_replace('+', ' ', $query);
        $words = explode(' ', strtolower($query));

        try {
            $results = DB::connection('mongodb')->table('word_images')
                ->raw(function ($collection) use ($words) {
                    $cursor = $collection->aggregate([
                        ['$match' => ['_id' => ['$in' => $words]]],
                        ['$unwind' => '$pages'],
                        [
                            '$group' => [
                                '_id' => '$pages.url',
                                'total_weight' => ['$sum' => '$pages.weight'],
                                'count' => ['$sum' => 1],
                            ]
                        ],
                        ['$sort' => ['total_weight' => -1]],
                    ]);

                    return iterator_to_array($cursor);
                });

            // Debug output
            error_log("Results count: " . count($results));
            // error_log(json_encode($results));

            $wordImagesHashmap = [];
            $urls = [];

            // Rank pages - Fix object/array access issue
            foreach ($results as $result) {
                // Check if result is object or array and access accordingly
                $url = is_object($result) ? $result->_id : $result['_id'];
                $count = is_object($result) ? $result->count : $result['count'];
                $total_weight = is_object($result) ? $result->total_weight : $result['total_weight'];

                $urls[$url] = true;
                if (!isset($wordImagesHashmap[$url])) {
                    $wordImagesHashmap[$url] = [
                        'count' => $count,
                        'weight' => $total_weight,
                        'rank' => $total_weight * $count,
                        'alt' => "",
                        'filename' => "",
                        'page_url' => "",
                        'page_title' => "",
                        'page_text' => "",
                    ];
                }
            }

            // Continue only if we have results
            if (empty($wordImagesHashmap)) {
                error_log("No image mappings found!");
                return collect();
            }

            $uniqueUrls = array_keys($urls);
            error_log("Unique URLs count: " . count($uniqueUrls));

            // Fetch image data
            $images_data = DB::connection('mongodb')->table('image')
                ->whereIn('_id', $uniqueUrls)
                ->get(); // Using get() instead of cursor() for simplicity

            $unique_page_urls = [];
            foreach ($images_data as $img_data) {
                $url = $img_data->_id ?? (is_array($img_data) ? $img_data['_id'] : null);
                if (!$url || !isset($wordImagesHashmap[$url])) {
                    continue;
                }

                $wordImagesHashmap[$url]['alt'] = $img_data->alt ?? $img_data['alt'] ?? '';
                $wordImagesHashmap[$url]['filename'] = $img_data->filename ?? $img_data['filename'] ?? '';
                $wordImagesHashmap[$url]['page_url'] = $img_data->page_url ?? $img_data['page_url'] ?? '';

                if (!empty($wordImagesHashmap[$url]['page_url'])) {
                    $unique_page_urls[$wordImagesHashmap[$url]['page_url']] = true;
                }
            }

            $unique_page_urls = array_keys($unique_page_urls);
            error_log("Unique page URLs count: " . count($unique_page_urls));

            if (!empty($unique_page_urls)) {
                // Fetch page metadata
                $pages_metadata = DB::connection('mongodb')->table('metadata')
                    ->whereIn('_id', $unique_page_urls)
                    ->get(); // Using get() instead of cursor() for simplicity

                $pageMetadataMap = [];
                foreach ($pages_metadata as $current_page_metadata) {
                    $id = $current_page_metadata->_id ?? (is_array($current_page_metadata) ? $current_page_metadata['_id'] : null);
                    if ($id) {
                        $pageMetadataMap[$id] = $current_page_metadata;
                    }
                }

                foreach ($wordImagesHashmap as $url => &$data) {
                    if (!empty($data['page_url']) && isset($pageMetadataMap[$data['page_url']])) {
                        $current_page_metadata = $pageMetadataMap[$data['page_url']];

                        // Handle both object and array access
                        if (is_object($current_page_metadata)) {
                            $data['page_title'] = $current_page_metadata->title ?? '';
                            $summary_text = $current_page_metadata->summary_text ?? '';
                        } else {
                            $data['page_title'] = $current_page_metadata['title'] ?? '';
                            $summary_text = $current_page_metadata['summary_text'] ?? '';
                        }

                        $data['page_text'] = strlen($summary_text) > 30
                            ? substr($summary_text, 0, 30) . '...'
                            : $summary_text;
                    }
                }
            }

            // Sort by rank (which now uses weight * count)
            $sortedWordImagesResults = collect($wordImagesHashmap)->sortByDesc('rank');

            // Add more debugging
            error_log("Final results count: " . $sortedWordImagesResults->count());

            // Return the top N images
            return $sortedWordImagesResults->take($limit);
        } catch (\Exception $e) {
            error_log("Exception in getTopImages: " . $e->getMessage());
            error_log("Trace: " . $e->getTraceAsString());
            return collect(); // Return an empty collection in case of error
        }
    }

    public function count_pages(Request $request)
    {
        $results = DB::connection('mongodb')->table('metadata')->count();
        return response()->json([
            'status' => 'up',
            'pages' => $results,
        ]);
    }

    public function search(Request $request)
    {
        $suggestions = $request->input('suggestions');
        $originalQuery = $request->input('q');
        $query = $request->input('processed_query');
        if (!$query) {
            return view('search-results', [
                'query' => $query,
                'results' => [],
                'total' => 0,
                'topImages' => [],
                'suggestions' => $suggestions,
                'originalQuery' => $originalQuery,
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
                    'suggestions' => $suggestions,
                    'originalQuery' => $originalQuery,
                ]);
            }

            // Rank pages
            foreach ($words as $word) {

                if ($results->where('id', $word)->count() <= 0) {
                    continue;
                }
                $pages = $results->where('id', $word)->first()->pages;

                foreach ($pages as ['url' => $url, 'weight' => $weight]) {
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

            // Get the pageranks
            $uniqueUrls = array_keys($urls);

            $pageranks = DB::connection('mongodb')->table('pagerank')
                ->whereIn('_id', $uniqueUrls)
                ->get();

            foreach ($pageranks as $pagerank) {
                $wordHashmap[$pagerank->id]['pagerank'] = $pagerank->rank;
                error_log('Pagerank added');
            }

            // Fetch page metadata
            $pages_metadata = DB::connection('mongodb')->table('metadata')
                ->whereIn('_id', $uniqueUrls)
                ->get();

            /*return $pages_metadata;*/
            foreach ($uniqueUrls as $url) {
                // $count = $wordHashmap[$url]['count'];
                $weight = $wordHashmap[$url]['weight'];
                // $backlinks = $wordHashmap[$url]['backlinks'];
                $pagerank = $wordHashmap[$url]['pagerank'] ?? 0.001;
                $wordHashmap[$url]['rank'] = 1000 * $weight + 10000 * $pagerank;

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

            // Return view for SSR
            return view('search-results', [
                'query' => $query,
                'results' => $paginatedResults,
                'total' => count($sortedWordResults),
                'topImages' => $topImages,
                'suggestions' => $suggestions,
                'originalQuery' => $originalQuery,
            ]);

        } catch (\Exception $e) {
            return redirect()->back()->with('error', $e->getMessage());
        }
    }

    public function search_images(Request $request)
    {
        error_log('Searching images');
        $suggestions = $request->input('suggestions');
        $originalQuery = $request->input('q');
        $query = $request->input('processed_query');

        if (!$query) {
            return view('search-results', [
                'query' => $query,
                'results' => [],
                'total' => 0,
                'topImages' => [],
                'suggestions' => $suggestions,
                'originalQuery' => $originalQuery,
            ]);
        }

        // Split the query string
        $query = str_replace('+', ' ', $query);
        $words = explode(' ', strtolower($query));

        // Set the number of results per page
        $perPage = 20;
        $page = $request->input('page', 1); // Default page 1

        try {
            // Add some debugging
            error_log("Searching for words: " . json_encode($words));

            // Option 1: Using the MongoDB aggregation pipeline like in getTopImages
            $results = DB::connection('mongodb')->table('word_images')
                ->raw(function ($collection) use ($words) {
                    $cursor = $collection->aggregate([
                        ['$match' => ['_id' => ['$in' => $words]]],
                        ['$unwind' => '$pages'],
                        [
                            '$group' => [
                                '_id' => '$pages.url',
                                'total_weight' => ['$sum' => '$pages.weight'],
                                'count' => ['$sum' => 1],
                            ]
                        ],
                        ['$sort' => ['total_weight' => -1]],
                    ]);

                    // Convert cursor to array
                    return iterator_to_array($cursor);
                });

            $wordImagesHashmap = [];
            $urls = [];

            // Process results - handle both object and array access
            foreach ($results as $result) {
                // Check if result is object or array and access accordingly
                $url = is_object($result) ? $result->_id : $result['_id'];
                $count = is_object($result) ? $result->count : $result['count'];
                $total_weight = is_object($result) ? $result->total_weight : $result['total_weight'];

                $urls[$url] = true;
                if (!isset($wordImagesHashmap[$url])) {
                    $wordImagesHashmap[$url] = [
                        'count' => $count,
                        'weight' => $total_weight,
                        'rank' => $total_weight, // Use weight for ranking
                        'alt' => "",
                        'filename' => "",
                        'page_url' => "",
                        'page_title' => "",
                        'page_text' => "",
                    ];
                }
            }

            // Continue only if we have results
            if (empty($wordImagesHashmap)) {
                error_log("No image mappings found");
                return view('search-image-results', [
                    'query' => $query,
                    'results' => [],
                    'total' => 0,
                    'topImages' => [],
                    'suggestions' => $suggestions,
                    'originalQuery' => $originalQuery,
                ]);
            }

            $uniqueUrls = array_keys($urls);
            error_log("Processing " . count($uniqueUrls) . " unique URLs");

            // Fetch image data
            $images_data = DB::connection('mongodb')->table('image')
                ->whereIn('_id', $uniqueUrls)
                ->select(['_id', 'alt', 'filename', 'page_url'])
                ->get();

            $unique_page_urls = [];
            foreach ($images_data as $img_data) {
                $url = $img_data->id;

                if (!isset($wordImagesHashmap[$url])) {
                    continue;
                }

                // Handle both object and array access
                if (is_object($img_data)) {
                    $wordImagesHashmap[$url]['alt'] = $img_data->alt ?? '';
                    $wordImagesHashmap[$url]['filename'] = $img_data->filename ?? '';
                    $wordImagesHashmap[$url]['page_url'] = $img_data->page_url ?? '';
                    if (!empty($img_data->page_url)) {
                        $unique_page_urls[$img_data->page_url] = true;
                    }
                } else {
                    $wordImagesHashmap[$url]['alt'] = $img_data['alt'] ?? '';
                    $wordImagesHashmap[$url]['filename'] = $img_data['filename'] ?? '';
                    $wordImagesHashmap[$url]['page_url'] = $img_data['page_url'] ?? '';
                    if (!empty($img_data['page_url'])) {
                        $unique_page_urls[$img_data['page_url']] = true;
                    }
                }
            }

            $unique_page_urls = array_keys($unique_page_urls);
            error_log("Found " . count($unique_page_urls) . " unique page URLs");

            if (!empty($unique_page_urls)) {
                // Fetch page metadata
                $pages_metadata = DB::connection('mongodb')->table('metadata')
                    ->whereIn('_id', $unique_page_urls)
                    ->select(['_id', 'title', 'summary_text'])
                    ->get(); // Using get() instead of cursor()

                $pageMetadataMap = [];
                foreach ($pages_metadata as $current_page_metadata) {
                    $id = is_object($current_page_metadata) ?
                        ($current_page_metadata->_id ?? null) :
                        ($current_page_metadata['_id'] ?? null);

                    if ($id) {
                        $pageMetadataMap[$id] = $current_page_metadata;
                    }
                }

                foreach ($wordImagesHashmap as $url => &$data) {
                    if (!empty($data['page_url']) && isset($pageMetadataMap[$data['page_url']])) {
                        $current_page_metadata = $pageMetadataMap[$data['page_url']];

                        // Handle both object and array access
                        if (is_object($current_page_metadata)) {
                            $data['page_title'] = $current_page_metadata->title ?? '';
                            $summary_text = $current_page_metadata->summary_text ?? '';
                        } else {
                            $data['page_title'] = $current_page_metadata['title'] ?? '';
                            $summary_text = $current_page_metadata['summary_text'] ?? '';
                        }

                        $data['page_text'] = strlen($summary_text) > 30
                            ? substr($summary_text, 0, 30) . '...'
                            : $summary_text;
                    }
                }
            }

            // Rank and paginate results
            $sortedWordImagesResults = collect($wordImagesHashmap)->sortByDesc('weight');
            $paginatedResults = $sortedWordImagesResults->forPage($page, $perPage);

            // error_log("Returning " . count($paginatedResults) . " results (page $page of " .
            // ceil(count($sortedWordImagesResults) / $perPage) . ")");

            // Return view for SSR
            return view('search-image-results', [
                'query' => $query,
                'results' => $paginatedResults,
                'total' => count($sortedWordImagesResults),
                'topImages' => [],
                'suggestions' => $suggestions,
                'originalQuery' => $originalQuery,
            ]);

        } catch (\Exception $e) {
            error_log("Exception in search_images: " . $e->getMessage());
            error_log("Trace: " . $e->getTraceAsString());
            return redirect()->back()->with('error', $e->getMessage());
        }
    }

    public function get_top_ranked_page(Request $request)
    {
        // Get the top ranked page from pagerank
        $results = DB::connection('mongodb')->table('pagerank')
            ->orderBy('rank', 'desc')
            ->limit(1)
            ->get();
        if ($results->count() <= 0) {
            return null;
        }

        // Fetch the page metadata
        $page_metadata = DB::connection('mongodb')->table('metadata')
            ->where('_id', $results[0]->id)
            ->first();
        if (!$page_metadata) {
            return null;
        }

        // Return the page metadata as an array
        return [
            'title' => $page_metadata->title,
            'url' => $page_metadata->id,
            'description' => $page_metadata->description,
            'last_crawled' => $page_metadata->last_crawled,
            'summary_text' => $page_metadata->summary_text,
        ];
    }

    // Make a function to get a random page from the metadata collection
    public function get_random_page(Request $request)
    {
        $results = DB::connection('mongodb')
            ->table('metadata')
            ->raw(function ($collection) {
                return $collection->aggregate([
                    ['$sample' => ['size' => 1]]
                ]);
            });

        $document = $results->toArray();

        if (empty($document)) {
            return null;
        }

        $doc = $document[0];

        // Return the page metadata as an array
        return [
            'title' => $doc['title'],
            'url' => $doc['_id'],
            'description' => $doc['description'],
            'last_crawled' => $doc['last_crawled'],
            'summary_text' => $doc['summary_text'],
        ];
    }


}
