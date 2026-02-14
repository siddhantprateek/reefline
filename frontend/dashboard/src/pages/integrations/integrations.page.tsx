import { useState, useEffect, useCallback } from "react";
import { IntegrationCard, IntegrationSetupDialog } from "@/components/custom";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Search, Filter } from "lucide-react";
import { integrationSchemas, type IntegrationSchema } from "@/types/integrations";
import {
  listIntegrations,
  connectIntegration,
  testCredentials,
  disconnectIntegration,
  type IntegrationStatus,
} from "@/api/integration.api";

export function IntegrationsPage() {
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedIntegration, setSelectedIntegration] = useState<IntegrationSchema | null>(null);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [statuses, setStatuses] = useState<Record<string, IntegrationStatus>>({});

  // Fetch integration statuses from the backend on mount
  const fetchStatuses = useCallback(async () => {
    try {
      const response = await listIntegrations();
      const statusMap: Record<string, IntegrationStatus> = {};
      response.integrations.forEach((s) => {
        statusMap[s.id] = s;
      });
      setStatuses(statusMap);
    } catch (err) {
      // Backend may not be running yet — silently fall back to default statuses
      console.warn("Failed to fetch integration statuses:", err);
    }
  }, []);

  useEffect(() => {
    fetchStatuses();
  }, [fetchStatuses]);

  // Merge backend statuses with the static schema definitions
  const integrationsWithStatus: IntegrationSchema[] = integrationSchemas.map((schema) => ({
    ...schema,
    status: statuses[schema.id]?.status === "connected" ? "connected" : "disconnected",
  }));

  const handleSetup = (integration: IntegrationSchema) => {
    // Kubernetes (and any noCredentials integration) is auto-detected — nothing to configure
    if (integration.noCredentials) return;
    setSelectedIntegration(integration);
    setIsDialogOpen(true);
  };

  const handleSave = async (values: Record<string, string>) => {
    if (!selectedIntegration) return;

    try {
      const response = await connectIntegration(selectedIntegration.id, values);

      if (response.status === "connected") {
        // Update local status immediately
        setStatuses((prev) => ({
          ...prev,
          [selectedIntegration.id]: {
            id: selectedIntegration.id,
            status: "connected",
            connected_at: new Date().toISOString(),
            metadata: response.metadata,
          },
        }));
        console.log(`✅ ${selectedIntegration.name} connected successfully`);
      } else {
        throw new Error(response.error || "Connection failed");
      }
    } catch (err) {
      console.error(`❌ Failed to connect ${selectedIntegration.name}:`, err);
      throw err; // Re-throw so the dialog shows the error state
    }
  };

  const handleTest = async (values: Record<string, string>): Promise<boolean> => {
    if (!selectedIntegration) return false;

    try {
      const response = await testCredentials(selectedIntegration.id, values);
      return response.status === "connected";
    } catch (err) {
      console.error("Credential test failed:", err);
      return false;
    }
  };

  const handleDisable = async (integration: IntegrationSchema) => {
    try {
      await disconnectIntegration(integration.id);

      // Update local status immediately
      setStatuses((prev) => ({
        ...prev,
        [integration.id]: {
          id: integration.id,
          status: "disconnected",
        },
      }));
      console.log(`🔌 ${integration.name} disconnected`);
    } catch (err) {
      console.error(`❌ Failed to disconnect ${integration.name}:`, err);
    }
  };

  const handleRemove = async (integration: IntegrationSchema) => {
    try {
      await disconnectIntegration(integration.id);

      // Update local status immediately
      setStatuses((prev) => ({
        ...prev,
        [integration.id]: {
          id: integration.id,
          status: "disconnected",
        },
      }));
      console.log(`🗑️ ${integration.name} removed`);
    } catch (err) {
      console.error(`❌ Failed to remove ${integration.name}:`, err);
    }
  };

  const filteredIntegrations = integrationsWithStatus.filter(
    (integration) =>
      integration.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      integration.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
      integration.category.toLowerCase().includes(searchQuery.toLowerCase()),
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
            noCredentials={integration.noCredentials}
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
