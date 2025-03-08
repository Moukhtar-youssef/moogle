import json
from typing import List, Dict, Any
from dataclasses import dataclass
from datetime import datetime
from email.utils import parsedate_to_datetime

@dataclass
class Metadata:
    title: str
    description: str
    summary_text: str
    last_crawled: str

    @classmethod
    def from_hash(cls, metadata: Dict[str, Any]) -> 'Metadata':
        if metadata == None:
            return None

        # Parse fields
        last_crawled = parsedate_to_datetime(metadata['last_crawled'])

        return cls (
            title=metadata['title'],
            description=metadata['description'],
            summary_text=metadata['summary_text'],
            last_crawled=last_crawled,
        )

