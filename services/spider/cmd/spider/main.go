package main

import (
    "log"
    "sync"
    "flag"

    "github.com/IonelPopJara/search-engine/services/spider/internal/pages"
    "github.com/IonelPopJara/search-engine/services/spider/internal/links"
    "github.com/IonelPopJara/search-engine/services/spider/internal/crawler"
    "github.com/IonelPopJara/search-engine/services/spider/internal/database"
    "github.com/IonelPopJara/search-engine/services/spider/internal/controllers"
)

func main() {

    // Parse flags
    maxConcurrency := flag.Int("max-concurrency", 10, "Maximum number of concurrenet workers")
    maxPages := flag.Int("max-pages", 100, "Maximum number of pages per batch")
    timeout := flag.Int("timeout", 5, "Maximum timeout allowed")
    queueKey := flag.String("url-queue-name", "url_queue", "Key of the redis url queue")

    flag.Parse()

    // Connect to Redis
    db := &database.Database{}
    err := db.ConnectToRedis()
    if err != nil {
        log.Printf("Error: Couldn't connect to Redis: %v\n", err)
        return
    }

    // Instantiate controllers
    pageController := controllers.NewPageController(db)
    outlinksController := controllers.NewOutlinksController(db)

    // Instantiate crawler
    crawler := &crawler.CrawlerConfig {
        Mu:             &sync.Mutex{},
        Wg:             &sync.WaitGroup{},
        Pages:          make(map[string]*pages.Page),
        Outlinks:       make(map[string]*links.Outlinks),
        MaxPages:       *maxPages,
        Timeout:        *timeout,
        MaxConcurrency: *maxConcurrency,
        QueueKey:       *queueKey,
        CachedPages:    make(map[string]*pages.Page),
    }

    // Populate the CachedPages map
    crawler.CachedPages = pageController.GetAllPages()

    // Infinite loop to crawl the web in batches
    for {
        log.Printf("Spawning workers...\n")
        for i := 0; i < crawler.MaxConcurrency; i++ {
            crawler.Wg.Add(1)
            go crawler.Crawl(db)
        }

        crawler.Wg.Wait()

        // Repopulate the CachedPages map to account for changes with other runners
        crawler.CachedPages = pageController.GetAllPages()

        // Write entries to the db
        pageController.SavePages(crawler)
        outlinksController.SaveOutlinks(crawler)
        log.Printf("DB updated!\n")

        // Repopulate the CachedPages map to account for changes with other runners
        crawler.CachedPages = pageController.GetAllPages()

        // Clean visited pages by this runner
        crawler.Pages = make(map[string]*pages.Page)
    }
}

