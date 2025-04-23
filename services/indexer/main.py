import logging
import signal
import sys
import os

from utils.constants import *
from models.image import Image
from data.redis_client import RedisClient
from data.mongo_client import MongoClient
from utils.utils import get_html_data, split_name, is_valid_image, split_url

import time
import os.path

from collections import Counter

# SETUP LOGGER
logger = logging.getLogger(__name__)
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# SHUTDOWN
running = True

def handle_exit(signum, frame):
    global running
    logger.info('Termination signal received - shutting down...')
    running = False
    sys.exit(0)

signal.signal(signal.SIGTERM, handle_exit)
signal.signal(signal.SIGINT, handle_exit)

if __name__ == "__main__":
    # REDIS ENV VARIABLES
    redis_host = os.getenv('REDIS_HOST', 'localhost')
    redis_port = int(os.getenv('REDIS_PORT', 6379))
    redis_password = os.getenv('REDIS_PASSWORD', None)
    redis_db = int(os.getenv('REDIS_DB', 0))

    # MONGO ENV VARIABLES
    mongo_host = os.getenv('MONGO_HOST', 'localhost')
    mongo_port = int(os.getenv('MONGO_PORT', 27017))
    mongo_password = os.getenv('MONGO_PASSWORD', None)
    mongo_db = os.getenv('MONGO_DB', 'test')
    mongo_username = os.getenv('MONGO_USERNAME', '')

    # CONNECT TO REDIS
    logger.info('Initializing Redis...')
    redis = RedisClient(
        host=redis_host,
        port=redis_port,
        password=redis_password,
        db=redis_db
    )

    if not redis.client or redis.client is None:
        logger.error('Could not initialize Redis...')
        logger.error('Exiting...')
        exit(1)

    # CONNECT TO MONGO
    mongo = MongoClient(
        host=mongo_host,
        port=mongo_port,
        password=mongo_password,
        db=mongo_db,
        username=mongo_username
    )

    if not mongo.client or mongo.client is None:
        logger.error('Could not initialize Mongo...')
        logger.error('Exiting...')
        exit(1)

    # Define thresholds for batch operations
    WORD_OP_THRESHOLD = 1000

    # Initialize operation buffers
    word_operations = []

    # Function to perform bulk operations when thresholds are met
    def perform_bulk_operations():
        global word_operations
        
        logger.info('Word operations: %d', len(word_operations))

        if len(word_operations) >= WORD_OP_THRESHOLD:
            logger.info('Performing word bulk operations...')
            mongo.add_words_bulk(word_operations)
            word_operations = []

    # INDEXING LOOP
    while running:
        queue_size = redis.get_queue_size()
        if queue_size == 0:
            redis.signal_crawler()
            logger.info(f'RESUME_CRAWL signal sent')

        logger.info(f'Waiting for message queue...')
        
        # Get the next page from the queue
        # page_id = redis.peek_page()
        page_id = redis.pop_page()
        if not page_id:
            logger.error('Could not fetch data from indexer queue')
            continue

        # Fetch page data
        logger.info(f'Fetching {page_id}...')
        page = redis.get_page_data(page_id)
        if page is None:
            logger.warning(f'Could not fetch {page_id}. Skipping...')
            continue
        
        logger.info(f'Page url: {page.normalized_url}')
        logger.info(f'Page html: {page.html[:15] + "..." if len(page.html) > 15 else page.html}')

        normalized_url = page.normalized_url

        logger.info(f'Getting {page_id} metadata...')
        old_metadata = mongo.get_metadata(normalized_url)
        if old_metadata and old_metadata.last_crawled == page.last_crawled:
            logger.info(f'No updates to {old_metadata._id}. Skipping...')
            continue

        logger.info(f'Parsing html data for {page_id}...')
        html_data = get_html_data(page.html)
        if not html_data:
            logger.error(f'Could not parse html data for {page_id}. Skipping...')
            continue

        logger.info(f'Parsed html data for {page_id}...')
        # logger.info(f'HTML data: {html_data}')
        
        if html_data['language'] != 'en':
            logger.info(f'{page_id} not english. Skipping...')
            continue

        text = html_data['text']
        if not text:
            logger.error(f'Could not process text {page_id}. Skipping...')
            continue

        # Make a dictionary with the words in the text and their frequency
        logger.info(f'Counting words from {page_id}...')
        words_weight = Counter(text)

        # Get the top MAX_INDEX_WORDS words
        keywords = dict(words_weight.most_common(MAX_INDEX_WORDS))
        
        logger.info(f'Storing metadata for {page_id}...')
        # TODO: keywords should also contain the title, description and other meta tags

        logger.info('Saving words and images...')
        # Create or update word entries and image entries

        logger.info(f'Check words in url {normalized_url}...')

        # Iterate through the url name to add more weight to some words
        words_in_url = split_url(normalized_url)
        for word in words_in_url:
            past_score = keywords.get(word, 0)
            logger.info(f'Word in url: {word} - past score: {past_score}')

            # If a word in the url was already in our registry of words we multiply it
            if past_score != 0:
                new_score = past_score * 50
                keywords[word] = new_score
            else:
                # If the word is not in our registry we add it with a score of 100
                keywords[word] = 10

        # Iterate through the images and add them to the word operations
        for word, weight in keywords.items():
            word_op = mongo.add_url_to_word_operation(word, normalized_url, weight)
            word_operations.append(word_op)

        # Save the metdata
        mongo.save_metadata(page, html_data, keywords)

        # Check if any thresholds are exceeded and perform bulk operations
        perform_bulk_operations()

        # Store all the words
        wordsSet = {word.lower() for word in text}
        mongo.add_words_to_dictionary(wordsSet)

        logger.info('Delete page data from redis...')
        redis.delete_page_data(page_id)

        logger.info('Saving Outlinks...')
        outlinks = redis.get_outlinks(normalized_url)
        res = mongo.save_outlinks(outlinks)
        redis.delete_outlinks(normalized_url)

        logger.info('Pushing to image indexer queue...')
        redis.push_to_image_indexer_queue(normalized_url)

    logger.info('Shutting down...')
    # Ensure any remaining operations are saved before exit
    logger.info('Final bulk operations before exit...')
    # Save all remaining operations regardless of threshold
    if word_operations:
        mongo.add_words_bulk(word_operations)
    
    sys.exit(0)