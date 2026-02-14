#!/usr/bin/env node

/**
 * Reefline MCP Server
 *
 * Exposes the Reefline container image security platform via the
 * Model Context Protocol, allowing AI assistants to trigger scans,
 * inspect jobs, manage registry integrations, and retrieve artifacts.
 *
 * Required environment variable:
 *   REEFLINE_API_URL  — Base URL of the Reefline server, e.g. http://localhost:8080
 */

import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import {
  CallToolRequestSchema,
  ErrorCode,
  ListToolsRequestSchema,
  McpError,
} from "@modelcontextprotocol/sdk/types.js";
import { z } from "zod";

import { analyzeTools, handleAnalyzeTool } from "./tools/analyze.js";
import { jobTools, handleJobTool } from "./tools/jobs.js";
import { integrationTools, handleIntegrationTool } from "./tools/integrations.js";
import { metricsTools, handleMetricsTool } from "./tools/metrics.js";

// ─── Aggregate all tools ──────────────────────────────────────────────────────

const ALL_TOOLS = [
  ...analyzeTools,
  ...jobTools,
  ...integrationTools,
  ...metricsTools,
];

// ─── Server ───────────────────────────────────────────────────────────────────

const server = new Server(
  { name: "reefline", version: "1.0.0" },
  { capabilities: { tools: {} } },
);

// List all available tools
server.setRequestHandler(ListToolsRequestSchema, async () => ({
  tools: ALL_TOOLS,
}));

// Dispatch tool calls to the appropriate module handler
server.setRequestHandler(CallToolRequestSchema, async (req) => {
  const { name, arguments: args } = req.params;

  if (!process.env.REEFLINE_API_URL) {
    process.stderr.write(
      "[reefline-mcp] Warning: REEFLINE_API_URL not set, defaulting to http://localhost:8080\n",
    );
  }

  try {
    const result =
      (await handleAnalyzeTool(name, args)) ??
      (await handleJobTool(name, args)) ??
      (await handleIntegrationTool(name, args)) ??
      (await handleMetricsTool(name, args));

    if (result === null) {
      throw new McpError(ErrorCode.MethodNotFound, `Unknown tool: ${name}`);
    }

    return { content: [{ type: "text", text: result }] };
  } catch (err) {
    if (err instanceof McpError) throw err;
    if (err instanceof z.ZodError) {
      throw new McpError(
        ErrorCode.InvalidParams,
        `Invalid arguments for '${name}': ${err.issues
          .map((i) => `${i.path.join(".")}: ${i.message}`)
          .join(", ")}`,
      );
    }
    throw new McpError(
      ErrorCode.InternalError,
      err instanceof Error ? err.message : String(err),
    );
  }
});

// ─── Start ────────────────────────────────────────────────────────────────────

async function main() {
  const apiUrl = process.env.REEFLINE_API_URL ?? "http://localhost:8080";
  const transport = new StdioServerTransport();
  await server.connect(transport);
  process.stderr.write(`[reefline-mcp] Server started. API: ${apiUrl}\n`);
}

main().catch((err) => {
  process.stderr.write(`[reefline-mcp] Fatal error: ${err}\n`);
  process.exit(1);
});
