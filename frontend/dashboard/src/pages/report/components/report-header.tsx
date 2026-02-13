import { useNavigate } from "react-router-dom";
import { ArrowLeft, Loader2, CheckCircle2, AlertCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { DottedBackground } from "@/components/custom/header/dotted-background";
import { cn } from "@/lib/utils";
import type { JobStatus } from "@/api/jobs.api";

const STATUS_CONFIG: Record<JobStatus, { icon: React.ElementType; label: string; color: string }> = {
  COMPLETED: { icon: CheckCircle2, label: "Completed", color: "text-green-600 dark:text-green-400" },
  RUNNING: { icon: Loader2, label: "Running", color: "text-blue-600 dark:text-blue-400 animate-spin" },
  PENDING: { icon: Loader2, label: "Pending", color: "text-yellow-600 dark:text-yellow-400" },
  FAILED: { icon: AlertCircle, label: "Failed", color: "text-red-600 dark:text-red-400" },
  CANCELLED: { icon: AlertCircle, label: "Cancelled", color: "text-gray-600 dark:text-gray-400" },
  SKIPPED: { icon: AlertCircle, label: "Skipped", color: "text-gray-500 dark:text-gray-500" },
  UNKNOWN: { icon: AlertCircle, label: "Unknown", color: "text-gray-400 dark:text-gray-600" },
};

interface ReportHeaderProps {
  jobId: string;
  status: JobStatus;
  imageRef?: string;
}

export function ReportHeader({ jobId, status, imageRef }: ReportHeaderProps) {
  const navigate = useNavigate();
  const statusConfig = STATUS_CONFIG[status] || STATUS_CONFIG.UNKNOWN;
  const StatusIcon = statusConfig.icon;

  return (
    <div className="border-b border-border bg-background/50 backdrop-blur-sm sticky top-0 z-10">
      <DottedBackground className="px-6 py-5" y={6}>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Button
              variant="ghost"
              size="icon"
              onClick={() => navigate("/jobs")}
              className="h-8 w-8"
            >
              <ArrowLeft className="h-4 w-4" />
            </Button>
            <div>
              <h1 className="text-2xl font-medium tracking-tight">Analysis Report</h1>
              <p className="text-sm text-muted-foreground mt-1 font-mono">
                {imageRef || jobId}
              </p>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2">
              <StatusIcon className={cn("h-5 w-5", statusConfig.color)} />
              <Badge
                variant={status === "COMPLETED" ? "default" : status === "FAILED" ? "destructive" : "secondary"}
              >
                {statusConfig.label}
              </Badge>
            </div>
          </div>
        </div>
      </DottedBackground>
    </div>
  );
}
