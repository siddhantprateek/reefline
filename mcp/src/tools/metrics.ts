import { z } from "zod";
import { apiRequest } from "../client.js";

// ─── Schemas ──────────────────────────────────────────────────────────────────

const JobMetricsArgs = z.object({
  time_range: z
    .enum(["24h", "7d", "30d"])
    .optional()
    .describe("Time window for aggregation — 24h (default), 7d, or 30d"),
});

// ─── Tool definitions ─────────────────────────────────────────────────────────

export const metricsTools = [
  {
    name: "get_queue_stats",
    description:
      "Get real-time job queue statistics: number of active, pending, completed, and failed jobs currently in the Reefline queue.",
    inputSchema: {
      type: "object",
      properties: {},
      required: [],
    },
  },
  {
    name: "get_job_metrics",
    description:
      "Get aggregated job metrics and trends over a time window (24h, 7d, or 30d): total scans, pass/fail rates, average duration, vulnerability counts, and more.",
    inputSchema: {
      type: "object",
      properties: {
        time_range: {
          type: "string",
          enum: ["24h", "7d", "30d"],
          description: "Time window for aggregation (default: 24h)",
        },
      },
      required: [],
    },
  },
  {
    name: "get_tool_performance",
    description:
      "Get performance metrics for the individual scanning tools (Grype, Dockle, Dive): average run time, success rate, and error counts.",
    inputSchema: {
      type: "object",
      properties: {},
      required: [],
    },
  },
] as const;

// ─── Handler ──────────────────────────────────────────────────────────────────

const METRICS_TOOL_NAMES = new Set([
  "get_queue_stats",
  "get_job_metrics",
  "get_tool_performance",
]);

export async function handleMetricsTool(
  name: string,
  args: unknown,
): Promise<string | null> {
  if (!METRICS_TOOL_NAMES.has(name)) return null;

  switch (name) {
    case "get_queue_stats": {
      const stats = await apiRequest<unknown>("/metrics/queue");
      return JSON.stringify(stats, null, 2);
    }

    case "get_job_metrics": {
      const { time_range } = JobMetricsArgs.parse(args ?? {});
      const qs = time_range ? `?time_range=${time_range}` : "";
      const metrics = await apiRequest<unknown>(`/metrics/jobs${qs}`);
      return JSON.stringify(metrics, null, 2);
    }

    case "get_tool_performance": {
      const perf = await apiRequest<unknown>("/metrics/tools");
      return JSON.stringify(perf, null, 2);
    }

    default:
      return null;
  }
}
