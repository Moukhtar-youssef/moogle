package main

import (
    "context"
    "log"
    "sync"
    "net/url"

    "github.com/redis/go-redis/v9"
    "github.com/IonelPopJara/search-engine/services/spider/internal/pages"
    "github.com/IonelPopJara/search-engine/services/spider/internal/crawler"
)

func newClient() (*redis.Client, error) {
    log.Println("Connecting to Redis...")

    rdb := redis.NewClient(&redis.Options{
        Addr:       "localhost:6379",
        Password:   "",
        DB:         0,
        Protocol:   2,
    })

    log.Println("Successfully connected to Redis!")

    return rdb, nil
}

func populateCachedPages(rdb *redis.Client, ctx context.Context) map[string]*pages.Page {
    log.Printf("Fetching data from Redis...\n")
    redisPages := make(map[string]*pages.Page)

    keys, err := rdb.Keys(ctx, "page:*").Result()
    if err != nil {
        log.Printf("Error fetching data from Redis: %v\n", err)
        return nil
    }

    // Process the redis data
    for _, key := range keys {
        data, err := rdb.HGetAll(ctx, key).Result()
        if err != nil {
            log.Printf("Error fetching data from Redis for %s: %v", key, err)
            return nil
        }

        page, err := pages.DehashPage(data)
        if err != nil {
            log.Printf("Error dehashing page from Redis for %v", err)
            return nil
        }

        redisPages[page.NormalizedURL] = page
    }

    return redisPages
}


func main() {

    // Connect to redis
    rdb, err := newClient()
    if err != nil {
        log.Printf("Error connecting to redis: %w", err)
    }

    // Get context background
    ctx := context.Background()

    parsedURL, err := url.Parse("https://wagslane.dev")
    if err != nil {
        log.Printf("Error parsing URL: %w", err)
    }

    crawler := &crawler.CrawlerConfig {
        StartURL:       parsedURL,
        Mu:             &sync.Mutex{},
        Wg:             &sync.WaitGroup{},
        Pages:          make(map[string]*pages.Page),
        MaxPages:       20,
        Timeout:        5,
        MaxConcurrency: 10,
        QueueKey:       "url_queue",
        CachedPages:    make(map[string]*pages.Page),
    }

    log.Printf("Adding starting URL to the shared queue...\n")
    err = rdb.LPush(ctx, crawler.QueueKey, crawler.StartURL.String()).Err()
    if err != nil {
        log.Printf("Error pushing URL to the queue: %v\n", err)
    }

    // Populate the CachedPages map
    crawler.CachedPages = populateCachedPages(rdb, ctx)

    // Infinite loop to crawl the web in batches
    for {
        log.Printf("Spawning workers...\n")
        for i := 0; i < crawler.MaxConcurrency; i++ {
            crawler.Wg.Add(1)
            go crawler.Crawl(rdb, &ctx)
        }

        crawler.Wg.Wait()

        // Repopulate the CachedPages map to account for changes with other runners
        crawler.CachedPages = populateCachedPages(rdb, ctx)

        // Write entries to the db
        log.Printf("Writing %d entries to the db...\n", len(crawler.Pages))
        for _, page := range crawler.Pages {
            // Compare the crawler.Pages entries to the crawler.CachedPages entries
            // If the entry doesn't exist
            if _, exists := crawler.CachedPages[page.NormalizedURL]; !exists {
                // Hash the structs

                // Store it in Redis
                pageHash, err := pages.HashPage(page)
                if err != nil {
                    log.Printf("Error hasing page: %w", err)
                }

                // Store the data in redis
                _, err = rdb.HSet(ctx, "page:"+page.NormalizedURL, pageHash).Result()
                if err != nil {
                    log.Printf("Error storing data in redis: %w", err)
                }
            }
        }
        log.Printf("DB updated!\n")

        // Repopulate the CachedPages map to account for changes with other runners
        crawler.CachedPages = populateCachedPages(rdb, ctx)

        // Clean visited pages by this runner
        crawler.Pages = make(map[string]*pages.Page)
    }
}

