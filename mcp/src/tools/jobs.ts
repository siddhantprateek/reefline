import { z } from "zod";
import { apiRequest } from "../client.js";

// ─── Schemas ──────────────────────────────────────────────────────────────────

const JobIdArgs = z.object({
  id: z.string().describe("Job UUID"),
});

const ArtifactArgs = z.object({
  id: z.string().describe("Job UUID"),
  artifact: z
    .enum(["report.md", "draft.md", "grype.json", "dive.json", "dockle.json"])
    .describe("Which artifact to download"),
});

const CompareArgs = z.object({
  job_id_a: z.string().describe("First job UUID"),
  job_id_b: z.string().describe("Second job UUID"),
});

// ─── Tool definitions ─────────────────────────────────────────────────────────

export const jobTools = [
  {
    name: "get_jobs",
    description:
      "List all analysis jobs submitted to Reefline. Returns job IDs, statuses (pending / running / completed / failed), image refs, and timestamps.",
    inputSchema: {
      type: "object",
      properties: {},
      required: [],
    },
  },
  {
    name: "get_job_by_id",
    description:
      "Get the full status and result summary of a specific Reefline analysis job by its UUID.",
    inputSchema: {
      type: "object",
      properties: {
        id: { type: "string", description: "Job UUID" },
      },
      required: ["id"],
    },
  },
  {
    name: "delete_job_by_id",
    description:
      "Delete a Reefline analysis job and purge all of its stored artifacts from object storage.",
    inputSchema: {
      type: "object",
      properties: {
        id: { type: "string", description: "Job UUID" },
      },
      required: ["id"],
    },
  },
  {
    name: "get_job_artifact",
    description: [
      "Download a specific artifact produced by a completed analysis job.",
      "Available artifacts:",
      "  • report.md   — full AI-generated security + optimization report (Markdown)",
      "  • draft.md    — supervisor first-pass draft (Markdown)",
      "  • grype.json  — Grype vulnerability scan results (JSON)",
      "  • dive.json   — Dive layer efficiency analysis (JSON)",
      "  • dockle.json — Dockle CIS Docker Benchmark results (JSON)",
    ].join("\n"),
    inputSchema: {
      type: "object",
      properties: {
        id: { type: "string", description: "Job UUID" },
        artifact: {
          type: "string",
          enum: ["report.md", "draft.md", "grype.json", "dive.json", "dockle.json"],
          description: "Which artifact to download",
        },
      },
      required: ["id", "artifact"],
    },
  },
  {
    name: "compare_jobs",
    description:
      "Compare the analysis results of two completed Reefline jobs side-by-side. Useful for evaluating before/after optimization changes.",
    inputSchema: {
      type: "object",
      properties: {
        job_id_a: { type: "string", description: "First job UUID" },
        job_id_b: { type: "string", description: "Second job UUID" },
      },
      required: ["job_id_a", "job_id_b"],
    },
  },
] as const;

// ─── Handler ──────────────────────────────────────────────────────────────────

const JOB_TOOL_NAMES = new Set([
  "get_jobs",
  "get_job_by_id",
  "delete_job_by_id",
  "get_job_artifact",
  "compare_jobs",
]);

export async function handleJobTool(
  name: string,
  args: unknown,
): Promise<string | null> {
  if (!JOB_TOOL_NAMES.has(name)) return null;

  switch (name) {
    case "get_jobs": {
      const jobs = await apiRequest<unknown[]>("/jobs");
      return JSON.stringify(jobs, null, 2);
    }

    case "get_job_by_id": {
      const { id } = JobIdArgs.parse(args);
      const job = await apiRequest<unknown>(`/jobs/${id}`);
      return JSON.stringify(job, null, 2);
    }

    case "delete_job_by_id": {
      const { id } = JobIdArgs.parse(args);
      await apiRequest<unknown>(`/jobs/${id}`, { method: "DELETE" });
      return `Job ${id} and all its artifacts have been deleted.`;
    }

    case "get_job_artifact": {
      const { id, artifact } = ArtifactArgs.parse(args);
      const content = await apiRequest<string>(`/jobs/${id}/${artifact}`);
      return typeof content === "string"
        ? content
        : JSON.stringify(content, null, 2);
    }

    case "compare_jobs": {
      const { job_id_a, job_id_b } = CompareArgs.parse(args);
      const result = await apiRequest<unknown>("/compare", {
        method: "POST",
        body: JSON.stringify({ job_id_a, job_id_b }),
      });
      return JSON.stringify(result, null, 2);
    }

    default:
      return null;
  }
}
