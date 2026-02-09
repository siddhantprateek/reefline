import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Loader2 } from "lucide-react";

export interface IntegrationSetupDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  integrationName: string;
  integrationIcon: React.ComponentType<{ size?: number; className?: string }>;
  onSave: (apiKey: string) => Promise<void>;
  onTest?: (apiKey: string) => Promise<boolean>;
}

export function IntegrationSetupDialog({
  open,
  onOpenChange,
  integrationName,
  integrationIcon: Icon,
  onSave,
  onTest,
}: IntegrationSetupDialogProps) {
  const [apiKey, setApiKey] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [isTesting, setIsTesting] = useState(false);
  const [testStatus, setTestStatus] = useState<"success" | "error" | null>(null);
  const [testMessage, setTestMessage] = useState("");

  const handleTestAndSave = async () => {
    if (!apiKey.trim()) return;

    // First test if onTest is provided
    if (onTest) {
      setIsTesting(true);
      setTestStatus(null);
      setTestMessage("");

      try {
        const result = await onTest(apiKey);
        setTestStatus(result ? "success" : "error");
        setTestMessage(
          result
            ? "Connection successful! API key is valid."
            : "Connection failed. Please check your API key."
        );

        // If test fails, don't proceed to save
        if (!result) {
          setIsTesting(false);
          return;
        }
      } catch (error) {
        setTestStatus("error");
        setTestMessage(error instanceof Error ? error.message : "Test failed");
        setIsTesting(false);
        return;
      } finally {
        setIsTesting(false);
      }
    }

    // Then save
    setIsLoading(true);
    try {
      await onSave(apiKey);
      setApiKey("");
      setTestStatus(null);
      setTestMessage("");
      onOpenChange(false);
    } catch (error) {
      setTestStatus("error");
      setTestMessage(error instanceof Error ? error.message : "Failed to save");
    } finally {
      setIsLoading(false);
    }
  };

  const handleCancel = () => {
    setApiKey("");
    setTestStatus(null);
    setTestMessage("");
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <div className="flex items-center gap-3 mb-2">
            <div className="p-2 rounded-lg bg-gradient-to-br from-primary/10 to-primary/5 border border-primary/20">
              <Icon size={24} className="text-primary" />
            </div>
            <DialogTitle className="text-xl">Setup {integrationName}</DialogTitle>
          </div>
          <DialogDescription className="text-sm text-muted-foreground">
            Enter your API key to connect {integrationName} to your account. Your
            credentials will be securely stored and encrypted.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <Label htmlFor="api-key" className="text-sm font-medium">
              API Key
            </Label>
            <Input
              id="api-key"
              type="password"
              placeholder="sk-..."
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              className="font-mono text-sm"
              disabled={isLoading || isTesting}
            />
            <p className="text-xs text-muted-foreground">
              You can find your API key in your {integrationName} dashboard.
            </p>
          </div>

          {testStatus && (
            <div
              className={`p-3 rounded-lg border text-sm ${testStatus === "success"
                ? "bg-green-50 border-green-200 text-green-800 dark:bg-green-950 dark:border-green-800 dark:text-green-200"
                : "bg-red-50 border-red-200 text-red-800 dark:bg-red-950 dark:border-red-800 dark:text-red-200"
                }`}
            >
              <div className="flex items-center gap-2">
                <Badge
                  variant={testStatus === "success" ? "default" : "destructive"}
                  className="text-xs"
                >
                  {testStatus === "success" ? "Success" : "Error"}
                </Badge>
                <span>{testMessage}</span>
              </div>
            </div>
          )}
        </div>

        <DialogFooter className="gap-3">
          <Button
            type="button"
            variant="outline"
            onClick={handleCancel}
            disabled={isLoading || isTesting}
          >
            Cancel
          </Button>
          <Button
            type="button"
            onClick={handleTestAndSave}
            disabled={!apiKey.trim() || isLoading || isTesting}
          >
            {isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Saving...
              </>
            ) : (
              "Test & Save"
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
