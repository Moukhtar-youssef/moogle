# Indexer

## TODO

- [x] Make a dictionary with the word count
- [x] Create a redis hash containing the word as the key, and the urls and weights as parameters
        word:some_word = {[google.com, 10], [reddit.com, 5]}
- [x] Read all entries periodically
    - [x] Compare last_crawled with the last_crawled in the db
    - [x] If it doesn't exist or the entry was updated, add entry
- [x] Ignore pages that are not in english
- [x] Index images
