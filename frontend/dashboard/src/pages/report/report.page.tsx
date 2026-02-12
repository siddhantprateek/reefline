import { useEffect, useState } from "react"
import { useParams, useNavigate } from "react-router-dom"
import {
  Loader2,
  ArrowLeft,
  Download,
  AlertCircle,
  CheckCircle2,
  TrendingDown,
  Package,
  Shield,
  Zap,
  FileText,
  Info,
  ChevronRight,
} from "lucide-react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import { DottedBackground } from "@/components/custom/header/dotted-background"
import { cn } from "@/lib/utils"
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion"

import {
  getJob,
  type JobReport,
  type JobStatus,
} from "@/api/jobs.api"

const API_BASE = "/api/v1"

const STATUS_CONFIG: Record<JobStatus, { icon: React.ElementType; label: string; color: string }> = {
  COMPLETED: { icon: CheckCircle2, label: "Completed", color: "text-green-600 dark:text-green-400" },
  RUNNING: { icon: Loader2, label: "Running", color: "text-blue-600 dark:text-blue-400 animate-spin" },
  PENDING: { icon: Loader2, label: "Pending", color: "text-yellow-600 dark:text-yellow-400" },
  FAILED: { icon: AlertCircle, label: "Failed", color: "text-red-600 dark:text-red-400" },
  CANCELLED: { icon: AlertCircle, label: "Cancelled", color: "text-gray-600 dark:text-gray-400" },
  SKIPPED: { icon: AlertCircle, label: "Skipped", color: "text-gray-500 dark:text-gray-500" },
  UNKNOWN: { icon: AlertCircle, label: "Unknown", color: "text-gray-400 dark:text-gray-600" },
}

const EFFORT_COLORS: Record<string, string> = {
  low: "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400",
  medium: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400",
  high: "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400",
}

function MetricCard({
  title,
  value,
  icon: Icon,
  description,
  className
}: {
  title: string
  value: string | number
  icon: React.ElementType
  description?: string
  className?: string
}) {
  return (
    <Card className={cn("", className)}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
        {description && (
          <p className="text-xs text-muted-foreground mt-1">{description}</p>
        )}
      </CardContent>
    </Card>
  )
}

function ScoreCard({
  current,
  estimated,
  gradeCurrent,
  gradeEstimated
}: {
  current: number
  estimated: number
  gradeCurrent: string
  gradeEstimated: string
}) {
  const improvement = estimated - current

  return (
    <Card className="border-2 border-primary/20">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Zap className="h-5 w-5 text-primary" />
          Security & Optimization Score
        </CardTitle>
        <CardDescription>Overall health and efficiency rating</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex items-center justify-between gap-8">
          <div className="flex-1">
            <div className="text-sm text-muted-foreground mb-2">Current Score</div>
            <div className="flex items-baseline gap-3">
              <span className="text-4xl font-bold">{current}</span>
              <Badge variant="outline" className="text-lg px-3 py-1">
                {gradeCurrent}
              </Badge>
            </div>
          </div>

          <ChevronRight className="h-8 w-8 text-muted-foreground" />

          <div className="flex-1">
            <div className="text-sm text-muted-foreground mb-2">Estimated After</div>
            <div className="flex items-baseline gap-3">
              <span className="text-4xl font-bold text-primary">{estimated}</span>
              <Badge className="text-lg px-3 py-1 bg-primary">
                {gradeEstimated}
              </Badge>
            </div>
          </div>
        </div>

        <div className="mt-4 flex items-center gap-2 text-sm">
          <TrendingDown className="h-4 w-4 text-green-600 dark:text-green-400" />
          <span className="text-green-600 dark:text-green-400 font-medium">
            +{improvement} point improvement
          </span>
        </div>
      </CardContent>
    </Card>
  )
}

