import re
import logging
from types.memory import ConversationMemory
from prompts import get_agent_system_prompt_template
from tools.tools import Tool
from models.llm_model import LLMModel
from typing import List, Dict, Any

logger = logging.getLogger(__name__)

class Agent:
    """The main agent class that handles user interactions."""

    def __init__(self, model: LLMModel, memory: ConversationMemory, tools: List[Tool]):
        self.model = model
        self.memory = memory
        self._tools = {tool.name: tool for tool in tools}

        tools_string = self._get_tools_string()
        tool_names = self._get_tool_names()
        system_prompt = get_agent_system_prompt_template().format(tools=tools_string, tool_names=tool_names)
        self.model.set_system_prompt(system_prompt)
        logger.debug(f"System prompt set: {system_prompt}")

    def _get_tool_names(self) -> str:
        """Returns a comma-separated string of tool names."""
        return ', '.join(self._tools.keys())

    def _get_tools_string(self) -> str:
        """Returns a formatted string of all available tools."""
        return "\n".join([
            f'{tool.name}: {tool.description}' for tool in self._tools.values()
        ])

    async def get_response(self, author: str, message: str, images: List[Dict[str, Any]] = None) -> str:
        """Gets a response from the agent."""
        self.memory.add_message(author, message)
        messages = self.memory.get_history()
        
        # Use LLMModel interface
        response = await self.model.generate_with_history_async(messages)
        
        logger.info(f"Model's raw response: {response}")

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
                prompt = f"{get_agent_system_prompt_template()}\n\nTOOLS:\n------\n{tools_string}\n\nPrevious conversation history:\n{history}\n\nNew input: {author}: {message}\nFinal Answer:"
                logger.debug(f"Prompt sent to model after tool use: {prompt}")
                
                # Use LLMModel interface for follow-up response
                response = await self.model.generate_async(prompt, images=images)
                
                logger.info(f"Model's raw response after tool use: {response}")

        self.memory.add_message("AI", response)
        logger.info(f"Agent's final response: {response}")
        return response