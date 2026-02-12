import { useEffect, useState, useMemo } from "react"
import { useNavigate } from "react-router-dom"
import {
  Loader2,
  RefreshCw,
  Check,
  Activity,
  XCircle,
  Clock,
  AlertCircle,
  PackageCheck,
  FileText,
  Calendar,
  Trash2,
} from "lucide-react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { DottedBackground } from "@/components/custom/header/dotted-background"
import { cn } from "@/lib/utils"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { useToast } from "@/hooks/use-toast"

import {
  listJobs,
  deleteJob,
  type Job,
  type JobStatus,
} from "@/api/jobs.api"

// --- Types ---

type ScenarioType = "dockerfile_only" | "image_only" | "both"

// --- Components ---

function FilterSection({
  title,
  icon: Icon,
  children
}: {
  title: string
  icon: React.ElementType
  children: React.ReactNode
}) {
  return (
    <div className="border-b border-border">
      <DottedBackground className="border-b border-border/50 dark:bg-neutral-900/20" cy={10}>
        <div className="flex items-center gap-2 px-4 py-2.5">
          <Icon className="h-3.5 w-3.5 text-muted-foreground" />
          <h3 className="text-xs font-semibold tracking-wider uppercase text-muted-foreground">{title}</h3>
        </div>
      </DottedBackground>
      <div className="px-4 py-3 flex flex-wrap gap-2">
        {children}
      </div>
    </div>
  )
}

function CheckboxItem({
  label,
  checked,
  onChange,
  count
}: {
  label: string
  checked: boolean
  onChange: (checked: boolean) => void
  count?: number
}) {
  return (
    <label className="flex items-center gap-3 cursor-pointer group select-none">
      <div
        className={cn(
          "h-4 w-4 rounded border border-input flex items-center justify-center transition-colors shadow-sm",
          checked
            ? "bg-primary border-primary text-primary-foreground"
            : "bg-transparent group-hover:border-primary/50"
        )}
      >
        {checked && <Check className="h-3 w-3" />}
        <input
          type="checkbox"
          className="sr-only"
          checked={checked}
          onChange={(e) => onChange(e.target.checked)}
        />
      </div>
      <span className={cn("text-sm transition-colors", checked ? "font-medium text-foreground" : "text-muted-foreground group-hover:text-foreground")}>
        {label}
      </span>
      {count !== undefined && (
        <span className="ml-auto text-xs text-muted-foreground border px-1.5 py-0.5 rounded-sm bg-muted/30">
          {count}
        </span>
      )}
    </label>
  )
}

function StatusBadge({
  label,
  checked,
  onChange,
  count,
  dotColor
}: {
  label: string
  checked: boolean
  onChange: (checked: boolean) => void
  count?: number
  dotColor?: string
}) {
  return (
    <label
      className={cn(
        "inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full cursor-pointer transition-all border",
        checked
          ? "bg-muted/60 border-border"
          : "bg-background border-border/50 hover:border-border hover:bg-muted/30"
      )}
    >
      <div
        className={cn(
          "h-2 w-2 rounded-full shrink-0",
          dotColor || "bg-muted-foreground"
        )}
      />
      <span className={cn("text-xs font-medium whitespace-nowrap", checked ? "text-foreground" : "text-muted-foreground")}>
        {label}
      </span>
      {count !== undefined && (
        <span className={cn("text-[10px] tabular-nums", checked ? "text-muted-foreground" : "text-muted-foreground/60")}>
          {count}
        </span>
      )}
      <input
        type="checkbox"
        className="sr-only"
        checked={checked}
        onChange={(e) => onChange(e.target.checked)}
      />
    </label>
  )
}

const STATUS_CONFIG: Record<JobStatus, { icon: React.ElementType; label: string; color: string; dotColor: string }> = {
  COMPLETED: { icon: Check, label: "Completed", color: "text-green-600 dark:text-green-400", dotColor: "bg-green-500" },
  RUNNING: { icon: Activity, label: "Running", color: "text-blue-600 dark:text-blue-400", dotColor: "bg-blue-500" },
  PENDING: { icon: Clock, label: "Pending", color: "text-yellow-600 dark:text-yellow-400", dotColor: "bg-yellow-500" },
  FAILED: { icon: XCircle, label: "Failed", color: "text-red-600 dark:text-red-400", dotColor: "bg-red-500" },
  CANCELLED: { icon: AlertCircle, label: "Cancelled", color: "text-gray-600 dark:text-gray-400", dotColor: "bg-gray-500" },
  SKIPPED: { icon: AlertCircle, label: "Skipped", color: "text-gray-500 dark:text-gray-500", dotColor: "bg-orange-500" },
  UNKNOWN: { icon: AlertCircle, label: "Unknown", color: "text-gray-400 dark:text-gray-600", dotColor: "bg-gray-400" },
}

