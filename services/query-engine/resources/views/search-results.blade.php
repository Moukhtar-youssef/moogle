<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    @vite('resources/css/app.css')
    <title>Search Results for "{{ $query }}"</title>
</head>

<body>
    <x-moogle-bar />
    <div class="results-counter">
        Found {{ $total }} results found for "{{ $query }}"...
    </div>

    <div class="results-container">
        <div class="pages-container col-span-2">
            @if (count($results) > 0)
                <ul>
                    @foreach ($results as $url => $data)
                        <x-search-result url="{{ $url }}" title="{{ $data['title'] }}"
                            text="{{ $data['summary_text'] }}" />
                    @endforeach
                </ul>
            @else
                <p>No results found</p>
            @endif
        </div>
        <div class="images-container flex flex-col items-center">
            <h2 class="top-images-text"> {{ ucwords($query) }} </h2>
            @foreach ($topImages as $url => $image)
                <div class="my-2">
                    <x-image-container url="{{ $url }}" alt="{{ $image['alt'] }}"
                        title="{{ $image['page_title'] }}" page="{{ $image['page_url'] }}"
                        text="{{ $image['page_text'] }}" />
                </div>
            @endforeach
        </div>
        <div class="flex flex-col justify-center items-center">
            <x-pagination-bar totalResults="{{ $total }}" />
        </div>
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
