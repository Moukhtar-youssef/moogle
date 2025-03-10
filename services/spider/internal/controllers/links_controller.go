package controllers

import(
    "log"

    "github.com/IonelPopJara/search-engine/services/spider/internal/database"
    "github.com/IonelPopJara/search-engine/services/spider/internal/crawler"
)

type OutlinksController struct {
    db *database.Database
}

func NewOutlinksController(db *database.Database) *OutlinksController {
    return &OutlinksController{
        db: db,
    }
}

func (pgc *OutlinksController) SaveOutlinks(crawcfg *crawler.CrawlerConfig) {
    log.Printf("Saving outlinks...\n")

    data := crawcfg.Outlinks
    pipeline := pgc.db.Client.Pipeline()

    for key, outlinks := range data {
        // check if it's not already crawled
        if _, exists := crawcfg.CachedPages[key]; !exists {
            for _, link := range outlinks.GetOutlinks() {
                pipeline.SAdd(pgc.db.Context, "outlinks:"+key, link)
            }
        }
    }

    _, err := pipeline.Exec(pgc.db.Context)
    if err != nil {
        log.Printf("Error executing pipeline: %v", err)
    } else {
        log.Printf("Successfully written %d entries to the db!", len(data))
    }
}
