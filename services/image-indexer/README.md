# Image Indexer

This service fetches images from the image message queue, processes them, and stores them in the database.

It first checks for the size of the image.
To do this, it fetches the keywords of the page where the image was found from mongo, and then it stores the image_words in the database.

Write the word operations in batches using multithreading