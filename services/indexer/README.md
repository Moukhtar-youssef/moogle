# Indexer

## Things that could be bottlenecks
- [ ] Loading nlptk, maybe I can just copy the array value since that's the only thing I need
- [ ] Not performing batches insert in mongo
- [ ] Too many loops
- [ ] Language detection
## TODO

- [x] Make a dictionary with the word count
- [x] Create a redis hash containing the word as the key, and the urls and weights as parameters
        word:some_word = {[google.com, 10], [reddit.com, 5]}
- [x] Read all entries periodically
    - [x] Compare last_crawled with the last_crawled in the db
    - [x] If it doesn't exist or the entry was updated, add entry
- [x] Ignore pages that are not in english
- [x] Index images

Make a new service for images. Make sure to store the most important keywords of each page in the database. Then I'll fetch for that in the image service.

Write the word operations in batches using multithreading