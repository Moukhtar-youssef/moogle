import re
import nltk
import string

from bs4 import BeautifulSoup

from nltk.corpus import stopwords
from nltk.tokenize import word_tokenize

nltk.download('stopwords')
nltk.download('punkt')
nltk.download('punkt_tab')

stop_words = set(stopwords.words('english'))

def get_meta_content(soup, property_value=None, name_value=None):
    tag = None
    if property_value:
        tag = soup.find('meta', property=property_value)
    if not tag and name_value:
        tag = soup.find('meta', attrs={'name': name_value})

    if tag and 'content' in tag.attrs:
        return tag['content']

    return None

def process_text(soup):
    unfiltered_text = soup.get_text()
    unfiltered_text = re.sub(r'\n+', ' ', unfiltered_text)    # Remove trailing newlines
    unfiltered_text = re.sub(r'\s+', ' ', unfiltered_text)    # Remove trailing spaces
    unfiltered_text.strip()                        # Remove extra spaces

    # FIXME: Check summary text for important parts?
    # Get the summary_text
    summary_text = unfiltered_text

    # Tokenize the text
    filtered_text = word_tokenize(unfiltered_text)

    # Check that we don't have any stop words or punctuation
    filtered_text = [word.lower() for word in filtered_text if word.lower() not in stop_words and word.lower().isalnum()]

    return {
        'summary_text': summary_text,
        'filtered_text': filtered_text
    }

def get_html_data(html: str):
    # Parse html
    soup = BeautifulSoup(html, 'html.parser')

    # Get metadata
    title = get_meta_content(soup, property_value='og:title', name_value='title')
    description = get_meta_content(soup, property_value='og:description', name_value='description')
    canonical_url = get_meta_content(soup, property_value='og:url', name_value='url')

    # Process text
    text = process_text(soup)
    filtered_text = text['filtered_text']
    summary_text = text['summary_text']

    return {
        'title': title,
        'description': description,
        'summary_text': summary_text,
        'text': filtered_text,
    }

