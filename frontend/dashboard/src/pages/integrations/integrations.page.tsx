import { useState } from "react";
import { IntegrationCard, IntegrationSetupDialog } from "@/components/custom";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Search, Filter } from "lucide-react";
import { integrationSchemas, type IntegrationSchema } from "@/types/integrations";

export function IntegrationsPage() {
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedIntegration, setSelectedIntegration] = useState<IntegrationSchema | null>(null);
  const [isDialogOpen, setIsDialogOpen] = useState(false);

  const handleSetup = (integration: IntegrationSchema) => {
    setSelectedIntegration(integration);
    setIsDialogOpen(true);
  };

  const handleSave = async (values: Record<string, string>) => {
    console.log("Saving credentials for", selectedIntegration?.name, ":", values);
    // TODO: Implement actual API call to save the integration
    await new Promise((resolve) => setTimeout(resolve, 1000));
  };

  const handleTest = async (values: Record<string, string>): Promise<boolean> => {
    console.log("Testing credentials for", selectedIntegration?.name, ":", values);
    // TODO: Implement actual API call to test the integration
    await new Promise((resolve) => setTimeout(resolve, 1500));
    // Simulate success/failure based on all required fields having values
    const allFieldsFilled = Object.values(values).every((v) => v && v.length > 3);
    return allFieldsFilled;
  };

  const handleDisable = (integration: IntegrationSchema) => {
    console.log("Disabling", integration.name);
    // TODO: Implement disable logic
  };

  const handleRemove = (integration: IntegrationSchema) => {
    console.log("Removing", integration.name);
    // TODO: Implement remove logic
  };

  const filteredIntegrations = integrationSchemas.filter(
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
            onDisable={() => handleDisable(integration)}
            onRemove={() => handleRemove(integration)}
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
          fields={selectedIntegration.fields}
          onSave={handleSave}
          onTest={handleTest}
        />
      )}
    </div>
  );
}
