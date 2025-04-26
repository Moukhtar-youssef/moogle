import logging
from typing import List
import pymongo

from pymongo import UpdateOne

# SETUP LOGGER
logger = logging.getLogger(__name__)
logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)

# COLLECTIONS
WORD_COLLECTION = "word"
WORD_IMAGES_COLLECTION = "word_images"

# NEW DB COLLECTIONS
NEW_WORDS_COLLECTION = "words"
NEW_WORD_IMAGES_COLLECTION = "word_images"


class MongoClient:
    def __init__(
        self, host="localhost", port=27017, password="", db="test", username=""
    ):
        try:
            self.client = pymongo.MongoClient(
                f"mongodb://{username}:{password}@{host}:{port}/{db}?authSource=admin"
            )
            self.db = self.client[db]
            self.new_db = self.client["flattened"]
            if self.db is None:
                logger.error(f"Mongo connection not initialized")
                return None

            logger.info("Creating indexes...")
            words = self.new_db[NEW_WORDS_COLLECTION]
            # Create a compound index to ensure uniqueness on word and url
            words.create_index([("word", 1), ("url", 1)], unique=True)
            # Create a compound index to easily sort by word and weight
            words.create_index([("word", 1), ("weight", -1)])

            # Create single field indexes
            words.create_index("word")
            words.create_index("url")

            word_images = self.new_db[NEW_WORD_IMAGES_COLLECTION]
            word_images.create_index([("word", 1), ("url", 1)], unique=True)
            word_images.create_index("word")
            word_images.create_index("url")

            self.client.admin.command("ping")
            logger.info("Successfully connected to mongo!")
        except Exception as e:
            logger.error(f"Failed to connect to mongo: {e}")
            self.client = None

    def perform_batch_operations(
        self, operations: List[UpdateOne], collection_name: str
    ):
        if self.client is None:
            logger.error(f"Mongo connection not initialized")
            return None

        if not operations:
            logger.warning(f"No operations to perform")
            return None

        try:
            res = self.new_db[collection_name].bulk_write(operations, ordered=False)
            return res
        except Exception as e:
            logger.error(f"Error performing batch operations: {e}")
            return None

    # --------------------- WORDS ---------------------
    def get_words_entries(self):
        if self.client is None:
            logger.error(f"Mongo connection not initialized")
            return None

        # Fetch Word from Mongo
        collection = self.db[WORD_COLLECTION]

        words_cursor = collection.find({}, batch_size=1000)
        if words_cursor is None:
            logger.warning(f"Word entries not found in MongoDB")
            return None

        return words_cursor

    def create_words_entry_operation(
        self, word: str, url: str, tf: int, weight: int
    ) -> None:
        if self.client is None:
            logger.error(f"Mongo connection not initialized")
            return None

        # Write the new word entry to the new database
        return UpdateOne(
            {"word": word, "url": url},
            {
                "$set": {
                    "tf": tf,
                    "weight": weight,
                }
            },
            upsert=True,
        )

    def create_words_bulk(self, operations: List[UpdateOne]):
        if not operations:
            return
        return self.perform_batch_operations(operations, NEW_WORDS_COLLECTION)

    # --------------------- WORDS ---------------------

    # --------------------- WORD IMAGES ---------------------
    def get_word_images_entries(self):
        if self.client is None:
            logger.error(f"Mongo connection not initialized")
            return None

        # Fetch Word from Mongo
        collection = self.db[WORD_IMAGES_COLLECTION]

        word_images_cursor = collection.find({}, batch_size=1000)
        if word_images_cursor is None:
            logger.warning(f"Word entries not found in MongoDB")
            return None

        return word_images_cursor

    def create_word_images_entry_operation(
        self, word: str, url: str, weight: int
    ) -> None:
        if self.client is None:
            logger.error(f"Mongo connection not initialized")
            return None

        # Write the new word entry to the new database
        return UpdateOne(
            {"word": word, "url": url},
            {
                "$set": {
                    "weight": weight,
                }
            },
            upsert=True,
        )

    def create_word_images_bulk(self, operations: List[UpdateOne]):
        if not operations:
            return
        return self.perform_batch_operations(operations, NEW_WORD_IMAGES_COLLECTION)

    # --------------------- WORD IMAGES ---------------------

    # --------------------- TESTING ---------------------#
    def get_word_entry(self, word: str):
        if self.client is None:
            logger.error(f"Mongo connection not initialized")
            return None

        # Fetch Word from Mongo
        collection = self.db[WORD_COLLECTION]

        word_entry = collection.find_one({"_id": word})
        if word_entry is None:
            logger.warning(f"Word {word} not found in MongoDB")
            return None

        return word_entry

    # --------------------- TESTING ---------------------#
