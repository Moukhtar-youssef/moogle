import logging
import signal
import math
import time
import sys
import os

from data.mongo_client import MongoClient

# SETUP LOGGER
logger = logging.getLogger(__name__)
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# SHUTDOWN
running = True

def handle_exit(signum, frame):
    global running
    logger.info('Termination signal received - shutting down...')
    running = False
    sys.exit(0)

signal.signal(signal.SIGTERM, handle_exit)
signal.signal(signal.SIGINT, handle_exit)

if __name__ == "__main__":
    # MONGO ENV VARIABLES
    mongo_host = os.getenv('MONGO_HOST', 'localhost')
    mongo_port = int(os.getenv('MONGO_PORT', 27017))
    mongo_password = os.getenv('MONGO_PASSWORD', None)
    mongo_db = os.getenv('MONGO_DB', 'test')
    mongo_username = os.getenv('MONGO_USERNAME', '')

    # CONNECT TO MONGO
    mongo = MongoClient(
        host=mongo_host,
        port=mongo_port,
        password=mongo_password,
        db=mongo_db,
        username=mongo_username
    )

    if not mongo.client or mongo.client is None:
        logger.error('Could not initialize Mongo...')
        logger.error('Exiting...')
        exit(1)


    # Get number of entries in the DB
    logger.info('Fetching number of documents...')

    total_number_of_documents = mongo.get_document_count()

    if total_number_of_documents == None:
        logger.error('No documents found - sleeping...')
        time.sleep(10000)

    logger.info(f'{total_number_of_documents} documents found!')

    logger.info(f'Fetching all the indexed words...')
    word_entries = mongo.get_all_word_entries()

    if word_entries == None:
        logger.error('No word entries found - sleeping...')
        time.sleep(10000)

    word_entries.batch_size(10) # Process in batches of 10 to avoid memory overhead

    bulk_operations = []
    processed_entries = 0
    for word_entry in word_entries:
        # Process word entry
        word = word_entry["_id"]
        pages = word_entry.get("pages", [])

        logger.info(f'Calculating TF-IDF for "{word}"...')

        df = len(pages)
        idf = math.log10(total_number_of_documents / (1 + df))

        # logger.info(f'{word} appears in {df} documents')

        for page in pages:
            tf = page["tf"]
            tfidf = tf * idf

            logger.info(f'TF-IDF({word}) in {page["url"]}: {tfidf}')

            # mongo.update_page_tfidf(word, page["url"], idf, tfidf)
            op = mongo.update_page_tfidf_op(word, page["url"], idf, tfidf)
            if op:
                bulk_operations.append(op)

        processed_entries += 1

        # Write in batches of 10000
        if processed_entries >= 10000:
            if bulk_operations:
                try:
                    result = mongo.perform_batch_operations(bulk_operations)
                    logger.info(f'Batch update for "{word}" applied: {result.bulk_api_result}')
                except Exception as e:
                    logger.error(f'Error in batch update for word "{word}": {e}')
                bulk_operations = []
                processed_entries = 0


    if bulk_operations:
        try:
            result = mongo.perform_batch_operations(bulk_operations)
            logger.info(f'Batch update for "{word}" applied: {result.bulk_api_result}')
        except Exception as e:
            logger.error(f'Error in batch update for word "{word}": {e}')

    logger.info('TF_IDF done...')
    logger.info('Shutting down...')
    sys.exit(0)

