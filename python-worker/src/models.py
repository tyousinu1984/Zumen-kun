"""Pydantic models for data transfer objects (DTOs)."""

from datetime import datetime
from typing import Optional

from pydantic import BaseModel


# DTO: Task received from Go via Redis
class ParsePDFTask(BaseModel):
    """Task pushed by Go Backend for PDF parsing."""

    property_id: str
    file_url: str
    timestamp: int


class VoiceCallTask(BaseModel):
    """Task pushed by Go Backend for voice call."""

    property_id: str
    phone_number: str
    property_name: str
    room_number: str
    timestamp: int


# DTO: Parsed property data
class ParsedPropertyData(BaseModel):
    """Extracted data from PDF Vision parsing."""

    property_name: Optional[str] = None  # 物件名
    room_number: Optional[str] = None  # 号室
    phone_number: Optional[str] = None  # 電話番号
    price: Optional[float] = None  # 賃料
    ad_rate: Optional[str] = None  # AD (広告料)
    floor_plan: Optional[str] = None  # 間取り
    area: Optional[str] = None  # 専有面積
    building_age: Optional[str] = None  # 築年数
    access: Optional[str] = None  # アクセス
    raw_text: Optional[str] = None  # 生テキスト（デバッグ用）


# Supabase property model
class Property(BaseModel):
    id: str
    name: Optional[str] = None
    room_number: Optional[str] = None
    phone_number: Optional[str] = None
    price: Optional[float] = None
    ad_rate: Optional[str] = None
    status: str = "pending"
    foreigner_ok: Optional[bool] = None
    pdf_url: Optional[str] = None
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None


# Call record model
class CallRecord(BaseModel):
    property_id: str
    transcript: Optional[str] = None
    ai_summary: Optional[str] = None
    recording_url: Optional[str] = None
    duration_seconds: Optional[int] = None
    status: str = "pending"


# Webhook payload from Retell AI
class RetellWebhookPayload(BaseModel):
    """Webhook data received from Retell AI after call completes."""

    call_id: str
    call_status: str  # ended, failed
    transcript: Optional[str] = None
    recording_url: Optional[str] = None
    call_ended_reason: Optional[str] = None
    metadata: Optional[dict] = None
    duration_ms: Optional[int] = None
