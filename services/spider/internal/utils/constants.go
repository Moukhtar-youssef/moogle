package utils

import(
    "time"
)

const (
    URLQueueKey = "set_url_queue"
    NormalizedURLPrefix = "normalized_url"
    PagePrefix = "page"
    ImagePrefix = "image"
    PageImagesPrefix = "page_images"
    BacklinksPrefix = "backlinks"
    OutlinksPrefix = "outlinks"
    Timeout = 5 * time.Second
    MaxScore = 10000
    MinScore = -1000
)

