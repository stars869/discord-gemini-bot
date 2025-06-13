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

# --- LangChain LLM and Memory Setup ---
# Initialize the Gemini LLM for LangChain
llm = ChatGoogleGenerativeAI(model=GEMINI_MODEL_NAME, google_api_key=GEMINI_API_KEY)

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
intents.members = True

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

    # Initialize conversation chain for the channel if it doesn't exist
    if channel_id not in conversation_chains:
        # ConversationBufferWindowMemory keeps a rolling window of past interactions
        memory = ConversationBufferWindowMemory(llm=llm, k=MEMORY_WINDOW_SIZE, return_messages=True)
        conversation_chains[channel_id] = ConversationChain(llm=llm, memory=memory, verbose=False)
        print(f"Initialized new ConversationChain for channel: {channel_id}")

    # --- Logic: Add ALL messages to memory, but only invoke LLM if bot is mentioned ---
    is_bot_mentioned = client.user.mentioned_in(message)
    text_content = message.content.strip()

    # Add any user message to the conversation history, regardless of whether the bot is mentioned.
    print(f"Adding message to memory (no immediate LLM invocation): {message.author}: {text_content}")
    conversation_chains[channel_id].memory.chat_memory.add_user_message(text_content)


    # Now, if the bot is mentioned, then we invoke the LLM
    if is_bot_mentioned:
        # Remove the bot's mention from the message content.
        cleaned_text_content = message.content.replace(f'<@{client.user.id}>', '').strip()

        if not cleaned_text_content:
            await message.channel.send("You mentioned me, but didn't say anything! What's on your mind?")
            return

        print(f"Received LLM invocation request from {message.author} in channel {channel_id}: {cleaned_text_content}")

        # Indicate that the bot is typing
        async with message.channel.typing():
            try:
                # LangChain's ConversationChain handles generating the response
                # using the existing memory and the current input.
                response_text = await asyncio.to_thread(
                    conversation_chains[channel_id].predict, input=cleaned_text_content
                )

                print(f"LLM generated response (length {len(response_text)}): {response_text[:100]}...") # Log first 100 chars

                # --- NEW LOGIC: Handle Discord's 2000 character limit ---
                if len(response_text) > DISCORD_MAX_MESSAGE_LENGTH:
                    # Split the response into chunks
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