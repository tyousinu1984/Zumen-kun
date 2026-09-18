"""
FastAPI application for Python Worker.

Provides:
  - Health check endpoint
  - Retell AI webhook endpoint
  - Worker lifecycle management
"""

import asyncio
import logging
import sys
from contextlib import asynccontextmanager

from fastapi import BackgroundTasks, FastAPI, HTTPException, Request
from fastapi.middleware.cors import CORSMiddleware

from .config import settings
from .models import RetellWebhookPayload
from .voice_agent.retell_client import summarize_call_transcript
from .voice_agent.webhook import (
    extract_foreigner_ok_from_summary,
    parse_call_duration,
)
from .worker import ZumenWorker

# Configure logging
logging.basicConfig(
    level=getattr(logging, settings.python_log_level.upper(), logging.INFO),
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
    stream=sys.stdout,
)
logger = logging.getLogger(__name__)

# Global worker instance
worker = ZumenWorker()
worker_task = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Manage worker lifecycle with FastAPI."""
    global worker_task

    # Start worker in background
    logger.info("🚀 Starting background worker...")
    worker_task = asyncio.create_task(worker.start())

    yield

    # Shutdown
    logger.info("⏳ Shutting down worker...")
    await worker.stop()
    if worker_task:
        worker_task.cancel()
        try:
            await worker_task
        except asyncio.CancelledError:
            pass


app = FastAPI(
    title="図面くん Python Worker",
    description="AI-powered PDF parser and voice call agent for Japanese real estate",
    version="1.0.0",
    lifespan=lifespan,
)

# CORS configuration
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:3000", "http://localhost:8080"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# ─────────────────────────────────────────────────────────
# Health Check
# ─────────────────────────────────────────────────────────


@app.get("/health")
async def health_check():
    """Health check endpoint."""
    redis_status = "not configured"

    if worker.redis_client:
        try:
            await worker.redis_client.ping()
            redis_status = "connected"
        except Exception:
            redis_status = "disconnected"

    return {
        "status": "ok",
        "service": "zumen-kun-python-worker",
        "worker_running": worker.running,
        "redis": redis_status,
        "supabase": (
            "configured" if worker.supabase_client else "not configured"
        ),
        "openai": (
            "configured" if settings.openai_api_key else "not configured"
        ),
        "retell": (
            "configured" if settings.retell_api_key else "not configured"
        ),
    }


# ─────────────────────────────────────────────────────────
# Retell AI Webhook
# ─────────────────────────────────────────────────────────


@app.post("/webhook/retell")
async def retell_webhook(
    payload: RetellWebhookPayload,
    background_tasks: BackgroundTasks,
):
    """
    Webhook endpoint for Retell AI call completion events.
    Called by Retell AI when a call ends.
    """
    logger.info(
        f"📞 Retell webhook received: call_id={payload.call_id}, "
        f"status={payload.call_status}"
    )

    if payload.call_status not in ("ended", "error"):
        return {"message": "acknowledged"}

    # Extract property_id from metadata
    property_id = None
    if payload.metadata:
        property_id = payload.metadata.get("property_id")

    if not property_id:
        logger.warning(
            f"No property_id in webhook metadata for call {payload.call_id}"
        )
        raise HTTPException(
            status_code=400, detail="Missing property_id in metadata"
        )

    # Process in background to return quickly
    background_tasks.add_task(
        _process_call_webhook,
        property_id=property_id,
        payload=payload,
    )

    return {"message": "webhook received", "call_id": payload.call_id}


async def _process_call_webhook(
    property_id: str,
    payload: RetellWebhookPayload,
):
    """Process call webhook asynchronously."""
    try:
        # Get property details for context
        property_name = ""
        room_number = ""

        if worker.supabase_client:
            try:
                result = (
                    worker.supabase_client.table("properties")
                    .select("name, room_number")
                    .eq("id", property_id)
                    .single()
                    .execute()
                )
                if result.data:
                    property_name = result.data.get("name", "")
                    room_number = result.data.get("room_number", "")
            except Exception as e:
                logger.warning(f"Could not fetch property details: {e}")

        # Generate AI summary
        ai_summary = ""
        foreigner_ok = None
        duration_seconds = parse_call_duration(payload.duration_ms)

        if payload.transcript:
            ai_summary = await summarize_call_transcript(
                payload.transcript,
                property_name,
                room_number,
            )
            foreigner_ok = extract_foreigner_ok_from_summary(ai_summary)

        # Determine status
        call_status = "success" if payload.call_status == "ended" else "failed"

        # Save call record
        await worker._save_call_record(
            property_id=property_id,
            transcript=payload.transcript or "",
            ai_summary=ai_summary,
            recording_url=payload.recording_url,
            duration_seconds=duration_seconds,
            status=call_status,
        )

        # Update property
        property_updates: dict = {"status": "completed"}
        if foreigner_ok is not None:
            property_updates["foreigner_ok"] = foreigner_ok

        await worker._update_property(property_id, property_updates)

        logger.info(
            f"✅ Call webhook processed: property={property_id}, "
            f"foreigner_ok={foreigner_ok}"
        )

    except Exception as e:
        logger.error(
            f"❌ Failed to process call webhook for {property_id}: {e}"
        )
        await worker._update_property(property_id, {"status": "call_failed"})
