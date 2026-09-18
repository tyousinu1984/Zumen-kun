"""Webhook handler for Retell AI call completion events."""

import logging
from typing import Optional

logger = logging.getLogger(__name__)


def extract_foreigner_ok_from_summary(summary: str) -> Optional[bool]:
    """
    Parse AI summary to determine foreigner_ok boolean status.

    Args:
        summary: AI-generated summary text

    Returns:
        True if foreigners are accepted, False if not, None if unclear
    """
    if not summary:
        return None

    summary_lower = summary.lower()

    # 外国人可の判定キーワード
    positive_keywords = ["外国人入居: 可", "外国人可", "入居可", "可能"]
    negative_keywords = [
        "外国人入居: 不可",
        "外国人不可",
        "外国人不可",
        "入居不可",
        "不可",
    ]

    for keyword in negative_keywords:
        if keyword in summary:
            return False

    for keyword in positive_keywords:
        if keyword in summary:
            return True

    # 条件付き可もTrueとして扱う
    if "条件付き可" in summary:
        return True

    return None


def parse_call_duration(duration_ms: Optional[int]) -> int:
    """Convert milliseconds to seconds, with None handling."""
    if duration_ms is None:
        return 0
    return duration_ms // 1000
