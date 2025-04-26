# MOOGLE - The Worst Best Search Engine

## TODO
- [ ] Test the flattened database in the query engine
- [ ] If everything works, rename the database and start the spiders, indexers, and other services to point to this new database
- [ ] Fix the query engine to use the new database

## For Future
- [ ] Use stems for words
- [ ] Check for words that could be hyphened or not (Megaman mega man mega-man).
- [x] Implement query filter service
- [ ] Handle weird queries like 'something+something' in the frontend

## Components

- [x] Crawler
- [x] Indexer
- [x] Search engine
- [x] Backlinks
- [x] TFIDF
- [x] PageRank
- [x] Query filter
- [x] Frontend
- [ ] Monitoring

## Top Priority
- [x] Images
- [x] Make frontend prettier
- [x] Docker images
- [x] Scaling up/down
- [x] Load balancing


## Install Redis
If you choose to use local Redis we strongly recommend using Docker. If you choose not to use Docker, use the following instructions based on your OS:

```bash
sudo docker run -d --name redis-stack -p 6379:6379 -p 8001:8001 redis/redis-stack:latest
sudo docker start redis-stack
```

SILENCE TEST DRIVEN DEVEVLOPER

## MESSAGE QUEUES
- `spider_queue`: ZSET
- `indexer_queue`: LIST
- `signal_queue`: LIST

## REDIS KEYS
- `page_data:<normalized_page_url>`: HASH
    - 1) "html" - string
    - 2) "last_crawled" - timestamp
- `page_images:<normalized_page_url>`: SET { `<image_url>` }
- `image_data:<image_url>`: HASH
    - 1) "normalized_page_url" - string
    - 2) "alt" - string
This one will be used to check if it has been visited before
- `normalized_url:<normalized_page_url>`: HASH
    - 1) "raw_url" - string
    - 2) "visited" - integer (0 = false, 1 = true)

## BOTH
Outlinks will be saved to storage db in the indexer
- `backlinks:<normalized_page_url>`: SET { `<normalized_page_url>` }
A different service will be in charge of saving and updating backlinks
- `outlinks:<normalized_page_url>`: SET { `<normalized_page_url>` }


## NOSQL DATA
- `word:<word>`: ZSET { `<normalized_page_url>`, SCORE }
- `word_images:<word>` ZSET { `<normalized_image_url>`, SCORE }
- `url_metadata:<normalized_page_url>`: HASH
    - 1) "title" - integer
    - 2) "summary_text" - string
    - 3) "description" - string
    - 4) "normalized_url" - string # I don't remember why I need this but I do
    - 4) "raw_url" - string # I don't remember why I need this but I do
    - 5) "last_crawled" - timestamp
- `image:<image_url>`: HASH
    - 1) "normalized_page_url" - string
    - 2) "alt" - string
    - 3) "keywords" - List[string]

## REDIS KEYS
- `backlinks:*`: SET { }
    - set{ "normalized_page_url" }
- `outlinks:*`: SET
    - set{ "normalized_page_url" }

Crawl links and push page data to redis
Also push backlinks, outlinks, page_image

These ones are only for transfering information to the indexer, can be deleted from the db so it's good to keep in redis
- `page_data:<normalized_page_url>`: HASH
    - 1) "status_code" - integer
    - 2) "content_type" - string
    - 3) "normalized_url" - string
    - 4) "html" - string
    - 5) "last_crawled" - timestamp
- `page_images:<normalized_page_url`: SET
    - set{ "normalized_image_url" }
After processing a page, you delete these ones from redis

These ones I can save 

These need to be in storage
- `image:<image_url>`: HASH
    - 1) "normalized_page_url" - string
    - 2) "alt" - string
- `normalized_url:<normalized_page_url>` STRING
    - "raw_url"

These should be saved in storage (Page Rank Algorithm)
- `backlinks:*`: SET
    - set{ "normalized_page_url" }
- `outlinks:*`: SET
    - set{ "normalized_page_url" }

These are given by the indexer (Page Rank Algorithm)
- `word:<word>`: ZSET
    - set{ "normalized_page_url", SCORE }
- `word_images:<word>` ZSET
    - set{ "normalized_image_url", SCORE }

This is to display results
- `url_metadata:<normalized_page_url>`: hash
    - 1) "title" - integer
    - 2) "summary_text" - string
    - 3) "description" - string
    - 4) "normalized_url" - string
    - 5) "last_crawled" - timestamp




	// Redis Data: some keys stay in Redis indefinitely, while others are transfer to MongoDB by other services
	NormalizedURLPrefix = "normalized_url"	// Stays in Redis indefinitely
	PagePrefix          = "page_data"		// Transferred by the indexer
	ImagePrefix         = "image_data"		// Transferred by the image indexer
	PageImagesPrefix    = "page_images"		// Transferred by the image indexer
	BacklinksPrefix		= "backlinks"		// Transferred by the backlinks processor
	OutlinksPrefix 		= "outlinks"		// Transferred by the indexer