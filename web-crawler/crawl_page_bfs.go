package main

import (
    "fmt"
    "net/url"
    "time"
)

func (cfg* config) crawlPageBFS() {
    defer cfg.wg.Done()
    fmt.Printf("Starting web crawler instance...\n")

    for {
        select {
        case rawCurrentURL, ok := <-cfg.queue:
            if !ok {
                fmt.Printf("\tQueue was closed. Stopping workers...\n")
                // Channel was closed
                return
            }

            // Parse current URL
            currentURL, err := url.Parse(rawCurrentURL)
            if err != nil {
                fmt.Printf("\tError parsing current URL: %v\n", err)
                return
            }

            // Check if the hosts match
            if currentURL.Hostname() != cfg.baseURL.Hostname() {
                if !cfg.multipleHosts {
                    fmt.Printf("\tNot the same base URL! %v | %v\n", cfg.baseURL, currentURL)
                    continue
                }
            }

            // Normalize current URL
            normalizedCurrentURL, err := normalizeURL(rawCurrentURL)
            if err != nil {
                fmt.Printf("\tError normalizing current URL: %v\n", err)
                continue
            }

            // Check for max pages
            enoughSpace := cfg.canAddPage()
            if !enoughSpace {
                fmt.Printf("\tMaximum number of pages reached!\n")
                return
            }

            // Check if the URL has been visited
            isPageNew := cfg.addPageVisit(normalizedCurrentURL)
            if !isPageNew {
                continue
            }

            fmt.Printf("BFSing %v...\n", normalizedCurrentURL)

            // Get the HTML of the current page
            html, err := getHTML(rawCurrentURL)
            if err != nil {
                fmt.Printf("\tError fetching HTML: %v\n", err)
                continue
            }

            // Fetch the links of the current page
            links, err := getURLsFromHTML(html, cfg.baseURL)
            if err != nil {
                fmt.Printf("\tError getting links from HTML: %v\n", err)
                continue
            }

            fmt.Printf("Adding found links to the queue...\n")

            for _, rawCurrentLink := range links {
                select {
                case cfg.queue <- rawCurrentLink:
                    // Successfully sent
                default:
                    fmt.Printf("\tQueue is full, skipping link: %s\n", rawCurrentLink)
                }
            }
        case <-time.After(time.Duration(cfg.timeOut) * time.Second):
            // Check if the queue is empty
            if len(cfg.queue) == 0 {
                fmt.Printf("\tQueue is empty after %v seconds, stopping worker...\n", cfg.timeOut)
                return
            }
        }
    }
    return
}

