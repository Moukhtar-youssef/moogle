package crawler

import (
    "log"
    "math"

    "github.com/IonelPopJara/search-engine/services/spider/internal/utils"
    "github.com/IonelPopJara/search-engine/services/spider/internal/pages"
    "github.com/IonelPopJara/search-engine/services/spider/internal/database"
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
        rawCurrentURL, depthLevel, err := db.PopURL()
        if err != nil {
            log.Printf("No more URLs in the queue: %v\n", err)
            return
        }

        log.Printf("\tURL '%v' has a depth level of %v\n", rawCurrentURL, depthLevel)

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

        log.Printf("\tCrawling from %v (%v)...\n", normalizedCurrentURL, rawCurrentURL)

        // Fetch HTML, Status Code, and Content-Type
        html, statusCode, contentType, err := getPageData(rawCurrentURL)
        if err != nil {
            // Skip if we couldn't fetch the data
            log.Printf("Error fetching %v data: %v\n", rawCurrentURL, err)
            continue
        }

        // Fetch the links of the current page
        // Change the cfg.BaseURL I don't get it but I need to change it
        outgoingLinks, imagesMap, err := getURLsFromHTML(html, rawCurrentURL)
        if err != nil {
            log.Printf("Error getting links from HTML: %v\n", err)
            continue
        }

        // Store images
        crawcfg.AddImages(normalizedCurrentURL, imagesMap)

        // Create outlinks and update backlinks
        crawcfg.UpdateLinks(normalizedCurrentURL, outgoingLinks)

        // Create Page struct
        pg := pages.CreatePage(normalizedCurrentURL, html, contentType, statusCode)

        // Add page visit
        if !crawcfg.addPageVisit(pg) {
            log.Printf("Error adding page visit\n")
            continue
        }

        log.Printf("\tAdding links from %v (%v)...\n", normalizedCurrentURL, rawCurrentURL)
        // Add links to url queue
        for _, rawCurrentLink := range outgoingLinks {
            // Check if the url is valid
            if !utils.IsValidURL(rawCurrentLink) {
                // If it's not valid, process the next link
                continue
            }

            // Check if the thing exists in the queue, and update weight
            score, exists := db.ExistsInQueue(rawCurrentLink)
            if exists {
                // // Already in queue so we update it's priority
                // fmt.Printf("\n--------------------( %v )---------------------------------------\n", score)
                // fmt.Printf("%v already in queue...\n", rawCurrentLink)
                // fmt.Printf("--------------------( %v )---------------------------------------\n\n", score - 1)
                score -= 0.001
                // time.Sleep(1 * time.Second)
            } else {
                score = depthLevel + 1
            }

            score = math.Max(utils.MinScore, math.Min(score, utils.MaxScore))

            // Update score based on depth
            err := db.PushURL(rawCurrentLink, score)
            if err != nil {
                log.Printf("Error pushing '%v' to the queue: %v\n", rawCurrentLink, err)
                log.Printf("\tSkipping...")
                continue
            }
        }
    }
}

