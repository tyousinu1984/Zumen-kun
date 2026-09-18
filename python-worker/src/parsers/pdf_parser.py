"""PDF to Image converter using pdf2image.

多ページ対応版：日本の物件図面リストは複数ページにまたがるため、
全ページを処理して各ページの base64 画像リストを返す。
"""

import io
import logging
from typing import List, Optional

logger = logging.getLogger(__name__)


def pdf_url_to_base64_images(pdf_url: str, dpi: int = 200) -> List[str]:
    """
    Download PDF from URL and convert ALL pages to base64 PNG images.

    Args:
        pdf_url: URL of the PDF file (http:// or local path)
        dpi: Resolution in DPI (200 is sufficient for GPT-4o Vision and faster)

    Returns:
        List of base64 encoded PNG image strings (one per page), or empty list on failure
    """
    try:
        import base64

        import requests
        from pdf2image import convert_from_bytes

        logger.info(f"Downloading PDF from: {pdf_url}")

        # Download PDF
        response = requests.get(pdf_url, timeout=60)
        response.raise_for_status()
        pdf_bytes = response.content

        logger.info(
            f"PDF downloaded ({len(pdf_bytes)} bytes), converting all pages to images..."
        )

        # Convert ALL pages to images
        images = convert_from_bytes(
            pdf_bytes,
            dpi=dpi,
            fmt="png",
        )

        if not images:
            logger.error("No pages found in PDF")
            return []

        logger.info(f"PDF has {len(images)} page(s)")

        base64_images = []
        for i, img in enumerate(images):
            img_buffer = io.BytesIO()
            img.save(img_buffer, format="PNG")
            img_bytes = img_buffer.getvalue()
            b64 = base64.b64encode(img_bytes).decode("utf-8")
            base64_images.append(b64)
            logger.info(
                f"Page {i+1}/{len(images)} converted ({len(b64)} chars)"
            )

        return base64_images

    except ImportError as e:
        logger.error(
            f"Missing dependency: {e}. Install: pip install pdf2image pillow requests"
        )
        return []
    except Exception as e:
        logger.error(f"Failed to convert PDF to images: {e}")
        return []


def pdf_url_to_base64_image(pdf_url: str, dpi: int = 200) -> Optional[str]:
    """
    Backward-compatible single-page version.
    Returns only the first page as base64.
    """
    images = pdf_url_to_base64_images(pdf_url, dpi=dpi)
    return images[0] if images else None
