package main

import (
    "log"
    "sync"
    "flag"
    "os"

    "github.com/IonelPopJara/search-engine/services/spider/internal/pages"
    "github.com/IonelPopJara/search-engine/services/spider/internal/crawler"
    "github.com/IonelPopJara/search-engine/services/spider/internal/database"
    "github.com/IonelPopJara/search-engine/services/spider/internal/controllers"
)

func getEnv(key, fallback string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }

    return fallback
}

func main() {

    // Parse flags
    maxConcurrency := flag.Int("max-concurrency", 20, "Maximum number of concurrenet workers")
    maxPages := flag.Int("max-pages", 100, "Maximum number of pages per batch")

    flag.Parse()

    // Retrive environment variables
    redisHost := getEnv("REDIS_HOST", "localhost")
    redisPort := getEnv("REDIS_PORT", "6379")
    redisPassword := getEnv("REDIS_PASSWORD", "")
    redisDB := getEnv("REDIS_DB", "0")

    // Connect to Redis
    db := &database.Database{}
    err := db.ConnectToRedis(redisHost, redisPort, redisPassword, redisDB)
    if err != nil {
        log.Printf("Error: %v\n", err)
        return
    }

    // Add an entry to the message queue with score 0 (high priority)
    err = db.PushURL("https://en.wikipedia.org/wiki/Kamen_Rider", 0)
    if err != nil {
        log.Printf("")
        return
    }

    // Instantiate controllers
    pageController := controllers.NewPageController(db)
    linksController := controllers.NewLinksController(db)
    imageController := controllers.NewImageController(db)

    // Instantiate crawler
    crawler := &crawler.CrawlerConfig {
        Mu:             &sync.Mutex{},
        Wg:             &sync.WaitGroup{},
        Pages:          make(map[string]*pages.Page),
        Outlinks:       make(map[string]*pages.PageNode),
        Backlinks:      make(map[string]*pages.PageNode),
        Images:         make(map[string][]*pages.Image),
        MaxPages:       *maxPages,
        MaxConcurrency: *maxConcurrency,
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
        linksController.SaveLinks(crawler)
        imageController.SaveImages(crawler)

        // Repopulate the CachedPages map to account for changes with other runners
        crawler.CachedPages = pageController.GetAllPages()

        // Clean visited pages by this runner
        crawler.Pages = make(map[string]*pages.Page)
        crawler.Outlinks = make(map[string]*pages.PageNode)
        crawler.Backlinks = make(map[string]*pages.PageNode)
        crawler.Images = make(map[string][]*pages.Image)
    }
}

