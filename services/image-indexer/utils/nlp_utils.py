import nltk

_initialized = False
STOP_WORDS = None

def initialize_nlp():
    # Download NLTK data files (if not already downloaded)
    global _initialized, STOP_WORDS
    
    if not _initialized:
        resources = ['stopwords', 'punkt']
        
        for resource in resources:
            try:
                nltk.data.find(f'tokenizers/{resource}' if resource == 'punkt' else f'corpora/{resource}')
            except LookupError:
                print(f"Downloading {resource}...")
                nltk.download(resource)
        
        # Initialize stop words
        from nltk.corpus import stopwords
        STOP_WORDS = set(stopwords.words('english'))
        _initialized = True
        
    return STOP_WORDS

# Initialize NLP resources on module load
initialize_nlp()