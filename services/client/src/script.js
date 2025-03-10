document.addEventListener("DOMContentLoaded", () => {
    const searchButton = document.getElementById("search-button");
    const searchBar = document.getElementById("search-bar");

    if (searchButton) {
        searchButton.addEventListener("click", () => {
            const query = searchBar.value.trim();
            if (query) {
                search(query);
            } else {
                alert('Please enter a search query');
            }
        });
    }

    if (searchBar) {
        searchBar.addEventListener("keydown", (event) => {
            if (event.key == 'Enter') {
                const query = searchBar.value.trim();
                if (query) {
                    search(query);
                } else {
                    alert('Please enter a search query');
                }
            }
        });
    }
});

async function search(query) {
    try {
        const encodedQuery = encodeURIComponent(query);
        console.log(`Query: ${query} | Encoded ${encodedQuery}`);

        // TODO: See how to replace this when it's hosted online
        const backendUrl = `http://127.0.0.1:8000`
        const requestUrl = `${backendUrl}/api/search?q=${encodedQuery}`

        window.location.href = requestUrl;
    } catch (error) {
        console.log(error.message);
    }
}
