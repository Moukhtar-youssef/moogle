import json
from typing import List, Dict, Any
from dataclasses import dataclass
from datetime import datetime
from email.utils import parsedate_to_datetime

@dataclass
class Image:
    page_url: str
    alt: str
    file_name: str

    @classmethod
    def from_hash(cls, image: Dict[str, Any]) -> 'Image':
        if image == None:
            return None

        return cls (
            page_url=image.get('page_url', ''),
            alt=image.get('alt', ''),
            file_name=image.get('file_name', '')
        )

