import re
import logging
from gemini_client import GeminiClient
from memory import ConversationMemory
from prompts import get_agent_prompt_template
from tools.tools import Tool
from typing import List, Dict, Any

logger = logging.getLogger(__name__)

class Agent:
    """The main agent class that handles user interactions."""

    def __init__(self, gemini_client: GeminiClient, memory: ConversationMemory, tools: List[Tool]):
        self.gemini_client = gemini_client
        self.memory = memory
        self._tools = {tool.name: tool for tool in tools}

    def _get_tools_string(self) -> str:
        """Returns a formatted string of all available tools."""
        return "\n".join([
            f'{tool.name}: {tool.description}' for tool in self._tools.values()
        ])

    async def get_response(self, author: str, message: str, images: List[Dict[str, Any]] = None) -> str:
        """Gets a response from the agent."""
        self.memory.add_message(author, message)
        history = self.memory.get_history()
        tools_string = self._get_tools_string()
        
        prompt = f"{get_agent_prompt_template()}\n\nTOOLS:\n------\n{tools_string}\n\nPrevious conversation history:\n{history}\n\nNew input: {author}: {message}\nFinal Answer:"        
        logger.debug(f"Prompt sent to Gemini: {prompt}")        
        
        response = await self.gemini_client.generate_content(prompt, images)        
        logger.info(f"Gemini's raw response: {response}")

        # Check for tool use
        tool_match = re.search(r"Action: (\w+)\nAction Input: (.*)", response)
        if tool_match:
            tool_name = tool_match.group(1)
            tool_input = tool_match.group(2)
            logger.info(f"Tool use detected: {tool_name} with input {tool_input}")
            tool = self._tools.get(tool_name)
            if tool:
                tool_result = await tool.arun(tool_input)
                observation = f"Tool {tool_name} used. Observation: {tool_result.return_display}"
                logger.info(f"Tool observation: {observation}")
                self.memory.add_message("AI", observation)
                
                # Get a new response with the tool's output
                history = self.memory.get_history()
                prompt = f"{get_agent_prompt_template()}\n\nTOOLS:\n------\n{tools_string}\n\nPrevious conversation history:\n{history}\n\nNew input: {author}: {message}\nFinal Answer:"
                logger.debug(f"Prompt sent to Gemini after tool use: {prompt}")
                response = await self.gemini_client.generate_content(prompt, images)
                logger.info(f"Gemini's raw response after tool use: {response}")

        self.memory.add_message("AI", response)
        logger.info(f"Agent's final response: {response}")
        return response