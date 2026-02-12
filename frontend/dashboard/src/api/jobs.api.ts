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

export type JobStatus = "PENDING" | "RUNNING" | "COMPLETED" | "FAILED" | "CANCELLED" | "SKIPPED" | "UNKNOWN";

export interface Job {
  id: string;
  job_id: string;
  image_ref: string;
  dockerfile?: string;
  status: JobStatus;
  scenario?: "dockerfile_only" | "image_only" | "both";
  created_at: string;
  updated_at: string;
  completed_at?: string;
  error_message?: string;
  progress?: number;
}

export interface JobReport {
  job_id: string;
  status: JobStatus;
  input_scenario: string;
  report?: {
    measured?: {
      current_size_mb?: number;
      layer_efficiency_pct?: number;
      total_packages?: number;
      total_cves?: number;
      critical_cves?: number;
      high_cves?: number;
      runs_as_root?: boolean;
      secrets_detected?: number;
      base_image_age_days?: number;
      lint_warnings?: number;
      lint_errors?: number;
    };
    proposed?: {
      estimated_size_mb?: number;
      estimated_cve_reduction?: number;
      estimated_remaining_cves?: number;
      actionable_cves?: number;
      score?: {
        current: number;
        estimated_after: number;
        grade_current: string;
        grade_estimated: string;
      };
      recommendations?: Array<{
        priority: number;
        category: string;
        title: string;
        description: string;
        effort: string;
        estimated_size_savings_mb?: number;
        estimated_cves_eliminated?: number;
        before?: string;
        after?: string;
      }>;
      optimized_dockerfile?: string;
      disclaimer?: string;
    };
    tool_data?: any;
  };
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

export async function deleteJob(jobId: string): Promise<void> {
  const res = await fetch(`${API_BASE}/jobs/${jobId}`, {
    method: "DELETE",
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.message || body.error || `Request failed: ${res.status}`);
  }
}
