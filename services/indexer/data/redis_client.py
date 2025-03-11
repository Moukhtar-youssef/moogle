import redis
import logging
from models.page import Page
from models.metadata import Metadata
from email.utils import format_datetime
from datetime import datetime
import time

logger = logging.getLogger(__name__)

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

class RedisClient:
    def __init__(self, host='localhost', port=6379, decode_responses=True):
        try:
            self.client = redis.Redis(
                host=host,
                port=port,
                decode_responses=decode_responses
            )

            self.client.ping()
            logger.info('Successfully connected to redis!')
        except Exception as e:
            logger.error(f'Failed to connect to redis')
            self.client = None

    def get_page(self, key: str) -> Page:
        if self.client is None:
            logger.error(f'Redis connection not initialized')
            return None

        try:
            page_hashed = self.client.hgetall(key)

            if not page_hashed:
                logger.warning(f'Page with key {key} not found in Redis')
                return None

            logger.info(f'Page with key {key} successfully fetched')
            return Page.from_hash(page_hashed)
        except Exception as e:
            logger.error(f'Unexpected error while fetching {key}: {e}')
            return None

    def get_metadata(self, key: str) -> Metadata:
        if self.client is None:
            logger.error(f'Redis connection not initialized')
            return None

        metadata_hashed = self.client.hgetall(key)

        if not metadata_hashed:
            logger.warning(f'Metadata with key {key} not found in Redis')
            return None

        logger.info(f'Metadata with key {key} successfully fetched')
        return Metadata.from_hash(metadata_hashed)

    def save_metadata(self, page_data: Page, html_data: Metadata) -> None:
        url = page_data.normalized_url
        title = html_data['title']
        description = html_data['description']
        text = html_data['summary_text']
        last_crawled = page_data.last_crawled.strftime("%a, %d %b %Y %H:%M:%S ") + time.tzname[0]

        if self.client is None:
            logger.error(f'Redis connection not initialized')
            return None

        key = f'url_metadata:{url}'

        metadata = {
            'title': title if title else '',
            'description': description if description else '',
            'summary_text': text if text else '',
            'last_crawled': last_crawled if last_crawled else '',
            'normalized_url': url
        }

        self.client.hset(key, mapping=metadata)

        logger.info(f'Metadata with key {key} successfully saved')

    def save_word(self, word: str, url: str, weight: int) -> None:
        if self.client is None:
            logger.error(f'Redis connection not initialized')
            return None

        # TODO: Read more about zadd
        key = f'word:{word}'
        self.client.zadd(key, {url: weight})



