import { Loader2, CheckCircle2, AlertCircle, Clock } from "lucide-react";
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
  scanTime?: string;
}

export function ReportHeader({ jobId, status, imageRef, scanTime }: ReportHeaderProps) {
  const statusConfig = STATUS_CONFIG[status] || STATUS_CONFIG.UNKNOWN;
  const StatusIcon = statusConfig.icon;

  return (
    <div className="border-b border-border bg-background/50 backdrop-blur-sm sticky top-0 z-10">
      <DottedBackground className="px-6 py-2" y={6}>
        <div className="flex items-center justify-between w-full">
          <div className="flex items-center gap-4">
            <div>
              <h1 className="text-2xl font-ligh tracking-tight">Report</h1>
              <p className="text-sm text-muted-foreground font-mono">
                {imageRef || jobId}
              </p>
            </div>
          </div>

          <div className="flex items-center gap-3">
            {scanTime && (
              <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                <Clock className="h-3.5 w-3.5" />
                <span>{new Date(scanTime).toLocaleString()}</span>
              </div>
            )}
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
