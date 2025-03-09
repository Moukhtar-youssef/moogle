<?php

use Illuminate\Http\Request;
use Illuminate\Support\Facades\Route;
use App\Http\Controllers\WordController;

Route::get('/words/{key}', [WordController::class, 'getWord']);
Route::get('/search', [WordController::class, 'search']);

