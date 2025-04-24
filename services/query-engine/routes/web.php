<?php

use Illuminate\Support\Facades\Route;

Route::get('/', function () {
    // Redirect to https://moogle.app
    return redirect('https://moogle.app');
});
