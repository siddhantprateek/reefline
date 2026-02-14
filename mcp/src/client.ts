import { ErrorCode, McpError } from "@modelcontextprotocol/sdk/types.js";

const API_URL = (
  process.env.REEFLINE_API_URL ?? "http://localhost:8080"
).replace(/\/$/, "");

export const BASE = `${API_URL}/api/v1`;

/**
 * Make a request to the Reefline API.
 * - JSON responses are parsed automatically.
 * - Non-JSON responses (e.g. Markdown files) are returned as plain strings.
 */
export async function apiRequest<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const url = `${BASE}${path}`;

  const res = await fetch(url, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
      ...(options.headers ?? {}),
    },
  });

  if (!res.ok) {
    let body = "";
    try {
      body = await res.text();
    } catch {
      // ignore
    }
    throw new McpError(
      ErrorCode.InternalError,
      `Reefline API ${res.status} ${res.statusText}: ${body}`,
    );
  }

  const contentType = res.headers.get("content-type") ?? "";
  if (contentType.includes("application/json")) {
    return res.json() as Promise<T>;
  }
  return res.text() as unknown as T;
}
