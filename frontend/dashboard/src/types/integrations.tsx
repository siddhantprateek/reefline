import {
  SiDocker,
  SiGithub,
  SiAnthropic,
  SiGoogle,
} from "@icons-pack/react-simple-icons";
import { HarborIcon, OpenRouterIcon } from "@/assets/icons";
import { Sparkles, Brain, Container } from "lucide-react";

// Field types for dynamic form generation
export type FieldType = "text" | "password" | "url";

export interface IntegrationField {
  name: string;
  label: string;
  type: FieldType;
  placeholder: string;
  helpText?: string;
  required: boolean;
}

export interface IntegrationSchema {
  id: string;
  name: string;
  description: string;
  category: "Container Registry" | "Version Control" | "AI Provider" | "Orchestration";
  icon: React.ComponentType<{ size?: number; className?: string }>;
  fields: IntegrationField[];
  status?: "connected" | "disconnected";
  /** When true, the integration is auto-detected — no credentials dialog is shown */
  noCredentials?: boolean;
}

// Integration schemas with their required fields
export const integrationSchemas: IntegrationSchema[] = [
  {
    id: "kubernetes",
    name: "Kubernetes",
    description: "Auto-detected when running inside a Kubernetes cluster. Lists all container images across namespaces — no credentials required.",
    category: "Orchestration",
    icon: Container,
    status: "disconnected",
    noCredentials: true,
    fields: [],
  },
  {
    id: "docker",
    name: "Docker",
    description: "Connect to Docker registries to manage and deploy container images efficiently.",
    category: "Container Registry",
    icon: SiDocker,
    status: "disconnected",
    fields: [
      {
        name: "patToken",
        label: "Personal Access Token",
        type: "password",
        placeholder: "dckr_pat_...",
        helpText: "Generate a PAT from Docker Hub > Account Settings > Security",
        required: true,
      },
      {
        name: "username",
        label: "Docker Username",
        type: "text",
        placeholder: "your-username",
        required: true,
      },
    ],
  },
  {
    id: "harbor",
    name: "Harbor",
    description: "Enterprise-grade container registry with security, policy, and lifecycle management.",
    category: "Container Registry",
    icon: HarborIcon,
    status: "disconnected",
    fields: [
      {
        name: "url",
        label: "Harbor URL",
        type: "url",
        placeholder: "https://harbor.example.com",
        required: true,
      },
      {
        name: "username",
        label: "Username",
        type: "text",
        placeholder: "admin",
        required: true,
      },
      {
        name: "password",
        label: "Password",
        type: "password",
        placeholder: "••••••••",
        required: true,
      },
    ],
  },
  {
    id: "github",
    name: "GitHub",
    description: "Integrate with GitHub for repository management, CI/CD, and package registry.",
    category: "Version Control",
    icon: SiGithub,
    status: "disconnected",
    fields: [
      {
        name: "patToken",
        label: "Personal Access Token",
        type: "password",
        placeholder: "ghp_...",
        helpText: "Generate a PAT from GitHub > Settings > Developer settings > Personal access tokens",
        required: true,
      },
    ],
  },
  {
    id: "openai",
    name: "OpenAI",
    description: "Access GPT models and other AI capabilities from OpenAI's API platform.",
    category: "AI Provider",
    icon: Sparkles,
    status: "disconnected",
    fields: [
      {
        name: "apiKey",
        label: "API Key",
        type: "password",
        placeholder: "sk-...",
        helpText: "Find your API key at platform.openai.com/api-keys",
        required: true,
      },
    ],
  },
  {
    id: "anthropic",
    name: "Anthropic",
    description: "Integrate Claude models for advanced AI-powered conversations and analysis.",
    category: "AI Provider",
    icon: SiAnthropic,
    status: "disconnected",
    fields: [
      {
        name: "apiKey",
        label: "API Key",
        type: "password",
        placeholder: "sk-ant-...",
        helpText: "Find your API key at console.anthropic.com",
        required: true,
      },
    ],
  },
  {
    id: "google",
    name: "Google AI",
    description: "Connect to Google's Gemini and other AI services for powerful language models.",
    category: "AI Provider",
    icon: SiGoogle,
    status: "disconnected",
    fields: [
      {
        name: "apiKey",
        label: "API Key",
        type: "password",
        placeholder: "AIza...",
        helpText: "Get your API key from Google AI Studio",
        required: true,
      },
    ],
  },
  {
    id: "openrouter",
    name: "OpenRouter",
    description: "Unified API for accessing multiple AI models from various providers.",
    category: "AI Provider",
    icon: OpenRouterIcon,
    status: "disconnected",
    fields: [
      {
        name: "apiKey",
        label: "API Key",
        type: "password",
        placeholder: "sk-or-...",
        helpText: "Get your API key from openrouter.ai/keys",
        required: true,
      },
    ],
  },
];
