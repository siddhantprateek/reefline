import { useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { FileText, Package, X } from "lucide-react";
import { ReportsTab } from "./reports-tab";
import { ArtifactsTab } from "./artifacts-tab";
import type { JobReport } from "@/api/jobs.api";

const TABS = [
  { id: "reports", label: "Report.md", icon: FileText },
  { id: "artifacts", label: "Artifacts", icon: Package },
] as const;

type TabId = typeof TABS[number]["id"];

interface TabsPanelProps {
  report: JobReport;
  jobId: string;
}

export function TabsPanel({ report, jobId }: TabsPanelProps) {
  const [activeTab, setActiveTab] = useState<TabId>("reports");

  return (
    <div className="h-full flex flex-col">
      {/* Tab Headers */}
      <div className="border-b border-border bg-muted/30 sticky top-0 z-10">
        <div className="flex gap-0.5 px-2 pt-1">
          {TABS.map((tab) => {
            const Icon = tab.icon;
            return (
              <div
                key={tab.id}
                className={cn(
                  "group flex items-center gap-2 px-3 py-2 text-sm border-t border-x rounded-t-md transition-colors relative",
                  activeTab === tab.id
                    ? "bg-background border-border text-foreground"
                    : "bg-transparent border-transparent text-muted-foreground hover:text-foreground hover:bg-muted/50"
                )}
              >
                <button
                  onClick={() => setActiveTab(tab.id)}
                  className="flex items-center gap-2 flex-1"
                >
                  <span className="text-sm">{tab.label}</span>
                </button>
                <Icon className="h-3.5 w-3.5 shrink-0" />
                <button
                  className={cn(
                    "h-4 w-4 rounded-sm hover:bg-muted flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity",
                    activeTab === tab.id && "opacity-70 hover:opacity-100"
                  )}
                  onClick={(e) => {
                    e.stopPropagation();
                    // Handle close if needed
                  }}
                >
                  <X className="h-3 w-3" />
                </button>
              </div>
            );
          })}
        </div>
      </div>

      {/* Tab Content */}
      <div className="flex-1 overflow-y-auto bg-background">
        {activeTab === "reports" && <ReportsTab report={report} />}
        {activeTab === "artifacts" && <ArtifactsTab jobId={jobId} />}
      </div>
    </div>
  );
}
