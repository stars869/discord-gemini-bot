import base64
import aiohttp

# Supported image content types for the LLM
SUPPORTED_IMAGE_TYPES = ['image/png', 'image/jpeg', 'image/webp', 'image/gif']

async def get_image_as_base64(url: str) -> dict | None:
    """
    Fetches an image from a URL and returns its base64 encoded string and MIME type.
    Returns None if fetching or encoding fails.
    """
    try:
        async with aiohttp.ClientSession() as session:
            async with session.get(url) as resp:
                if resp.status == 200:
                    content_type = resp.headers.get('Content-Type')
                    if content_type and content_type in SUPPORTED_IMAGE_TYPES:
                        image_bytes = await resp.read()
                        base64_encoded_image = base64.b64encode(image_bytes).decode('utf-8')
                        return {"mime_type": content_type, "data": base64_encoded_image}
                    else:
                        return None
                else:
                    return None
    except Exception as e:
        print(f"Error fetching or encoding image from {url}: {e}")
        return None

def split_long_text(text: str, max_length: int = 2000) -> list[str]:
    """
    Splits a long text into chunks, prioritizing newline characters to avoid breaking sentences.
    If a single line is longer than max_length, it will be split by character count.

    Args:
        text (str): The long text to be split.
        max_length (int): The maximum length for each chunk. Defaults to 2000 for Discord.

    Returns:
        list[str]: A list of text chunks.
    """
    if len(text) <= max_length:
        return [text]

    chunks = []
    current_chunk = ""
    lines = text.split('\n')

    for line in lines:
        # Check if adding the current line (plus a newline if it's not the first part of the chunk)
        # would exceed the max length
        # The '1' accounts for the newline character that will be added if current_chunk is not empty.
        if len(current_chunk) + len(line) + (1 if current_chunk else 0) <= max_length:
            if current_chunk:
                current_chunk += '\n'
            current_chunk += line
        else:
            # If the current chunk is not empty, add it to chunks
            if current_chunk:
                chunks.append(current_chunk)
            
            # Now handle the line that didn't fit
            # If the line itself is too long, split it by character count
            if len(line) > max_length:
                # Split this long line into sub-chunks
                for i in range(0, len(line), max_length):
                    chunks.append(line[i:i + max_length])
                current_chunk = "" # Reset current_chunk after splitting a very long line
            else:
                # If the line fits in a new chunk, start a new chunk with it
                current_chunk = line

    # Add any remaining content in current_chunk
    if current_chunk:
        chunks.append(current_chunk)

    return chunks
