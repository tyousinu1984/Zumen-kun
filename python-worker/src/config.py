"""Configuration management for Python Worker."""

import os

from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """Application settings loaded from environment variables."""

    # Supabase
    supabase_url: str = ""
    supabase_key: str = ""

    # OpenAI
    openai_api_key: str = ""

    # Gemini
    gemini_api_key: str = ""

    # Vision AI Provider: "openai" | "gemini" | "mock"
    # openai → GPT-4o Vision（画像ベース）
    # gemini → Gemini 1.5 Pro（PDF直接対応・無料枠あり）
    # mock   → APIキー不要（開発・テスト用）
    vision_provider: str = "mock"

    # Retell AI
    retell_api_key: str = ""
    retell_agent_id: str = ""

    # Redis
    redis_url: str = "redis://localhost:6379"

    # App
    python_log_level: str = "INFO"
    environment: str = "development"

    class Config:
        env_file = ".env"
        case_sensitive = False


settings = Settings()
