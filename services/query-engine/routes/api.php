<?php

use Illuminate\Http\Request;
use Illuminate\Support\Facades\Route;
use Illuminate\Support\Facades\DB;

use App\Http\Controllers\QuerySearchController;

Route::get('/test-mongo', function (Request $request) {
    $connection = DB::connection('mongodb');
    $msg = 'MongoDB is online!';
    try {
        $connection->getMongoClient()->selectDatabase('test')->command(['ping' => 1]);
    } catch (\Exception $e) {
        $msg = 'Mongo is not online. Error: ' . $e->getMessage();
    }

    return ['msg' => $msg];
});

Route::get('/search', [QuerySearchController::class, 'search']);
Route::get('/search_images', [QuerySearchController::class, 'search_images']);
