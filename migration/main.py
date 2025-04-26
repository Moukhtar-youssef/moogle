import logging
import signal
import sys
import os

from data.mongo_client import MongoClient

# SETUP LOGGER
logger = logging.getLogger(__name__)
logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)

# SHUTDOWN
running = True


# FIX: Remove this later
from dotenv import load_dotenv

load_dotenv("variables.env")  # Add this near the top of your file


def handle_exit(signum, frame):
    global running
    logger.info("Termination signal received - shutting down...")
    running = False

    # Perform batch operations to add the new word entries to the new database
    logger.info("Performing final bulk operations...")
    mongo.create_words_bulk(create_words_entry_operations)
    mongo.create_word_images_bulk(create_word_images_entry_operations)

    logger.info("Shutting down...")

    sys.exit(0)


signal.signal(signal.SIGTERM, handle_exit)
signal.signal(signal.SIGINT, handle_exit)


OPERATIONS_THRESHOLD = 10000


# Function to perform bulk operations when thresholds are met
def perform_bulk_operations():
    global create_words_entry_operations, create_word_images_entry_operations

    if len(create_words_entry_operations) >= OPERATIONS_THRESHOLD:
        logger.info("Performing words bulk operations...")
        mongo.create_words_bulk(create_words_entry_operations)
        create_words_entry_operations = []

    if len(create_word_images_entry_operations) >= OPERATIONS_THRESHOLD:
        logger.info("Performing word images bulk operations...")
        mongo.create_word_images_bulk(create_word_images_entry_operations)
        create_word_images_entry_operations = []


if __name__ == "__main__":
    # MONGO ENV VARIABLES
    mongo_host = os.getenv("MONGO_HOST", "localhost")
    mongo_password = os.getenv("MONGO_PASSWORD", "")
    mongo_db = os.getenv("MONGO_DB", "test")
    mongo_username = os.getenv("MONGO_USERNAME", "")

    # CONNECT TO MONGO
    mongo = MongoClient(
        host=mongo_host,
        password=mongo_password,
        db=mongo_db,
        username=mongo_username,
    )

    if not mongo.client or mongo.client is None:
        logger.error("Could not initialize Mongo...")
        logger.error("Exiting...")
        exit(1)

    logger.info("Mongo initialized successfully!")

    logger.info("Starting migration...")
    create_words_entry_operations = []
    create_word_images_entry_operations = []

    logger.info("Migrating words...")
    # Get all words from the old database (using a cursor to avoid loading everything into memory)
    words_entries = mongo.get_words_entries()
    count = 0
    for word_entry in words_entries:
        # Iterate through all words and migrate them to the new format
        count += 1
        word = word_entry["_id"]
        pages = word_entry["pages"]

        logger.info(f"\tProcessing word: {word} with {len(pages)} pages.")

        for page in pages:
            try:
                # Check if the keys "tf" and "weight" exist in the page dictionary
                url = page["url"] if "url" in page else ""
                tf = page["tf"] if "tf" in page else 0
                weight = page["weight"] if "weight" in page else 0

                op = mongo.create_words_entry_operation(word, url, tf, weight)
                create_words_entry_operations.append(op)
            except KeyError as e:
                logger.error(f"KeyError: {e} - {word} - {page}")
                sys.exit(1)

            perform_bulk_operations()

    logger.info(f"Processed {count} words.")

    logger.info(f"Migrating word images...")
    # Get all images from the old database (using a cursor to avoid loading everything into memory)
    word_images_entries = mongo.get_word_images_entries()
    count = 0
    for word_image_entry in word_images_entries:
        count += 1
        word = word_image_entry["_id"]
        pages = word_image_entry["pages"]

        logger.info(f"\tProcessing word image: {word} with {len(pages)} pages.")
        for page in pages:
            try:
                # Check if the keys "tf" and "weight" exist in the page dictionary
                url = page["url"] if "url" in page else ""
                weight = page["weight"] if "weight" in page else 0

                op = mongo.create_word_images_entry_operation(word, url, weight)
                create_word_images_entry_operations.append(op)
            except KeyError as e:
                logger.error(f"KeyError: {e} - {word} - {page}")
                sys.exit(1)

            perform_bulk_operations()

    # Perform batch operations to add the new word entries to the new database
    logger.info("Performing final bulk operations...")
    perform_bulk_operations()

    logger.info("Shutting down...")

    sys.exit(0)
