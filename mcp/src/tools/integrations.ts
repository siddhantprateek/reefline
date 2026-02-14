import { z } from "zod";
import { apiRequest } from "../client.js";

// ─── Schemas ──────────────────────────────────────────────────────────────────

const IntegrationIdArgs = z.object({
  id: z
    .string()
    .describe(
      "Integration ID — one of: docker, github, harbor, kubernetes, openai, anthropic, google, openrouter",
    ),
});

const ConnectArgs = z.object({
  id: z.string().describe("Integration ID to connect"),
  credentials: z
    .record(z.string())
    .describe(
      "Key-value credential map. Required keys by provider: " +
        "GitHub → token; " +
        "Docker Hub → username, password; " +
        "Harbor → url, username, password; " +
        "AI providers → api_key; " +
        "Kubernetes → (none, auto-detected from in-cluster config)",
    ),
});

const ListImagesArgs = z.object({
  provider: z
    .enum(["github", "docker", "harbor", "kubernetes"])
    .describe("Which registry or cluster to list images from"),
  namespace: z
    .string()
    .optional()
    .describe(
      "Kubernetes only — filter to a specific namespace (omit for all namespaces)",
    ),
});

// ─── Tool definitions ─────────────────────────────────────────────────────────

export const integrationTools = [
  {
    name: "get_integrations",
    description:
      "List all configured integrations (Docker Hub, GitHub, Harbor, Kubernetes, AI providers) and their connection statuses.",
    inputSchema: {
      type: "object",
      properties: {},
      required: [],
    },
  },
  {
    name: "get_integration_by_id",
    description:
      "Get details and connection status of a specific integration by ID.",
    inputSchema: {
      type: "object",
      properties: {
        id: {
          type: "string",
          description:
            "Integration ID — one of: docker, github, harbor, kubernetes, openai, anthropic, google, openrouter",
        },
      },
      required: ["id"],
    },
  },
  {
    name: "connect_integration",
    description: [
      "Connect (configure) an integration by providing its credentials.",
      "Required credential keys by provider:",
      "  • github      → token",
      "  • docker      → username, password",
      "  • harbor      → url, username, password",
      "  • openai / anthropic / google / openrouter → api_key",
      "  • kubernetes  → (no credentials — auto-detected from in-cluster config)",
    ].join("\n"),
    inputSchema: {
      type: "object",
      properties: {
        id: { type: "string", description: "Integration ID to connect" },
        credentials: {
          type: "object",
          description: "Key-value credential map (keys depend on provider)",
          additionalProperties: { type: "string" },
        },
      },
      required: ["id", "credentials"],
    },
  },
  {
    name: "disconnect_integration",
    description:
      "Disconnect an integration and remove its stored credentials from Reefline.",
    inputSchema: {
      type: "object",
      properties: {
        id: { type: "string", description: "Integration ID to disconnect" },
      },
      required: ["id"],
    },
  },
  {
    name: "test_integration",
    description:
      "Re-validate stored credentials for an integration by making a live call to the provider API.",
    inputSchema: {
      type: "object",
      properties: {
        id: { type: "string", description: "Integration ID to test" },
      },
      required: ["id"],
    },
  },
  {
    name: "list_images",
    description: [
      "List container images from any connected registry or cluster.",
      "Provider routing:",
      "  • github     → GHCR container images for the authenticated user/org",
      "  • docker     → Docker Hub repositories for the authenticated user/org",
      "  • harbor     → Projects in the connected Harbor registry",
      "  • kubernetes → All running container images in the cluster (all pods + init containers)",
      "Pass 'namespace' only when provider is 'kubernetes' to scope results to one namespace.",
    ].join("\n"),
    inputSchema: {
      type: "object",
      properties: {
        provider: {
          type: "string",
          enum: ["github", "docker", "harbor", "kubernetes"],
          description: "Which registry or cluster to list images from",
        },
        namespace: {
          type: "string",
          description:
            "Kubernetes only — filter to a specific namespace (omit for all namespaces)",
        },
      },
      required: ["provider"],
    },
  },
  {
    name: "get_kubernetes_status",
    description:
      "Check if Reefline is running inside a Kubernetes cluster and retrieve cluster metadata: server version, node count, namespace count. No credentials needed — auto-detected via the mounted service account token.",
    inputSchema: {
      type: "object",
      properties: {},
      required: [],
    },
  },
  {
    name: "list_kubernetes_namespaces",
    description: "List all namespace names in the connected Kubernetes cluster.",
    inputSchema: {
      type: "object",
      properties: {},
      required: [],
    },
  },
] as const;

// ─── Handler ──────────────────────────────────────────────────────────────────

const INTEGRATION_TOOL_NAMES = new Set([
  "get_integrations",
  "get_integration_by_id",
  "connect_integration",
  "disconnect_integration",
  "test_integration",
  "list_images",
  "get_kubernetes_status",
  "list_kubernetes_namespaces",
]);

export async function handleIntegrationTool(
  name: string,
  args: unknown,
): Promise<string | null> {
  if (!INTEGRATION_TOOL_NAMES.has(name)) return null;

  switch (name) {
    case "get_integrations": {
      const result = await apiRequest<unknown>("/integrations");
      return JSON.stringify(result, null, 2);
    }

    case "get_integration_by_id": {
      const { id } = IntegrationIdArgs.parse(args);
      const result = await apiRequest<unknown>(`/integrations/${id}`);
      return JSON.stringify(result, null, 2);
    }

    case "connect_integration": {
      const { id, credentials } = ConnectArgs.parse(args);
      const result = await apiRequest<unknown>(`/integrations/${id}/connect`, {
        method: "POST",
        body: JSON.stringify(credentials),
      });
      return JSON.stringify(result, null, 2);
    }

    case "disconnect_integration": {
      const { id } = IntegrationIdArgs.parse(args);
      const result = await apiRequest<unknown>(
        `/integrations/${id}/disconnect`,
        { method: "POST" },
      );
      return JSON.stringify(result, null, 2);
    }

    case "test_integration": {
      const { id } = IntegrationIdArgs.parse(args);
      const result = await apiRequest<unknown>(`/integrations/${id}/test`, {
        method: "POST",
      });
      return JSON.stringify(result, null, 2);
    }

    case "list_images": {
      const { provider, namespace } = ListImagesArgs.parse(args);

      const providerRoutes: Record<typeof provider, string> = {
        github:     "/integrations/github/images",
        docker:     "/integrations/docker/repos",
        harbor:     "/integrations/harbor/projects",
        kubernetes: `/integrations/kubernetes/images${namespace ? `?namespace=${encodeURIComponent(namespace)}` : ""}`,
      };

      const result = await apiRequest<unknown>(providerRoutes[provider]);
      return JSON.stringify(result, null, 2);
    }

    case "get_kubernetes_status": {
      const status = await apiRequest<unknown>(
        "/integrations/kubernetes/status",
      );
      return JSON.stringify(status, null, 2);
    }

    case "list_kubernetes_namespaces": {
      const result = await apiRequest<unknown>(
        "/integrations/kubernetes/namespaces",
      );
      return JSON.stringify(result, null, 2);
    }

    default:
      return null;
  }
}
