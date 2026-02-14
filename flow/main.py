import logging
import os
from contextlib import asynccontextmanager

from agents import set_tracing_disabled, set_tracing_export_api_key
from dotenv import load_dotenv
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

from flow import run_flow
from integration.db import get_ai_credentials, get_job
from provider import ProviderConfig

# Load environment variables from .env file
load_dotenv()

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")

# Enable tracing if an OpenAI API key is available for export;
# non-OpenAI providers can still trace via the OpenAI Traces dashboard.
_tracing_key = os.getenv("OPENAI_API_KEY", "")
if _tracing_key:
    set_tracing_export_api_key(_tracing_key)
    logging.info("Tracing enabled — exporting to OpenAI Traces dashboard")
else:
    set_tracing_disabled(True)
    logging.info("Tracing disabled — set OPENAI_API_KEY to enable")


@asynccontextmanager
async def lifespan(app: FastAPI):
    yield


app = FastAPI(title="Reefline Flow Service", lifespan=lifespan)


# ── Request / Response ────────────────────────────────────────────────────────

class ReportRequest(BaseModel):
    job_id: str
    provider: str = "openai"  # used to pick the integration row if multiple exist


class ReportResponse(BaseModel):
    job_id: str
    report: str
    bytes: int


# ── Routes ────────────────────────────────────────────────────────────────────

@app.get("/health")
def health():
    return {"status": "ok"}


@app.post("/report", response_model=ReportResponse)
async def generate_report(req: ReportRequest):
    # 1. Verify job exists
    job = get_job(req.job_id)
    if not job:
        raise HTTPException(status_code=404, detail=f"job {req.job_id!r} not found")

    # 2. Load AI credentials from DB
    creds = get_ai_credentials(req.provider)
    if not creds:
        raise HTTPException(status_code=400, detail="no connected AI integration found")

    cfg = ProviderConfig.from_db_row(creds)

    # 3. Run the Supervisor → Critique flow
    try:
        report = await run_flow(req.job_id, cfg)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

    return ReportResponse(job_id=req.job_id, report=report, bytes=len(report))
