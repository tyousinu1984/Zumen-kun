"""GPT-4o Vision parser for Japanese real estate property sheets (図面)."""

import json
import logging
from typing import Optional

from ..config import settings
from ..models import ParsedPropertyData

logger = logging.getLogger(__name__)

# System prompt for parsing Japanese real estate documents
PROPERTY_PARSE_SYSTEM_PROMPT = """あなたは日本の不動産図面（物件概要書）を解析する専門AIです。

# タスク
提供された画像から物件情報を正確に抽出し、JSON形式で返してください。

# 抽出する情報
以下のフィールドを探してください：

1. **property_name** (物件名): マンション名・アパート名（例：鈴木ハイム、〇〇マンション）
2. **room_number** (号室): 部屋番号（例：102、305号室）
3. **phone_number** (電話番号): 管理会社・仲介業者の電話番号
4. **price** (賃料): 月額家賃（数字のみ、単位：万円）
5. **ad_rate** (広告料・AD): 
   - 「広告料」「AD」「客付手数料」「客付AD」などのキーワードを探す
   - 帯（Obi、図面下部の仲介業者情報部分）に記載されることが多い
   - 例：「1ヶ月」「2ヶ月」「0.5ヶ月」「50%」など
6. **floor_plan** (間取り): 部屋のタイプ（例：1K、1LDK、2LDK）
7. **area** (専有面積): 例：「25.5㎡」「30.00m²」
8. **building_age** (築年数・築年月): 例：「2010年3月」「築15年」
9. **access** (アクセス): 最寄り駅・徒歩時間（例：「渋谷駅 徒歩5分」）

# 重要な注意点
- 「帯（Obi）」とは図面の下部にある細長い情報バー。仲介情報が記載されることが多い
- **広告料（AD）の検索ワード**: 「客付手数料」「広告料」「AD」「client fee」
- 情報が見つからない場合はnullを返す
- 数字はできるだけ数値型で返す（賃料は万円単位の数値）

# 出力形式（JSON）
```json
{
  "property_name": "物件名またはnull",
  "room_number": "号室またはnull",
  "phone_number": "電話番号またはnull",
  "price": 数値またはnull,
  "ad_rate": "AD表示またはnull",
  "floor_plan": "間取りまたはnull",
  "area": "面積またはnull",
  "building_age": "築年数またはnull",
  "access": "アクセスまたはnull",
  "raw_text": "画像から読み取ったメインテキストの要約"
}
```

JSON以外の出力は一切不要です。"""


async def parse_property_from_image(
    base64_image: str,
) -> Optional[ParsedPropertyData]:
    """
    Use GPT-4o Vision to extract property information from image.

    Args:
        base64_image: Base64 encoded PNG image

    Returns:
        ParsedPropertyData object with extracted information
    """
    if not settings.openai_api_key:
        logger.warning("OpenAI API key not set, returning mock data")
        return _mock_parsed_data()

    try:
        from openai import AsyncOpenAI

        client = AsyncOpenAI(api_key=settings.openai_api_key)

        logger.info("Calling GPT-4o Vision API...")

        response = await client.chat.completions.create(
            model="gpt-4o",
            messages=[
                {
                    "role": "system",
                    "content": PROPERTY_PARSE_SYSTEM_PROMPT,
                },
                {
                    "role": "user",
                    "content": [
                        {
                            "type": "image_url",
                            "image_url": {
                                "url": f"data:image/png;base64,{base64_image}",
                                "detail": "high",  # High detail for small Japanese text
                            },
                        },
                        {
                            "type": "text",
                            "text": "この不動産図面から物件情報を抽出してください。特に「広告料」「客付手数料」「AD」の情報に注意してください。",
                        },
                    ],
                },
            ],
            max_tokens=1000,
            temperature=0.1,  # Low temperature for consistent extraction
        )

        content = response.choices[0].message.content
        logger.info(f"GPT-4o response received: {content[:200]}...")

        # Parse JSON response
        # Remove markdown code blocks if present
        content = content.strip()
        if content.startswith("```"):
            content = content.split("```")[1]
            if content.startswith("json"):
                content = content[4:]

        data = json.loads(content)
        parsed = ParsedPropertyData(**data)

        logger.info(
            f"Parsed property: {parsed.property_name}, AD: {parsed.ad_rate}"
        )
        return parsed

    except json.JSONDecodeError as e:
        logger.error(f"Failed to parse GPT-4o JSON response: {e}")
        return None
    except Exception as e:
        logger.error(f"GPT-4o Vision API error: {e}")
        return None


def _mock_parsed_data() -> ParsedPropertyData:
    """Return mock data for development without OpenAI API key."""
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
        raw_text="[開発モード: OpenAI APIキー未設定のためモックデータを返しています]",
    )
