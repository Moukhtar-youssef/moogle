package utils

import(
    "time"
)

const (
    // Crawler constants
    Timeout = 5 * time.Second
    MaxScore = 10000
    MinScore = -1000

    // Message Queues
    SpiderQueueKey = "spider_queue"
    IndexerQueueKey = "pages_queue"
    SignalQueueKey = "signal_queue"
    ResumeCrawl = "RESUME_CRAWL"
    MaxIndexerQueueSize = 100

    // Redis Data
    NormalizedURLPrefix = "normalized_url"
    PagePrefix = "page_data"
    ImagePrefix = "image_data"
    PageImagesPrefix = "page_images"

    // These ones will be saved by the indexer and the backlinks server
    BacklinksPrefix = "backlinks"
    OutlinksPrefix = "outlinks"
)

