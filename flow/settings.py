from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    # PostgreSQL
    db_host: str = "localhost"
    db_port: int = 5432
    db_user: str = "postgres"
    db_password: str = "postgres"
    db_name: str = "reefline"

    # Encryption (AES-256-GCM — must match Go server's ENCRYPTION_KEY)
    encryption_key: str = ""

    # MinIO
    minio_endpoint: str = "localhost:9000"
    minio_access_key: str = "minioadmin"
    minio_secret_key: str = "minioadmin"
    minio_use_ssl: bool = False
    minio_default_bucket: str = "reefline"

    class Config:
        env_file = "../.env"
        env_file_encoding = "utf-8"
        extra = "ignore"
        # Only read from .env — never fall back to system env vars
        env_ignore_empty = False


settings = Settings()
