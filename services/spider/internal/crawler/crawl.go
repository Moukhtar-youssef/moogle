package crawler

import (
    "log"
    "math"
    // "time"

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

        // time.Sleep(1 * time.Second)
        // Get the next URL from the queue
        log.Printf("Waiting for message queue...\n")
        rawCurrentURL, depthLevel, err := db.PopURL()
        if err != nil {
            log.Printf("No more URLs in the queue: %v\n", err)
            return
        }

        // Normalize current URL
        normalizedCurrentURL, err := utils.NormalizeURL(rawCurrentURL)
        if err != nil {
            // Skip URL if we can't normalize it
            log.Printf("Error normalizing current URL: %v\n", err)
            continue
        }


        log.Printf("Popped URL: %v | Normalized URL: %v\n", rawCurrentURL, normalizedCurrentURL)
        // time.Sleep(1 * time.Second)

        // Check if the URL has been visited
        visited, err := db.HasURLBeenVisited(normalizedCurrentURL)
        if err != nil {
            log.Printf("Error: [%v] - skipping...\n", err)
            continue
        }

        if visited {
            log.Printf("Skipping %v - already visited\n", normalizedCurrentURL)
            continue
        }

        log.Printf("Crawling from %v (%v)...\n", normalizedCurrentURL, rawCurrentURL)
        // time.Sleep(1 * time.Second)

        // Fetch HTML, Status Code, and Content-Type
        html, statusCode, contentType, err := getPageData(rawCurrentURL)
        if err != nil {
            // Skip if we couldn't fetch the data
            log.Printf("Error fetching %v data: %v\n", rawCurrentURL, err)
            continue
        }

        // Fetch the links of the current page
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
        err = crawcfg.addPage(pg)
        if err != nil {
            log.Printf("\tError adding page visit: %v\n", err)
            continue
        }

        err = db.VisitPage(normalizedCurrentURL)
        if err != nil {
            log.Printf("\tError adding page visit: %v\n", err)
            continue
        }

        log.Printf("Adding links from %v (%v)...\n", normalizedCurrentURL, rawCurrentURL)
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
                score -= 0.001
            } else {
                score = depthLevel + 1
            }

            score = math.Max(utils.MinScore, math.Min(score, utils.MaxScore))

            // Update score based on depth
            _ = db.PushURL(rawCurrentLink, score)
            // if err == nil {
            //     log.Printf("\tPUSH => %v\n", rawCurrentLink)
            // }
            // if err != nil {
            //     log.Printf("\tError pushing '%v' to the queue: %v\n", rawCurrentLink, err)
            //     log.Printf("\t\tSkipping...")
            //     continue
            // }
        }

        // time.Sleep(1 * time.Second)
    }
}

