import re
import nltk
import string
import py3langid as langid

from bs4 import BeautifulSoup

from nltk.corpus import stopwords
from nltk.tokenize import word_tokenize

from utils.constants import FILE_TYPES, POPULAR_DOMAINS

nltk.download('stopwords')
nltk.download('punkt')
nltk.download('punkt_tab')

stop_words = set(stopwords.words('english'))

def split_name(filename: str):
    pattern = r'[-_./\s]+'
    parts = re.split(pattern, filename)
    parts = [part.lower() for part in parts if part and part.lower() not in FILE_TYPES and "px" not in part.lower()]
    return parts

def split_url(url: str):
    pattern = r'[.,\-_\/+\:\(\)]'
    parts = re.split(pattern, url)
    parts = [part.lower() for part in parts if part and part.lower() not in POPULAR_DOMAINS]
    return parts

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

    # Get the summary_text
    summary_text = unfiltered_text

    # Tokenize the text
    filtered_text = word_tokenize(unfiltered_text)

    # Check that we don't have any stop words or punctuation
    filtered_text = [word.lower() for word in filtered_text if word.lower() not in stop_words and word.lower().isalnum()]

    # TODO: Return language as well
    return {
        'summary_text': summary_text,
        'filtered_text': filtered_text,
        # RETURN LANGUAGE HERE
    }

def get_html_data(html: str):
    # Parse html
    soup = BeautifulSoup(html, 'html.parser')

    # Get metadata
    title = get_meta_content(soup, property_value='og:title', name_value='title')
    description = get_meta_content(soup, property_value='og:description', name_value='description')
    canonical_url = get_meta_content(soup, property_value='og:url', name_value='url')

    page_text = ""
    paragraphs = soup.find_all('p')
    for p in paragraphs:
        page_text += re.sub(r'\[.*?\]', '', p.getText()).strip()

    # Summary text
    # Take first 500 words
    summary_text = page_text.split()
    summary_text = " ".join(summary_text) if len(summary_text) < 500 else " ".join(summary_text[:500])
    # print(f'Summary text: {summary_text}')

    # Processed text should remove all thet ",' for now
    # Process text
    filtered_text = word_tokenize(page_text)

    # Check that we don't have any stop words or punctuation
    filtered_text = [word.lower() for word in filtered_text if word.lower() not in stop_words and word.lower().isalnum()]

    # print(f'Filtered text: {filtered_text}')

    language, confidence = langid.classify(summary_text)

    return {
        'title': title,
        'description': description,
        'summary_text': summary_text,
        'text': filtered_text,
        'language': language
    }

from PIL import Image
import requests
from io import BytesIO

def is_valid_image(url, min_width=100, min_height=100):
    try:
        url = 'https://' + url
        response = requests.get(url, timeout=5)
        response.raise_for_status()
        img = Image.open(BytesIO(response.content))
        width, height = img.size
        return width >= min_width and height >= min_height
    except Exception as e:
        # print(f"Skipping {url}: {e}")
        return False

