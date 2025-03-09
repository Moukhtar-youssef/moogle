<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;

class WordController extends Controller
{
    public function getWord($key)
    {
        return response()->json([
            'username' => 'Bro',
            'id' => 42069,
        ]);
    }
}

