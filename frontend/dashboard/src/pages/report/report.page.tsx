import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { Loader2, AlertCircle } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { getJob, type JobReport } from "@/api/jobs.api";
import {
  ReportHeader,
  ReportLayout,
  PlanPanel,
  TabsPanel,
} from "./components";

export function ReportPage() {
  const { jobId } = useParams<{ jobId: string }>();
  const [report, setReport] = useState<JobReport | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!jobId) return;

    const fetchReport = async () => {
      try {
        setError(null);
        const data = await getJob(jobId);
        setReport(data);
      } catch (err: any) {
        setError(err.message || "Failed to load report");
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    fetchReport();
  }, [jobId]);

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center flex-col gap-4">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
        <p className="text-sm text-muted-foreground">Loading report...</p>
      </div>
    );
  }

  if (error || !report || !jobId) {
    return (
      <div className="flex h-screen items-center justify-center flex-col gap-4">
        <div className="h-16 w-16 rounded-full bg-destructive/10 flex items-center justify-center mb-2">
          <AlertCircle className="h-8 w-8 text-destructive" />
        </div>
        <h3 className="text-lg font-medium">Failed to load report</h3>
        <p className="text-muted-foreground text-sm max-w-md text-center">
          {error || "The requested report could not be found."}
        </p>
      </div>
    );
  }

  const proposed = report.report?.proposed;

  // Show different UI based on job status
  if (report.status === "RUNNING" || report.status === "PENDING") {
    return (
      <div className="flex flex-col h-screen">
        <ReportHeader jobId={jobId} status={report.status} imageRef={report.input_scenario} />
        <div className="flex-1 flex items-center justify-center">
          <Card className="max-w-md">
            <CardContent className="flex flex-col items-center justify-center py-12">
              <Loader2 className="h-12 w-12 animate-spin text-primary mb-4" />
              <h3 className="text-lg font-medium mb-2">Analysis in progress</h3>
              <p className="text-sm text-muted-foreground text-center">
                Your container image is being analyzed. This page will automatically update when complete.
              </p>
            </CardContent>
          </Card>
        </div>
      </div>
    );
  }

  if (report.status === "FAILED") {
    return (
      <div className="flex flex-col h-screen">
        <ReportHeader jobId={jobId} status={report.status} imageRef={report.input_scenario} />
        <div className="flex-1 flex items-center justify-center">
          <Card className="max-w-md border-destructive">
            <CardContent className="flex flex-col items-center justify-center py-12">
              <div className="h-16 w-16 rounded-full bg-destructive/10 flex items-center justify-center mb-4">
                <AlertCircle className="h-8 w-8 text-destructive" />
              </div>
              <h3 className="text-lg font-medium mb-2">Analysis Failed</h3>
              <p className="text-sm text-muted-foreground text-center">
                The analysis job encountered an error and could not be completed.
              </p>
            </CardContent>
          </Card>
        </div>
      </div>
    );
  }

  // Completed status - show full report
  return (
    <div className="flex flex-col h-screen overflow-hidden">
      <ReportHeader jobId={jobId} status={report.status} imageRef={report.input_scenario} />

      <ReportLayout
        left={
          <PlanPanel
            recommendations={proposed?.recommendations}
            score={proposed?.score}
          />
        }
        right={<TabsPanel report={report} jobId={jobId} />}
      />
    </div>
  );
}
