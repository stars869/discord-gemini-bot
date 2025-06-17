import os
import discord
import asyncio
# Removed: from discord.ext import commands - no longer needed for commands

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
# A window of 10 means the last 5 user messages and 5 AI responses are remembered.
MEMORY_WINDOW_SIZE = 20

# Discord's maximum message length
DISCORD_MAX_MESSAGE_LENGTH = 2000

# Check if environment variables are set
if not DISCORD_BOT_TOKEN:
    print("Error: DISCORD_BOT_TOKEN environment variable not set.")
    print("Please set it before running the bot.")
    exit(1)
if not GEMINI_API_KEY:
    print("Error: GEMINI_API_KEY environment variable not set.")
    print("Please set it before running the bot.")
    exit(1)

# --- NEW: LLM System Instruction ---
# This instruction guides the LLM's behavior, personality, and response format.
LLM_SYSTEM_INSTRUCTION = """
你是一位知识渊博、经验丰富的中文专家和老师。你现在在一个中文Discord群组中，你将看到群组内的聊天记录，每条聊天内容的最前面是发送者的名字，请用中文回答所有成员的问题。
"""

# --- LangChain LLM and Memory Setup ---
# Initialize the Gemini LLM for LangChain
llm = ChatGoogleGenerativeAI(
    model=GEMINI_MODEL_NAME, 
    google_api_key=GEMINI_API_KEY,
    model_kwargs={"system_instruction": LLM_SYSTEM_INSTRUCTION} # --- NEW: Pass the system instruction here ---
)

# Dictionary to store a ConversationChain instance for each channel.
# Each ConversationChain will have its own memory.
# Key: channel_id (int)
# Value: langchain.chains.ConversationChain
conversation_chains = {}

# --- Discord Bot Setup ---
# Define intents needed for the bot.
# MESSAGE_CONTENT is required to read message content from Discord.
intents = discord.Intents.default()
intents.message_content = True
intents.members = True # Ensure this intent is enabled to access member information like display_name

# Initialize the Discord bot using discord.Client directly, as commands.Bot is not needed.
client = discord.Client(intents=intents)

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
    Handles general chat messages where the bot is mentioned.
    """
    # Ignore messages sent by the bot itself to prevent infinite loops.
    if message.author == client.user:
        return

    channel_id = message.channel.id

    # Get the sender's display name or username
    # message.author.display_name is the server nickname if set, otherwise username.
    # message.author.global_name is the global display name if set.
    # message.author.name is the original username.
    # We prioritize display_name, then global_name, then name.
    sender_name = message.author.display_name if message.author.display_name else \
                  message.author.global_name if message.author.global_name else \
                  message.author.name

    # Initialize conversation chain for the channel if it doesn't exist
    if channel_id not in conversation_chains:
        # ConversationBufferWindowMemory keeps a rolling window of past interactions
        memory = ConversationBufferWindowMemory(llm=llm, k=MEMORY_WINDOW_SIZE, return_messages=True)
        conversation_chains[channel_id] = ConversationChain(llm=llm, memory=memory, verbose=False)
        print(f"Initialized new ConversationChain for channel: {channel_id}")

    # --- Logic: Add ALL messages to memory, but only invoke LLM if bot is mentioned ---
    is_bot_mentioned = client.user.mentioned_in(message)
    text_content = message.content.strip()

    # Prepend the sender's name to the message content before adding to memory
    # This helps the LLM differentiate between speakers.
    formatted_message_for_memory = f"{sender_name}: {text_content}"
    print(f"Adding message to memory (no immediate LLM invocation): {formatted_message_for_memory}")
    conversation_chains[channel_id].memory.chat_memory.add_user_message(formatted_message_for_memory)


    # Now, if the bot is mentioned, then we invoke the LLM
    if is_bot_mentioned:
        # Remove the bot's mention from the message content.
        cleaned_text_content = message.content.replace(f'<@{client.user.id}>', '').strip()

        # if not cleaned_text_content:
        #     await message.channel.send("You mentioned me, but didn't say anything! What's on your mind?")
        #     return

        # Prepend the sender's name to the cleaned input for the LLM
        # This ensures the LLM knows who is asking the current question.
        formatted_input_for_llm = f"{sender_name}: {cleaned_text_content}"
        print(f"Received LLM invocation request from {message.author} in channel {channel_id}: {formatted_input_for_llm}")

        # Indicate that the bot is typing
        async with message.channel.typing():
            try:
                # LangChain's ConversationChain handles generating the response
                # using the existing memory and the current input.
                response_text = await asyncio.to_thread(
                    conversation_chains[channel_id].predict, input=formatted_input_for_llm
                )

                print(f"LLM generated response (length {len(response_text)}): {response_text[:100]}...") # Log first 100 chars

                # --- NEW LOGIC: Handle Discord's 2000 character limit ---
                if len(response_text) > DISCORD_MAX_MESSAGE_LENGTH:
                    # Split the response into chunks
                    # TODO: Split at \n
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
                # --- END NEW LOGIC ---

                print(f"Sent response(s) to channel {channel_id}.")

            except Exception as e:
                print(f"Error during LangChain conversation for channel {channel_id}: {e}")
                await message.channel.send(
                    "Oops! Something went wrong while trying to process your chat. "
                    "Please try again later."
                )



# --- Run the Bot ---
if __name__ == '__main__':
    print("Starting Discord bot...")
    client.run(DISCORD_BOT_TOKEN)
