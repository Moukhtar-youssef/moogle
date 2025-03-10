package crawler

import (
    "sync"

    "github.com/IonelPopJara/search-engine/services/spider/internal/pages"
    "github.com/IonelPopJara/search-engine/services/spider/internal/links"
)

// When the pages reaches a length of maxPages, stop the cycle, fetch/write data, and start again
type CrawlerConfig struct {
    Mu                  *sync.Mutex         // Sync
    Wg                  *sync.WaitGroup     // Sync
    Pages               map[string]*pages.Page// Discovered pages
    Outlinks            map[string]*links.Outlinks// Discovered outlinks
    MaxPages            int                 // Max discovered pages
    Timeout             int                 // Timeout in seconds
    MaxConcurrency      int                 // Maximum concurrent workers in the pool
    QueueKey            string              // Redis queue key
    CachedPages         map[string]*pages.Page// All the db pages cached
}

func (crawcfg *CrawlerConfig) lenPages() int {
    crawcfg.Mu.Lock()
    defer crawcfg.Mu.Unlock()

    return len(crawcfg.Pages)
}

func (crawcfg *CrawlerConfig) maxPagesReached() (bool) {
    crawcfg.Mu.Lock()
    defer crawcfg.Mu.Unlock()

    if len(crawcfg.Pages) >= crawcfg.MaxPages {
        // Can't add more pages because max pages has been reached
        return true
    }

    // Max pages has not been reached
    return false
}


func (crawcfg *CrawlerConfig) canVisitPage(normalizedURL string) (bool) {
    crawcfg.Mu.Lock()
    defer crawcfg.Mu.Unlock()

    if _, visited := crawcfg.Pages[normalizedURL]; visited {
        return false
    }

    if _, visited := crawcfg.CachedPages[normalizedURL]; visited {
        // TODO: Check timestamp
        return false
    }

    return true
}

func (crawcfg *CrawlerConfig) addPageVisit(page *pages.Page, outlinks *links.Outlinks) (bool) {
    crawcfg.Mu.Lock()
    defer crawcfg.Mu.Unlock()

    normalizedURL := page.NormalizedURL

    if _, visited := crawcfg.Pages[normalizedURL]; visited {
        return false
    }

    if _, visited := crawcfg.CachedPages[normalizedURL]; visited {
        // TODO: Check timestamp
        return false
    }

    if len(crawcfg.Pages) >= crawcfg.MaxPages {
        // Can't add more pages because max pages has been reached
        return false
    }

    crawcfg.Pages[normalizedURL] = page
    crawcfg.Outlinks[normalizedURL] = outlinks
    return true
}

