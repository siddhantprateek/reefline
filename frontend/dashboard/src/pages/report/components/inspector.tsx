import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Layers, FileCode, Hash, AlertTriangle } from "lucide-react";
import { cn } from "@/lib/utils";
import type { DiveReport } from "@/types/jobs";

interface InspectorProps {
  dive: DiveReport;
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`;
}

export function Inspector({ dive }: InspectorProps) {
  const layers = dive.layers ?? [];
  const totalSize = layers.reduce((sum, l) => sum + l.sizeBytes, 0);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          <Layers className="h-5 w-5" />
          Image Layers
        </CardTitle>
        <CardDescription>
          {layers.length} layers · {formatBytes(dive.sizeBytes)} total ·{" "}
          {(dive.efficiency * 100).toFixed(1)}% efficient ·{" "}
          {formatBytes(dive.wastedBytes)} wasted
        </CardDescription>
      </CardHeader>
      <CardContent>
        <ScrollArea className="h-[400px] pr-4">
          <div className="space-y-2">
            {layers.map((layer) => (
              <div
                key={layer.digestId}
                className="border rounded-lg p-3 hover:bg-muted/50 transition-colors"
              >
                <div className="flex items-start justify-between gap-3 mb-2">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <Badge variant="outline" className="text-xs shrink-0">
                        Layer {layer.index + 1}
                      </Badge>
                      <div className="flex items-center gap-1 text-xs text-muted-foreground font-mono truncate">
                        <Hash className="h-3 w-3 shrink-0" />
                        <span className="truncate">{layer.digestId.slice(7, 19)}</span>
                      </div>
                      {layer.fileCount > 0 && (
                        <span className="text-xs text-muted-foreground shrink-0">
                          {layer.fileCount} files
                        </span>
                      )}
                    </div>
                    <div className="flex items-start gap-2">
                      <FileCode className="h-4 w-4 text-muted-foreground shrink-0 mt-0.5" />
                      <code className="text-xs break-all text-muted-foreground line-clamp-2">
                        {layer.command}
                      </code>
                    </div>
                  </div>
                  <div className="text-right shrink-0">
                    <div className="text-sm font-medium">{formatBytes(layer.sizeBytes)}</div>
                    {layer.sizeBytes > 0 && totalSize > 0 && (
                      <div className="text-xs text-muted-foreground">
                        {((layer.sizeBytes / totalSize) * 100).toFixed(1)}%
                      </div>
                    )}
                  </div>
                </div>

                {layer.sizeBytes > 0 && totalSize > 0 && (
                  <div className="mt-2">
                    <div className="h-1.5 bg-muted rounded-full overflow-hidden">
                      <div
                        className={cn(
                          "h-full rounded-full",
                          layer.sizeBytes > totalSize * 0.4
                            ? "bg-red-500"
                            : layer.sizeBytes > totalSize * 0.15
                            ? "bg-yellow-500"
                            : "bg-green-500"
                        )}
                        style={{ width: `${(layer.sizeBytes / totalSize) * 100}%` }}
                      />
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        </ScrollArea>

        {dive.inefficiencies && dive.inefficiencies.length > 0 && (
          <div className="mt-4 border rounded-lg p-3 border-yellow-500/30 bg-yellow-500/5">
            <div className="flex items-center gap-2 mb-2 text-sm font-medium text-yellow-600 dark:text-yellow-400">
              <AlertTriangle className="h-4 w-4" />
              {dive.inefficiencies.length} inefficient path{dive.inefficiencies.length !== 1 ? "s" : ""} detected
            </div>
            <div className="space-y-1 max-h-32 overflow-y-auto">
              {dive.inefficiencies.map((item, idx) => (
                <div key={idx} className="flex justify-between items-center text-xs">
                  <code className="text-muted-foreground truncate max-w-[200px]">{item.path}</code>
                  <span className="text-muted-foreground shrink-0 ml-2">
                    {item.removedOperations}x removed · {formatBytes(item.sizeBytes)}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
