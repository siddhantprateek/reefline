/**
 * Integration API client
 *
 * All calls go through the Vite dev proxy (/api/v1 → http://localhost:8080/api/v1)
 * to avoid CORS issues during development.
 */

const API_BASE = "/api/v1/integrations";

// ─── Types ───────────────────────────────────────────────────────────────────

export interface IntegrationStatus {
  id: string;
  status: "connected" | "disconnected" | "error";
  connected_at?: string;
  metadata?: Record<string, unknown>;
}

export interface IntegrationListResponse {
  integrations: IntegrationStatus[];
}

export interface ConnectResponse {
  id: string;
  status: "connected" | "error";
  metadata?: Record<string, unknown>;
  error?: string;
}

export interface DisconnectResponse {
  id: string;
  status: "disconnected";
}

export interface TestConnectionResponse {
  id: string;
  status: "connected" | "error";
  latency_ms?: number;
  error?: string;
}

// GitHub-specific types
export interface GitHubRepo {
  id: number;
  name: string;
  full_name: string;
  description: string;
  private: boolean;
  html_url: string;
  default_branch: string;
}

export interface GitHubContainerImage {
  id: number;
  name: string;
  package_type: string;
  html_url: string;
  tags: string[];
}

export interface GitHubIssue {
  id: number;
  number: number;
  title: string;
  html_url: string;
}

// Docker Hub-specific types
export interface DockerHubRepo {
  name: string;
  namespace: string;
  description: string;
  is_private: boolean;
  star_count: number;
  pull_count: number;
  last_updated: string;
}

export interface DockerHubTag {
  name: string;
  full_size: number;
  last_updated: string;
  digest: string;
}

// Kubernetes-specific types
export interface KubernetesContainerImage {
  image: string;
  pod_name: string;
  namespace: string;
  container_name: string;
  is_init: boolean;
}

export interface KubernetesStatus {
  id: string;
  status: "connected" | "disconnected" | "error";
  available: boolean;
  message?: string;
  error?: string;
  metadata?: {
    server_version: string;
    node_count: number;
    namespace_count: number;
  };
}

// Harbor-specific types
export interface HarborProject {
  project_id: number;
  name: string;
  repo_count: number;
  owner_name: string;
}

export interface HarborArtifact {
  id: number;
  digest: string;
  size: number;
  tags: Array<{ id: number; name: string }>;
  push_time: string;
}

// ─── Helper ──────────────────────────────────────────────────────────────────

/**
 * Base64-encode a credentials object for secure transport.
 * The backend expects: { "data": "<base64-encoded JSON>", "test_only": bool }
 */
function encodeCredentials(credentials: Record<string, string>): string {
  const json = JSON.stringify(credentials);
  return btoa(json);
}

class ApiError extends Error {
  statusCode: number;

  constructor(
    statusCode: number,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
    this.statusCode = statusCode;
  }
}

async function request<T>(
  url: string,
  options: RequestInit = {},
): Promise<T> {
  const res = await fetch(url, {
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
    ...options,
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({ message: res.statusText }));
    throw new ApiError(
      res.status,
      body.message || body.error || `Request failed: ${res.status}`,
    );
  }

  return res.json() as Promise<T>;
}

// ─── Generic Integration CRUD ────────────────────────────────────────────────

/**
 * List all integrations with their connection status.
 * Merges backend state with the known integration schemas.
 */
export async function listIntegrations(): Promise<IntegrationListResponse> {
  return request<IntegrationListResponse>(API_BASE);
}

/**
 * Get a specific integration's details.
 */
export async function getIntegration(id: string): Promise<IntegrationStatus> {
  return request<IntegrationStatus>(`${API_BASE}/${id}`);
}

/**
 * Connect an integration by providing credentials.
 * The backend validates the credentials before saving.
 *
 * @param id - Integration ID (e.g., "github", "docker", "harbor", "openai")
 * @param credentials - Key/value pairs matching the integration's field schema
 */
export async function connectIntegration(
  id: string,
  credentials: Record<string, string>,
): Promise<ConnectResponse> {
  return request<ConnectResponse>(`${API_BASE}/${id}/connect`, {
    method: "POST",
    body: JSON.stringify({
      data: encodeCredentials(credentials),
      test_only: false,
    }),
  });
}

/**
 * Disconnect an integration, removing stored credentials.
 */
export async function disconnectIntegration(
  id: string,
): Promise<DisconnectResponse> {
  return request<DisconnectResponse>(`${API_BASE}/${id}/disconnect`, {
    method: "POST",
  });
}

/**
 * Test an existing integration's credentials without modifying them.
 */
export async function testIntegrationConnection(
  id: string,
): Promise<TestConnectionResponse> {
  return request<TestConnectionResponse>(`${API_BASE}/${id}/test`, {
    method: "POST",
  });
}

/**
 * Test credentials before saving — sends the credentials to the connect endpoint
 * but with a test-only flag. Falls back to a direct test if available.
 *
 * @param id - Integration ID
 * @param credentials - Credentials to test
 */
