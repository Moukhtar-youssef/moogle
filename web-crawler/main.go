package main

import (
    "fmt"
    "os"
    "net/url"
    "sync"
    "flag"
)

func main() {

    maxConcurrency := flag.Int("c", 10, "Maximum concurrent workers")
    maxPages := flag.Int("p", 10, "Maximum pages allowed to crawl")
    queueSize := flag.Int("q", 10000, "Queue size")
    timeOut := flag.Int("t", 5, "Maximum timeout allowed in seconds")
    rawBaseURL := flag.String("u", "", "Base URL to start crawling from [REQUIRED]")
    DFS := flag.Bool("DFS", false, "DFS mode enabled. By default the crawler will use BFS")
    multipleHostsAllowed := flag.Bool("mHost", false, "Allows the crawler to fetch URLs from different hosts")
    shouldPrintReport := flag.Bool("pReport", true, "Prints a report with the links found")

    // Parse the flags
    flag.Parse()

    if *rawBaseURL == "" {
        fmt.Printf("Error! Please provide a base url.\n")
        flag.Usage()
        os.Exit(1)
    }

    // Parse base URL
    parsedURL, err := url.Parse(*rawBaseURL)
    if err != nil {
        fmt.Printf("Error parsing base URL: %v\n", err)
        os.Exit(1)
    }

    // Initialize the config struct
    cfg := &config{
        maxPages:           *maxPages,
        timeOut:            *timeOut,
        pages:              make(map[string]int),
        baseURL:            parsedURL,
        mu:                 &sync.Mutex{},
        concurrencyControl: make(chan struct{}, *maxConcurrency),
        wg:                 &sync.WaitGroup{},
        queue:              make(chan string, *queueSize),
        multipleHosts:      *multipleHostsAllowed,
    }

    fmt.Printf("Starting crawl from: %v...\n", rawBaseURL)

    if *DFS {
        // Use DFS to crawl
        cfg.wg.Add(1)
        go cfg.crawlPage(*rawBaseURL)
        cfg.wg.Wait()
    } else {
        // Use BFS to crawl
        fmt.Printf("Adding baseURL: '%v' to the queue...\n", cfg.baseURL)
        cfg.queue <- cfg.baseURL.String()

        // Instantiate worker pool
        for i := 0; i < 10; i++ {
            cfg.wg.Add(1)
            go cfg.crawlPageBFS()
        }

        cfg.wg.Wait()
    }

    if *shouldPrintReport {
        printReport(cfg.pages, cfg.baseURL.String())
    }

    return
}

