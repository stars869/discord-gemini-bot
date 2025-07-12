import os
import aiohttp
from tools.tools import Tool, ToolResult

class GoogleSearchTool(Tool):
    """A tool for performing Google searches."""

    def __init__(self):
        super().__init__(
            name="google_search",
            description="Searches Google for the given query."
        )
        self.api_key = os.getenv("GOOGLE_API_KEY")
        self.cse_id = os.getenv("GOOGLE_CSE_ID")
        self.url = "https://www.googleapis.com/customsearch/v1"

    async def arun(self, query: str) -> ToolResult:
        """Runs the Google search tool."""
        params = {
            'key': self.api_key,
            'cx': self.cse_id,
            'q': query
        }
        async with aiohttp.ClientSession() as session:
            async with session.get(self.url, params=params) as response:
                if response.status == 200:
                    results = await response.json()
                    snippets = [item['snippet'] for item in results.get('items', [])]
                    return ToolResult(return_display="\n".join(snippets))
                else:
                    return ToolResult(return_display=f"Error: {response.status}")
