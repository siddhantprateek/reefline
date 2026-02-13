import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Layers, FileCode, Hash } from "lucide-react";
import { cn } from "@/lib/utils";
import type { JobReport } from "@/api/jobs.api";

interface Layer {
  id: string;
  size: number;
  command: string;
  created: string;
}

interface InspectorProps {
  report: JobReport;
}

export function Inspector({ report }: InspectorProps) {
  // Mock layer data - in production this would come from Skopeo/Dive metadata
  const mockLayers: Layer[] = [
    {
      id: "sha256:e692418e...1c8d",
      size: 27100000,
      command: "FROM ubuntu:22.04",
      created: "2024-01-15T10:30:00Z",
    },
    {
      id: "sha256:a3ed95ca...b5e2",
      size: 4200000,
      command: "RUN apt-get update && apt-get install -y curl",
      created: "2024-01-15T10:31:23Z",
    },
    {
      id: "sha256:5f70bf18...3ac4",
      size: 8900000,
      command: "COPY . /app",
      created: "2024-01-15T10:32:45Z",
    },
    {
      id: "sha256:9c27e219...7f91",
      size: 156000,
      command: "RUN npm install --production",
      created: "2024-01-15T10:35:12Z",
    },
    {
      id: "sha256:f1b5933f...8d2a",
      size: 0,
      command: "WORKDIR /app",
      created: "2024-01-15T10:35:13Z",
    },
    {
      id: "sha256:c2d5e8a1...9b3c",
      size: 0,
      command: 'CMD ["node", "index.js"]',
      created: "2024-01-15T10:35:14Z",
    },
  ];

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`;
  };

  const totalSize = mockLayers.reduce((sum, layer) => sum + layer.size, 0);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          <Layers className="h-5 w-5" />
          Image Layers
        </CardTitle>
        <CardDescription>
          {mockLayers.length} layers • Total size: {formatBytes(totalSize)}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <ScrollArea className="h-[400px] pr-4">
          <div className="space-y-2">
            {mockLayers.map((layer, idx) => (
              <div
                key={layer.id}
                className="border rounded-lg p-3 hover:bg-muted/50 transition-colors"
              >
                <div className="flex items-start justify-between gap-3 mb-2">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <Badge variant="outline" className="text-xs shrink-0">
                        Layer {idx + 1}
                      </Badge>
                      <div className="flex items-center gap-1 text-xs text-muted-foreground font-mono truncate">
                        <Hash className="h-3 w-3 shrink-0" />
                        <span className="truncate">{layer.id}</span>
                      </div>
                    </div>
                    <div className="flex items-start gap-2">
                      <FileCode className="h-4 w-4 text-muted-foreground shrink-0 mt-0.5" />
                      <code className="text-sm break-all">{layer.command}</code>
                    </div>
                  </div>
                  <div className="text-right shrink-0">
                    <div className="text-sm font-medium">
                      {formatBytes(layer.size)}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {layer.size > 0 && (
                        <>{((layer.size / totalSize) * 100).toFixed(1)}%</>
                      )}
                    </div>
                  </div>
                </div>

                {/* Size bar */}
                {layer.size > 0 && (
                  <div className="mt-2">
                    <div className="h-1.5 bg-muted rounded-full overflow-hidden">
                      <div
                        className={cn(
                          "h-full rounded-full",
                          layer.size > totalSize * 0.5
                            ? "bg-red-500"
                            : layer.size > totalSize * 0.2
                            ? "bg-yellow-500"
                            : "bg-green-500"
                        )}
                        style={{ width: `${(layer.size / totalSize) * 100}%` }}
                      />
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        </ScrollArea>
      </CardContent>
    </Card>
  );
}
