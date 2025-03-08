<?php

use Illuminate\Support\Facades\Route;
use App\Http\Controllers\RedisController;

Route::get('/', function () {
    error_log('Hello');
    return view('welcome');
});

Route::get('/get/{key}', [RedisController::class, 'get']);
Route::get('/set/{key}/{value}', [RedisController::class, 'get']);
