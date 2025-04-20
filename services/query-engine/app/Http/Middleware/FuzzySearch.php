<?php

namespace App\Http\Middleware;

use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Cache;

class FuzzySearch
{
    /**
     * Handle an incoming request.
     *
     * @param  \Closure(\Illuminate\Http\Request): (\Symfony\Component\HttpFoundation\Response)  $next
     */
    public function handle(Request $request, Closure $next): Response
    {
        $query = $request->query('q');

        if (!$query) {
            return $next($request);
        }

        // Split the query string
        $query = str_replace('+', ' ', $query);
        $query = explode(' ', strtolower($query));

        $processedQuery = [];
        $dictionary = $this->getDictionaryWords();

        // Get the dictionary from MongoDB
        try {
            // Iterate through all the words in the query
            foreach ($query as $word) {
                // Fetch the word from the dictionary in MongoDB
                $result = DB::connection('mongodb')->table('dictionary')->where('_id', $word)->first();

                // If the word is not found, perform a fuzzy search
                if (!$result) {
                    $suggestion = $this->fuzzySearch($word, $dictionary);
                    $processedQuery[] = empty($suggestion) ? $word : $suggestion[0];
                } else {
                    // Append the word to the processed query
                    $processedQuery[] = $word;
                }
            }

            $processedQuery = array_unique($processedQuery);
            $processedQuery = implode(' ', $processedQuery);

            $request->merge([
                'processed_query' => $processedQuery,
                'suggestions' => $processedQuery !== implode(' ', $query) ? true : false
            ]);

        } catch (\Exception $e) {
            $request->merge([
                'processed_query' => implode(' ', $query),
                'suggestions' => false
            ]);
        }

        return $next($request);
    }

    private function getDictionaryWords()
    {
        if (Cache::has('dictionary')) {
            return Cache::get('dictionary');
        }

        $words = $this->fetchdictionarywords();

        // Cache the dictionary for 24 hours
        Cache::put('dictionary', $words, 60 * 24);
        return $words;
    }

    private function fetchDictionaryWords()
    {
        $words = DB::connection('mongodb')->table('dictionary')->get();
        return $words;
    }

    private function fuzzySearch($word, $dictionary)
    {
        $suggestions = [];
        $wordLength = strlen($word);

        foreach ($dictionary as $dictWord) {
            $dictWord = $dictWord->id;
            $dictWordLength = strlen($dictWord);

            if ($dictWordLength < $wordLength - 1 || $dictWordLength > $wordLength + 1) {
                continue;
            }

            $levenshteinDistance = levenshtein($word, $dictWord);

            if ($levenshteinDistance <= 2) {
                $suggestions[] = $dictWord;
            }
        }

        return $suggestions;
    }
}
