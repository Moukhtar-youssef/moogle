<?php

namespace App\Http\Controllers;

use Illuminate\Support\Facades\Redis;
use Illuminate\Http\Request;

class RedisController extends Controller
{
    // Retrieve a value from redis
    public function get($key)
    {
        error_log($key);
        $value = Redis::zrange($key, 0, -1);
        if ($value) {
            return response()->json(['value' => $value], 200);
        }

        return response()->json(['error' => 'Key not found'], 200);
    }

    // Set a value in Redis
    public function set($key, $value)
    {
        Redis::set($key, $value);
        return response()->json(['message' => 'Key set successfully'], 200);
    }
}
