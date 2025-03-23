import logging
import signal
import sys
import os

from utils.constants import *
from models.image import Image
from data.redis_client import RedisClient
from data.mongo_client import MongoClient
from utils.utils import get_html_data, split_name

import time

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

    # INDEXING LOOP
    while running:
        queue_size = redis.get_queue_size()
        if queue_size == 0:
            redis.signal_crawler()
            logger.info(f'RESUME_CRAWL signal sent')

        logger.info(f'Waiting for message queue...')
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

        # TODO: This may fail sometimes
        if html_data['language'] != 'en':
            logger.info(f'{page_id} not english. Skipping...')
            continue

        logger.info(f'Storing metadata for {page_id}...')
        mongo.save_metadata(page, html_data)

        text = html_data['text']
        if not text:
            logger.error(f'Could not process text {page_id}. Skipping...')
            continue

        # Make a dictionary with the words in the text and their frequency
        logger.info(f'Storing words from {page_id}...')
        words_weight = {}
        for word in text:
            words_weight[word] = words_weight.get(word, 0) + 1

        # Process images
        images_urls = redis.get_page_images(normalized_url)

        # What if I just save the first 100 words?
        top_words = dict(sorted(words_weight.items(), key=lambda item: item[1], reverse=True)[:MAX_INDEX_WORDS])

        word_operations = []
        word_images_operations = []

        # Create or update word entries and image entries
        for word, weight in top_words.items():
            print(f'adding word to mongo: {word} | {weight}')
            # Save the word image entry to a batch of operations
            word_op = mongo.add_url_to_word_operation(word, normalized_url, weight)
            word_operations.append(word_op)

            for image_url in images_urls:
                file_extension = image_url.split('/')[-1].split('.')[-1]
                if file_extension == "svg" or "icons" in image_url:
                    continue

                # Save the word image entry to a batch of operations
                word_image_op = mongo.add_url_to_word_images_operation(word, image_url, weight)
                word_images_operations.append(word_image_op)

        mongo.add_words_bulk(word_operations)
        mongo.add_word_images_bulk(word_images_operations)

        # I could put this loop inside the other one but it would look too bloated
        # Update image hash to include file_name
        # and update word_image to include entries with the words in the file name
        for image_url in images_urls:
            # Ignore svg files since they are usually just graphics
            image_name = image_url.split('/')[-1]
            file_extension = image_name.split('.')[-1]
            if file_extension == "svg" or "icons" in image_url:
                continue

            # Fetch image data from redis
            image_data = redis.pop_image(image_url)

            if not image_data:
                continue

            # Update image data with file_name
            image_data.file_name = image_name

            # Save image to mongo
            mongo.save_image(image_data)

            # Get words in image_name
            words_in_filename = split_name(image_name)

            # Get words from alt
            words_in_alt = split_name(image_data.alt)

            # Combine them
            total_words = words_in_filename + words_in_alt

            # print(f'Iterating through words: {total_words}')
            # This is a stupid algorithm but whatever
            # Iterate through all the words in the image_name and the alt
            for word in total_words:
                # Get the score of the word
                past_score = words_weight.get(word, 0)

                if past_score == 0:
                    new_score = 30
                else:
                    new_score = past_score * 100

                # print(f'Saving {image_name} with score: {new_score}')
                mongo.add_url_to_word_images(word, image_url, new_score)

        # Delete page_data from redis
        redis.delete_page_data(page_id)
        # Delete page_images
        redis.delete_page_images(normalized_url)

        # Save outlinks to mongo
        outlinks = redis.get_outlinks(normalized_url)
        mongo.save_outlinks(outlinks)
        # Delete outlinks from redis
        redis.delete_outlinks(normalized_url)

        time.sleep(0.5)

    logger.info('Shutting down...')
    sys.exit(0)

