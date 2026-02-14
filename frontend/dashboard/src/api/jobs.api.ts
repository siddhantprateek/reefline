import type {
  AnalysisRequest,
  JobResponse,
  Job,
  JobReport,
  GrypeReport,
  DiveReport,
  DockleReport,
} from "@/types/jobs";

export type { AnalysisRequest, JobResponse, Job, JobReport, GrypeReport, DiveReport, DockleReport };
export type { JobStatus } from "@/types/jobs";

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

export async function listJobs(): Promise<Job[]> {
  const res = await fetch(`${API_BASE}/jobs`, {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.message || body.error || `Request failed: ${res.status}`);
  }

  return res.json();
}

export async function getJob(jobId: string): Promise<JobReport> {
  const res = await fetch(`${API_BASE}/jobs/${jobId}`, {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.message || body.error || `Request failed: ${res.status}`);
  }

  return res.json();
}

export async function getGrypeReport(jobId: string): Promise<GrypeReport> {
  const res = await fetch(`${API_BASE}/jobs/${jobId}/grype.json`);
  if (!res.ok) throw new Error(`Failed to fetch grype report: ${res.status}`);
  return res.json();
}

export async function getDiveReport(jobId: string): Promise<DiveReport> {
  const res = await fetch(`${API_BASE}/jobs/${jobId}/dive.json`);
  if (!res.ok) throw new Error(`Failed to fetch dive report: ${res.status}`);
  return res.json();
}

export async function getDockleReport(jobId: string): Promise<DockleReport> {
  const res = await fetch(`${API_BASE}/jobs/${jobId}/dockle.json`);
  if (!res.ok) throw new Error(`Failed to fetch dockle report: ${res.status}`);
  return res.json();
}

export async function deleteJob(jobId: string): Promise<void> {
  const res = await fetch(`${API_BASE}/jobs/${jobId}`, {
    method: "DELETE",
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.message || body.error || `Request failed: ${res.status}`);
  }
}
