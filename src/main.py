import os
import discord
import asyncio
import logging

from gemini_client import GeminiClient
from memory import ConversationMemory
from agent import Agent
from tools.google_search import GoogleSearchTool
from tools.url_fetch import UrlFetchTool
from utils import get_image_as_base64, split_long_text

# --- Logging Configuration ---
logging.basicConfig(
    level=logging.DEBUG,  # Set the logging level (e.g., INFO, DEBUG, WARNING, ERROR)
    format='[%(asctime)s] %(levelname)s: %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)
logger = logging.getLogger(__name__)


# --- Configuration ---
from dotenv import load_dotenv

load_dotenv()

DISCORD_BOT_TOKEN = os.getenv('DISCORD_BOT_TOKEN')
GEMINI_API_KEY = os.getenv('GEMINI_API_KEY')

MEMORY_WINDOW_SIZE = 20
DISCORD_MAX_MESSAGE_LENGTH = 2000

# Supported image content types for the LLM
SUPPORTED_IMAGE_TYPES = ['image/png', 'image/jpeg', 'image/webp', 'image/gif']


# --- Initialize components ---
gemini_client = GeminiClient(api_key=GEMINI_API_KEY)

# Initialize tools
tools = [
    GoogleSearchTool(),
    UrlFetchTool()
]

# Dictionary to hold agent instances for each channel
channel_agents = {}

# --- Discord Bot Setup ---
intents = discord.Intents.default()
intents.message_content = True
intents.members = True

client = discord.Client(intents=intents)


# --- Event Handlers ---
@client.event
async def on_ready():
    """
    Called when the bot successfully connects to Discord.
    """
    logger.info(f'Logged in as {client.user} (ID: {client.user.id})')
    logger.info('------')

@client.event
async def on_message(message):
    """
    Called every time a message is sent in a channel the bot can see.
    Handles general chat messages and multimodal inputs where the bot is mentioned.
    """
    # Ignore messages sent by the bot itself to prevent infinite loops.
    if message.author == client.user:
        return

    is_bot_mentioned = client.user.mentioned_in(message)
    if is_bot_mentioned:
        logger.info(f"Received message from {message.author.display_name} in channel {message.channel.id}: {message.content}")
        original_text_content = message.content.strip()
        # Remove the bot's mention from the message content for the LLM input.
        cleaned_text_for_llm_input = original_text_content.replace(f'<@{client.user.id}>', '').strip()

        images = []
        # Process attachments for images
        if message.attachments:
            for attachment in message.attachments:
                if attachment.content_type and attachment.content_type in SUPPORTED_IMAGE_TYPES:
                    image_data = await get_image_as_base64(attachment.url)
                    if image_data:
                        images.append(image_data)

        # If no text and no images, ignore the message
        if not cleaned_text_for_llm_input and not images:
            logger.info(f"Ignoring empty message from {message.author.display_name} in channel {message.channel.id}")
            return

        # Indicate that the bot is typing
        async with message.channel.typing():
            try:
                channel_id = str(message.channel.id)
                if channel_id not in channel_agents:
                    logger.info(f"Creating new agent for channel {channel_id}")
                    # Create a new memory and agent for this channel
                    memory = ConversationMemory(window_size=MEMORY_WINDOW_SIZE)
                    channel_agents[channel_id] = Agent(gemini_client=gemini_client, memory=memory, tools=tools)
                
                current_agent = channel_agents[channel_id]

                response_text = await current_agent.get_response(
                    author=message.author.display_name,
                    message=cleaned_text_for_llm_input,
                    images=images
                )

                # Handle Discord's 2000 character limit
                if len(response_text) > DISCORD_MAX_MESSAGE_LENGTH:
                    chunks = split_long_text(response_text, DISCORD_MAX_MESSAGE_LENGTH)
                    for chunk in chunks:
                        await message.channel.send(chunk)
                        # Optional: Add a small delay between messages to avoid rate limits
                        await asyncio.sleep(0.5)
                else:
                    # Send the Gemini's response back to the Discord channel normally
                    await message.channel.send(response_text)
                logger.info(f"Sent response to channel {message.channel.id}")

            except Exception as e:
                logger.error(f"Error during agent execution for channel {message.channel.id}: {e}", exc_info=True)
                await message.channel.send("Sorry! Something went wrong while processing your request. Please try again later.")

# --- Run the Bot ---
if __name__ == '__main__':
    print("Starting Discord bot...")
    client.run(DISCORD_BOT_TOKEN)