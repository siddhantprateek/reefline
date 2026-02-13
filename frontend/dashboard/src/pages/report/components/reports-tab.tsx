import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Package, Shield, Zap, AlertTriangle, ChevronDown, ChevronRight, Layers } from "lucide-react";
import { Inspector } from "./inspector";
import { ImageInfoCard } from "./image-info-card";
import type { JobReport } from "@/api/jobs.api";

interface ReportsTabProps {
  report: JobReport;
}

export function ReportsTab({ report }: ReportsTabProps) {
  const [inspectorOpen, setInspectorOpen] = useState(false);
  const measured = report.report?.measured;

  return (
    <div className="p-6 space-y-6">
      {/* Inspector Toggle / Image Info */}
      <div>
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold flex items-center gap-2">
            <Layers className="h-5 w-5" />
            Image Inspector
          </h3>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setInspectorOpen(!inspectorOpen)}
          >
            {inspectorOpen ? (
              <>
                <ChevronDown className="h-4 w-4 mr-2" />
                Collapse
              </>
            ) : (
              <>
                <ChevronRight className="h-4 w-4 mr-2" />
                Inspect Layers
              </>
            )}
          </Button>
        </div>

        {inspectorOpen ? (
          <Inspector report={report} />
        ) : (
          <ImageInfoCard report={report} />
        )}
      </div>

      {/* Metrics Grid */}
      {measured && (
        <>
          <div>
            <h3 className="text-lg font-semibold mb-4">Key Metrics</h3>
            <div className="grid gap-4 md:grid-cols-2">
              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Image Size</CardTitle>
                  <Package className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">
                    {measured.current_size_mb ? `${measured.current_size_mb} MB` : "N/A"}
                  </div>
                  <p className="text-xs text-muted-foreground mt-1">
                    Current image size
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Vulnerabilities</CardTitle>
                  <Shield className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{measured.total_cves || 0}</div>
                  <p className="text-xs text-muted-foreground mt-1">
                    {measured.critical_cves || 0} critical, {measured.high_cves || 0} high
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Layer Efficiency</CardTitle>
                  <Zap className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">
                    {measured.layer_efficiency_pct ? `${measured.layer_efficiency_pct}%` : "N/A"}
                  </div>
                  <p className="text-xs text-muted-foreground mt-1">
                    Storage optimization
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Total Packages</CardTitle>
                  <Package className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{measured.total_packages || 0}</div>
                  <p className="text-xs text-muted-foreground mt-1">
                    Installed packages
                  </p>
                </CardContent>
              </Card>
            </div>
          </div>

          {/* Security Analysis */}
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">Security Analysis</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Runs as root</span>
                <Badge variant={measured.runs_as_root ? "destructive" : "outline"}>
                  {measured.runs_as_root ? "Yes" : "No"}
                </Badge>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Secrets detected</span>
                <Badge variant={measured.secrets_detected && measured.secrets_detected > 0 ? "destructive" : "outline"}>
                  {measured.secrets_detected || 0}
                </Badge>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Base image age</span>
                <span className="font-mono text-xs">
                  {measured.base_image_age_days ? `${measured.base_image_age_days} days` : "N/A"}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Lint warnings</span>
                <Badge variant="outline">{measured.lint_warnings || 0}</Badge>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Lint errors</span>
                <Badge variant={measured.lint_errors && measured.lint_errors > 0 ? "destructive" : "outline"}>
                  {measured.lint_errors || 0}
                </Badge>
              </div>
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}
