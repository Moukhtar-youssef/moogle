<div class="pagination-bar-container">
    <div class="flex gap-1 m-5">
        @php
            $osToRender = max(min($totalPages, 10), 2); // Clamp max pages
            $currentPage = request()->query('page', 1);
            $currentQuery = request()->query('q');
            $currentPath = request()->path();
        @endphp
        <span class="moogle-letter">M</span>
        @for ($i = 1; $i <= $osToRender; $i++)
            @php
                $activeClass = $i == $currentPage ? 'active' : 'inactive';
                $url = url($currentPath) . '?q=' . $currentQuery . '&page=' . $i;
            @endphp
            <a href="{{ $url }}" class="moogle-page {{ $activeClass }}">
                <span class="moogle-letter">o</span>
                <p>{{ $i }} </p>
            </a>
        @endfor
        <span class="moogle-letter">g</span>
        <span class="moogle-letter">l</span>
        <span class="moogle-letter">e</span>
    </div>
</div>
