from typing import Any, NamedTuple

class ToolResult(NamedTuple):
    """The result of a tool execution."""
    return_display: str

class Tool:
    """The base class for all tools."""
    def __init__(self, name: str, description: str):
        self.name = name
        self.description = description

    def run(self, *args: Any, **kwargs: Any) -> Any:
        raise NotImplementedError("Tool does not support sync execution.")

    async def arun(self, *args: Any, **kwargs: Any) -> Any:
        raise NotImplementedError("Tool does not support async execution.")