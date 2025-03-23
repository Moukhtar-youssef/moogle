<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    @vite('resources/css/app.css')
    <title>Search Image Results for "{{ $query }}"</title>
</head>

<body>
    <x-moogle-bar />
    <div class="results-counter">
        Found {{ $total }} results found for "{{ $query }}"...
    </div>
    <div class="results-images-container">
        @if (count($results) > 0)
            <ul class="flex flex-wrap gap-4">
                @foreach ($results as $url => $data)
                    <li class="m-2 flex-shrink-0 w-1/6">
                        <x-image-container url="{{ $url }}" alt="{{ $data['alt'] }}"
                            title="{{ $data['page_title'] }}" page="{{ $data['page_url'] }}"
                            text="{{ $data['page_text'] }}" />
                    </li>
                @endforeach
            </ul>
        @else
            <p>No results found</p>
        @endif
    </div>
    <div class="flex flex-col justify-center items-center">
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
@vite('resources/js/search.js')

</html>
