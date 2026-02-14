import json
from minio import Minio
from settings import settings

_client: Minio | None = None


def get_client() -> Minio:
    global _client
    if _client is None:
        _client = Minio(
            settings.minio_endpoint,
            access_key=settings.minio_access_key,
            secret_key=settings.minio_secret_key,
            secure=settings.minio_use_ssl,
        )
    return _client


def read_artifact(job_id: str, filename: str) -> bytes:
    bucket = settings.minio_default_bucket
    object_name = f"{job_id}/artifacts/{filename}"
    response = get_client().get_object(bucket, object_name)
    try:
        return response.read()
    finally:
        response.close()
        response.release_conn()


def write_artifact(job_id: str, filename: str, content: bytes, content_type: str = "text/plain") -> None:
    import io
    bucket = settings.minio_default_bucket
    object_name = f"{job_id}/artifacts/{filename}"
    get_client().put_object(
        bucket, object_name,
        io.BytesIO(content), len(content),
        content_type=content_type,
    )
