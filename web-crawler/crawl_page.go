package main

import (
    "fmt"
    "net/url"
)

func (cfg* config) crawlPage(rawCurrentURL string) {
    // Put a new empty struct in the channel
    // If there is not enough space, it waits
    cfg.concurrencyControl <- struct{}{}

    // When the function returns,
    // remove the element from the channel
    // and stop the wait group
    defer func() {
        <-cfg.concurrencyControl
        cfg.wg.Done()
    }()

    // Return immediately if we exceed the max pages
    if cfg.pagesLen() >= cfg.maxPages {
        return
    }

    // Parse current URL
    currentURL, err := url.Parse(rawCurrentURL)
    if err != nil {
        fmt.Printf("Error parsing current URL: %v\n", err)
        return
    }

    // Skip other websites
    if currentURL.Hostname() != cfg.baseURL.Hostname() {
        return
    }

    // Normalize current URL
    normalizedCurrentURL, err := normalizeURL(rawCurrentURL)
    if err != nil {
        fmt.Printf("Error normalizing current URL: %v\n", err)
        return
    }

    isPageNew := cfg.addPageVisit(normalizedCurrentURL)
    if !isPageNew {
        return
    }

    fmt.Printf("DFSing %v!\n", normalizedCurrentURL)

    // Get the HTML of the current page
    html, err := getHTML(rawCurrentURL)
    if err != nil {
        fmt.Printf("Error fetching HTML: %v\n", err)
        return
    }

    // Fetch the links of the current page
    links, err := getURLsFromHTML(html, cfg.baseURL)
    if err != nil {
        fmt.Printf("Error fetching links from HTML: %v\n", err)
        return
    }

    fmt.Printf("Links found:\n")
    for _, rawCurrentLink := range links {
        // Add a new  fg
        cfg.wg.Add(1)
        go cfg.crawlPage(rawCurrentLink)
    }
}

