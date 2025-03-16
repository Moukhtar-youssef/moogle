# MOOGLE (The worst best search engine)

# TODO:

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
- [ ] Setup github pipeline and pass db credentials as secrets
- [ ] If I have time, adding images would be sick

- [x] Update search-engine to use pipelines
- [ ] Load balancing
- [ ] Implement query filter service
- [ ] Handle weird queries like 'something+something' in the frontend
- [ ] Fix the links not updating to purple
- [ ] Add search suggeetsions based on most popular searches
- [ ] Add "more than x amount of entries" at the begining of the client. Like when google said "Soon 2.5 million entries or something"
- [ ] Check for words that could be hyphened or not (Megaman mega man mega-man). Idk how to check this yet.
    - [ ] Just check all possible combintions like "mega-man" "megaman" "mega" "man"
- [ ] Try to implement page rank algorithm. But I need to check for inner links or something...
- [ ] Add favicon icons for preview

## Components

- [x] Crawler
- [x] Indexer
- [x] Search engine
- [ ] Query filter
- [x] Frontend


## Top Priority
- [ ] Images
- [ ] Make frontend prettier
- [ ] Docker images
- [ ] Scaling up/down
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
