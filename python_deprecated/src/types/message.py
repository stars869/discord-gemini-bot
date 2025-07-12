from dataclasses import dataclass
from datetime import datetime

@dataclass
class Message:
    timestamp: datetime
    role: str
    type: str
    content: str
