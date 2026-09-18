"""Vision AI Router - 環境変数 VISION_PROVIDER で OpenAI / Gemini を切り替え。

VISION_PROVIDER=openai  → GPT-4o Vision（画像ベース）
VISION_PROVIDER=gemini  → Gemini 1.5 Pro（PDF直接 or 画像）
VISION_PROVIDER=mock    → モックデータ（APIキー不要・開発用）
"""

import logging
from typing import List, Optional

from ..config import settings
from ..models import ParsedPropertyData

logger = logging.getLogger(__name__)


async def parse_property_from_image(
    base64_image: str,
    pdf_bytes: Optional[bytes] = None,
) -> Optional[ParsedPropertyData]:
    """
    Parse property info from image (or PDF bytes for Gemini).

    Args:
        base64_image: Base64 PNG image (used by OpenAI and Gemini image mode)
        pdf_bytes: Raw PDF bytes (used by Gemini PDF direct mode)

    Returns:
        ParsedPropertyData or None
    """
    provider = settings.vision_provider.lower()
    logger.info(f"🤖 Vision provider: {provider}")

    if provider == "gemini":
        if pdf_bytes and settings.gemini_api_key:
            # Gemini PDF direct mode (fastest - no image conversion needed)
            logger.info("Using Gemini PDF direct mode")
            from .gemini_vision import parse_with_gemini_pdf

            results = await parse_with_gemini_pdf(
                pdf_bytes, settings.gemini_api_key
            )
            # Return first result for single-page compatibility
            return results[0] if results else None
        elif settings.gemini_api_key:
            # Gemini image mode
            logger.info("Using Gemini image mode")
            from .gemini_vision import parse_with_gemini_image

            return await parse_with_gemini_image(
                base64_image, settings.gemini_api_key
            )
        else:
            logger.warning("GEMINI_API_KEY not set, falling back to mock")
            return _mock_parsed_data()

    elif provider == "openai":
        if settings.openai_api_key:
            logger.info("Using OpenAI GPT-4o Vision")
            from .openai_vision import parse_with_openai

            return await parse_with_openai(
                base64_image, settings.openai_api_key
            )
        else:
            logger.warning("OPENAI_API_KEY not set, falling back to mock")
            return _mock_parsed_data()

    else:
        # mock mode or unknown provider
        logger.info("Using mock data (VISION_PROVIDER=mock or not set)")
        return _mock_parsed_data()


async def parse_all_pages_from_pdf(
    pdf_bytes: bytes,
    base64_images: List[str],
) -> List[Optional[ParsedPropertyData]]:
    """
    Parse all pages from a PDF.

    Gemini PDF direct mode: 全ページを一括処理（高速）
    OpenAI / Gemini image mode: ページごとに処理

    Args:
        pdf_bytes: Raw PDF bytes
        base64_images: List of base64 PNG images (one per page)

    Returns:
        List of ParsedPropertyData (one per page)
    """
    provider = settings.vision_provider.lower()

    if provider == "gemini" and settings.gemini_api_key:
        # Gemini PDF direct: send entire PDF at once → get all pages back
        logger.info(
            f"🚀 Gemini PDF direct mode: processing all {len(base64_images)} pages at once"
        )
        from .gemini_vision import parse_with_gemini_pdf

        return await parse_with_gemini_pdf(pdf_bytes, settings.gemini_api_key)

    else:
        # OpenAI / Gemini image / mock: process page by page
        logger.info(
            f"📄 Processing {len(base64_images)} pages one by one (provider: {provider})"
        )
        results = []
        for i, b64 in enumerate(base64_images):
            logger.info(f"Parsing page {i+1}/{len(base64_images)}...")
            result = await parse_property_from_image(b64)
            results.append(result)
        return results


def _mock_parsed_data() -> ParsedPropertyData:
    """Return mock data for development without API key."""
    return ParsedPropertyData(
        property_name="サンプルマンション",
        room_number="101",
        phone_number="03-1234-5678",
        price=8.5,
        ad_rate="1ヶ月",
        floor_plan="1K",
        area="25.0㎡",
        building_age="2018年3月",
        access="渋谷駅 徒歩5分",
        raw_text="[開発モード: APIキー未設定のためモックデータを返しています]",
    )
