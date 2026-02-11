export interface AnalysisRequest {
  image_ref: string;
  app_context?: string;
}

export interface JobResponse {
  job_id: string;
  status: string;
  stream_url?: string;
  error?: string;
}

const API_BASE = "/api/v1";

export async function analyzeImage(imageRef: string): Promise<JobResponse> {
  const res = await fetch(`${API_BASE}/analyze`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ image_ref: imageRef }),
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.message || body.error || `Request failed: ${res.status}`);
  }

  return res.json();
}
