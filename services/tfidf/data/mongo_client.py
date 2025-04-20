import time
import logging
import pymongo

from typing import Optional, List, Set
from pymongo import UpdateOne

# SETUP LOGGER
logger = logging.getLogger(__name__)
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# COLLECTIONS
METADATA_COLLECTION = 'metadata'
WORD_COLLECTION = 'word'

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


    def perform_batch_operations(self, operations: List[UpdateOne]):
        if self.client is None:
            logger.error(f'Mongo connection not initialized')
            return None

        res = self.db[WORD_COLLECTION].bulk_write(operations)
        return res

    # --------------------- METADATA ---------------------
    def get_document_count(self) -> Optional[int]:
        if self.client is None:
            logger.error(f'Mongo connection not initialized')
            return None

        collection = self.db[METADATA_COLLECTION]
        result = collection.count_documents({})

        return result
    # --------------------- METADATA ---------------------

    # --------------------- WORD ---------------------
    def get_all_word_entries(self) -> Optional[List]:
        if self.client is None:
            logger.error(f'Mongo connection not initialized')
            return None

        collection = self.db[WORD_COLLECTION]

        result = collection.find({})

        return result

    def update_page_tfidf(self, word: str, url: str, idf: float, tfidf: float):
        if self.client is None:
            logger.error(f'Mongo connection not initialized')
            return None

        collection = self.db[WORD_COLLECTION]

        try:
            result = collection.update_one(
                {
                    "_id": word,
                    "pages.url": url
                },
                {
                    "$set": {
                        "pages.$.idf": idf,
                        "pages.$.weight": tfidf
                    }
                }
            )
            if result.matched_count == 0:
                logger.warning(f"No match found for word '{word}' and url '{url}'")
        except Exception as e:
            logger.error(f"Error updating IDF/TFIDF for word '{word}' and url '{url}': {e}")


    def update_page_tfidf_op(self, word: str, url: str, idf: float, tfidf: float) -> UpdateOne:
        if self.client is None:
            logger.error(f'Mongo connection not initialized')
            return None

        return UpdateOne(
            {"_id": word, "pages.url": url},
            {"$set": {
                "pages.$.idf": idf,
                "pages.$.weight": tfidf
            }}
        )
    # --------------------- WORD ---------------------
