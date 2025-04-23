import time
import logging
import pymongo

from typing import Optional, List, Set
from models.page import Page
from models.metadata import Metadata
from models.image import Image
from models.outlinks import Outlinks

from pymongo import UpdateOne

from collections import Counter

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

STOP_WORDS = [
    "the", "to", "is", "in", "for", "on", "and", "a", "an", "of", "with", "as", "by", "at", "this", "that", "their",
    "there", "it", "its", "they", "them", "he", "she", "we", "you", "your", "my", "me", "us", "our", "hers",
    "him", "his", "her", "them", "they're", "we're", "you're", "i'm", "it's", "that's", "who", "what", "where",
    "when", "why", "how", "which", "whom", "whose", "if", "then", "else", "but", "or", "not", "so", "than", "too",
    "very", "just", "only", "also", "such", "more", "most", "some", "any", "all", "each", "every", "few", "less",
    "least", "many", "much", "more", "most", "several", "both", "either", "neither", "one", "two", "three", "four",
    "a", "about", "above", "after", "again", "against", "all", "am", "an", "and", "any", "are", "aren't", "as", "at",
    "be", "because", "been", "before", "being", "below", "between", "both", "but", "by", "can't", "cannot", "could",
    "couldn't", "did", "didn't", "do", "does", "doesn't", "doing", "don't", "down", "during", "each", "few", "for",
    "from", "further", "had", "hadn't", "has", "hasn't", "have", "haven't", "having", "he", "he'd", "he'll", "he's",
    "her", "here", "here's", "hers", "herself", "him", "himself", "his", "how", "how's", "i", "i'd", "i'll", "i'm",
    "i've", "if", "in", "into", "is", "isn't", "it", "it's", "its", "itself", "let's", "me", "more", "most", "mustn't",
    "my", "myself", "no", "nor", "not", "of", "off", "on", "once", "only", "or", "other", "ought", "our", "ours",
    "ourselves", "out", "over", "own", "same", "shan't", "she", "she'd", "she'll", "she's", "should", "shouldn't", "so",
    "some"
    ]

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
        
        if not operations:
            logger.warning(f'No operations to perform')
            return None
        
        try:
            res = self.db[collection_name].bulk_write(operations, ordered=False)
            return res
        except Exception as e:
            logger.error(f'Error performing batch operations: {e}')
            return None

    # --------------------- METADATA ---------------------
    def get_keywords(self, mongo_id: str) -> Optional[List[str]]:
        if self.client is None:
            logger.error(f'Mongo connection not initialized')
            return None

        logger.info(f'Fetching keywords for {mongo_id}')
        collection = self.db[METADATA_COLLECTION]
        result = collection.find_one(
            {"_id": mongo_id},
            {"keywords": 1}
        )

        # Check if keywords exist
        if result is None or 'keywords' not in result:
            logger.error(f'No keywords found for {mongo_id}')
            logger.info(f'Fetching for the whole document')

            result = collection.find_one(
                {"_id": mongo_id}
            )
            if result is None:
                logger.error(f'No result found for {mongo_id}')
                return []
            
            # Total words
            total_words = []
            # Get summary text
            summary_text = result.get('summary_text', '')
            # convert the summary text to a list of words and remove stop words
            summary_text = summary_text.lower()
            summary_text = summary_text.split()
            summary_text = [word for word in summary_text if word not in STOP_WORDS]
            
            # Get description
            description = result.get('description', '') or ''
            # convert the description text to a list of words and remove stop words
            description = description.lower()
            description = description.split()
            description = [word for word in description if word not in STOP_WORDS]

            # Get title
            title = result.get('title', '')
            # convert the title text to a list of words and remove stop words
            title = title.lower()
            title = title.split()
            title = [word for word in title if word not in STOP_WORDS]
            
            # Add all words to the total words list
            total_words += summary_text
            total_words += description
            total_words += title
            
            # Count the words
            word_count = Counter(total_words)
            # Get the most common words
            most_common_words = dict(word_count.most_common(1000))

            return most_common_words
        
        # Extract keywords
        keywords = result['keywords']
        return keywords

    
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
    
    def save_image_operation(self, image: Image) -> UpdateOne:
        if self.client is None:
            logger.error(f'Mongo connection not initialized')
            return None

        return UpdateOne(
            {"_id": image._id},
            {"$set": image.to_dict()},
            upsert=True
        )

    def save_images_bulk(self, save_image_operations: List[UpdateOne]):
        if not save_image_operations:
            return
        return self.perform_batch_operations(save_image_operations, IMAGE_COLLECTION)

    # --------------------- IMAGE ---------------------

    # --------------------- WORD ---------------------
    def add_url_to_word(self, word: str, url: str, weight: int) -> None:
        if self.client is None:
            logger.error(f'Mongo connection not initialized')
            return None

        # Fetch Word from Mongo
        collection = self.db[WORD_COLLECTION]

        update_operation = self.add_url_to_word_operation(word, url, weight)
        if update_operation is None:
            return

        collection.bulk_write([update_operation])

    def add_url_to_word_operation(self, word: str, url: str, weight: int) -> UpdateOne:
        if self.client is None:
            logger.error(f'Mongo connection not initialized')
            return None

        return UpdateOne(
            {"_id": word},
            [
                {
                    "$set": {
                        "pages": {
                            "$cond": {
                                "if": {"$isArray": "$pages"},
                                "then": {
                                    "$concatArrays": [
                                        # Filter out any existing entry with this URL
                                        {"$filter": {
                                            "input": "$pages",
                                            "cond": {"$ne": ["$$this.url", url]}
                                        }},
                                        # Add the new entry
                                        # Here weight is actually tf-idf but it is not calculated yet
                                        [{"url": url, "tf": weight, "weight": 0}]
                                    ]
                                },
                                # Here weight is actually tf-idf but it is not calculated yet
                                # I should change the names
                                "else": [{"url": url, "tf": weight, "weight": 0}]
                            }
                        }
                    }
                },
                # Sort the pages array by tf in descending order
                {
                    "$set": {
                        "pages": {
                            "$sortArray": {
                                "input": "$pages",
                                "sortBy": {"tf": -1}
                            }
                        }
                    }
                }
            ],
            upsert=True
        )

    def add_words_bulk(self, operations: List[UpdateOne]):
        if not operations:
            return
        return self.perform_batch_operations(operations, 'word')

    def add_words_to_dictionary(self, words: Set[str]):
        if self.client is None:
            logger.error(f'Mongo connection not initialized')
            return None
        collection = self.db["dictionary"]  # Assuming a dictionary collection
        try:
            operations = []
            for word in words:
                operations.append(UpdateOne(
                    {"_id": word},
                    {"$set": {"_id": word}},
                    upsert=True
                ))
            if operations:
                collection.bulk_write(operations)
        except Exception as e:
            logger.error(f"Error adding words to dictionary: {e}")
            return None

    # --------------------- WORD ---------------------

    # --------------------- WORD IMAGES ---------------------
    def add_url_to_word_images(self, word: str, image_url: str, weight: int) -> None:
        if self.client is None:
            logger.error(f'Mongo connection not initialized')
            return None

        # Fetch Word from Mongo
        collection = self.db[WORD_IMAGES_COLLECTION]

        update_operation = self.add_url_to_word_images_operation(word, image_url, weight)
        if update_operation is None:
            return

        collection.bulk_write([update_operation])

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

        return result
    # --------------------- OUTLINKS ---------------------
