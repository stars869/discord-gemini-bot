import os
import aiohttp

class GeminiClient:
    """A client for interacting with the Google Gemini API."""

    def __init__(self, api_key: str):
        self.api_key = api_key
        self.base_url = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"

    async def generate_content(self, prompt: str) -> str:
        """Generates content using the Gemini API."""
        headers = {"Content-Type": "application/json", "X-goog-api-key": self.api_key}
        
        data = {
            "contents": [
                {
                    "parts": [
                        {
                            "text": prompt
                        }
                    ]
                }
            ]
        }

        async with aiohttp.ClientSession() as session:
            async with session.post(self.base_url, headers=headers, json=data) as response:
                if response.status == 200:
                    result = await response.json()
                    # Check for 'parts' in the response, which is the expected structure
                    if 'candidates' in result and result['candidates'][0]['content'].get('parts'):
                        return result['candidates'][0]['content']['parts'][0]['text']
                    else:
                        return "No content generated from the prompt."
                else:
                    # It's helpful to see the error from the API
                    error_text = await response.text()
                    print(f"Error from Gemini API: {error_text}")
                    response.raise_for_status()