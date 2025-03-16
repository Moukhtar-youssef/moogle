import logging
import os

from models.page import Page
from models.image import Image
from data.redis_client import RedisClient
from utils.utils import get_html_data
import time

logger = logging.getLogger(__name__)

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

MESSAGE_QUEUE = 'indexer_queue'

file_types = ["png", "svg", "ico", "gif", "jpeg", "jpg"]

import re
def split_name(filename: str):
    pattern = r'[-_./\s]+'
    parts = re.split(pattern, filename)
    parts = [part.lower() for part in parts if part and part.lower() not in file_types and "px" not in part.lower()]
    return parts

if __name__ == "__main__":
    redis_host = os.getenv('REDIS_HOST', 'localhost')
    redis_port = int(os.getenv('REDIS_PORT', 6379))
    redis_password = os.getenv('REDIS_PASSWORD', None)
    redis_db = int(os.getenv('REDIS_DB', 0))

    # Connect to redis
    logger.info('Initializing Redis...')
    redis = RedisClient(host=redis_host, port=redis_port, password=redis_password, db=redis_db)

    print(f'Redis: {redis.client}')

    if not redis.client or redis.client is None:
        logger.error('Could not initialize Redis...')
        logger.error('Exiting...')

        exit(1)


    while(True):
        # Wait until we get an entry from the message queue
        logger.info(f'Waiting for message queue...')
        popped = redis.client.brpop(MESSAGE_QUEUE)
        if not popped:
            loger.error(f'Could not fetch from message queue')
            continue

        _, page_id = popped

        # Fetch page data
        logger.info(f'Fetching {page_id}...')
        page = redis.get_page(page_id)
        if page is None:
            logger.warning(f'Could not fetch {page_id}. Skipping...')
            continue

        # Check if the page has been crawled before
        old_metadata_id = f'url_metadata:{page.normalized_url}'
        old_metadata = redis.get_metadata(old_metadata_id)
        if old_metadata and old_metadata.last_crawled == page.last_crawled:
            logger.info(f'No updates to {old_metadata_id}. Skipping...')
            continue

        logger.info(f'Parsing html data for {page_id}...')
        # Extract html data from page
        html_data = get_html_data(page.html)
        if not html_data:
            logger.error(f'Could not parse html data for {page_id}. Skipping...')
            continue

        # TODO: This may fail sometimes
        if html_data['language'] != 'en':
            logger.info(f'{page_id} not english. Skipping...')
            continue

        logger.info(f'Storing metadata for {page_id} in Redis...')
        # Store the metadata in a separate hash
        redis.save_metadata(page, html_data)

        text = html_data['text']
        if not text:
            logger.error(f'Could not process text {page_id}. Skipping...')
            continue

        logger.info(f'Storing words from {page_id} in Redis...')
        # Make a dictionary with the words in the text and their frequency
        words_weight = {}

        for word in text:
            words_weight[word] = words_weight.get(word, 0) + 1

        # Process images
        page_images_id = f'page_images:{page.normalized_url}'
        images_urls = redis.get_page_images_urls(page_images_id)

        # Create or update word entries and image entries
        for word, weight in words_weight.items():
            redis.save_word(word, page.normalized_url, weight)

            for image_url in images_urls:
                file_extension = image_url.split('/')[-1].split('.')[-1]
                if file_extension == "svg" or "icons" in image_url:
                    continue

                # Save the word entry
                redis.save_word_images(word, image_url, weight)

        # Update image hash to include file_name
        # and update word_image to include entries with the words in the file name
        for image_url in images_urls:
            # Ignore svg files since they are usually just graphics
            image_name = image_url.split('/')[-1]
            file_extension = image_name.split('.')[-1]
            if file_extension == "svg" or "icons" in image_url:
                continue

            # Fetch image data
            image_data = redis.get_image(f'image:{image_url}')
            # Update image data with file_name
            image_data.file_name = image_name
            redis.update_image(f'image:{image_url}', image_data)

            # Get words in image_name
            words_in_filename = split_name(image_name)

            # Get words from alt
            words_in_alt = split_name(image_data.alt)

            # Combine them
            total_words = words_in_filename + words_in_alt

            # Iterate through all the words in the image_name and the alt
            for word in total_words:
                word_images_key = f'word_images:{word}'
                results = redis.get_images_from_word(word_images_key)

                # Check if there is already an entry for this world and url
                is_present = any(url == image_url for url, score in results)
                if is_present:
                    # If there is one, multiply it by 100 or something idk
                    current_score = next(score for url, score in results if url == image_url)
                    new_score = current_score * 100
                    new_score = min(new_score, 10000000)

                    # Update word entry
                    redis.save_word_images(word, image_url, new_score)
                else:
                    # If it's not present, just add it with a weight of 10 or something
                    redis.save_word_images(word, image_url, 30)

