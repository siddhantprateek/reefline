export interface QueueStats {
  active: number;
  pending: number;
  scheduled: number;
  completed: number;
  failed: number;
  throughput_per_hour: number;
}

export interface JobMetrics {
  summary: {
    total: number;
    completed: number;
    failed: number;
    running: number;
    success_rate_pct: number;
    avg_queue_wait_ms: number;
    avg_processing_ms: number;
    avg_total_ms: number;
  };
  time_series: Array<{
    timestamp: string;
    completed: number;
    failed: number;
  }>;
  duration_breakdown: {
    avg_queue_ms: number;
    avg_grype_ms: number;
    avg_dockle_ms: number;
    avg_dive_ms: number;
  };
  status_distribution: {
    COMPLETED: number;
    FAILED: number;
    RUNNING: number;
    PENDING: number;
  };
}

export interface ToolPerformance {
  tools: {
    [toolName: string]: {
      avg_duration_ms: number;
      success_rate_pct: number;
      total_runs: number;
      p50_ms: number;
      p95_ms: number;
      p99_ms: number;
    };
  };
}

export type TimeRange = "24h" | "7d" | "30d";

const API_BASE = "/api/v1";

export async function getQueueStats(): Promise<QueueStats> {
  const res = await fetch(`${API_BASE}/metrics/queue`, {
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

export async function getJobMetrics(timeRange: TimeRange = "24h"): Promise<JobMetrics> {
  const res = await fetch(`${API_BASE}/metrics/jobs?time_range=${timeRange}`, {
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

export async function getToolPerformance(): Promise<ToolPerformance> {
  const res = await fetch(`${API_BASE}/metrics/tools`, {
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
