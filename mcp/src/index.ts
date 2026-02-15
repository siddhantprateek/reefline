#!/usr/bin/env node

/**
 * Reefline MCP Server
 *
 * Exposes the Reefline container image security platform via the
 * Model Context Protocol, allowing AI assistants to trigger scans,
 * inspect jobs, manage registry integrations, and retrieve artifacts.
 *
 * Environment variables:
 *   REEFLINE_API_URL  — Base URL of the Reefline server (default: http://localhost:8080)
 *   MCP_TRANSPORT     — "stdio" (default) or "http"
 *   MCP_PORT          — HTTP port when transport=http (default: 4000)
 *
 * CLI flags (override env vars):
 *   --stdio           — Use stdio transport
 *   --http            — Use HTTP/SSE transport
 *   --port <n>        — HTTP port (default: 4000)
 */

import { createServer } from "node:http";
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { StreamableHTTPServerTransport } from "@modelcontextprotocol/sdk/server/streamableHttp.js";
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

// ─── Server factory ───────────────────────────────────────────────────────────

function createMcpServer() {
  const server = new Server(
    { name: "reefline", version: "1.0.0" },
    { capabilities: { tools: {} } },
  );

  server.setRequestHandler(ListToolsRequestSchema, async () => ({
    tools: ALL_TOOLS,
  }));

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

  return server;
}

// ─── Transport selection ──────────────────────────────────────────────────────

function parseArgs() {
  const args = process.argv.slice(2);
  let useHttp = false;
  let port = parseInt(process.env.MCP_PORT ?? "4000", 10);

  if (process.env.MCP_TRANSPORT === "http") useHttp = true;

  for (let i = 0; i < args.length; i++) {
    if (args[i] === "--http") useHttp = true;
    if (args[i] === "--stdio") useHttp = false;
    if (args[i] === "--port" && args[i + 1]) {
      port = parseInt(args[++i], 10);
    }
  }

  return { useHttp, port };
}

// ─── Start ────────────────────────────────────────────────────────────────────

async function startStdio() {
  const server = createMcpServer();
  const transport = new StdioServerTransport();
  await server.connect(transport);
  process.stderr.write(
    `[reefline-mcp] stdio transport ready. API: ${process.env.REEFLINE_API_URL ?? "http://localhost:8080"}\n`,
  );
}

async function startHttp(port: number) {
  // Each HTTP request gets its own stateless transport + server instance
  // so that multiple clients can connect independently.
  const httpServer = createServer(async (req, res) => {
    if (req.url !== "/mcp" || req.method !== "POST") {
      // Also handle GET /mcp for SSE streaming (handled inside the transport)
      if (req.url === "/mcp" && req.method === "GET") {
        // Fall through to transport
      } else if (req.url === "/mcp" && req.method === "DELETE") {
        // Fall through to transport
      } else if (req.url === "/" || req.url === "/health") {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ status: "ok", server: "reefline-mcp" }));
        return;
      } else {
        res.writeHead(404);
        res.end();
        return;
      }
    }

    // sessionIdGenerator: undefined → stateless mode (no session tracking).
    // Each request creates its own transport; sessions are not persisted.
    const transport = new StreamableHTTPServerTransport({
      sessionIdGenerator: undefined,
    });

    const server = createMcpServer();
    await server.connect(transport);
    await transport.handleRequest(req, res);
  });

  httpServer.listen(port, () => {
    process.stderr.write(
      `[reefline-mcp] HTTP transport listening on http://0.0.0.0:${port}/mcp\n`,
    );
    process.stderr.write(
      `[reefline-mcp] API: ${process.env.REEFLINE_API_URL ?? "http://localhost:8080"}\n`,
    );
  });
}

async function main() {
  const { useHttp, port } = parseArgs();
  if (useHttp) {
    await startHttp(port);
  } else {
    await startStdio();
  }
}

main().catch((err) => {
  process.stderr.write(`[reefline-mcp] Fatal error: ${err}\n`);
  process.exit(1);
});
