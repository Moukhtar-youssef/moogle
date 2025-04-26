import logging
import signal
import math
import time
import sys
import os
import threading
from queue import Queue
from data.mongo_client import MongoClient

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
    # Perform final bulk operations regardless of the threshold
    logger.info("Performing final bulk operations...")
    mongo.perform_batch_operations(bulk_operations)
    sys.exit(0)


signal.signal(signal.SIGTERM, handle_exit)
signal.signal(signal.SIGINT, handle_exit)


# Thread function to process words
def process_words(thread_id, word_queue, total_number_of_documents):
    global bulk_operations
    processed_count = 0

    while running:
        try:
            # Get a word from the queue
            word_item = word_queue.get(block=False)
            if word_item is None:
                break

            word = word_item["word"]
            word_document_count = mongo.get_word_document_count(word)
            word_entries = mongo.get_word_documents(word)

            # Calculate IDF for this word
            idf = math.log10(total_number_of_documents / (1 + word_document_count))
            logger.info(
                f'Thread-{thread_id}: Calculating TF-IDF for "{word}" (appears in {word_document_count} documents)'
            )

            inner_count = 0
            local_ops = []
            for word_entry in word_entries:
                url = word_entry["url"]
                tf = word_entry["tf"]
                tfidf = tf * idf
                op = mongo.update_page_tfidf_op(word, url, idf, tfidf)
                local_ops.append(op)
                inner_count += 1

            # Add operations to bulk list with lock
            with operations_lock:
                bulk_operations.extend(local_ops)
                # Check if we need to perform bulk operations
                if len(bulk_operations) > OPERATIONS_THRESHOLD:
                    logger.info(f"Thread-{thread_id}: Performing bulk operations...")
                    mongo.update_page_tfidf_bulk(bulk_operations)
                    bulk_operations = []

            logger.info(
                f"Thread-{thread_id}: Processed {inner_count} entries for word: {word}"
            )
            processed_count += 1

            # Mark task as done
            word_queue.task_done()

        except Exception as e:
            if "Empty" in str(e):
                break
            logger.error(f"Thread-{thread_id} error: {e}")
            break

    logger.info(f"Thread-{thread_id}: Processed {processed_count} words total")


if __name__ == "__main__":
    # MONGO ENV VARIABLES
    mongo_host = os.getenv("MONGO_HOST", "localhost")
    mongo_port = int(os.getenv("MONGO_PORT", 27017))
    mongo_password = os.getenv("MONGO_PASSWORD", None)
    mongo_db = os.getenv("MONGO_DB", "test")
    mongo_username = os.getenv("MONGO_USERNAME", "")

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

    # NUMBER OF OPERATIONS BEFORE WRITING TO MONGO
    OPERATIONS_THRESHOLD = int(os.getenv("OPERATIONS_THRESHOLD", 1000))
    NUM_THREADS = int(os.getenv("NUM_THREADS", 4))  # Number of worker threads

    bulk_operations = []
    operations_lock = (
        threading.Lock()
    )  # Lock for synchronizing access to bulk_operations

    # Get number of pages in the database
    logger.info("Fetching number of documents...")
    total_number_of_documents = mongo.get_document_count()
    if total_number_of_documents is None:
        logger.error("No documents found - sleeping...")
        time.sleep(10000)

    logger.info(f"{total_number_of_documents} documents found!")

    logger.info("Get unique words...")
    unique_words = mongo.get_unique_words()
    if unique_words is None:
        logger.error("No word entries found - sleeping...")
        time.sleep(10000)

    # Create a queue to hold all words
    word_queue = Queue()

    # Put all words in the queue
    for item in unique_words:
        word_queue.put(item)

    initial_queue_size = (
        word_queue.qsize() if hasattr(word_queue, "qsize") else "unknown"
    )
    logger.info(
        f"Starting processing of {initial_queue_size} unique words with {NUM_THREADS} threads"
    )

    # Create and start worker threads
    threads = []
    for i in range(NUM_THREADS):
        t = threading.Thread(
            target=process_words, args=(i + 1, word_queue, total_number_of_documents)
        )
        t.daemon = True
        threads.append(t)
        t.start()

    # Wait for all threads to complete
    for t in threads:
        t.join()

    # Perform any remaining bulk operations
    logger.info("Performing final bulk operations...")
    if bulk_operations:
        mongo.update_page_tfidf_bulk(bulk_operations)

    logger.info("TF-IDF processing complete")
    logger.info("Shutting down...")
    sys.exit(0)
