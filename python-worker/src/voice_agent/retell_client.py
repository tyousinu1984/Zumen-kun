"""Retell AI voice agent integration for automated property inquiry calls (物確)."""

import logging
from typing import Optional

from ..config import settings
from ..models import VoiceCallTask

logger = logging.getLogger(__name__)

# System prompt for the AI voice agent conducting property inquiry calls
PROPERTY_INQUIRY_SYSTEM_PROMPT_TEMPLATE = """あなたは不動産会社の事務担当者「田中」です。
管理会社に電話して、物件の空室状況と外国人入居者の可否を確認する役割です。

## 確認する物件情報
- 物件名: {property_name}
- 号室: {room_number}

## 確認事項（必ず確認すること）
1. {property_name} {room_number}の現在の空室状況（空室かどうか）
2. 外国人留学生の入居可否
3. 外国人入居可の場合、必要な条件（保証人、保証会社、審査書類など）
4. 広告料（AD）の条件（既に把握している場合は確認のみ）

## 会話のポリシー
- 常に丁寧な敬語を使う
- 簡潔に要件を伝える
- 相手の回答を正確に聞き取る
- 不明な点は聞き返す
- 確認が完了したら丁寧にお礼を言って終話する

## 会話の流れ例
1. 挨拶と自己紹介
2. 確認事項を順番に質問
3. 回答を確認・復唱
4. お礼を言って終話"""


async def initiate_property_inquiry_call(
    task: VoiceCallTask,
) -> Optional[str]:
    """
    Initiate an automated property inquiry call using Retell AI.

    Args:
        task: VoiceCallTask containing property and phone details

    Returns:
        call_id if successful, None otherwise
    """
    if not settings.retell_api_key:
        logger.warning("Retell AI API key not set, simulating call")
        return f"mock_call_{task.property_id}"

    try:
        import httpx

        system_prompt = PROPERTY_INQUIRY_SYSTEM_PROMPT_TEMPLATE.format(
            property_name=task.property_name or "（物件名不明）",
            room_number=task.room_number or "（号室不明）",
        )

        # Retell AI Create Phone Call API
        payload = {
            "agent_id": settings.retell_agent_id,
            "to_number": task.phone_number,
            "retell_llm_dynamic_variables": {
                "property_name": task.property_name or "",
                "room_number": task.room_number or "",
            },
            "override_agent_general_prompt": system_prompt,
            "metadata": {
                "property_id": task.property_id,
            },
        }

        headers = {
            "Authorization": f"Bearer {settings.retell_api_key}",
            "Content-Type": "application/json",
        }

        async with httpx.AsyncClient(timeout=30) as client:
            response = await client.post(
                "https://api.retellai.com/v2/create-phone-call",
                json=payload,
                headers=headers,
            )
            response.raise_for_status()
            data = response.json()

        call_id = data.get("call_id")
        logger.info(
            f"Retell AI call initiated: {call_id} for property {task.property_id}"
        )
        return call_id

    except Exception as e:
        logger.error(f"Failed to initiate Retell AI call: {e}")
        return None


async def summarize_call_transcript(
    transcript: str,
    property_name: str,
    room_number: str,
) -> str:
    """
    Use GPT to summarize the call transcript and extract key findings.

    Args:
        transcript: Full call transcript
        property_name: Property name for context
        room_number: Room number for context

    Returns:
        AI-generated summary of the call
    """
    if not settings.openai_api_key:
        return "[開発モード: 要約なし - OpenAI APIキー未設定]"

    try:
        from openai import AsyncOpenAI

        client = AsyncOpenAI(api_key=settings.openai_api_key)

        response = await client.chat.completions.create(
            model="gpt-4o-mini",
            messages=[
                {
                    "role": "system",
                    "content": """不動産の物確電話の通話記録から重要情報を抽出して要約してください。

以下の形式で出力：
- 空室状況: 空室 / 成約済み / 不明
- 外国人入居: 可 / 不可 / 条件付き可 / 不明  
- 条件: （保証人要、保証会社要、審査ありなど）
- その他: （特記事項があれば）""",
                },
                {
                    "role": "user",
                    "content": f"物件: {property_name} {room_number}\n\n通話記録:\n{transcript}",
                },
            ],
            max_tokens=300,
            temperature=0.3,
        )

        return response.choices[0].message.content

    except Exception as e:
        logger.error(f"Failed to summarize transcript: {e}")
        return f"要約失敗: {str(e)}"
