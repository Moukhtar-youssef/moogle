import time
import logging
import pymongo

from typing import Optional, List
from models.page import Page
from models.metadata import Metadata
from models.image import Image
from models.outlinks import Outlinks

from pymongo import UpdateOne

# SETUP LOGGER
logger = logging.getLogger(__name__)
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# COLLECTIONS
IMAGE_COLLECTION = 'image'
METADATA_COLLECTION = 'metadata'
OUTLINKS_COLLECTION = 'outlinks'
WORD_COLLECTION = 'word'
WORD_IMAGES_COLLECTION = 'word_images'

class MongoClient:
    def __init__(self, host='localhost', port=27017, password="", db="test", username=""):
        try:
            self.client = pymongo.MongoClient(f'mongodb://{username}:{password}@{host}:{port}/{db}?authSource=admin')
            self.db = self.client[db]
            self.client.admin.command("ping")
            logger.info('Successfully connected to mongo!')
        except Exception as e:
            logger.error(f'Failed to connect to mongo')
            self.client = None


    def perform_batch_operations(self, operations: List[UpdateOne], collection_name: str):
        if self.client is None:
            logger.error(f'Mongo connection not initialized')
            return None

        res = self.db[collection_name].bulk_write(operations)
        return res

    # --------------------- METADATA ---------------------
    def get_metadata(self, normalized_url: str) -> Optional[Metadata]:
        if self.client is None:
            logger.error(f'Mongo connection not initialized')
            return None

        collection = self.db[METADATA_COLLECTION]
        result = collection.find_one(
            {"_id": normalized_url},
        )

        return Metadata.from_dict(result)

    def save_metadata(self, page_data: Page, html_data: Metadata) -> None:
        if self.client is None:
            logger.error(f'Mongo connection not initialized')
            return None

        # Create Metadata object
        normalized_url = page_data.normalized_url
        metadata = Metadata(
            _id=normalized_url,
            title=html_data['title'],
            description=html_data['description'],
            summary_text=html_data['summary_text'],
            last_crawled=page_data.last_crawled
        )

        # Save to mongo
        collection = self.db[METADATA_COLLECTION]

        result = collection.update_one(
            {"_id": normalized_url},
            {"$set": metadata.to_dict()},
            upsert=True
        )

    # --------------------- METADATA ---------------------

    # --------------------- IMAGE ---------------------
    def save_image(self, image: Image) -> int:
        if self.client is None:
            logger.error(f'Mongo connection not initialized')
            return None

        # Use collection
        collection = self.db[IMAGE_COLLECTION]

        # Save to mongo
        result = collection.update_one(
            {"_id": image._id},
            {"$set": image.to_dict()},
            upsert=True
        )

    # --------------------- IMAGE ---------------------

    # --------------------- WORD ---------------------
    def add_url_to_word(self, word: str, url: str, weight: int) -> None:
        if self.client is None:
            logger.error(f'Mongo connection not initialized')
            return None

        # Fetch Word from Mongo
        collection = self.db[WORD_COLLECTION]

        result = collection.update_one(
            {"_id": word},
            {
                "$push": {
                    "pages": {"url": url, "weight": weight}
                }
            },
            upsert=True
        )

    def add_url_to_word_operation(self, word: str, url: str, weight: int) -> UpdateOne:
        if self.client is None:
            logger.error(f'Mongo connection not initialized')
            return None

        return UpdateOne(
            {"_id": word},
            {
                "$push": {
                    "pages": {
                        "$each": [{"url": url, "weight": weight}],
                        "$sort": {"weight": -1}, # Sort by weight in descending order
                    }
                }
            },
            upsert=True
        )

    def add_words_bulk(self, operations: List[UpdateOne]):
        if not operations:
            return
        return self.perform_batch_operations(operations, 'word')

    # --------------------- WORD ---------------------

    # --------------------- WORD IMAGES ---------------------
    def add_url_to_word_images(self, word: str, image_url: str, weight: int) -> None:
        if self.client is None:
            logger.error(f'Mongo connection not initialized')
            return None

        # Fetch Word from Mongo
        collection = self.db[WORD_IMAGES_COLLECTION]

        result = collection.update_one(
            {"_id": word},
            {
                "$push": {
                    "pages": {"url": image_url, "weight": weight}
                }
            },
            upsert=True
        )


    def add_url_to_word_images_operation(self, word: str, image_url: str, weight: int) -> UpdateOne:
        if self.client is None:
            logger.error(f'Mongo connection not initialized')
            return None

        return UpdateOne(
            {"_id": word},
            {
                "$push": {
                    "pages": {
                        "$each": [{"url": image_url, "weight": weight}],
                        "$sort": {"weight": -1}, # Sort by weight in descending order
                    }
                }
            },
            upsert=True
        )


    def add_word_images_bulk(self, operations: List[UpdateOne]):
        if not operations:
            return
        return self.perform_batch_operations(operations, 'word_images')

    # --------------------- WORD IMAGES ---------------------

    # --------------------- OUTLINKS ---------------------
    def save_outlinks(self, outlinks: Outlinks) -> None:
        if self.client is None:
            logger.error(f'Mongo connection not initialized')
            return None

        if not outlinks:
            logger.error(f'Outlinks is None')
            return

        # Use collection
        collection = self.db[OUTLINKS_COLLECTION]

        result = collection.update_one(
            {"_id": outlinks._id},
            {"$set": outlinks.to_dict()},
            upsert=True
        )

        print(f'save_outlinks: {result}')
    # --------------------- OUTLINKS ---------------------
