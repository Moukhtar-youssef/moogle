# Indexer

## TODO

- [x] Make a dictionary with the word count
- [x] Create a redis hash containing the word as the key, and the urls and weights as parameters
        word:some_word = {[google.com, 10], [reddit.com, 5]}
- [ ] Read all entries periodically
    - [ ] Compare last_crawled with the last_crawled in the db
    - [ ] If it doesn't exist or the entry was updated, add entry
- [ ] Ignore pages that are not in english
