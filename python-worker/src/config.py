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
