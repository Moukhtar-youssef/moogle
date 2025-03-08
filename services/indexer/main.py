import logging

from models.page import Page
from data.redis_client import RedisClient
from utils.utils import get_html_data

logger = logging.getLogger(__name__)

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

MESSAGE_QUEUE = 'indexer_queue'

if __name__ == "__main__":

    # Connect to redis
    logger.info('Initializing Redis...')
    redis = RedisClient()

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

        # Sort the dictionary just to debug
        # words_weight = dict(sorted(words_weight.items(), key=lambda item: item[1], reverse=True))

        # Create or update word entries
        for word, weight in words_weight.items():
            redis.save_word(word, page.normalized_url, weight)

