package main

import (
    "log"
    "sync"
    "flag"
    "os"

    "github.com/IonelPopJara/search-engine/services/spider/internal/pages"
    "github.com/IonelPopJara/search-engine/services/spider/internal/utils"
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
    maxConcurrency := flag.Int("max-concurrency", 5, "Maximum number of concurrenet workers")
    maxPages := flag.Int("max-pages", 30, "Maximum number of pages per batch")
    startingURL := flag.String("starting-url", "https://en.wikipedia.org/wiki/Kamen_Rider", "Starting URL for this spider")

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
    // PushURL also creates a lookup entry
    msg := db.PushURL(*startingURL, 0)
    log.Printf("PUSH %v - %v\n", *startingURL, msg)

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
    }

    // Infinite loop to crawl the web in batches
    for {

        // Check how busy the indexer queue is
        log.Printf("Checking number of entries...\n")
        // If we have reached the maximum number of entries in the spider queue
        queueSize, err := db.GetIndexerQueueSize()
        if err != nil {
            log.Printf("Error getting indexer queue: %v\n", err)
            return
        }

        if queueSize >= utils.MaxIndexerQueueSize {
            log.Printf("Indexer queue is full. Waiting...\n")
            // Wait until we receive a signal to start crawling again
            for {
                sig, err := db.PopSignalQueue()
                if err != nil {
                    log.Printf("could not get signal: %v\n", err)
                    return
                }

                if sig == utils.ResumeCrawl {
                    log.Printf("Resume crawl!\n")
                    break
                }
            }
        }

        log.Printf("Spawning workers...\n")
        for i := 0; i < crawler.MaxConcurrency; i++ {
            crawler.Wg.Add(1)
            go crawler.Crawl(db)
        }

        crawler.Wg.Wait()

        // Write entries to the db
        pageController.SavePages(crawler)
        linksController.SaveLinks(crawler)
        imageController.SaveImages(crawler)

        // Clean visited pages by this runner
        crawler.Pages = make(map[string]*pages.Page)
        crawler.Outlinks = make(map[string]*pages.PageNode)
        crawler.Backlinks = make(map[string]*pages.PageNode)
        crawler.Images = make(map[string][]*pages.Image)
    }
}

