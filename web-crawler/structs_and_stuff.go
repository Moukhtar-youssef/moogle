package main

import (
    "net/url"
    "sync"
)

type config struct {
    maxPages            int
    timeOut             int
    pages               map[string]int
    baseURL             *url.URL
    mu                  *sync.Mutex
    concurrencyControl  chan struct {}
    wg                  *sync.WaitGroup
    queue               chan string
    multipleHosts       bool
}

func (cfg *config) pagesLen() int {
    cfg.mu.Lock()
    defer cfg.mu.Unlock()
    return len(cfg.pages)
}

func (cfg *config) addPageVisit(normalizedURL string) (isFirst bool) {
    cfg.mu.Lock()
    defer cfg.mu.Unlock()

    if _, visited := cfg.pages[normalizedURL]; visited {
        cfg.pages[normalizedURL] += 1
        return false
    }

    cfg.pages[normalizedURL] = 1
    return true
}

func (cfg *config) canAddPage() (canAdd bool) {
    cfg.mu.Lock()
    defer cfg.mu.Unlock()

    // Check if we have enough space to add the visit
    if cfg.maxPages > 0 && len(cfg.pages) >= cfg.maxPages {
        return false
    }

    return true
}

