import json
from datetime import datetime
from typing import List, Dict, Any
from dataclasses import dataclass
from email.utils import parsedate_to_datetime

@dataclass
class Page:
    normalized_url: str
    html: str
    content_type: str
    outgoing_links: List[str]
    status_code: int
    last_crawled: datetime

    @classmethod
    def from_hash(cls, page_data: Dict[str, Any]) -> 'Page':

        if page_data == None:
            return None

        # Parse fields
        outgoing_links = json.loads(page_data['outgoing_links'])
        last_crawled = parsedate_to_datetime(page_data['last_crawled'])

        return cls (
            normalized_url=page_data['normalized_url'],
            html=page_data['html'],
            content_type=page_data['content_type'],
            outgoing_links=outgoing_links,
            status_code=int(page_data['status_code']),
            last_crawled=last_crawled,
        )

    def prettify(self) -> str:
        return f"""
        -----------------------------------------------------
        URL: {self.normalized_url}
        HTML: {self.html[:15] + "..." if len(self.html) > 15 else self.html}
        Content Type: {self.content_type}
        Outgoing Links: {self.outgoing_links if len(self.outgoing_links) < 3 else self.outgoing_links[:3]}
        Status Code: {self.status_code}
        Last Crawled: {self.last_crawled}
        -----------------------------------------------------
        """

