#!/usr/bin/env python3
"""
Example usage of the LLMModel abstract class and Gemini implementation
"""

import os
from src.models import LLMModel, Gemini

def main():
    # Example API key - replace with your actual API key
    api_key = os.getenv("GEMINI_API_KEY", "your-api-key-here")
    
    # Create Gemini model instance
    gemini_model = Gemini(api_key=api_key)
    
    # Set system prompt
    gemini_model.set_system_prompt("You are a helpful AI assistant.")
    
    # Get model info
    print("Model Info:")
    print(gemini_model.get_model_info())
    print()
    
    # Generate simple response
    print("Simple Generation:")
    response = gemini_model.generate("What is the capital of France?")
    print(response)
    print()
    
    # Generate with conversation history
    print("Generation with History:")
    messages = [
        {"role": "user", "content": "Hello, what's your name?"},
        {"role": "assistant", "content": "I'm Claude, an AI assistant."},
        {"role": "user", "content": "What can you help me with?"}
    ]
    response = gemini_model.generate_with_history(messages)
    print(response)

if __name__ == "__main__":
    main()
