<?php

namespace App\Http\Middleware;

use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Cache;

class FuzzySearch
{
    public function handle(Request $request, Closure $next): Response
    {
        error_log('FuzzySearch middleware called');
        $query = $request->query('q');
        if (!$query) {
            return $next($request);
        }

        // Split the query string
        $query = str_replace('+', ' ', $query);
        $queryWords = explode(' ', strtolower($query));
        $processedQuery = [];
        $hasSuggestions = false;

        try {
            foreach ($queryWords as $word) {
                if (empty(trim($word))) {
                    continue;
                }
                $suggestion = $this->checkOrSuggestWord($word);
                error_log('Processing word: ' . $word . ' => ' . ($suggestion ?? 'no suggestion'));

                if ($suggestion && $suggestion !== $word) {
                    $processedQuery[] = $suggestion;
                    $hasSuggestions = true;
                } else {
                    $processedQuery[] = $word;
                }
            }

            $processedQueryString = implode(' ', $processedQuery);
            $request->merge(['processedQuery' => $processedQueryString]);
            if ($hasSuggestions) {
                $request->merge(['hasSuggestions' => true]);
            }

            error_log('Processed query: ' . $processedQueryString);

        } catch (\Exception $e) {
            \Log::error('Spell check error: ' . $e->getMessage());
        }

        return $next($request);
    }

    private function checkOrSuggestWord(string $word): ?string
    {
        $cacheKey = 'spellcheck:' . $word;

        // Check cache first
        if (Cache::has($cacheKey)) {
            error_log('Cache hit for word: ' . $word);
            return Cache::get($cacheKey);
        }

        try {
            $collection = DB::connection('mongodb')->table('dictionary');
            // Check DB directly for exact word
            $exists = $collection->find($word);

            if ($exists) {
                Cache::put($cacheKey, $word, 3600); // Cache for 1 hour
                error_log('Word found in DB: ' . $word);
                return $word;
            }

            error_log('Word not found in DB: ' . $word);

            $length = strlen($word);
            $searchLength = $length - 3 > 0 ? $length - 2 : 1;
            $firstTwoChars = substr($word, 0, $searchLength);

            $cursor = DB::connection('mongodb')
                ->table('dictionary')
                ->raw(function ($collection) use ($firstTwoChars, $length) {
                    return $collection->aggregate([
                        [
                            '$match' => [
                                '_id' => ['$regex' => '^' . $firstTwoChars, '$options' => 'i']
                            ]
                        ],
                        [
                            '$addFields' => [
                                'length' => ['$strLenCP' => '$_id']
                            ]
                        ],
                        [
                            '$match' => [
                                'length' => ['$gte' => $length - 1, '$lte' => $length + 1]
                            ]
                        ]
                    ]);
                });

            // Find best match by Levenshtein distance
            $bestMatch = null;
            $minDistance = PHP_INT_MAX;
            $wordLength = strlen($word);

            foreach ($cursor as $document) {
                $candidate = $document->_id;
                $candidateLength = strlen($candidate);

                if (abs($candidateLength - $wordLength) > 2) {
                    continue; // Too different
                }

                $distance = levenshtein($word, $candidate);

                $maxDistance = $wordLength <= 4 ? 1 : min(2, floor($wordLength / 4));
                if ($distance <= $maxDistance && $distance < $minDistance) {
                    $minDistance = $distance;
                    $bestMatch = $candidate;
                }
            }

            $finalSuggestion = $bestMatch ?? $word;
            error_log('Best match found: ' . $finalSuggestion);

            // Save to cache
            Cache::put($cacheKey, $finalSuggestion, 3600);

            return $finalSuggestion;

        } catch (\Exception $e) {
            error_log('Suggestion lookup error: ' . $e->getMessage());
            return $word; // fallback to original word
        }
    }
}
