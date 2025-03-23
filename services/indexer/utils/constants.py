# Constants

# Message Queues
INDEXER_QUEUE_KEY = "pages_queue"
SIGNAL_QUEUE_KEY = "signal_queue"
RESUME_CRAWL = "RESUME_CRAWL"

# Redis Data
NORMALIZED_URL_PREFIX = "normalized_url"
URL_METADATA_PREFIX = "url_metadata"
PAGE_IMAGES_PREFIX = "page_images"
WORD_IMAGES_PREFIX = "word_images"
IMAGE_PREFIX = "image_data"
PAGE_PREFIX = "page_data"
WORD_PREFIX = "word"

# These ones will be saved by the indexer and the backlinks server
BACKLINKS_PREFIX = "backlinks"
OUTLINKS_PREFIX = "outlinks"

# Maximum words to index
MAX_INDEX_WORDS = 1000

# File extensions (used to omit them when indexing images)
FILE_TYPES = ["png", "svg", "ico", "gif", "jpeg", "jpg"]
