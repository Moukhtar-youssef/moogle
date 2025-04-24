<?php

use App\Http\Middleware\FuzzySearch;
use App\Http\Middleware\StoreSearchTerm;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Route;
use Illuminate\Support\Facades\DB;

use App\Http\Controllers\QuerySearchController;
use App\Http\Controllers\RedisController;

// Route::get('/test-mongo', function (Request $request) {
//     $connection = DB::connection('mongodb');
//     $msg = 'MongoDB is online!';
//     try {
//         $connection->getMongoClient()->selectDatabase('test')->command(['ping' => 1]);
//     } catch (\Exception $e) {
//         $msg = 'Mongo is not online. Error: ' . $e->getMessage();
//     }

//     return ['msg' => $msg];
// });

Route::get('/search', [QuerySearchController::class, 'search'])->middleware([FuzzySearch::class, StoreSearchTerm::class]);
Route::get('/search_force', [QuerySearchController::class, 'search'])->name('search_force');
Route::get('/search_images', [QuerySearchController::class, 'search_images'])->middleware(FuzzySearch::class);
Route::get('/search_images_force', [QuerySearchController::class, 'search_images'])->name('search_images_force');
Route::get('/count_pages', [QuerySearchController::class, 'count_pages']);
Route::get('/get_top_searches', [RedisController::class, 'get_top_searches'])->name('get.top.searches');
Route::get('/get_search_suggestions', [RedisController::class, 'get_search_suggestions'])->name('get.search.suggestions');
Route::get('/cringe', [RedisController::class, 'cringe'])->name('cringe');
Route::get('/top_ranked_pages', [QuerySearchController::class, 'get_top_ranked_page'])->name('top_ranked_page');

// Return a secret message when the url is /secret
Route::get('/secret', function () {
    return response()->json(['message' => 'Congratulations! You have found the secret message! It does nothing :)']);
});
