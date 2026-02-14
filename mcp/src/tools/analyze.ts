import { z } from "zod";
import { apiRequest } from "../client.js";

// ─── Schema ───────────────────────────────────────────────────────────────────

export const AnalyzeImageArgs = z.object({
  image: z.string().describe(
    "Full image reference to analyze, e.g. 'nginx:1.25' or 'ghcr.io/org/app:latest'",
  ),
  dockerfile: z
    .string()
    .optional()
    .describe("Contents of a Dockerfile to include in the analysis (optional)"),
});

// ─── Tool definition ──────────────────────────────────────────────────────────

export const analyzeTools = [
  {
    name: "analyze_image",
    description:
      "Submit a container image (and optionally a Dockerfile) to Reefline for a full security and optimization analysis — vulnerability scan (Grype), CIS benchmark (Dockle), layer efficiency (Dive), and AI-generated recommendations. Returns the created job ID.",
    inputSchema: {
      type: "object",
      properties: {
        image: {
          type: "string",
          description:
            "Full image reference to analyze, e.g. 'nginx:1.25' or 'ghcr.io/org/app:latest'",
        },
        dockerfile: {
          type: "string",
          description:
            "Contents of a Dockerfile to include in the analysis (optional)",
        },
      },
      required: ["image"],
    },
  },
] as const;

// ─── Handler ──────────────────────────────────────────────────────────────────

export async function handleAnalyzeTool(
  name: string,
  args: unknown,
): Promise<string | null> {
  if (name !== "analyze_image") return null;

  const { image, dockerfile } = AnalyzeImageArgs.parse(args);
  const body: Record<string, string> = { image };
  if (dockerfile) body.dockerfile = dockerfile;

  const result = await apiRequest<Record<string, unknown>>("/analyze", {
    method: "POST",
    body: JSON.stringify(body),
  });

  return JSON.stringify(result, null, 2);
}
