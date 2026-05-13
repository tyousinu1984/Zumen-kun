"""
Redis Worker - Consumes tasks from Go Backend and executes AI processing.

Queues:
  - task:parse_pdf   → PDF parsing with GPT-4o Vision
  - task:voice_call  → Automated property inquiry calls via Retell AI
"""

import asyncio
import json
import logging
import signal
import sys
import time
from typing import Optional

import redis.asyncio as redis

from .config import settings
from .models import ParsePDFTask, VoiceCallTask
from .parsers.pdf_parser import pdf_url_to_base64_images
from .parsers.vision_parser import parse_all_pages_from_pdf
from .voice_agent.retell_client import (
    initiate_property_inquiry_call,
    summarize_call_transcript,
)
from .voice_agent.webhook import extract_foreigner_ok_from_summary

logger = logging.getLogger(__name__)

# Queue names (must match Go Backend constants)
QUEUE_PARSE_PDF = "task:parse_pdf"
QUEUE_VOICE_CALL = "task:voice_call"

# Timeout monitoring: if no status update within 60 seconds, mark as failed
TASK_TIMEOUT_SECONDS = 60


class ZumenWorker:
    """Main worker class that processes tasks from Redis queues."""

    def __init__(self):
        self.redis_client: Optional[redis.Redis] = None
        self.supabase_client = None
        self.running = False

    async def start(self):
        """Initialize connections and start processing."""
        logger.info("🚀 Starting 図面くん Python Worker...")

        # Connect to Redis
        try:
            self.redis_client = redis.from_url(
                settings.redis_url,
                encoding="utf-8",
                decode_responses=True,
            )
            await self.redis_client.ping()
            logger.info("✅ Connected to Redis")
        except Exception as e:
            logger.error(f"❌ Redis connection failed: {e}")
            logger.warning(
                "⚠️  Continuing without Redis (tasks won't be processed)"
            )

        # Connect to Supabase
        if settings.supabase_url and settings.supabase_key:
            try:
                from supabase import create_client

                self.supabase_client = create_client(
                    settings.supabase_url,
                    settings.supabase_key,
                )
                logger.info("✅ Connected to Supabase")
            except Exception as e:
                logger.warning(f"⚠️  Supabase connection failed: {e}")
        else:
            logger.warning(
                "⚠️  Supabase not configured - DB updates will be skipped"
            )

        self.running = True
        logger.info(
            f"👷 Worker ready. Listening on queues: {QUEUE_PARSE_PDF}, {QUEUE_VOICE_CALL}"
        )

        # Start parallel listeners
        await asyncio.gather(
            self._listen_parse_pdf(),
            self._listen_voice_call(),
        )

    async def stop(self):
        """Gracefully stop the worker."""
        self.running = False
        if self.redis_client:
            await self.redis_client.close()
        logger.info("✅ Worker stopped")

    # ─────────────────────────────────────────────────────────
    # PDF Parse Queue Listener
    # ─────────────────────────────────────────────────────────

    async def _listen_parse_pdf(self):
        """Continuously consume tasks from task:parse_pdf queue."""
        logger.info(f"📄 Listening on {QUEUE_PARSE_PDF}...")

        while self.running:
            try:
                if not self.redis_client:
                    await asyncio.sleep(5)
                    continue

                # BLPOP blocks until a message arrives (timeout=5s to allow shutdown)
                result = await self.redis_client.blpop(
                    QUEUE_PARSE_PDF, timeout=5
                )
                if result is None:
                    continue

                _, raw_data = result
                logger.info(f"📄 Received parse task: {raw_data[:100]}...")

                task_data = json.loads(raw_data)
                task = ParsePDFTask(**task_data)

                await self._process_parse_pdf(task)

            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error(f"❌ Parse PDF listener error: {e}")
                await asyncio.sleep(1)

    async def _process_parse_pdf(self, task: ParsePDFTask):
        """Process a PDF parsing task (multi-page support).

        多ページ PDF の場合：
          - 1ページ目 → task.property_id の既存レコードを更新
          - 2ページ目以降 → Supabase に新規物件レコードを追加
        """
        logger.info(
            f"🔍 Processing PDF parse for property: {task.property_id}"
        )
        start_time = time.time()

        try:
            # Step 1: Convert ALL PDF pages to base64 images
            base64_images = pdf_url_to_base64_images(task.file_url, dpi=200)

            if not base64_images:
                raise ValueError("Failed to convert PDF to images")

            logger.info(
                f"📄 Processing {len(base64_images)} page(s) for {task.property_id}"
            )

            for page_idx, base64_image in enumerate(base64_images):
                page_num = page_idx + 1
                logger.info(
                    f"🔍 Parsing page {page_num}/{len(base64_images)}..."
                )

                try:
                    # Step 2: Parse each page with GPT-4o Vision
                    parsed_data = await parse_property_from_image(base64_image)

                    if not parsed_data:
                        logger.warning(
                            f"⚠️  No data extracted from page {page_num}, skipping"
                        )
                        continue

                    updates = {
                        "status": "completed",
                        "name": parsed_data.property_name,
                        "room_number": parsed_data.room_number,
                        "phone_number": parsed_data.phone_number,
                        "price": parsed_data.price,
                        "ad_rate": parsed_data.ad_rate,
                        "pdf_url": task.file_url,
                    }
                    # Remove None values
                    updates = {
                        k: v for k, v in updates.items() if v is not None
                    }

                    if page_idx == 0:
                        # First page: update the existing property record
                        await self._update_property(task.property_id, updates)
                        logger.info(
                            f"✅ Page 1 → updated property {task.property_id}: {parsed_data.property_name}"
                        )
                    else:
                        # Additional pages: create new property records
                        new_id = await self._create_property(updates)
                        if new_id:
                            logger.info(
                                f"✅ Page {page_num} → created new property {new_id}: {parsed_data.property_name}"
                            )
                        else:
                            logger.warning(
                                f"⚠️  Page {page_num}: could not create new property (dev mode)"
                            )

                except Exception as page_err:
                    logger.error(
                        f"❌ Failed to parse page {page_num}: {page_err}"
                    )
                    continue

            elapsed = time.time() - start_time
            logger.info(
                f"✅ PDF multi-page parse completed for {task.property_id} ({elapsed:.1f}s)"
            )

        except Exception as e:
            logger.error(f"❌ PDF parse failed for {task.property_id}: {e}")
            await self._update_property(
                task.property_id,
                {"status": "parse_failed"},
            )

    # ─────────────────────────────────────────────────────────
    # Voice Call Queue Listener
    # ─────────────────────────────────────────────────────────

    async def _listen_voice_call(self):
        """Continuously consume tasks from task:voice_call queue."""
        logger.info(f"📞 Listening on {QUEUE_VOICE_CALL}...")

        while self.running:
            try:
                if not self.redis_client:
                    await asyncio.sleep(5)
                    continue

                result = await self.redis_client.blpop(
                    QUEUE_VOICE_CALL, timeout=5
                )
                if result is None:
                    continue

                _, raw_data = result
                logger.info(
                    f"📞 Received voice call task: {raw_data[:100]}..."
                )

                task_data = json.loads(raw_data)
                task = VoiceCallTask(**task_data)

                await self._process_voice_call(task)

            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error(f"❌ Voice call listener error: {e}")
                await asyncio.sleep(1)

    async def _process_voice_call(self, task: VoiceCallTask):
        """Process a single voice call task."""
        logger.info(
            f"📞 Initiating call for property: {task.property_id} "
            f"({task.property_name} {task.room_number}) → {task.phone_number}"
        )

        try:
            # Initiate call via Retell AI
            call_id = await initiate_property_inquiry_call(task)

            if not call_id:
                raise ValueError(
                    "Failed to initiate call - no call_id returned"
                )

            logger.info(f"📞 Call initiated: {call_id}")

            # In development mode (mock call), simulate call completion
            if call_id.startswith("mock_call_"):
                await asyncio.sleep(2)
                mock_transcript = f"""
田中: はい、〇〇不動産の田中と申します。{task.property_name}の{task.room_number}号室について確認したくお電話しました。
担当者: はい、どのようなことでしょうか？
田中: 現在空室かどうかと、外国人留学生の入居は可能かどうかを確認させていただきたいのですが。
担当者: {task.property_name}の{task.room_number}号室は現在空室となっております。外国人の方については、保証会社ご利用いただければ入居可能です。
田中: ありがとうございます。保証会社はどちらをご利用でしょうか？
担当者: 全保連と提携しております。
田中: 承知いたしました。大変ありがとうございました。
"""
                ai_summary = await summarize_call_transcript(
                    mock_transcript,
                    task.property_name,
                    task.room_number,
                )
                foreigner_ok = extract_foreigner_ok_from_summary(ai_summary)

                await self._save_call_record(
                    property_id=task.property_id,
                    transcript=mock_transcript,
                    ai_summary=ai_summary,
                    recording_url=None,
                    duration_seconds=30,
                    status="success",
                )

                await self._update_property(
                    task.property_id,
                    {
                        "status": "completed",
                        "foreigner_ok": foreigner_ok,
                    },
                )
                logger.info(f"✅ Mock call completed for {task.property_id}")

        except Exception as e:
            logger.error(f"❌ Voice call failed for {task.property_id}: {e}")
            await self._update_property(
                task.property_id,
                {"status": "call_failed"},
            )

    # ─────────────────────────────────────────────────────────
    # Supabase Helper Methods
    # ─────────────────────────────────────────────────────────

    async def _update_property(self, property_id: str, updates: dict):
        """Update property record in Supabase."""
        if not self.supabase_client:
            logger.info(
                f"[DEV] Would update property {property_id}: {updates}"
            )
            return

        try:
            self.supabase_client.table("properties").update(updates).eq(
                "id", property_id
            ).execute()
            logger.info(
                f"✅ Property {property_id} updated: {list(updates.keys())}"
            )
        except Exception as e:
            logger.error(f"❌ Failed to update property {property_id}: {e}")

    async def _create_property(self, data: dict) -> Optional[str]:
        """Create a new property record in Supabase (for multi-page PDFs).
        Returns the new property_id, or None in dev mode."""
        import uuid as uuid_lib

        new_id = str(uuid_lib.uuid4())
        record = {"id": new_id, "status": "completed", **data}

        if not self.supabase_client:
            logger.info(
                f"[DEV] Would create property {new_id}: {data.get('name', '?')}"
            )
            return None  # dev mode: no DB

        try:
            self.supabase_client.table("properties").insert(record).execute()
            logger.info(f"✅ New property created: {new_id}")
            return new_id
        except Exception as e:
            logger.error(f"❌ Failed to create property: {e}")
            return None

    async def _save_call_record(
        self,
        property_id: str,
        transcript: str,
        ai_summary: str,
        recording_url: Optional[str],
        duration_seconds: int,
        status: str,
    ):
        """Save call record to Supabase."""
        record = {
            "property_id": property_id,
            "transcript": transcript,
            "ai_summary": ai_summary,
            "recording_url": recording_url,
            "duration_seconds": duration_seconds,
            "status": status,
        }

        if not self.supabase_client:
            logger.info(f"[DEV] Would save call record: {record}")
            return

        try:
            self.supabase_client.table("call_records").insert(record).execute()
            logger.info(f"✅ Call record saved for property {property_id}")
        except Exception as e:
            logger.error(f"❌ Failed to save call record: {e}")
