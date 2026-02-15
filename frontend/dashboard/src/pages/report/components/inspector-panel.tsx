import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Package, Shield, Zap, ChevronDown, ChevronRight, Layers } from "lucide-react";
import { Inspector } from "./inspector";
import { ImageInfoCard } from "./image-info-card";
import type { JobReport, GrypeReport, DiveReport, DockleReport } from "@/types/jobs";

interface InspectorPanelProps {
  report: JobReport;
  grype: GrypeReport | null;
  dive: DiveReport | null;
  dockle: DockleReport | null;
  onShowAllVulnerabilities?: () => void;
}

const SEVERITY_CONFIG = {
  Critical: { color: "bg-red-600", textColor: "text-red-600 dark:text-red-400", badge: "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400" },
  High: { color: "bg-orange-500", textColor: "text-orange-500 dark:text-orange-400", badge: "bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400" },
  Medium: { color: "bg-yellow-500", textColor: "text-yellow-500 dark:text-yellow-400", badge: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400" },
  Low: { color: "bg-blue-500", textColor: "text-blue-400", badge: "bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400" },
  Unknown: { color: "bg-muted-foreground", textColor: "text-muted-foreground", badge: "bg-secondary text-muted-foreground" },
} as const;

function getGrypeSeverityCounts(grype: GrypeReport) {
  const counts: Record<string, number> = { Critical: 0, High: 0, Medium: 0, Low: 0, Unknown: 0 };
  for (const row of grype.Table.Rows ?? []) {
    const severity = row[5];
    if (severity in counts) counts[severity]++;
    else counts.Unknown++;
  }
  return counts;
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`;
}

export function InspectorPanel({ report, grype, dive, dockle, onShowAllVulnerabilities }: InspectorPanelProps) {
  const [inspectorOpen, setInspectorOpen] = useState(false);
  const measured = report.report?.measured;
  const severityCounts = grype ? getGrypeSeverityCounts(grype) : null;
  const totalCves = severityCounts ? Object.values(severityCounts).reduce((a, b) => a + b, 0) : 0;
  return (
    <div className="p-6 space-y-6">
      {/* Inspector Toggle / Image Info */}
      <div>
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-medium flex items-center gap-2">
            <Layers className="h-5 w-5" />
            Image Inspector
          </h3>
          {dive && (
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
          )}
        </div>

        {inspectorOpen && dive ? (
          <Inspector dive={dive} />
        ) : (
          <ImageInfoCard report={report} dive={dive} />
        )}
      </div>

      {/* Vulnerability Summary from Grype */}
      {severityCounts && totalCves > 0 && (
        <div>
          <div className="flex items-center justify-between mb-3">
            <h3 className="text-lg font-semibold flex items-center gap-2">
              <Shield className="h-5 w-5" />
              Vulnerability Summary
              <Badge variant="outline" className="ml-1">{totalCves} total</Badge>
            </h3>
            {onShowAllVulnerabilities && (
              <Button variant="outline" size="sm" onClick={onShowAllVulnerabilities}>
                Show all
              </Button>
            )}
          </div>
          <div className="grid grid-cols-5 gap-2">
            {(Object.keys(SEVERITY_CONFIG) as Array<keyof typeof SEVERITY_CONFIG>).map((sev) => (
              <Card key={sev} className="shadow-none border border-border">
                <CardContent className="py-0 px-3 flex flex-col h-full min-h-24">
                  <span className="text-xs uppercase font-medium text-muted-foreground mb-auto">{sev}</span>
                  <div className="mt-auto">
                    <p className={`text-3xl font-light mb-1 ${SEVERITY_CONFIG[sev].textColor}`}>
                      {severityCounts[sev] ?? 0}
                    </p>
                    <div className="w-full h-1 bg-secondary rounded-full">
                      <div className={`h-1 rounded-full ${SEVERITY_CONFIG[sev].color}`} style={{ width: "100%" }} />
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          {/* Top affected packages */}
          {grype && grype.Table.Rows && grype.Table.Rows.length > 0 && (
            <div className="mt-3 border rounded-lg overflow-hidden">
              <div className="px-3 py-2 bg-muted/30 text-xs font-medium text-muted-foreground">
                Top CVEs
              </div>
              <div className="divide-y">
                {grype.Table.Rows.slice(0, 5).map((row, idx) => (
                  <div key={idx} className="flex items-center justify-between px-3 py-2 text-xs">
                    <div className="flex items-center gap-2 min-w-0">
                      <span className="font-medium truncate">{row[0]}</span>
                      <span className="text-muted-foreground shrink-0">{row[4]}</span>
                    </div>
                    <Badge
                      className={`shrink-0 ml-2 border-0 ${SEVERITY_CONFIG[row[5] as keyof typeof SEVERITY_CONFIG]?.badge ?? "bg-secondary text-muted-foreground"}`}
                    >
                      {row[5]}
                    </Badge>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Key Metrics */}
      {measured && (
        <>
          <div>
            <h3 className="text-lg font-semibold mb-4">Key Metrics</h3>
            <div className="grid gap-4 grid-cols-2">
              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Image Size</CardTitle>
                  <Package className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">
                    {dive ? formatBytes(dive.sizeBytes) : measured.current_size_mb ? `${measured.current_size_mb} MB` : "N/A"}
                  </div>
                  <p className="text-xs text-muted-foreground mt-1">Current image size</p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Vulnerabilities</CardTitle>
                  <Shield className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{totalCves || measured.total_cves || 0}</div>
                  <p className="text-xs text-muted-foreground mt-1">
                    {severityCounts
                      ? `${severityCounts.Critical} critical, ${severityCounts.High} high`
                      : `${measured.critical_cves || 0} critical, ${measured.high_cves || 0} high`}
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
                    {dive ? `${(dive.efficiency * 100).toFixed(1)}%` : measured.layer_efficiency_pct ? `${measured.layer_efficiency_pct}%` : "N/A"}
                  </div>
                  <p className="text-xs text-muted-foreground mt-1">
                    {dive ? `${formatBytes(dive.wastedBytes)} wasted` : "Storage optimization"}
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
                  <p className="text-xs text-muted-foreground mt-1">Installed packages</p>
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
