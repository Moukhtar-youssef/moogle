import logging
import signal
import sys
import os

from utils.constants import *
from models.image import Image
from data.redis_client import RedisClient
from data.mongo_client import MongoClient
from utils.utils import is_valid_image, split_name

import time
import os.path

from concurrent.futures import ThreadPoolExecutor

# SETUP LOGGER
logger = logging.getLogger(__name__)
logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)

# SHUTDOWN
running = True


def handle_exit(signum, frame):
    global running
    logger.info("Termination signal received - shutting down...")
    running = False
    sys.exit(0)


signal.signal(signal.SIGTERM, handle_exit)
signal.signal(signal.SIGINT, handle_exit)

if __name__ == "__main__":
    # REDIS ENV VARIABLES
    redis_host = os.getenv("REDIS_HOST", "localhost")
    redis_port = int(os.getenv("REDIS_PORT", 6379))
    redis_password = os.getenv("REDIS_PASSWORD", None)
    redis_db = int(os.getenv("REDIS_DB", 0))

    # MONGO ENV VARIABLES
    mongo_host = os.getenv("MONGO_HOST", "localhost")
    mongo_port = int(os.getenv("MONGO_PORT", 27017))
    mongo_password = os.getenv("MONGO_PASSWORD", None)
    mongo_db = os.getenv("MONGO_DB", "test")
    mongo_username = os.getenv("MONGO_USERNAME", "")

    # CONNECT TO REDIS
    logger.info("Initializing Redis...")
    redis = RedisClient(
        host=redis_host, port=redis_port, password=redis_password, db=redis_db
    )

    if not redis.client or redis.client is None:
        logger.error("Could not initialize Redis...")
        logger.error("Exiting...")
        exit(1)

    # CONNECT TO MONGO
    mongo = MongoClient(
        host=mongo_host,
        port=mongo_port,
        password=mongo_password,
        db=mongo_db,
        username=mongo_username,
    )

    if not mongo.client or mongo.client is None:
        logger.error("Could not initialize Mongo...")
        logger.error("Exiting...")
        exit(1)

    # Define thresholds for batch operations
    WORD_IMAGE_OP_THRESHOLD = 500
    SAVE_IMAGE_OP_THRESHOLD = 100  # Why did I set this to 10000 :( that was too much, and it was never going to reach it

    # Initialize operation buffers
    word_images_operations = []
    save_image_operations = []

    # Function to perform bulk operations when thresholds are met
    def perform_bulk_operations():
        global word_images_operations, save_image_operations

        logger.info("Word images operations: %d", len(word_images_operations))
        logger.info("Save image operations: %d", len(save_image_operations))

        if len(word_images_operations) >= WORD_IMAGE_OP_THRESHOLD:
            logger.info("Performing word images bulk operations...")
            mongo.add_word_images_bulk(word_images_operations)
            word_images_operations = []

        if len(save_image_operations) >= SAVE_IMAGE_OP_THRESHOLD:
            logger.info("Performing save image bulk operations...")
            mongo.save_images_bulk(save_image_operations)
            save_image_operations = []

    # INDEXING LOOP
    while running:
        logger.info(f"Checking for message queue...")

        # Get the next image from the queue
        page_id = redis.pop_image()
        # page_id = redis.peek_page()
        if not page_id:
            logger.error("Could not fetch data from indexer queue")
            continue

        logger.info(f"Got {page_id} from queue...")

        # Fetch keywords from Mongo
        mongo_id = page_id.split("page_images:")[-1]
        keywords = mongo.get_keywords(mongo_id)

        logger.info(f"Got {len(keywords)} keywords from Mongo...")

        # Get the members of the page_id
        page_images = redis.get_page_images(page_id)
        if page_images is None:
            logger.warning(f"Could not fetch {page_id} images. Skipping...")
            continue

        if len(page_images) == 0:
            logger.warning(f"No images found for {page_id}. Skipping...")
            # continue
        logger.info(f"Got {len(page_images)} images from Redis...")

        # Check if the urls are valid using workers
        logger.info("Processing images...")

        def process_image_url(image_url):
            if is_valid_image(image_url):
                # Check file extension
                _, file_extension = os.path.splitext(image_url.split("/")[-1])
                file_extension = file_extension.lstrip(".")

                # Skip SVG files and icons
                if file_extension == "svg" or "icons" in image_url:
                    # Delete image from Redis
                    redis.delete_image_data(image_url)
                    return None

                # Fetch image data from redis
                image_data = redis.pop_image_data(image_url)
                if not image_data:
                    logger.error(f"Could not fetch image {image_url} from Redis")
                    return None

                # Update image data with filename
                image_data.filename = image_url.split("/")[-1]

                # Extract words from the image name
                filename_words_op = []
                words_in_filename = split_name(image_data.filename)
                for word in words_in_filename:
                    if word in keywords:
                        # Get the score of the word
                        past_score = keywords[word]
                        new_score = past_score * 100
                    else:
                        new_score = 30

                    # Create operation to add image URL to word images
                    image_op = mongo.add_url_to_word_images_operation(
                        word, image_url, new_score
                    )
                    filename_words_op.append(image_op)

                # Add save image operation to the buffer
                save_image_op = mongo.save_image_operation(image_data)
                return (image_url, save_image_op, filename_words_op)

            # If the image is not valid, delete it from Redis
            redis.delete_image_data(image_url)
            return None

        with ThreadPoolExecutor() as executor:
            # First parallel operation - processing image URLs
            results = list(executor.map(process_image_url, page_images))
            valid_results = [result for result in results if result is not None]

            # Unpack the results into image URLs and save image operations
            if valid_results:
                images_urls = []
                save_image_ops = []
                filename_words_ops_all = []

                for url, save_op, word_ops in valid_results:
                    images_urls.append(url)
                    save_image_ops.append(save_op)
                    filename_words_ops_all.extend(word_ops)

                save_image_operations.extend(save_image_ops)
                word_images_operations.extend(filename_words_ops_all)
            else:
                images_urls = []
            # images_urls = [url for url in results if url is not None]

            logger.info(f"Got {len(images_urls)} valid images from Redis...")

            # Second parallel operation - processing word-image operations
            # Create a list of all the (word, image_url, weight) combinations
            operations = [
                (word, image_url, weight)
                for word, weight in keywords.items()
                for image_url in images_urls
            ]

            # Define a function to process each operation
            def process_word_image(operation):
                word, image_url, weight = operation

                return mongo.add_url_to_word_images_operation(word, image_url, weight)

            # Execute the operations in parallel and wait for completion
            keyword_ops = list(executor.map(process_word_image, operations))
            word_images_operations.extend(keyword_ops)

        # Check if any thresholds are exceeded and perform bulk operations
        perform_bulk_operations()

        # Delete page_images
        # It's stupid to pass the mongo_id here but lol
        redis.delete_page_images(mongo_id)

    logger.info("Shutting down...")
    # Ensure any remaining operations are saved before exit
    logger.info("Final bulk operations before exit...")
    # Save all remaining operations regardless of threshold
    if word_images_operations:
        mongo.add_word_images_bulk(word_images_operations)
    if save_image_operations:
        mongo.save_images_bulk(save_image_operations)

    sys.exit(0)
