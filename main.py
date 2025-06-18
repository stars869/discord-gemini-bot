import os
import discord
import asyncio
import base64
import aiohttp # For asynchronously fetching image data

# Import LangChain components for Gemini and conversation management
from langchain_google_genai import ChatGoogleGenerativeAI
from langchain.memory import ConversationBufferWindowMemory
from langchain.chains import ConversationChain
from langchain_core.messages import HumanMessage, AIMessage

# --- Configuration ---
# It's recommended to set these as environment variables for security.
# Example: export DISCORD_BOT_TOKEN="YOUR_DISCORD_BOT_TOKEN"
# Example: export GEMINI_API_KEY="YOUR_GEMINI_API_KEY"
DISCORD_BOT_TOKEN = os.getenv('DISCORD_BOT_TOKEN')
GEMINI_API_KEY = os.getenv('GEMINI_API_KEY')
GEMINI_MODEL_NAME = "gemini-2.0-flash" # Using gemini-2.0-flash as per instructions

# Max number of messages (user + AI) to keep in the rolling conversation window.
# A window of 20 means the last 10 user messages and 10 AI responses are remembered.
MEMORY_WINDOW_SIZE = 20

# Discord's maximum message length
DISCORD_MAX_MESSAGE_LENGTH = 2000

# Supported image content types for the LLM
SUPPORTED_IMAGE_TYPES = ['image/png', 'image/jpeg', 'image/webp', 'image/gif']

# Check if environment variables are set
if not DISCORD_BOT_TOKEN:
    print("Error: DISCORD_BOT_TOKEN environment variable not set.")
    print("Please set it before running the bot.")
    exit(1)
if not GEMINI_API_KEY:
    print("Error: GEMINI_API_KEY environment variable not set.")
    print("Please set it before running the bot.")
    exit(1)

# --- LLM System Instruction ---
# This instruction guides the LLM's behavior, personality, and response format.
LLM_SYSTEM_INSTRUCTION = """
你是一位知识渊博、经验丰富的中文专家和老师。你现在在一个中文Discord群组中，你将看到群组内的聊天记录，每条聊天内容的最前面是发送者的名字，请用中文回答所有成员的问题。
你也可以理解并回答关于图片内容的问题。
"""

# --- LangChain LLM and Memory Setup ---
# Initialize the Gemini LLM for LangChain
llm = ChatGoogleGenerativeAI(
    model=GEMINI_MODEL_NAME,
    google_api_key=GEMINI_API_KEY,
    model_kwargs={"system_instruction": LLM_SYSTEM_INSTRUCTION}
)

# Dictionary to store a ConversationChain instance for each channel.
# Each ConversationChain will hold its own memory, which we will manually manage.
# Key: channel_id (int)
# Value: langchain.chains.ConversationChain
conversation_chains = {}

# --- Discord Bot Setup ---
# Define intents needed for the bot.
# MESSAGE_CONTENT is required to read message content from Discord.
intents = discord.Intents.default()
intents.message_content = True
intents.members = True # Ensure this intent is enabled to access member information like display_name

# Initialize the Discord bot using discord.Client directly.
client = discord.Client(intents=intents)

# --- Helper Functions ---

async def get_image_as_base64(url: str) -> tuple[str, str] | None:
    """
    Fetches an image from a URL and returns its base64 encoded string and MIME type.
    Returns None if fetching or encoding fails.
    """
    try:
        async with aiohttp.ClientSession() as session:
            async with session.get(url) as resp:
                if resp.status == 200:
                    content_type = resp.headers.get('Content-Type')
                    if content_type and content_type.startswith('image/'):
                        image_bytes = await resp.read()
                        base64_encoded_image = base64.b64encode(image_bytes).decode('utf-8')
                        return base64_encoded_image, content_type
                    else:
                        print(f"Skipping non-image attachment: {url} (Content-Type: {content_type})")
                        return None
                else:
                    print(f"Failed to download image from {url}: HTTP {resp.status}")
                    return None
    except Exception as e:
        print(f"Error fetching or encoding image from {url}: {e}")
        return None

# --- Event Handlers ---

@client.event
async def on_ready():
    """
    Called when the bot successfully connects to Discord.
    """
    print(f'Logged in as {client.user} (ID: {client.user.id})')
    print('------')

