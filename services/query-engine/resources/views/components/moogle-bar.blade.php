<div class="moogle-bar-container">
    <div class="top-part">
        @php
            $currentPage = request()->query('page', 1);
            $currentQuery = request()->query('q');
        @endphp
        {{-- TODO: Change this for the actual url --}}
        <a href="http://localhost:5173/" class="logo-container">
            Moogle!
        </a>
        <form action="/api/search" method="GET" class="px-3">
            <input type="text" id="search-bar" name="q" placeholder="Search..." />
            <button type="submit" class="btn" id="search-button">
                Moogle it!
            </button>
            <button type="button" class="btn">
                Life ain't cringe!
            </button>
        </form>
    </div>
    <div class="bottom-part">
        @php
            $currentPath = request()->path();
            $currentSearchFunction = explode('/', $currentPath)[1];
            $images_url = '/api/search_images?q=' . $currentQuery;
            $pages_url = '/api/search?q=' . $currentQuery;

            $imagesActive = '';
            $pagesActive = '';
            if ($currentSearchFunction == 'search') {
                $imagesActive = '';
                $pagesActive = 'active';
            } else {
                $imagesActive = 'active';
                $pagesActive = '';
            }
        @endphp
        <a href="{{ $pages_url }}" class="tab {{ $pagesActive }}">
            <span> PAGES </span>
        </a>
        <a href="{{ $images_url }}" class="tab {{ $imagesActive }}">
            <span> IMAGES </span>
        </a>
    </div>
</div>
