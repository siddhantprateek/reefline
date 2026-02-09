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
  <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className}>
    <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8zm-5-9h10v2H7v-2z" />
  </svg>
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
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
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