@client.event
async def on_message(message):
    """
    Called every time a message is sent in a channel the bot can see.
    Handles general chat messages and multimodal inputs where the bot is mentioned,
    and stores images in memory regardless of mention.
    """
    # Ignore messages sent by the bot itself to prevent infinite loops.
    if message.author == client.user:
        return

    channel_id = message.channel.id

    # Get the sender's display name or username
    sender_name = message.author.display_name if message.author.display_name else \
                  message.author.global_name if message.author.global_name else \
                  message.author.name

    # Initialize conversation chain (and its memory) for the channel if it doesn't exist
    if channel_id not in conversation_chains:
        # ConversationBufferWindowMemory keeps a rolling window of past interactions
        # We will directly append to its chat_memory.messages and trim it manually.
        memory = ConversationBufferWindowMemory(llm=llm, k=MEMORY_WINDOW_SIZE, return_messages=True)
        conversation_chains[channel_id] = ConversationChain(llm=llm, memory=memory, verbose=False)
        print(f"Initialized new ConversationChain for channel: {channel_id}")

    is_bot_mentioned = client.user.mentioned_in(message)
    original_text_content = message.content.strip()

    # --- Step 1: Prepare parts for the HumanMessage (text and images) for memory ---
    human_message_parts_for_memory = []

    # Add text part to the human message
    # Prepend sender's name to the current textual input for context in memory
    message_text_for_memory = f"{sender_name}: {original_text_content}"
    if message_text_for_memory.strip() != f"{sender_name}:": # Only add if there's actual text content
        human_message_parts_for_memory.append({"type": "text", "text": message_text_for_memory})
    elif message.attachments: # If no text, but attachments, add a placeholder
        human_message_parts_for_memory.append({"type": "text", "text": f"{sender_name}: (仅图片)"}) # (Image only)

    # Process attachments for images and add them to the human message parts
    images_attached_to_current_message = False
    for attachment in message.attachments:
        if attachment.content_type and attachment.content_type.startswith('image/'):
            print(f"Found image attachment: {attachment.url} (Type: {attachment.content_type})")
            image_data_b64, mime_type = await get_image_as_base64(attachment.url)
            if image_data_b64 and mime_type in SUPPORTED_IMAGE_TYPES:
                human_message_parts_for_memory.append({
                    "type": "image_url",
                    "image_url": {
                        "url": f"data:{mime_type};base64,{image_data_b64}"
                    }
                })
                images_attached_to_current_message = True
            else:
                print(f"Skipping unsupported or failed image attachment: {attachment.url}")
        else:
            print(f"Skipping non-image attachment: {attachment.url} (Type: {attachment.content_type})")

    # If the message is completely empty (no text, no supported images), ignore it for memory
    if not human_message_parts_for_memory:
        return

    # Create the HumanMessage object for the current message with its full content
    current_human_message = HumanMessage(content=human_message_parts_for_memory)

    # --- Step 2: Add the current multimodal message to memory and trim if necessary ---
    conversation_chains[channel_id].memory.chat_memory.messages.append(current_human_message)

    # Manually trim the memory to adhere to MEMORY_WINDOW_SIZE
    memory_messages = conversation_chains[channel_id].memory.chat_memory.messages
    if len(memory_messages) > MEMORY_WINDOW_SIZE:
        conversation_chains[channel_id].memory.chat_memory.messages = memory_messages[-MEMORY_WINDOW_SIZE:]
        print(f"Trimmed memory for channel {channel_id}. Current size: {len(conversation_chains[channel_id].memory.chat_memory.messages)}")
    else:
        print(f"Added message to channel {channel_id} memory. Current size: {len(memory_messages)}")


    # --- Step 3: If the bot is mentioned, invoke LLM with the full multimodal history ---
    if is_bot_mentioned:
        # Remove the bot's mention from the message content for the LLM input if it exists.
        # This is primarily for the *current turn's* text input to the LLM.
        cleaned_text_for_llm_input = original_text_content.replace(f'<@{client.user.id}>', '').strip()

        # If after removing mention, there's no text and no images were attached, prompt user.
        # if not cleaned_text_for_llm_input and not images_attached_to_current_message:
        #      await message.channel.send("您提到了我，但没有说任何话或附加支持的图片！您想说什么？")
        #      return

        # Indicate that the bot is typing
        async with message.channel.typing():
            try:
                # The 'current_chat_history' already contains the latest message (current_human_message)
                # with its multimodal content and is already trimmed.
                current_chat_history_for_llm = conversation_chains[channel_id].memory.chat_memory.messages

                print(f"Invoking LLM with {len(current_chat_history_for_llm)} messages (last message contains {len(current_human_message.content)} parts).")
                # Directly invoke the LLM with the full list of messages from memory
                ai_response_message = await llm.ainvoke(current_chat_history_for_llm)
                response_text = ai_response_message.content

                print(f"LLM generated response (length {len(response_text)}): {response_text[:100]}...") # Log first 100 chars

                # --- Handle Discord's 2000 character limit ---
                if len(response_text) > DISCORD_MAX_MESSAGE_LENGTH:
                    chunks = [response_text[i:i + DISCORD_MAX_MESSAGE_LENGTH]
                              for i in range(0, len(response_text), DISCORD_MAX_MESSAGE_LENGTH)]
                    print(f"Response too long, splitting into {len(chunks)} chunks.")
                    for i, chunk in enumerate(chunks):
                        await message.channel.send(chunk)
                        # Optional: Add a small delay between messages to avoid rate limits
                        await asyncio.sleep(0.5)
                else:
                    # Send the Gemini's response back to the Discord channel normally
                    await message.channel.send(response_text)

                # --- Crucial: Add the AI response to memory and trim again ---
                ai_message_for_memory = AIMessage(content=response_text)
                conversation_chains[channel_id].memory.chat_memory.messages.append(ai_message_for_memory)

                # Trim memory after adding AI message
                memory_messages = conversation_chains[channel_id].memory.chat_memory.messages
                if len(memory_messages) > MEMORY_WINDOW_SIZE:
                    conversation_chains[channel_id].memory.chat_memory.messages = memory_messages[-MEMORY_WINDOW_SIZE:]
                    print(f"Trimmed memory for channel {channel_id}. Current size: {len(conversation_chains[channel_id].memory.chat_memory.messages)}")
                else:
                    print(f"Added AI response to channel {channel_id} memory. Current size: {len(memory_messages)}")

            except Exception as e:
                print(f"Error during LangChain multimodal conversation for channel {channel_id}: {e}")
                await message.channel.send(
                    "抱歉！处理您的聊天或图片时出现了一些问题。请稍后再试。"
                )

# --- Run the Bot ---
if __name__ == '__main__':
    print("Starting Discord bot...")
    client.run(DISCORD_BOT_TOKEN)
