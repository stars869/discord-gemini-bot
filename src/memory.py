from collections import deque
from typing import List, Dict

class ConversationMemory:
    """A class to manage conversation history for a single channel."""

    def __init__(self, window_size: int = 20):
        self.window_size = window_size
        self.history: deque = deque(maxlen=self.window_size)

    def add_message(self, author: str, content: str):
        """Adds a message to the conversation history."""
        self.history.append(f"{author}: {content}")

    def get_history(self) -> str:
        """Retrieves the conversation history."""
        return "\n".join(self.history)

