document.addEventListener("DOMContentLoaded", () => {
    const searchButton = document.getElementById("search-button");
    const searchBar = document.getElementById("search-bar");

    if (searchButton) {
        searchButton.addEventListener("click", () => {
            const query = searchBar.value.trim();
            // TODO: Check url
            if (query) {
                window.location.href = `/api/search?q=${encodeURIComponent(query)}`
            } else {
                alert("Please enter a search query");
            }
        });
    }

    if (searchBar) {
        searchBar.addEventListener("keydown", (event) => {
            if (event.key == 'Enter') {
                const query = searchBar.value.trim();
                // TODO: Check url
                if (query) {
                    window.location.href = `/api/search?q=${encodeURIComponent(query)}`
                } else {
                    alert("Please enter a search query");
                }
            }
        });
    }
});