export function ReportPage() {
  const { jobId } = useParams<{ jobId: string }>()
  const navigate = useNavigate()

  const [report, setReport] = useState<JobReport | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!jobId) return

    const fetchReport = async () => {
      try {
        setError(null)
        const data = await getJob(jobId)
        setReport(data)
      } catch (err: any) {
        setError(err.message || "Failed to load report")
        console.error(err)
      } finally {
        setLoading(false)
      }
    }

    fetchReport()
  }, [jobId])

  const handleDownload = (type: string) => {
    if (!jobId) return
    const url = `${API_BASE}/jobs/${jobId}/${type}`
    window.open(url, "_blank")
  }

  if (loading) {
    return (
      <div className="flex h-[calc(100vh-theme(spacing.16))] items-center justify-center flex-col gap-4 text-muted-foreground">
        <Loader2 className="h-8 w-8 animate-spin" />
        <p className="text-sm">Loading report...</p>
      </div>
    )
  }

  if (error || !report) {
    return (
      <div className="flex h-[calc(100vh-theme(spacing.16))] items-center justify-center flex-col gap-4">
        <div className="h-16 w-16 rounded-full bg-destructive/10 flex items-center justify-center mb-2">
          <AlertCircle className="h-8 w-8 text-destructive" />
        </div>
        <h3 className="text-lg font-medium">Failed to load report</h3>
        <p className="text-muted-foreground text-sm max-w-md text-center">
          {error || "The requested report could not be found."}
        </p>
        <Button onClick={() => navigate("/jobs")} variant="outline" className="mt-4">
          <ArrowLeft className="mr-2 h-4 w-4" />
          Back to Jobs
        </Button>
      </div>
    )
  }

  const statusConfig = STATUS_CONFIG[report.status] || STATUS_CONFIG.UNKNOWN
  const StatusIcon = statusConfig.icon

  const measured = report.report?.measured
  const proposed = report.report?.proposed

  return (
    <div className="flex flex-col w-full pb-8">
      {/* Header */}
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
                <h1 className="text-2xl font-medium tracking-tight">Report</h1>
                <p className="text-sm text-muted-foreground mt-1 font-mono">
                  Job ID: {report.job_id}
                </p>
              </div>
            </div>

            <div className="flex items-center gap-3">
              <div className="flex items-center gap-2">
                <StatusIcon className={cn("h-5 w-5", statusConfig.color)} />
                <Badge
                  variant={report.status === "COMPLETED" ? "default" : report.status === "FAILED" ? "destructive" : "secondary"}
                >
                  {statusConfig.label}
                </Badge>
              </div>
            </div>
          </div>
        </DottedBackground>
      </div>

      {/* Content */}
      <div className="px-6 pt-6 max-w-7xl mx-auto w-full space-y-6">

        {report.status === "COMPLETED" && measured && proposed ? (
          <>
            {/* Score Card */}
            {proposed.score && (
              <ScoreCard
                current={proposed.score.current}
                estimated={proposed.score.estimated_after}
                gradeCurrent={proposed.score.grade_current}
                gradeEstimated={proposed.score.grade_estimated}
              />
            )}

            {/* Metrics Grid */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
              <MetricCard
                title="Image Size"
                value={measured.current_size_mb ? `${measured.current_size_mb} MB` : "N/A"}
                icon={Package}
                description={proposed.estimated_size_mb ? `Can be reduced to ~${proposed.estimated_size_mb} MB` : undefined}
              />
              <MetricCard
                title="Vulnerabilities"
                value={measured.total_cves || 0}
                icon={Shield}
                description={`${measured.critical_cves || 0} critical, ${measured.high_cves || 0} high`}
                className={measured.critical_cves && measured.critical_cves > 0 ? "border-red-200 dark:border-red-900" : undefined}
              />
              <MetricCard
                title="Layer Efficiency"
                value={measured.layer_efficiency_pct ? `${measured.layer_efficiency_pct}%` : "N/A"}
                icon={Zap}
                description="Storage optimization"
              />
              {/* <MetricCard
                title="Packages"
                value={measured.total_packages || 0}
                icon={Package}
                description={`${proposed.packages_removable || 0} can be removed`}
              /> */}
            </div>

            {/* Additional Metrics */}
            <div className="grid gap-4 md:grid-cols-3">
              <Card>
                <CardHeader>
                  <CardTitle className="text-sm">Security Issues</CardTitle>
                </CardHeader>
                <CardContent className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Runs as root</span>
                    <Badge variant={measured.runs_as_root ? "destructive" : "outline"}>
                      {measured.runs_as_root ? "Yes" : "No"}
                    </Badge>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Secrets detected</span>
                    <Badge variant={measured.secrets_detected && measured.secrets_detected > 0 ? "destructive" : "outline"}>
                      {measured.secrets_detected || 0}
                    </Badge>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Base image age</span>
                    <span className="font-mono text-xs">
                      {measured.base_image_age_days ? `${measured.base_image_age_days} days` : "N/A"}
                    </span>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="text-sm">Lint Results</CardTitle>
                </CardHeader>
                <CardContent className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Warnings</span>
                    <Badge variant="outline">{measured.lint_warnings || 0}</Badge>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Errors</span>
                    <Badge variant={measured.lint_errors && measured.lint_errors > 0 ? "destructive" : "outline"}>
                      {measured.lint_errors || 0}
                    </Badge>
                  </div>
                </CardContent>
              </Card>

              {/* <Card>
                <CardHeader>
                  <CardTitle className="text-sm">CVE Breakdown</CardTitle>
                </CardHeader>
                <CardContent className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Actionable</span>
                    <Badge variant="outline">{proposed.actionable_cves || 0}</Badge>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Auto-eliminated</span>
                    <Badge variant="outline" className="bg-green-50 dark:bg-green-950">
                      {proposed.auto_eliminated_cves_by_restructure || 0}
                    </Badge>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">No fix available</span>
                    <span className="font-mono text-xs text-muted-foreground">
                      {proposed.no_fix_available_cves || 0}
                    </span>
                  </div>
                </CardContent>
              </Card> */}
            </div>

            {/* Recommendations */}
            {proposed.recommendations && proposed.recommendations.length > 0 && (
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Info className="h-5 w-5" />
                    Recommendations
                  </CardTitle>
                  <CardDescription>
                    Prioritized optimizations to improve your container image
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <Accordion type="single" collapsible className="w-full">
                    {proposed.recommendations.map((rec, idx) => (
                      <AccordionItem key={idx} value={`item-${idx}`}>
                        <AccordionTrigger className="hover:no-underline">
                          <div className="flex items-center gap-3 w-full pr-4">
                            <Badge variant="outline" className="shrink-0">
                              P{rec.priority}
                            </Badge>
                            <div className="flex-1 text-left">
                              <div className="font-medium">{rec.title}</div>
                              <div className="text-xs text-muted-foreground mt-1">
                                {rec.category} • {rec.effort} effort
                                {rec.estimated_size_savings_mb && (
                                  <> • ~{rec.estimated_size_savings_mb}MB savings</>
                                )}
                              </div>
                            </div>
                            <Badge
                              className={cn("shrink-0", EFFORT_COLORS[rec.effort] || "")}
                              variant="outline"
                            >
                              {rec.effort}
                            </Badge>
                          </div>
                        </AccordionTrigger>
                        <AccordionContent className="pt-4 space-y-4">
                          <p className="text-sm text-muted-foreground">{rec.description}</p>

                          {rec.estimated_cves_eliminated && rec.estimated_cves_eliminated > 0 && (
                            <div className="text-sm flex items-center gap-2 text-green-600 dark:text-green-400">
                              <Shield className="h-4 w-4" />
                              Eliminates ~{rec.estimated_cves_eliminated} vulnerabilities
                            </div>
                          )}

                          {rec.before && rec.after && (
                            <div className="space-y-2">
                              <Separator />
                              <div className="grid md:grid-cols-2 gap-4">
                                <div>
                                  <div className="text-xs font-medium mb-2 text-muted-foreground">Before</div>
                                  <pre className="text-xs bg-muted p-3 rounded-md overflow-x-auto">
                                    {rec.before}
                                  </pre>
                                </div>
                                <div>
                                  <div className="text-xs font-medium mb-2 text-primary">After</div>
                                  <pre className="text-xs bg-primary/5 p-3 rounded-md overflow-x-auto border border-primary/20">
                                    {rec.after}
                                  </pre>
                                </div>
                              </div>
                            </div>
                          )}
                        </AccordionContent>
                      </AccordionItem>
                    ))}
                  </Accordion>
                </CardContent>
              </Card>
            )}

            {/* Disclaimer */}
            {proposed.disclaimer && (
              <Card className="border-yellow-200 dark:border-yellow-900 bg-yellow-50/50 dark:bg-yellow-950/20">
                <CardHeader>
                  <CardTitle className="text-sm flex items-center gap-2">
                    <AlertCircle className="h-4 w-4 text-yellow-600 dark:text-yellow-400" />
                    Important Notice
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-muted-foreground">{proposed.disclaimer}</p>
                </CardContent>
              </Card>
            )}

            {/* Download Artifacts */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <FileText className="h-5 w-5" />
                  Download Artifacts
                </CardTitle>
                <CardDescription>
                  Access generated files and detailed analysis data
                </CardDescription>
              </CardHeader>
              <CardContent className="grid gap-3 md:grid-cols-2 lg:grid-cols-4">
                <Button
                  variant="outline"
                  className="justify-start"
                  onClick={() => handleDownload("report")}
                >
                  <Download className="mr-2 h-4 w-4" />
                  Full Report (JSON)
                </Button>
                <Button
                  variant="outline"
                  className="justify-start"
                  onClick={() => handleDownload("dockerfile")}
                >
                  <Download className="mr-2 h-4 w-4" />
                  Optimized Dockerfile
                </Button>
                <Button
                  variant="outline"
                  className="justify-start"
                  onClick={() => handleDownload("sbom")}
                >
                  <Download className="mr-2 h-4 w-4" />
                  SBOM
                </Button>
                <Button
                  variant="outline"
                  className="justify-start"
                  onClick={() => handleDownload("graph")}
                >
                  <Download className="mr-2 h-4 w-4" />
                  Build Graph (SVG)
                </Button>
              </CardContent>
            </Card>
          </>
        ) : report.status === "RUNNING" || report.status === "PENDING" ? (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <Loader2 className="h-12 w-12 animate-spin text-primary mb-4" />
              <h3 className="text-lg font-medium mb-2">Analysis in progress</h3>
              <p className="text-sm text-muted-foreground text-center max-w-md">
                Your container image is being analyzed. This page will automatically update when the analysis is complete.
              </p>
            </CardContent>
          </Card>
        ) : report.status === "FAILED" ? (
          <Card className="border-destructive">
            <CardContent className="flex flex-col items-center justify-center py-12">
              <div className="h-16 w-16 rounded-full bg-destructive/10 flex items-center justify-center mb-4">
                <AlertCircle className="h-8 w-8 text-destructive" />
              </div>
              <h3 className="text-lg font-medium mb-2">Analysis Failed</h3>
              <p className="text-sm text-muted-foreground text-center max-w-md">
                The analysis job encountered an error and could not be completed.
              </p>
            </CardContent>
          </Card>
        ) : (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <AlertCircle className="h-12 w-12 text-muted-foreground mb-4" />
              <h3 className="text-lg font-medium mb-2">No report data available</h3>
              <p className="text-sm text-muted-foreground text-center max-w-md">
                This job has not generated a report yet.
              </p>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  )
}
