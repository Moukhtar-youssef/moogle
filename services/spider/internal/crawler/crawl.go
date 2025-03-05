package crawler

import (
    "log"
    "context"
    "time"

    "github.com/redis/go-redis/v9"
    "github.com/IonelPopJara/search-engine/services/spider/internal/utils"
    "github.com/IonelPopJara/search-engine/services/spider/internal/pages"
)

// BFS crawling
func (crawcfg *CrawlerConfig) Crawl(rdb *redis.Client, ctx *context.Context) {
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
        result, err := rdb.BLPop(*ctx, time.Duration(crawcfg.Timeout) * time.Second, crawcfg.QueueKey).Result()
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
        links, err := getURLsFromHTML(html, rawCurrentURL)
        if err != nil {
            log.Printf("Error getting links from HTML: %v\n", err)
            continue
        }

        // Create Page struct
        // NOTE: I'll add the data to redis in batches, so not here
        pg := pages.CreatePage(normalizedCurrentURL, html, contentType, links, statusCode)
        // log.Printf("Page Data:\n%v\n", pg)

        // log.Printf("\n\nAdding %v when there is %d elements\n\n", normalizedCurrentURL, crawcfg.lenPages())
        // Add page visit
        if !crawcfg.addPageVisit(pg) {
            log.Printf("Error adding page visit\n")
            continue
        }

        log.Printf("Adding links from %v (%v)...\n", normalizedCurrentURL, rawCurrentURL)
        // Add links to queue
        for _, rawCurrentLink := range links {
            // log.Printf("Adding %v to the queue\n", rawCurrentLink)
            rdb.LPush(*ctx, crawcfg.QueueKey, rawCurrentLink)
        }
    }
}

