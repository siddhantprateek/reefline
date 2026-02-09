import { useState } from "react";
import { IntegrationCard, IntegrationSetupDialog } from "@/components/custom";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Search, Filter } from "lucide-react";
import {
  SiDocker,
  SiGithub,
  SiOpenai,
  SiAnthropic,
  SiGoogle,
} from "@icons-pack/react-simple-icons";

// Custom icons for Harbor and OpenRouter (not in react-simple-icons)
const HarborIcon = ({ size = 24, className = "" }: { size?: number; className?: string }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className}>
    <path d="M12 2L2 7v10c0 5.55 3.84 10.74 10 12 6.16-1.26 10-6.45 10-12V7l-10-5zm0 2.18l8 4v8.82c0 4.52-2.98 8.69-8 9.92-5.02-1.23-8-5.4-8-9.92V8.18l8-4z" />
  </svg>
);

const OpenRouterIcon = ({ size = 24, className = "" }: { size?: number; className?: string }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg"><title>OpenRouter</title><path d="M16.778 1.844v1.919q-.569-.026-1.138-.032-.708-.008-1.415.037c-1.93.126-4.023.728-6.149 2.237-2.911 2.066-2.731 1.95-4.14 2.75-.396.223-1.342.574-2.185.798-.841.225-1.753.333-1.751.333v4.229s.768.108 1.61.333c.842.224 1.789.575 2.185.799 1.41.798 1.228.683 4.14 2.75 2.126 1.509 4.22 2.11 6.148 2.236.88.058 1.716.041 2.555.005v1.918l7.222-4.168-7.222-4.17v2.176c-.86.038-1.611.065-2.278.021-1.364-.09-2.417-.357-3.979-1.465-2.244-1.593-2.866-2.027-3.68-2.508.889-.518 1.449-.906 3.822-2.59 1.56-1.109 2.614-1.377 3.978-1.466.667-.044 1.418-.017 2.278.02v2.176L24 6.014Z" /></svg>
);

interface Integration {
  id: string;
  name: string;
  description: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
  category: string;
  status?: "connected" | "disconnected";
}

const integrations: Integration[] = [
  {
    id: "docker",
    name: "Docker",
    description: "Connect to Docker registries to manage and deploy container images efficiently.",
    icon: SiDocker,
    category: "Container Registry",
    status: "disconnected",
  },
  {
    id: "harbor",
    name: "Harbor",
    description: "Enterprise-grade container registry with security, policy, and lifecycle management.",
    icon: HarborIcon,
    category: "Container Registry",
    status: "disconnected",
  },
  {
    id: "github",
    name: "GitHub",
    description: "Integrate with GitHub for repository management, CI/CD, and package registry.",
    icon: SiGithub,
    category: "Version Control",
    status: "disconnected",
  },
  {
    id: "openai",
    name: "OpenAI",
    description: "Access GPT models and other AI capabilities from OpenAI's API platform.",
    icon: SiOpenai,
    category: "AI Provider",
    status: "disconnected",
  },
  {
    id: "anthropic",
    name: "Anthropic",
    description: "Integrate Claude models for advanced AI-powered conversations and analysis.",
    icon: SiAnthropic,
    category: "AI Provider",
    status: "disconnected",
  },
  {
    id: "google",
    name: "Google AI",
    description: "Connect to Google's Gemini and other AI services for powerful language models.",
    icon: SiGoogle,
    category: "AI Provider",
    status: "disconnected",
  },
  {
    id: "openrouter",
    name: "OpenRouter",
    description: "Unified API for accessing multiple AI models from various providers.",
    icon: OpenRouterIcon,
    category: "AI Provider",
    status: "disconnected",
  },
];

export function IntegrationsPage() {
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedIntegration, setSelectedIntegration] = useState<Integration | null>(null);
  const [isDialogOpen, setIsDialogOpen] = useState(false);

  const handleSetup = (integration: Integration) => {
    setSelectedIntegration(integration);
    setIsDialogOpen(true);
  };

  const handleSave = async (apiKey: string) => {
    console.log("Saving API key for", selectedIntegration?.name, ":", apiKey);
    // TODO: Implement actual API call to save the integration
    await new Promise((resolve) => setTimeout(resolve, 1000));
  };

  const handleTest = async (apiKey: string): Promise<boolean> => {
    console.log("Testing API key for", selectedIntegration?.name, ":", apiKey);
    // TODO: Implement actual API call to test the integration
    await new Promise((resolve) => setTimeout(resolve, 1500));
    // Simulate success/failure
    return apiKey.length > 10;
  };

  const filteredIntegrations = integrations.filter(
    (integration) =>
      integration.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      integration.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
      integration.category.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="flex flex-col">
      {/* Header */}
      <div className="p-4 md:p-6">
        <h1 className="text-3xl font-medium tracking-tight bg-gradient-to-br from-foreground to-foreground/70 bg-clip-text">
          Integrations
        </h1>
        <p className="text-muted-foreground">
          Connect your tools and services to streamline your workflow
        </p>
      </div>

      {/* Search and Filter Bar */}
      <div className="flex">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search for integration - search bar"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9 rounded-none border-t-2"
          />
        </div>
        <Button variant="outline">
          <Filter className="h-4 w-4" />
          Filter
        </Button>
      </div>

      {/* Integration Cards Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 border-l border-t border-border">
        {filteredIntegrations.map((integration) => (
          <IntegrationCard
            key={integration.id}
            name={integration.name}
            description={integration.description}
            icon={integration.icon}
            category={integration.category}
            status={integration.status}
            onSetup={() => handleSetup(integration)}
          />
        ))}
      </div>

      {/* No Results */}
      {filteredIntegrations.length === 0 && (
        <div className="text-center">
          <p className="text-muted-foreground">
            No integrations found matching "{searchQuery}"
          </p>
        </div>
      )}

      {/* Setup Dialog */}
      {selectedIntegration && (
        <IntegrationSetupDialog
          open={isDialogOpen}
          onOpenChange={setIsDialogOpen}
          integrationName={selectedIntegration.name}
          integrationIcon={selectedIntegration.icon}
          onSave={handleSave}
          onTest={handleTest}
        />
      )}
    </div>
  );
}

