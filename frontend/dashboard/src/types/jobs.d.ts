export interface AnalysisRequest {
  image_ref: string;
  app_context?: string;
}

export interface JobResponse {
  job_id: string;
  status: string;
  stream_url?: string;
  error?: string;
}

export type JobStatus = "PENDING" | "RUNNING" | "COMPLETED" | "FAILED" | "CANCELLED" | "SKIPPED" | "UNKNOWN";

export interface Job {
  id: string;
  job_id: string;
  image_ref: string;
  dockerfile?: string;
  status: JobStatus;
  scenario?: "image" | "dockerfile" | "both";
  created_at: string;
  updated_at: string;
  completed_at?: string;
  error_message?: string;
  progress?: number;
}

export interface JobReport {
  job_id: string;
  status: JobStatus;
  input_scenario: string;
  report?: {
    measured?: {
      current_size_mb?: number;
      layer_efficiency_pct?: number;
      total_packages?: number;
      total_cves?: number;
      critical_cves?: number;
      high_cves?: number;
      runs_as_root?: boolean;
      secrets_detected?: number;
      base_image_age_days?: number;
      lint_warnings?: number;
      lint_errors?: number;
    };
    proposed?: {
      estimated_size_mb?: number;
      estimated_cve_reduction?: number;
      estimated_remaining_cves?: number;
      actionable_cves?: number;
      score?: {
        current: number;
        estimated_after: number;
        grade_current: string;
        grade_estimated: string;
      };
      recommendations?: Array<{
        priority: number;
        category: string;
        title: string;
        description: string;
        effort: string;
        estimated_size_savings_mb?: number;
        estimated_cves_eliminated?: number;
        before?: string;
        after?: string;
      }>;
      optimized_dockerfile?: string;
      disclaimer?: string;
    };
    tool_data?: any;
  };
}

// Grype vulnerability scan
export interface GrypeReport {
  ID: string;
  Table: {
    Rows: string[][] | null; // [package, installed, fixed, type, CVE, severity]
  };
}

// Dockle CIS benchmark scan
export interface DockleAssessment {
  code: string;
  title: string;
  level: string;
  levelInt: number;
}

export interface DockleReport {
  image: string;
  scanTime: string;
  status: string;
  summary: {
    fatal: number;
    warn: number;
    info: number;
    skip: number;
    pass: number;
    total: number;
  };
  assessments: DockleAssessment[];
}

// Dive layer efficiency analysis
export interface DiveLayer {
  index: number;
  id: string;
  digestId: string;
  command: string;
  sizeBytes: number;
  fileCount: number;
}

export interface DiveInefficiency {
  path: string;
  sizeBytes: number;
  removedOperations: number;
}

export interface DiveReport {
  image: string;
  efficiency: number;
  status: string;
  sizeBytes: number;
  wastedBytes: number;
  wastedUserPercent: number;
  layers: DiveLayer[];
  inefficiencies: DiveInefficiency[] | null;
}
