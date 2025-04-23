import json
from typing import List, Dict, Any
from dataclasses import dataclass, asdict

@dataclass
class Image:
    _id:        str # image_url
    page_url:   str
    alt:        str
    filename:  str

    @classmethod
    def from_hash(cls, image: Dict[str, Any], image_url: str) -> 'Image':
        if image == None:
            return None

        return cls (
            _id=image_url,
            page_url=image.get('page_url', ''),
            alt=image.get('alt', ''),
            filename=image.get('filename', '')
        )

    def to_dict(self) -> Dict[str, Any]:
        # Convert to dictionary
        data = asdict(self)
        return data

    def prettify(self) -> str:
        return f"""
        -----------------------------------------------------
        IMAGE URL: {self._id}
        PAGE URL: {self.page_url}
        ALT: {self.alt}
        FILENAME: {self.filename}
        -----------------------------------------------------
        """

