<?php

use Illuminate\Http\Request;
use Illuminate\Support\Facades\Route;
use App\Http\Controllers\WordController;

/*Route::get('/user', function (Request $request) {*/
/*    return $request->user();*/
/*})->middleware('auth:sanctum');*/

Route::get('/get/{key}', [WordController::class, 'getWord']);

