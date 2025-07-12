from .llm_model import LLMModel
from google import genai
from google.genai import types
from typing import List, Dict, Any, Optional
import asyncio

class Gemini(LLMModel):
    """Gemini model implementation using the new Google GenAI SDK"""
    
    def __init__(self, api_key: str, model_name: str = "gemini-2.0-flash-001", 
                 temperature: float = 1.0, max_output_tokens: int = 8192,
                 system_prompt: Optional[str] = None):
        self.api_key = api_key
        self.model_name = model_name
        self.system_prompt = system_prompt
        
        # Create client with API key
        self.client = genai.Client(api_key=self.api_key)
        
        # Set default configuration
        self.config = types.GenerateContentConfig(
            temperature=temperature,
            max_output_tokens=max_output_tokens,
            system_instruction=system_prompt
        )
        
    async def generate_async(self, prompt: str, images: Optional[List[Dict[str, Any]]] = None) -> str:
        """Generate text asynchronously, optionally with images"""
        try:
            # Build content based on whether images are provided
            if images:
                # Build parts list with text and images
                parts = [types.Part.from_text(text=prompt)]
                
                # Add image parts
                for image in images:
                    if 'data' in image and 'mime_type' in image:
                        parts.append(types.Part.from_bytes(
                            data=image['data'],
                            mime_type=image['mime_type']
                        ))
                    elif 'uri' in image and 'mime_type' in image:
                        parts.append(types.Part.from_uri(
                            file_uri=image['uri'],
                            mime_type=image['mime_type']
                        ))
                
                contents = types.UserContent(parts=parts)
            else:
                # Text only
                contents = prompt
            
            response = await self.client.aio.models.generate_content(
                model=self.model_name,
                contents=contents,
                config=self.config
            )
            return response.text
        except Exception as e:
            return f"Error generating response: {str(e)}"
    
    async def generate_with_history_async(self, messages: List[Dict[str, str]]) -> str:
        """Generate text asynchronously with conversation history"""
        try:
            # Convert messages to the new format
            contents = []
            for message in messages:
                if message["role"] == "user":
                    contents.append(types.UserContent(
                        parts=[types.Part.from_text(text=message["content"])]
                    ))
                elif message["role"] == "assistant":
                    contents.append(types.ModelContent(
                        parts=[types.Part.from_text(text=message["content"])]
                    ))
            
            response = await self.client.aio.models.generate_content(
                model=self.model_name,
                contents=contents,
                config=self.config
            )
            return response.text
        except Exception as e:
            return f"Error generating response with history: {str(e)}"
    
    def set_system_prompt(self, system_prompt: str) -> None:
        """Set the system prompt for the model"""
        self.system_prompt = system_prompt