<a href="https://{{ $url }}">
    <li class="result-container">
        <h3 class="result-title">{{ $title }}</h3>
        <p class="result-text">{{ Str::limit($text, 200) }}</p>
        <p class="result-url">{{ $url }}</p>
    </li>
</a>
