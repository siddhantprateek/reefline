import base64
import json

import psycopg2
import psycopg2.extras
from cryptography.hazmat.primitives.ciphers.aead import AESGCM

from settings import settings


def get_connection():
    return psycopg2.connect(
        host=settings.db_host,
        port=settings.db_port,
        user=settings.db_user,
        password=settings.db_password,
        dbname=settings.db_name,
    )


def _decrypt_credentials(encrypted: str) -> dict:
    """
    Decrypt AES-256-GCM credentials stored by the Go server.
    Format: base64(nonce[12] || ciphertext+tag)
    Key: base64-encoded 32-byte key from ENCRYPTION_KEY env var.
    """
    key = base64.b64decode(settings.encryption_key)
    data = base64.b64decode(encrypted)
    nonce, ciphertext = data[:12], data[12:]
    plaintext = AESGCM(key).decrypt(nonce, ciphertext, None)
    return json.loads(plaintext)


def get_ai_credentials(provider: str, user_id: str = "default-user") -> dict | None:
    """
    Returns integration_id, api_key, model_id for the given provider + user.
    Decrypts the credentials blob stored by the Go server.
    """
    with get_connection() as conn:
        with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
            cur.execute(
                """
                SELECT integration_id, credentials
                FROM integrations
                WHERE user_id = %s
                  AND integration_id = %s
                  AND status = 'connected'
                ORDER BY updated_at DESC
                LIMIT 1
                """,
                (user_id, provider),
            )
            row = cur.fetchone()
            if not row:
                return None

            creds = _decrypt_credentials(row["credentials"])
            return {
                "integration_id": row["integration_id"],
                "api_key": creds.get("apiKey", ""),
                "model_id": creds.get("modelId"),
            }


def get_job(job_id: str) -> dict | None:
    with get_connection() as conn:
        with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
            cur.execute(
                "SELECT * FROM jobs WHERE job_id = %s AND deleted_at IS NULL LIMIT 1",
                (job_id,),
            )
            row = cur.fetchone()
            return dict(row) if row else None
