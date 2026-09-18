"""GPT-4o Vision parser for Japanese real estate property sheets."""

import json
import logging
from typing import Optional

from ..models import ParsedPropertyData
from .prompts import SINGLE_PAGE_SYSTEM_PROMPT, SINGLE_PAGE_USER_MESSAGE

logger = logging.getLogger(__name__)


async def parse_with_openai(
    base64_image: str,
    api_key: str,
) -> Optional[ParsedPropertyData]:
    """
    Use GPT-4o Vision to extract property information from a base64 image.

    Args:
        base64_image: Base64 encoded PNG image
        api_key: OpenAI API key

    Returns:
        ParsedPropertyData or None
    """
    try:
        from openai import AsyncOpenAI

        client = AsyncOpenAI(api_key=api_key)

        logger.info("Calling GPT-4o Vision API...")

        response = await client.chat.completions.create(
            model="gpt-4o",
            messages=[
                {
                    "role": "system",
                    "content": SINGLE_PAGE_SYSTEM_PROMPT,
                },
                {
                    "role": "user",
                    "content": [
                        {
                            "type": "image_url",
                            "image_url": {
                                "url": f"data:image/png;base64,{base64_image}",
                                "detail": "high",
                            },
                        },
                        {
                            "type": "text",
                            "text": SINGLE_PAGE_USER_MESSAGE,
                        },
                    ],
                },
            ],
            max_tokens=1000,
            temperature=0.1,
        )

        content = response.choices[0].message.content
        logger.info(f"GPT-4o response: {content[:200]}...")
        return _parse_json_response(content)

    except Exception as e:
        logger.error(f"GPT-4o Vision API error: {e}")
        return None


def _parse_json_response(content: str) -> Optional[ParsedPropertyData]:
    """Parse JSON from AI response."""
    try:
        content = content.strip()
        if content.startswith("```"):
            content = content.split("```")[1]
            if content.startswith("json"):
                content = content[4:]
        data = json.loads(content)
        parsed = ParsedPropertyData(**data)
        logger.info(f"Parsed: {parsed.property_name}, AD: {parsed.ad_rate}")
        return parsed
    except json.JSONDecodeError as e:
        logger.error(f"Failed to parse JSON response: {e}")
        return None