export async function testCredentials(
  id: string,
  credentials: Record<string, string>,
): Promise<TestConnectionResponse> {
  // Send base64-encoded credentials with test_only flag
  return request<TestConnectionResponse>(`${API_BASE}/${id}/connect`, {
    method: "POST",
    body: JSON.stringify({
      data: encodeCredentials(credentials),
      test_only: true,
    }),
  });
}

// ─── GitHub-specific endpoints ───────────────────────────────────────────────

/**
 * List GitHub repositories accessible to the connected PAT.
 */
export async function listGitHubRepos(
  page = 1,
  perPage = 20,
): Promise<GitHubRepo[]> {
  return request<GitHubRepo[]>(
    `${API_BASE}/github/repos?page=${page}&per_page=${perPage}`,
  );
}

/**
 * Fetch a Dockerfile from a GitHub repository.
 *
 * @param owner - Repository owner (user or org)
 * @param repo  - Repository name
 * @param path  - Path to Dockerfile within the repo (optional, defaults to "Dockerfile")
 * @param ref   - Branch, tag, or commit SHA (optional, defaults to default branch)
 */
export async function getGitHubDockerfile(
  owner: string,
  repo: string,
  path?: string,
  ref?: string,
): Promise<{ content: string; path: string; sha: string }> {
  const params = new URLSearchParams();
  if (path) params.set("path", path);
  if (ref) params.set("ref", ref);
  const qs = params.toString();
  return request(
    `${API_BASE}/github/repos/${owner}/${repo}/dockerfile${qs ? `?${qs}` : ""}`,
  );
}

/**
 * List container images published to GHCR.
 */
export async function listGitHubContainerImages(
  owner?: string,
  page = 1,
  perPage = 20,
): Promise<GitHubContainerImage[]> {
  const params = new URLSearchParams({
    page: String(page),
    per_page: String(perPage),
  });
  if (owner) params.set("owner", owner);
  return request<GitHubContainerImage[]>(
    `${API_BASE}/github/images?${params}`,
  );
}

/**
 * Create a GitHub issue with optimization recommendations from an analysis job.
 */
export async function createGitHubIssue(
  owner: string,
  repo: string,
  jobId: string,
  title?: string,
  labels?: string[],
): Promise<GitHubIssue> {
  return request<GitHubIssue>(
    `${API_BASE}/github/repos/${owner}/${repo}/issues`,
    {
      method: "POST",
      body: JSON.stringify({
        job_id: jobId,
        title: title || "Container Image Optimization Report",
        labels: labels || ["optimization"],
      }),
    },
  );
}

// ─── Docker Hub-specific endpoints ───────────────────────────────────────────

/**
 * List Docker Hub repositories for the connected account.
 */
export async function listDockerHubRepos(
  page = 1,
  pageSize = 20,
): Promise<DockerHubRepo[]> {
  return request<DockerHubRepo[]>(
    `${API_BASE}/docker/repos?page=${page}&page_size=${pageSize}`,
  );
}

/**
 * List tags for a Docker Hub repository.
 */
export async function listDockerHubTags(
  namespace: string,
  repo: string,
  page = 1,
  pageSize = 20,
): Promise<DockerHubTag[]> {
  return request<DockerHubTag[]>(
    `${API_BASE}/docker/repos/${namespace}/${repo}/tags?page=${page}&page_size=${pageSize}`,
  );
}

// ─── Harbor-specific endpoints ───────────────────────────────────────────────

/**
 * List projects from the connected Harbor instance.
 */
export async function listHarborProjects(
  page = 1,
  pageSize = 20,
): Promise<HarborProject[]> {
  return request<HarborProject[]>(
    `${API_BASE}/harbor/projects?page=${page}&page_size=${pageSize}`,
  );
}

/**
 * List artifacts for a Harbor repository.
 */
export async function listHarborArtifacts(
  project: string,
  repo: string,
  page = 1,
  pageSize = 20,
): Promise<HarborArtifact[]> {

  return request<HarborArtifact[]>(
    `${API_BASE}/harbor/projects/${project}/repos/${repo}/artifacts?page=${page}&page_size=${pageSize}`,
  );
}

// ─── Kubernetes-specific endpoints ───────────────────────────────────────────

/**
 * Get Kubernetes in-cluster status and cluster metadata.
 * Returns connected/disconnected based on whether the app is running in-cluster.
 */
export async function getKubernetesStatus(): Promise<KubernetesStatus> {
  return request<KubernetesStatus>(`${API_BASE}/kubernetes/status`);
}

/**
 * List all container images currently running in the Kubernetes cluster.
 * Optionally filter by namespace.
 */
export async function listKubernetesImages(
  namespace?: string,
): Promise<KubernetesContainerImage[]> {
  const params = namespace ? `?namespace=${encodeURIComponent(namespace)}` : "";
  return request<KubernetesContainerImage[]>(
    `${API_BASE}/kubernetes/images${params}`,
  );
}

/**
 * List all namespaces in the Kubernetes cluster.
 */
export async function listKubernetesNamespaces(): Promise<{ namespaces: string[] }> {
  return request<{ namespaces: string[] }>(`${API_BASE}/kubernetes/namespaces`);
}
