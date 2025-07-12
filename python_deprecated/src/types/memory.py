from collections import deque
from typing import List
from .message import Message


class ConversationMemory:
    """A class to manage conversation history for a single channel."""

    def __init__(self, window_size: int = 20):
        self.window_size = window_size
        self.history: deque = deque(maxlen=self.window_size)

    def add_message(self, message: Message):
        """Adds a message to the conversation history as a Message object."""
        self.history.append(message)

    def get_history(self) -> List[Message]:
        """Retrieves the conversation history as a list of Message objects."""
        return list(self.history)