interface JobRowProps {
  job: Job
  onClick: (job: Job) => void
  onDelete: (jobId: string, e: React.MouseEvent) => void
}

function JobRow({ job, onClick, onDelete }: JobRowProps) {
  const statusConfig = STATUS_CONFIG[job.status] || STATUS_CONFIG.UNKNOWN
  const StatusIcon = statusConfig.icon

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
  }

  const getDuration = () => {
    if (!job.created_at) return null
    const start = new Date(job.created_at).getTime()
    const end = job.completed_at ? new Date(job.completed_at).getTime() : Date.now()
    const durationMs = end - start
    const seconds = Math.floor(durationMs / 1000)
    const minutes = Math.floor(seconds / 60)

    if (minutes > 0) return `${minutes}m ${seconds % 60}s`
    return `${seconds}s`
  }

  return (
    <div
      onClick={() => onClick(job)}
      className="group flex items-center justify-between gap-4 border-b last:border-0 border-border p-4 hover:bg-muted/40 transition-colors cursor-pointer"
    >
      {/* Left: Status & Image */}
      <div className="flex items-start gap-4 min-w-[350px] max-w-[40%]">
        <div className={cn(
          "flex h-10 w-10 shrink-0 items-center justify-center rounded-full border-2 bg-background",
          job.status === "COMPLETED" && "border-green-500/20 bg-green-500/5",
          job.status === "RUNNING" && "border-blue-500/20 bg-blue-500/5",
          job.status === "FAILED" && "border-red-500/20 bg-red-500/5",
          job.status === "PENDING" && "border-yellow-500/20 bg-yellow-500/5",
          job.status === "UNKNOWN" && "border-gray-500/20 bg-gray-500/5"
        )}>
          <StatusIcon className={cn("h-5 w-5", statusConfig.color)} />
        </div>

        <div className="space-y-1 min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <h3 className="font-medium text-sm truncate font-mono" title={job.job_id}>
              {job.job_id}
            </h3>
          </div>
          <div className="flex items-center gap-2">
            <p className="text-xs text-muted-foreground truncate font-mono" title={job.image_ref}>
              {job.image_ref || "No image reference"}
            </p>
          </div>
        </div>
      </div>

      {/* Middle: Metadata */}
      <div className="hidden md:flex flex-1 items-center gap-4 px-4">
        {job.scenario && (
          <Badge variant="outline" className="px-2 py-0 text-[10px] h-5 font-normal border-border bg-background capitalize">
            {job.scenario.replace('_', ' ')}
          </Badge>
        )}
        {job.status === "RUNNING" && job.progress !== undefined && (
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <div className="w-20 h-1.5 bg-muted rounded-full overflow-hidden">
              <div
                className="h-full bg-primary transition-all duration-300"
                style={{ width: `${job.progress}%` }}
              />
            </div>
            <span>{job.progress}%</span>
          </div>
        )}
      </div>

      {/* Right: Time & Actions */}
      <div className="flex items-center gap-6 shrink-0 text-sm text-muted-foreground">
        <div className="hidden lg:flex flex-col items-end gap-1 text-xs min-w-[140px]">
          <div className="flex items-center gap-1.5">
            <Calendar className="h-3 w-3" />
            <span>{job.created_at ? formatDate(job.created_at) : "N/A"}</span>
          </div>
          {getDuration() && (
            <div className="flex items-center gap-1.5 text-muted-foreground/70">
              <Clock className="h-3 w-3" />
              <span>{getDuration()}</span>
            </div>
          )}
        </div>

        <Badge
          variant={job.status === "COMPLETED" ? "default" : job.status === "FAILED" ? "destructive" : "secondary"}
          className="text-xs min-w-[80px] justify-center"
        >
          {statusConfig.label}
        </Badge>

        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity hover:text-destructive"
                onClick={(e) => onDelete(job.job_id, e)}
                aria-label="Delete job"
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              <p>Delete job</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </div>
    </div>
  )
}

