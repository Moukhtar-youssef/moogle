# MOOGLE (The worst best search engine)

# TODO:

- [ ] Process text more efficiently so summary text is more readable
- [ ] Update metadata struct in the DB
    - [ ] include url
    - [ ] Store <h1>s
- [ ] Update search-engine to use pipelines
- [ ] Add authentication to the DB
- [ ] Try deploying the database making docker images of this
- [ ] Load balancing
- [ ] Implement query filter service
- [ ] Handle weird queries like 'something+something' in the frontend
- [ ] Fix the links not updating to purple
- [ ] Remove cites because it produces duplicated content en.wikipedia.org/wiki/Anime#cite_ref-bbc_43-0o
- [ ] Add search suggeetsions based on most popular searches
- [ ] Check for pages encoding
- [ ] Add "more than x amount of entries" at the begining of the client. Like when google said "Soon 2.5 million entries or something"
- [ ] If I have time, adding images would be sick
- [ ] Add favicon icons for preview

## Components

- [x] Crawler
- [x] Indexer
- [x] Search engine
- [ ] Query filter
- [x] Frontend

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
