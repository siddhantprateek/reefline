import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Container, Layers, Calendar, Cpu } from "lucide-react";
import type { JobReport } from "@/api/jobs.api";

interface ImageInfoCardProps {
  report: JobReport;
}

export function ImageInfoCard({ report }: ImageInfoCardProps) {
  // Mock metadata - in production this would come from Skopeo inspection
  const metadata = {
    name: report.input_scenario === "image" ? "ubuntu" : "custom-image",
    tag: "22.04",
    architecture: "amd64",
    os: "linux",
    layers: 6,
    created: "2024-01-15T10:30:00Z",
    size: "40.36 MB",
  };

  const infoItems = [
    {
      icon: Container,
      label: "Image",
      value: `${metadata.name}:${metadata.tag}`,
    },
    {
      icon: Layers,
      label: "Layers",
      value: metadata.layers.toString(),
    },
    {
      icon: Cpu,
      label: "Platform",
      value: `${metadata.os}/${metadata.architecture}`,
    },
    {
      icon: Calendar,
      label: "Created",
      value: new Date(metadata.created).toLocaleDateString(),
    },
  ];

  return (
    <Card>
      <CardContent className="pt-6">
        <div className="grid grid-cols-2 gap-4">
          {infoItems.map((item) => {
            const Icon = item.icon;
            return (
              <div key={item.label} className="flex items-start gap-3">
                <div className="p-2 rounded-lg bg-primary/10">
                  <Icon className="h-4 w-4 text-primary" />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="text-xs text-muted-foreground mb-1">
                    {item.label}
                  </div>
                  <div className="text-sm font-medium truncate">
                    {item.value}
                  </div>
                </div>
              </div>
            );
          })}
        </div>

        <div className="mt-4 pt-4 border-t flex items-center justify-between">
          <span className="text-sm text-muted-foreground">Total Size</span>
          <Badge variant="outline" className="font-mono">
            {metadata.size}
          </Badge>
        </div>
      </CardContent>
    </Card>
  );
}
