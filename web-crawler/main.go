package main

import (
    "fmt"
    "os"
    "net/url"
    "sync"
    "flag"
)

func main() {
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, `Usage: %s <URL> [options]

        Options:
        -concurrency <num>         Maximum concurrent workers (default: 10)
        -max-pages <num>           Maximum pages allowed to crawl (default: 10)
        -queue-size <num>          Queue size (default: 10000)
        -timeout <seconds>         Maximum timeout allowed (default: 5)
        -use-dfs                   Enable DFS mode (default: BFS)
        -allow-multiple-hosts      Allow fetching URLs from different hosts (default: true)
        -print-report              Print a report of links found (default: true)

        Example:
        %s https://example.com -concurrency 20 -max-pages 50 -use-dfs
        `, os.Args[0], os.Args[0])
    }

    maxConcurrency := flag.Int("concurrency", 10, "Maximum concurrent workers")
    maxPages := flag.Int("max-pages", 10, "Maximum pages allowed to crawl")
    queueSize := flag.Int("queue-size", 10000, "Queue size")
    timeout := flag.Int("timeout", 5, "Maximum timeout allowed in seconds")
    DFS := flag.Bool("use-dfs", false, "DFS mode enabled. By default the crawler will use BFS")
    multipleHostsAllowed := flag.Bool("allow-multiple-hosts", true, "Allows the crawler to fetch URLs from different hosts")
    shouldPrintReport := flag.Bool("print-report", true, "Prints a report with the links found")

    // Parse the flags
    flag.Parse()

    args := flag.Args()

    if len(args) == 0 {
        fmt.Printf("Error! Please provide a base url.\n")
        flag.Usage()
        os.Exit(1)
    }

    rawBaseURL := args[0]

    // Parse base URL
    parsedURL, err := url.Parse(rawBaseURL)
    if err != nil {
        fmt.Printf("Error parsing base URL: %v\n", err)
        os.Exit(1)
    }

    // Initialize the config struct
    cfg := &config{
        maxPages:           *maxPages,
        timeout:            *timeout,
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
        go cfg.crawlPage(rawBaseURL)
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

