# MOOGLE (The worst best search engine)

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

## TODO:

- [x] Add server side rendering again
- [x] Implement pagination
- [x] Make placeholder frontend again
    - [x] Landing page
    - [ ] Results page (this also shows the first 5 images found)
    - [x] Images tab (just like google search images)

- [ ] Omit images that are less than 100x100 px
- [ ] Check description in the indexer
- [ ] Modify indexer to also account for page url words for the weight
- [ ] Filename is not being stored by the indexer
- [ ] Load balancing
- [ ] Implement query filter service
- [ ] Handle weird queries like 'something+something' in the frontend
- [ ] Fix the links not updating to purple
- [ ] Add search suggeetsions based on most popular searches
- [ ] Add "more than x amount of entries" at the begining of the client. Like when google said "Soon 2.5 million entries or something"
- [ ] Check for words that could be hyphened or not (Megaman mega man mega-man). Idk how to check this yet.
    - [ ] Just check all possible combintions like "mega-man" "megaman" "mega" "man"
- [ ] Add favicon icons for preview
- [x] Add mongodb support
- [x] Refactor crawler code and normalize the names for the datasets
- [x] Remove datasets after being crawled: CHECK THIS ONE
- [x] Wait for message queue to restart when the queue is full
- [x] Pause the crawlers when there is more than a certain amount of pages in the indexer_queue. Wait until the pages reach <500 or until the monitoring service sends a message
- [x] Implement mongodb storage
- [x] Make a monitoring service to scale crawlers dynamically
- [x] Add images to indexer
- [x] Ignore .svg and .ico
- [x] Dockerize the indexer so it can be scaled
- [x] Update metadata struct in the DB
    - [x] include url
- [x] Finish the monitoring service
- [x] Clean up the codebase
    - [x] Remove environment variables from code
- [x] Fix crawler overhead
    - [x] Normalize urls before pushing
    - [x] Convert the queue to a set
    - [x] Use a priority queue based on weights and depth level
- [x] Set up building pipeline
- [x] Finish seting up the server
- [x] Add authentication to the DB
- [x] Remove cites because it produces duplicated content en.wikipedia.org/wiki/Anime#cite_ref-bbc_43-0o
- [x] Check for pages encoding
- [x] Omit pages that are not in english
- [x] Process text more efficiently so summary text is more readable
- [x] Setup github pipeline and pass db credentials as secrets
- [x] If I have time, adding images would be sick

- [x] Update search-engine to use pipelines
- [x] Try to implement page rank algorithm. But I need to check for inner links or something...

## Components

- [x] Crawler
- [x] Indexer
- [x] Search engine
- [ ] Query filter
- [x] Frontend


## Top Priority
- [x] Images
- [ ] Make frontend prettier
- [x] Docker images
- [x] Scaling up/down
- [ ] Load balancing

    parsedURL, err := url.Parse("https://en.wikipedia.org/wiki/Mega_Man_X")

If you choose to use local Redis we strongly recommend using Docker. If you choose not to use Docker, use the following instructions based on your OS:

sudo docker run -d --name redis-stack -p 6379:6379 -p 8001:8001 redis/redis-stack:latest

❯ go run ./cmd/spider/main.go
Connecting to Redis...
Successfully connected to Redis!
Page:
-------------------------------------
URL: hello.com
HTML: <h1>WTF</h1>
Last Crawled: Tue, 04 Mar 2025 20:49:07 CET
Outgoing Links: 0
Status Code: 404
Content Type: text/html
-------------------------------------

Page Hash:
map[content_type:text/html html:<h1>WTF</h1> last_crawled:<h1>WTF</h1> outgoing_links:[] status_code:404 url:hello.com]
panic: redis: can't marshal []string (implement encoding.BinaryMarshaler)

goroutine 1 [running]:
main.main()
	/home/mults/Development/search-engine/services/spider/cmd/spider/main.go:49 +0x669
exit status 2

sudo docker run -d --name redis-stack -p 6379:6379 -p 8001:8001 redis/redis-stack:latest

sudo docker start redis-stack

I hate race conditions :(


SILENE TEST DRIVEN DEVEVLOPER


The main problem that docker solves is hte "It works on my machine"
