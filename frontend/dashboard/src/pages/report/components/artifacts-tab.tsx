import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Download, FileJson, FileCode, FileText, Image } from "lucide-react";

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
    size: "~45 KB",
  },
  {
    id: "dockerfile",
    name: "Optimized Dockerfile",
    description: "Generated Dockerfile with recommended optimizations applied",
    icon: FileCode,
    type: "Dockerfile",
    size: "~2 KB",
  },
  {
    id: "sbom",
    name: "Software Bill of Materials",
    description: "SPDX-format SBOM listing all packages and dependencies",
    icon: FileText,
    type: "JSON",
    size: "~120 KB",
  },
  {
    id: "graph",
    name: "Build Graph",
    description: "Visual representation of the image build process",
    icon: Image,
    type: "SVG",
    size: "~8 KB",
  },
];

export function ArtifactsTab({ jobId }: ArtifactsTabProps) {
  const handleDownload = (artifactId: string) => {
    const url = `${API_BASE}/jobs/${jobId}/${artifactId}`;
    window.open(url, "_blank");
  };

  return (
    <div className="p-6">
      <Card>
        <CardHeader>
          <CardTitle>Download Artifacts</CardTitle>
          <CardDescription>
            Access generated files and detailed analysis data
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
                    <div className="text-xs text-muted-foreground mt-1">
                      Size: {artifact.size}
                    </div>
                  </div>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleDownload(artifact.id)}
                  className="ml-4"
                >
                  <Download className="h-4 w-4 mr-2" />
                  Download
                </Button>
              </div>
            );
          })}
        </CardContent>
      </Card>
    </div>
  );
}
