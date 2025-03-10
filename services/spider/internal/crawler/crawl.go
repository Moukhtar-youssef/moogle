package crawler

import (
    "log"
    "time"

    "github.com/IonelPopJara/search-engine/services/spider/internal/utils"
    "github.com/IonelPopJara/search-engine/services/spider/internal/pages"
    "github.com/IonelPopJara/search-engine/services/spider/internal/database"
    "github.com/IonelPopJara/search-engine/services/spider/internal/links"
)

// BFS crawling
func (crawcfg *CrawlerConfig) Crawl(db *database.Database) {
    // Starting a new webcrawler instance
    defer crawcfg.Wg.Done()

    // BFS loop
    for {
        // Check if we have reached the maximum number of pages
        if crawcfg.maxPagesReached() {
            log.Printf("Maximum number of pages reached\n")
            return
        }

        // Get the next URL from the queue
        log.Printf("Waiting for message queue...\n")
        result, err := db.Client.BRPop(db.Context, time.Duration(crawcfg.Timeout) * time.Second, crawcfg.QueueKey).Result()
        if err != nil {
            log.Printf("No more URLs in the queue: %v\n", err)
            return
        }

        // Get current URL
        rawCurrentURL := result[1]

        // Normalize current URL
        normalizedCurrentURL, err := utils.NormalizeURL(rawCurrentURL)
        if err != nil {
            // Skip URL if we can't normalize it
            log.Printf("Error normalizing current URL: %v\n", err)
            continue
        }

        // Check if the URL has been visited
        if !crawcfg.canVisitPage(normalizedCurrentURL) {
            log.Printf("Skipping %v. Already visited\n", normalizedCurrentURL)
            continue
        }

        log.Printf("Crawling from %v (%v)...\n", normalizedCurrentURL, rawCurrentURL)

        // Fetch HTML, Status Code, and Content-Type
        html, statusCode, contentType, err := getPageData(rawCurrentURL)
        if err != nil {
            // Skip if we couldn't fetch the data
            log.Printf("Error fetching %v data: %v\n", rawCurrentURL, err)
            continue
        }

        // Fetch the links of the current page
        // Change the cfg.BaseURL I don't get it but I need to change it
        outgoingLinks, err := getURLsFromHTML(html, rawCurrentURL)
        if err != nil {
            log.Printf("Error getting links from HTML: %v\n", err)
            continue
        }

        // Create Page struct
        pg := pages.CreatePage(normalizedCurrentURL, html, contentType, statusCode)
        ol := links.CreateOutlinks(normalizedCurrentURL, outgoingLinks)

        // Add page visit
        if !crawcfg.addPageVisit(pg, ol) {
            log.Printf("Error adding page visit\n")
            continue
        }

        log.Printf("Adding links from %v (%v)...\n", normalizedCurrentURL, rawCurrentURL)
        // Add links to url queue
        for _, rawCurrentLink := range outgoingLinks {
            db.Client.LPush(db.Context, crawcfg.QueueKey, rawCurrentLink)
        }
    }
}

