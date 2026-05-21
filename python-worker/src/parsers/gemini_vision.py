"""Gemini Vision parser for Japanese real estate property sheets.

Gemini 1.5 Pro の特徴：
- PDF をそのまま渡せる（pdf2image 変換不要）
- 複数ページを一括処理できる
- 無料枠あり（開発テストに最適）
"""

import json
import logging
from typing import List, Optional

from ..models import ParsedPropertyData
from .prompts import (
    MULTI_PAGE_PDF_PROMPT,
    SINGLE_PAGE_SYSTEM_PROMPT,
    SINGLE_PAGE_USER_MESSAGE,
)

logger = logging.getLogger(__name__)


async def parse_with_gemini_image(
    base64_image: str,
    api_key: str,
) -> Optional[ParsedPropertyData]:
    """
    Use Gemini Vision to extract property information from a base64 image.

    Args:
        base64_image: Base64 encoded PNG image
        api_key: Gemini API key

    Returns:
        ParsedPropertyData or None
    """
    try:
        import base64

        import google.generativeai as genai

        genai.configure(api_key=api_key)
        model = genai.GenerativeModel("gemini-2.0-flash")

        logger.info("Calling Gemini Vision API (image mode)...")

        # Decode base64 to bytes for Gemini
        image_bytes = base64.b64decode(base64_image)

        response = model.generate_content(
            [
                {
                    "mime_type": "image/png",
                    "data": image_bytes,
                },
                SINGLE_PAGE_SYSTEM_PROMPT + "\n\n" + SINGLE_PAGE_USER_MESSAGE,
            ]
        )

        content = response.text
        logger.info(f"Gemini response: {content[:200]}...")
        return _parse_json_response(content)

    except Exception as e:
        logger.error(f"Gemini Vision API error: {e}")
        return None


async def parse_with_gemini_pdf(
    pdf_bytes: bytes,
    api_key: str,
) -> List[Optional[ParsedPropertyData]]:
    """
    Use Gemini Vision to extract property information directly from PDF bytes.
    Returns a list (one entry per page).

    Gemini の PDF 直接入力機能を使用 → pdf2image 変換不要！

    Args:
        pdf_bytes: Raw PDF bytes
        api_key: Gemini API key

    Returns:
        List of ParsedPropertyData (one per page), None entries for failed pages
    """
    try:
        import base64

        import google.generativeai as genai

        genai.configure(api_key=api_key)
        model = genai.GenerativeModel("gemini-2.0-flash")

        logger.info(
            f"Calling Gemini Vision API (PDF direct mode, {len(pdf_bytes)} bytes)..."
        )

        # Gemini supports PDF directly via base64
        pdf_b64 = base64.b64encode(pdf_bytes).decode("utf-8")

        response = model.generate_content(
            [
                {
                    "mime_type": "application/pdf",
                    "data": pdf_b64,
                },
                MULTI_PAGE_PDF_PROMPT,
            ]
        )

        content = response.text
        logger.info(f"Gemini PDF response: {content[:300]}...")
        return _parse_json_array_response(content)

    except Exception as e:
        logger.error(f"Gemini PDF Vision API error: {e}")
        return []


def _parse_json_response(content: str) -> Optional[ParsedPropertyData]:
    """Parse single JSON object from AI response."""
    try:
        content = content.strip()
        if content.startswith("```"):
            content = content.split("```")[1]
            if content.startswith("json"):
                content = content[4:]
        data = json.loads(content)
        parsed = ParsedPropertyData(**data)
        logger.info(
            f"Gemini parsed: {parsed.property_name}, AD: {parsed.ad_rate}"
        )
        return parsed
    except json.JSONDecodeError as e:
        logger.error(f"Failed to parse Gemini JSON response: {e}")
        return None


def _parse_json_array_response(
    content: str,
) -> List[Optional[ParsedPropertyData]]:
    """Parse JSON array from AI response (multi-page PDF)."""
    try:
        content = content.strip()
        if content.startswith("```"):
            content = content.split("```")[1]
            if content.startswith("json"):
                content = content[4:]
        data_list = json.loads(content)
        if not isinstance(data_list, list):
            # Single object returned - wrap in list
            data_list = [data_list]

        results = []
        for i, data in enumerate(data_list):
            try:
                parsed = ParsedPropertyData(**data)
                logger.info(
                    f"Gemini page {i+1}: {parsed.property_name}, AD: {parsed.ad_rate}"
                )
                results.append(parsed)
            except Exception as e:
                logger.warning(f"Failed to parse page {i+1} data: {e}")
                results.append(None)
        return results

    except json.JSONDecodeError as e:
        logger.error(f"Failed to parse Gemini JSON array response: {e}")
        return []