export function JobsPage() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const [jobs, setJobs] = useState<Job[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Delete confirmation dialog
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [jobToDelete, setJobToDelete] = useState<string | null>(null)

  // Filters
  const [searchQuery, setSearchQuery] = useState("")
  const [selectedStatuses, setSelectedStatuses] = useState<JobStatus[]>([])
  const [selectedScenarios, setSelectedScenarios] = useState<ScenarioType[]>([])

  // Fetch logic
  const fetchJobs = async () => {
    try {
      setError(null)
      const data = await listJobs()
      setJobs(data)
    } catch (err: any) {
      setError(err.message || "Failed to load jobs")
      console.error(err)
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }

  useEffect(() => {
    fetchJobs()
  }, [])

  const handleRefresh = async () => {
    setRefreshing(true)
    await fetchJobs()
    if (!error) {
      toast({
        title: "Jobs refreshed",
        description: "The job list has been updated successfully.",
      })
    }
  }

  const handleJobClick = (job: Job) => {
    navigate(`/jobs/${job.job_id}`)
  }

  const handleDeleteClick = (jobId: string, e: React.MouseEvent) => {
    e.stopPropagation()
    setJobToDelete(jobId)
    setDeleteDialogOpen(true)
  }

  const handleDeleteConfirm = async () => {
    if (!jobToDelete) return

    try {
      await deleteJob(jobToDelete)
      setJobs(prev => prev.filter(j => j.job_id !== jobToDelete))
      toast({
        title: "Job deleted",
        description: "The analysis job has been deleted successfully.",
      })
    } catch (err: any) {
      toast({
        title: "Failed to delete job",
        description: err.message || "An error occurred while deleting the job.",
        variant: "destructive",
      })
    } finally {
      setDeleteDialogOpen(false)
      setJobToDelete(null)
    }
  }

  // Filtering
  const filteredJobs = useMemo(() => {
    return jobs.filter(job => {
      // 1. Search Query
      if (searchQuery && !job.image_ref?.toLowerCase().includes(searchQuery.toLowerCase()) && !job.job_id.toLowerCase().includes(searchQuery.toLowerCase())) {
        return false
      }
      // 2. Status Filter
      if (selectedStatuses.length > 0 && !selectedStatuses.includes(job.status)) {
        return false
      }
      // 3. Scenario Filter
      if (selectedScenarios.length > 0 && job.scenario && !selectedScenarios.includes(job.scenario)) {
        return false
      }
      return true
    })
  }, [jobs, searchQuery, selectedStatuses, selectedScenarios])

  // Stats for sidebar counts
  const statusCounts = useMemo(() => {
    const counts: Record<JobStatus, number> = {
      COMPLETED: 0,
      RUNNING: 0,
      PENDING: 0,
      FAILED: 0,
      CANCELLED: 0,
      SKIPPED: 0,
      UNKNOWN: 0,
    }
    jobs.forEach(job => {
      counts[job.status] = (counts[job.status] || 0) + 1
    })
    return counts
  }, [jobs])

  const scenarioCounts = useMemo(() => {
    const counts: Record<string, number> = {}
    jobs.forEach(job => {
      if (job.scenario) {
        counts[job.scenario] = (counts[job.scenario] || 0) + 1
      }
    })
    return counts
  }, [jobs])

  const handleStatusToggle = (status: JobStatus, checked: boolean) => {
    setSelectedStatuses(prev => {
      if (checked) {
        return [...prev, status]
      } else {
        return prev.filter(s => s !== status)
      }
    })
  }

  const handleScenarioToggle = (scenario: ScenarioType, checked: boolean) => {
    setSelectedScenarios(prev => {
      if (checked) {
        return [...prev, scenario]
      } else {
        return prev.filter(s => s !== scenario)
      }
    })
  }

  const isAllStatuses = selectedStatuses.length === 0
  const isAllScenarios = selectedScenarios.length === 0

  return (
    <div className="flex flex-col h-[calc(100vh-theme(spacing.16))] w-full">
      {/* Page Header */}
      <div className="flex items-center justify-between px-6 py-5 border-b border-border bg-background/50 backdrop-blur-sm sticky top-0 z-10">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Analysis Jobs</h1>
          <p className="text-sm text-muted-foreground mt-1">
            View and manage all container image analysis jobs.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Input
            placeholder="Search jobs..."
            className="w-64 h-9"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
          <Button
            variant="outline"
            size="sm"
            onClick={handleRefresh}
            disabled={loading || refreshing}
            className="h-9"
          >
            <RefreshCw
              className={cn("mr-2 h-3.5 w-3.5", refreshing && "animate-spin")}
            />
            {refreshing ? "Refreshing..." : "Refresh"}
          </Button>
        </div>
      </div>

      <div className="flex flex-1 overflow-hidden">
        {/* Sidebar Filters */}
        <aside className="w-64 border-r border-border bg-card/30 flex flex-col overflow-y-auto [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-accent/30 [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb:hover]:bg-muted-background/50">

          <FilterSection title="Status" icon={Activity}>
            <StatusBadge
              label="All"
              checked={isAllStatuses}
              onChange={(_checked) => setSelectedStatuses([])}
              count={jobs.length}
              dotColor="bg-muted-foreground"
            />
            <StatusBadge
              label="Success"
              checked={selectedStatuses.includes('COMPLETED')}
              onChange={(c) => handleStatusToggle('COMPLETED', c)}
              count={statusCounts.COMPLETED}
              dotColor={STATUS_CONFIG.COMPLETED.dotColor}
            />
            <StatusBadge
              label="Failed"
              checked={selectedStatuses.includes('FAILED')}
              onChange={(c) => handleStatusToggle('FAILED', c)}
              count={statusCounts.FAILED}
              dotColor={STATUS_CONFIG.FAILED.dotColor}
            />
            <StatusBadge
              label="Running"
              checked={selectedStatuses.includes('RUNNING')}
              onChange={(c) => handleStatusToggle('RUNNING', c)}
              count={statusCounts.RUNNING}
              dotColor={STATUS_CONFIG.RUNNING.dotColor}
            />
            <StatusBadge
              label="Pending"
              checked={selectedStatuses.includes('PENDING')}
              onChange={(c) => handleStatusToggle('PENDING', c)}
              count={statusCounts.PENDING}
              dotColor={STATUS_CONFIG.PENDING.dotColor}
            />
            <StatusBadge
              label="Cancelled"
              checked={selectedStatuses.includes('CANCELLED')}
              onChange={(c) => handleStatusToggle('CANCELLED', c)}
              count={statusCounts.CANCELLED}
              dotColor={STATUS_CONFIG.CANCELLED.dotColor}
            />
            <StatusBadge
              label="Skipped"
              checked={selectedStatuses.includes('SKIPPED')}
              onChange={(c) => handleStatusToggle('SKIPPED', c)}
              count={statusCounts.SKIPPED}
              dotColor={STATUS_CONFIG.SKIPPED.dotColor}
            />
          </FilterSection>

          <FilterSection title="Scenario" icon={FileText}>
            <CheckboxItem
              label="All"
              checked={isAllScenarios}
              onChange={() => setSelectedScenarios([])}
              count={jobs.length}
            />
            <CheckboxItem
              label="Dockerfile Only"
              checked={selectedScenarios.includes('dockerfile_only')}
              onChange={(c) => handleScenarioToggle('dockerfile_only', c)}
              count={scenarioCounts.dockerfile_only || 0}
            />
            <CheckboxItem
              label="Image Only"
              checked={selectedScenarios.includes('image_only')}
              onChange={(c) => handleScenarioToggle('image_only', c)}
              count={scenarioCounts.image_only || 0}
            />
            <CheckboxItem
              label="Both"
              checked={selectedScenarios.includes('both')}
              onChange={(c) => handleScenarioToggle('both', c)}
              count={scenarioCounts.both || 0}
            />
          </FilterSection>

        </aside>

        {/* Main List Content */}
        <main className="flex-1 overflow-y-auto [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-accent/30 [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb:hover]:bg-muted-background/50 dark:bg-[#151716]">
          {error && (
            <div className="m-6 mb-2 rounded-md bg-destructive/10 p-4 text-sm text-destructive border border-destructive/20 flex items-center gap-2">
              <span className="font-semibold">Error:</span> {error}
            </div>
          )}

          {loading ? (
            <div className="flex h-64 items-center justify-center flex-col gap-4 text-muted-foreground">
              <Loader2 className="h-8 w-8 animate-spin" />
              <p className="text-sm">Loading jobs...</p>
            </div>
          ) : filteredJobs.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-64 text-center">
              <div className="h-16 w-16 rounded-full bg-muted/50 flex items-center justify-center mb-4">
                <PackageCheck className="h-8 w-8 text-muted-foreground" />
              </div>
              <h3 className="text-lg font-medium">No jobs found</h3>
              <p className="text-muted-foreground text-sm max-w-sm mt-2">
                {jobs.length === 0
                  ? "You haven't run any analysis jobs yet. Start by analyzing a container image from the Overview page."
                  : "We couldn't find any jobs matching your current filters. Try adjusting your search or filters."}
              </p>
              {(searchQuery || selectedStatuses.length > 0 || selectedScenarios.length > 0) && (
                <Button
                  variant="link"
                  onClick={() => {
                    setSearchQuery("");
                    setSelectedStatuses([]);
                    setSelectedScenarios([])
                  }}
                  className="mt-4"
                >
                  Clear all filters
                </Button>
              )}
            </div>
          ) : (
            <div className="flex flex-col">
              {filteredJobs.map(job => (
                <JobRow
                  key={job.id}
                  job={job}
                  onClick={handleJobClick}
                  onDelete={handleDeleteClick}
                />
              ))}
            </div>
          )}
        </main>
      </div>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Job</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete this job? This action cannot be undone. All associated artifacts and results will be permanently deleted.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setDeleteDialogOpen(false)}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleDeleteConfirm}
            >
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
