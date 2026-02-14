import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Download, Eye, FileJson, FileText } from "lucide-react";

const API_BASE = "/api/v1";

interface ArtifactsTabProps {
  jobId: string;
}

const ARTIFACTS = [
  {
    id: "report",
    name: "Full Report",
    description: "Complete analysis report with all metrics and recommendations",
    icon: FileJson,
    type: "JSON",
  },
  {
    id: "grype.json",
    name: "Vulnerability Scan",
    description: "Grype CVE scan results for all packages in the image",
    icon: FileJson,
    type: "JSON",
  },
  {
    id: "dive.json",
    name: "Layer Efficiency",
    description: "Dive analysis of image layers and wasted space",
    icon: FileText,
    type: "JSON",
  },
  {
    id: "dockle.json",
    name: "CIS Benchmark",
    description: "Dockle CIS Docker Benchmark compliance check results",
    icon: FileText,
    type: "JSON",
  },
];

export function ArtifactsTab({ jobId }: ArtifactsTabProps) {
  const getUrl = (artifactId: string) => `${API_BASE}/jobs/${jobId}/${artifactId}`;

  return (
    <div className="p-6">
      <Card>
        <CardHeader>
          <CardTitle>Artifacts</CardTitle>
          <CardDescription>
            Preview or download generated analysis files
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {ARTIFACTS.map((artifact) => {
            const Icon = artifact.icon;
            return (
              <div
                key={artifact.id}
                className="flex items-center justify-between p-4 border rounded-lg hover:bg-muted/50 transition-colors"
              >
                <div className="flex items-start gap-4 flex-1">
                  <div className="p-2 rounded-lg bg-primary/10">
                    <Icon className="h-5 w-5 text-primary" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <h4 className="font-medium text-sm">{artifact.name}</h4>
                      <span className="text-xs text-muted-foreground">
                        ({artifact.type})
                      </span>
                    </div>
                    <p className="text-xs text-muted-foreground">
                      {artifact.description}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-1.5 ml-4 shrink-0">
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={() => window.open(getUrl(artifact.id), "_blank")}
                    title="Preview"
                  >
                    <Eye className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={() => window.open(`${getUrl(artifact.id)}?download=true`, "_blank")}
                    title="Download"
                  >
                    <Download className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            );
          })}
        </CardContent>
      </Card>
    </div>
  );
}
