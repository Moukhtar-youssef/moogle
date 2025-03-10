<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    @vite('resources/css/app.css')
    <title>Search Results for "{{ $query }}"</title>
</head>

<body class="bg-slate-100">
    <div class="main-container">
        <div class="search-container">
            {{-- TODO: Change this for the actual url --}}
            <a href="http://localhost:5174/">
                <div class="logo-container">
                    <h1 class="drop-shadow-lg">
                        Moogle!
                    </h1>
                </div>
            </a>
            <input type="text" id="search-bar" name="search" />
            <div class="button-container">
                <button class="btn" id="search-button">
                    Moogle it!
                </button>
                <button class="btn">
                    Life ain't cringe!
                </button>
            </div>
        </div>
        <div class="results-container">
            @if (count($results) > 0)
                <p>{{ count($results) }} results found for "{{ $query }}"</p>
                <ul>
                    @foreach ($results as $url => $data)
                        <li class="result-container">
                            <a href="https://{{ $url }}">
                                <h3 class="result-title">{{ $data['title'] }}</h3>
                            </a>
                            <p class="result-text">{{ Str::limit($data['text'], 200) }}</p>
                            <a href="https://{{ $url }}">
                                <p class="result-url">{{ $url }}</p>
                            </a>
                        </li>
                    @endforeach
                </ul>
            @else
                <p>No results found</p>
            @endif
        </div>
    </div>
</body>
@vite('resources/js/search.js')
</html>
