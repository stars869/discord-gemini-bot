import aiohttp
from tools.tools import Tool, ToolResult

class UrlFetchTool(Tool):
    """A tool for fetching content from a URL."""

    def __init__(self):
        super().__init__(
            name="url_fetch",
            description="Fetches the content of a given URL. Input should be a valid URL string."
        )

    async def arun(self, url: str) -> ToolResult:
        """Runs the URL fetch tool."""
        try:
            async with aiohttp.ClientSession() as session:
                async with session.get(url) as response:
                    response.raise_for_status()  # Raise an exception for HTTP errors (4xx or 5xx)
                    content = await response.text()
                    return ToolResult(return_display=f"Content from {url}:\n{content[:1000]}...") # Return first 1000 chars
        except aiohttp.ClientError as e:
            return ToolResult(return_display=f"Error fetching URL {url}: {e}")
        except Exception as e:
            return ToolResult(return_display=f"An unexpected error occurred: {e}")

