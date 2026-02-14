import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Container, Layers, HardDrive, Activity } from "lucide-react";
import type { JobReport, DiveReport } from "@/types/jobs";

interface ImageInfoCardProps {
  report: JobReport;
  dive?: DiveReport | null;
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`;
}

export function ImageInfoCard({ report, dive }: ImageInfoCardProps) {
  const imageRef = dive?.image ?? report.input_scenario;
  const layerCount = dive?.layers?.length ?? null;
  const efficiency = dive ? `${(dive.efficiency * 100).toFixed(1)}%` : null;
  const wastedBytes = dive ? formatBytes(dive.wastedBytes) : null;
  const totalSize = dive ? formatBytes(dive.sizeBytes) : null;

  const infoItems = [
    {
      icon: Container,
      label: "Image",
      value: imageRef,
    },
    ...(layerCount !== null ? [{
      icon: Layers,
      label: "Layers",
      value: `${layerCount} layers`,
    }] : []),
    ...(efficiency ? [{
      icon: Activity,
      label: "Efficiency",
      value: efficiency,
    }] : []),
    ...(wastedBytes ? [{
      icon: HardDrive,
      label: "Wasted",
      value: wastedBytes,
    }] : []),
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

        {totalSize && (
          <div className="mt-4 pt-4 border-t flex items-center justify-between">
            <span className="text-sm text-muted-foreground">Total Size</span>
            <Badge variant="outline" className="font-mono">
              {totalSize}
            </Badge>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
