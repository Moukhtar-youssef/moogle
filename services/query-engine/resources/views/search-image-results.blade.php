<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    @vite('resources/css/app.css')
    <link rel="alternate icon" href="{{ asset('favicon.ico') }}" type="image/x-icon">
    <link rel="icon" type="image/svg+xml" href="{{ asset('favicon.svg') }}">
    <link rel="apple-touch-icon" type="image/svg+xml" href="{{ asset('favicon.svg') }}">
    <link rel="mask-icon" type="image/svg+xml" href="{{asset('favicon.svg')}}" color="#000000">
    <title>Search Image Results for "{{ $query }}"</title>
</head>

<body class="min-h-screen flex flex-col">
    <x-moogle-bar />

    <div class="results-counter">
        @php
            $suggestion = '';
            if ($suggestions) {
                $suggestion = "Did you mean $query? ";
            }
        @endphp
        <span id="suggestion">{{ $suggestion }}</span><span>Showing {{ $total }} results for
            {{ $query }}</span>
        @if ($suggestions)
            <p class="mt-2">Search for <a id="suggestion-link"
                    href="{{ route('search_images_force', ['processed_query' => $originalQuery]) }}">{{ $originalQuery }}</a>
                instead
            </p>
        @endif
    </div>
    <div class="results-images-container">
        @if (count($results) > 0)
            <ul class="flex flex-wrap justify-center">
                @foreach ($results as $res)
                    <li >
                        <x-image-container url="{{ $res->_id }}" alt="{{ $res->alt }}"
                            title="{{ $res->page_title }}" page="{{ $res->page_url }}" text="{{ $res->page_text }}" />
                    </li>
                @endforeach
            </ul>
        @else
            <p class="text-center">No results found</p>
        @endif
    </div>
    <div class="flex flex-col justify-center items-center mt-8">
        <x-pagination-bar totalResults="{{ $total }}" />
    </div>
    <!-- Footer -->
    <footer>
        <p> <a href="https://github.com/IonelPopJara/search-engine">Support the project!</a></p>
        <p>
            <a href="https://x.com/ionelalexandr12">Twitter</a> -
            <a href="https://www.youtube.com/multselmesco">YouTube</a> -
            <a href="https://github.com/IonelPopJara">GitHub</a>
        </p>
        <p id="copyright">©2025</p>
    </footer>
</body>

</html>
