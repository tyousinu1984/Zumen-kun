"""PDF to Image converter using pdf2image."""

import io
import logging
from typing import Optional

logger = logging.getLogger(__name__)


def pdf_url_to_base64_image(pdf_url: str, dpi: int = 300) -> Optional[str]:
    """
    Download PDF from URL and convert first page to base64 PNG image.

    Args:
        pdf_url: URL of the PDF file
        dpi: Resolution in DPI (300 recommended for OCR quality)

    Returns:
        Base64 encoded PNG image string, or None on failure
    """
    try:
        import base64

        import requests
        from pdf2image import convert_from_bytes

        logger.info(f"Downloading PDF from: {pdf_url}")

        # Download PDF
        response = requests.get(pdf_url, timeout=30)
        response.raise_for_status()
        pdf_bytes = response.content

        logger.info(
            f"PDF downloaded ({len(pdf_bytes)} bytes), converting to image..."
        )

        # Convert first page to image at 300 DPI
        images = convert_from_bytes(
            pdf_bytes,
            dpi=dpi,
            first_page=1,
            last_page=1,  # Only first page (物件図面は通常1ページ)
            fmt="png",
        )

        if not images:
            logger.error("No pages found in PDF")
            return None

        # Convert PIL Image to base64
        img_buffer = io.BytesIO()
        images[0].save(img_buffer, format="PNG")
        img_bytes = img_buffer.getvalue()
        base64_image = base64.b64encode(img_bytes).decode("utf-8")

        logger.info(
            f"PDF converted to base64 image ({len(base64_image)} chars)"
        )
        return base64_image

    except ImportError as e:
        logger.error(
            f"Missing dependency: {e}. Install: pip install pdf2image pillow"
        )
        return None
    except Exception as e:
        logger.error(f"Failed to convert PDF to image: {e}")
        return None
