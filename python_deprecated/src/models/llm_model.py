from abc import ABC, abstractmethod
from typing import List, Optional, Dict, Any

class LLMModel(ABC):
    """Abstract base class for Large Language Models"""
    
    @abstractmethod
    async def generate_async(self, prompt: str, **kwargs) -> str:
        """Generate text asynchronously based on the given prompt"""
        pass
    
    @abstractmethod
    async def generate_with_history_async(self, messages: List[Dict[str, str]], **kwargs) -> str:
        """Generate text asynchronously with conversation history"""
        pass
    
    @abstractmethod
    def set_system_prompt(self, system_prompt: str) -> None:
        """Set the system prompt for the model"""
        pass
    