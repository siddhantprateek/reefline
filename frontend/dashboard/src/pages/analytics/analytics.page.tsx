import { useState, useEffect } from "react";
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";
import {
  getQueueStats,
  getJobMetrics,
  getToolPerformance,
  type QueueStats,
  type JobMetrics,
  type ToolPerformance,
  type TimeRange,
} from "@/api/metrics.api";
import { Activity, CheckCircle2, XCircle, Clock, TrendingUp } from "lucide-react";

const TIME_RANGES: TimeRange[] = ["24h", "7d", "30d"];

const STATUS_COLORS = {
  COMPLETED: "#22c55e",
  FAILED: "#ef4444",
  RUNNING: "#3b82f6",
  PENDING: "#f59e0b",
};

export function AnalyticsPage() {
  const [timeRange, setTimeRange] = useState<TimeRange>("24h");
  const [queueStats, setQueueStats] = useState<QueueStats | null>(null);
  const [jobMetrics, setJobMetrics] = useState<JobMetrics | null>(null);
  const [toolPerformance, setToolPerformance] = useState<ToolPerformance | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchMetrics();
  }, [timeRange]);

  const fetchMetrics = async () => {
    setLoading(true);
    setError(null);
    try {
      const [queue, jobs, tools] = await Promise.all([
        getQueueStats(),
        getJobMetrics(timeRange),
        getToolPerformance(),
      ]);
      setQueueStats(queue);
      setJobMetrics(jobs);
      setToolPerformance(tools);
    } catch (err: any) {
      setError(err.message || "Failed to fetch metrics");
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-center">
          <Activity className="h-8 w-8 animate-spin mx-auto mb-4 text-primary" />
          <p className="text-muted-foreground">Loading analytics...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-center">
          <XCircle className="h-8 w-8 mx-auto mb-4 text-destructive" />
          <p className="text-destructive">{error}</p>
          <button
            onClick={fetchMetrics}
            className="mt-4 px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  // Prepare pie chart data
  const statusDistributionData = jobMetrics
    ? Object.entries(jobMetrics.status_distribution).map(([name, value]) => ({
        name,
        value,
      }))
    : [];

  // Prepare duration breakdown data
  const durationBreakdownData = jobMetrics
    ? [
        { name: "Queue Wait", value: Math.round(jobMetrics.duration_breakdown.avg_queue_ms / 1000) },
        { name: "Grype", value: Math.round(jobMetrics.duration_breakdown.avg_grype_ms / 1000) },
        { name: "Dockle", value: Math.round(jobMetrics.duration_breakdown.avg_dockle_ms / 1000) },
        { name: "Dive", value: Math.round(jobMetrics.duration_breakdown.avg_dive_ms / 1000) },
      ]
    : [];

  // Prepare tool performance data
  const toolPerformanceData = toolPerformance
    ? Object.entries(toolPerformance.tools).map(([name, stats]) => ({
        name: name.charAt(0).toUpperCase() + name.slice(1),
        duration: Math.round(stats.avg_duration_ms / 1000),
        successRate: stats.success_rate_pct,
      }))
    : [];

  // Prepare time series data
  const timeSeriesData = jobMetrics?.time_series.map((point) => ({
    timestamp: new Date(point.timestamp).toLocaleString("en-US", {
      month: "short",
      day: "numeric",
      hour: timeRange === "24h" ? "numeric" : undefined,
    }),
    completed: point.completed,
    failed: point.failed,
  }));

  return (
    <div className="container mx-auto px-6 py-8">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Analytics Dashboard</h1>
        <p className="text-muted-foreground">
          Performance metrics and insights for your container analysis jobs
        </p>
      </div>

      {/* Time Range Selector */}
      <div className="flex gap-2 mb-6">
        {TIME_RANGES.map((range) => (
          <button
            key={range}
            onClick={() => setTimeRange(range)}
            className={`px-4 py-2 rounded-md transition-colors ${
              timeRange === range
                ? "bg-primary text-primary-foreground"
                : "bg-secondary text-secondary-foreground hover:bg-secondary/80"
            }`}
          >
            {range}
          </button>
        ))}
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4 mb-8">
        <SummaryCard
          icon={Activity}
          title="Total Jobs"
          value={jobMetrics?.summary.total || 0}
          iconColor="text-blue-500"
        />
        <SummaryCard
          icon={CheckCircle2}
          title="Success Rate"
          value={`${jobMetrics?.summary.success_rate_pct.toFixed(1) || 0}%`}
          iconColor="text-green-500"
        />
        <SummaryCard
          icon={Clock}
          title="Avg Duration"
          value={`${Math.round((jobMetrics?.summary.avg_total_ms || 0) / 1000)}s`}
          iconColor="text-purple-500"
        />
        <SummaryCard
          icon={TrendingUp}
          title="Queue Depth"
          value={queueStats?.pending || 0}
          iconColor="text-orange-500"
        />
        <SummaryCard
          icon={Activity}
          title="Throughput"
          value={`${queueStats?.throughput_per_hour.toFixed(1) || 0}/hr`}
          iconColor="text-cyan-500"
        />
      </div>

      {/* Charts Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        {/* Jobs Over Time */}
        <ChartCard title="Jobs Over Time">
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={timeSeriesData}>
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
              <XAxis dataKey="timestamp" className="text-xs" />
              <YAxis className="text-xs" />
              <Tooltip
                contentStyle={{
                  backgroundColor: "hsl(var(--background))",
                  border: "1px solid hsl(var(--border))",
                }}
              />
              <Legend />
              <Line
                type="monotone"
                dataKey="completed"
                stroke="#22c55e"
                strokeWidth={2}
                name="Completed"
              />
              <Line
                type="monotone"
                dataKey="failed"
                stroke="#ef4444"
                strokeWidth={2}
                name="Failed"
              />
            </LineChart>
          </ResponsiveContainer>
        </ChartCard>

        {/* Status Distribution */}
        <ChartCard title="Status Distribution">
          <ResponsiveContainer width="100%" height={300}>
            <PieChart>
              <Pie
                data={statusDistributionData}
                cx="50%"
                cy="50%"
                labelLine={false}
                label={(entry) => `${entry.name}: ${entry.value}`}
                outerRadius={100}
                fill="#8884d8"
                dataKey="value"
              >
                {statusDistributionData.map((entry) => (
                  <Cell
                    key={`cell-${entry.name}`}
                    fill={STATUS_COLORS[entry.name as keyof typeof STATUS_COLORS]}
                  />
                ))}
              </Pie>
              <Tooltip
                contentStyle={{
                  backgroundColor: "hsl(var(--background))",
                  border: "1px solid hsl(var(--border))",
                }}
              />
            </PieChart>
          </ResponsiveContainer>
        </ChartCard>

        {/* Duration Breakdown */}
        <ChartCard title="Average Duration Breakdown">
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={durationBreakdownData}>
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
              <XAxis dataKey="name" className="text-xs" />
              <YAxis className="text-xs" label={{ value: "Seconds", angle: -90, position: "insideLeft" }} />
              <Tooltip
                contentStyle={{
                  backgroundColor: "hsl(var(--background))",
                  border: "1px solid hsl(var(--border))",
                }}
              />
              <Bar dataKey="value" fill="#3b82f6" />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>

        {/* Tool Performance */}
        <ChartCard title="Tool Performance">
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={toolPerformanceData} layout="vertical">
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
              <XAxis type="number" className="text-xs" label={{ value: "Seconds", position: "insideBottom" }} />
              <YAxis dataKey="name" type="category" className="text-xs" />
              <Tooltip
                contentStyle={{
                  backgroundColor: "hsl(var(--background))",
                  border: "1px solid hsl(var(--border))",
                }}
              />
              <Bar dataKey="duration" fill="#8b5cf6" />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>
      </div>

      {/* Queue Statistics */}
      <div className="bg-card border rounded-lg p-6">
        <h3 className="text-lg font-semibold mb-4">Queue Statistics</h3>
        <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
          <QueueStat label="Active" value={queueStats?.active || 0} />
          <QueueStat label="Pending" value={queueStats?.pending || 0} />
          <QueueStat label="Scheduled" value={queueStats?.scheduled || 0} />
          <QueueStat label="Completed" value={queueStats?.completed || 0} />
          <QueueStat label="Failed" value={queueStats?.failed || 0} />
        </div>
      </div>
    </div>
  );
}

function SummaryCard({
  icon: Icon,
  title,
  value,
  iconColor,
}: {
  icon: any;
  title: string;
  value: string | number;
  iconColor: string;
}) {
  return (
    <div className="bg-card border rounded-lg p-4">
      <div className="flex items-center justify-between mb-2">
        <span className="text-sm text-muted-foreground">{title}</span>
        <Icon className={`h-4 w-4 ${iconColor}`} />
      </div>
      <p className="text-2xl font-bold">{value}</p>
    </div>
  );
}

function ChartCard({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="bg-card border rounded-lg p-6">
      <h3 className="text-lg font-semibold mb-4">{title}</h3>
      {children}
    </div>
  );
}

function QueueStat({ label, value }: { label: string; value: number }) {
  return (
    <div className="text-center">
      <p className="text-sm text-muted-foreground mb-1">{label}</p>
      <p className="text-xl font-semibold">{value}</p>
    </div>
  );
}
