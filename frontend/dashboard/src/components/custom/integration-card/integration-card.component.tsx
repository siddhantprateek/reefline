import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

export interface IntegrationCardProps {
  name: string;
  description: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
  category: string;
  status?: "connected" | "disconnected";
  onSetup: () => void;
}

export function IntegrationCard({
  name,
  description,
  icon: Icon,
  category,
  status = "disconnected",
  onSetup,
}: IntegrationCardProps) {
  return (
    <Card className="group rounded-none relative overflow-hidden transition-all duration-300">
      <div className="absolute inset-0 bg-gradient-to-t from-primary/5 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />

      <CardHeader className="relative">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-gradient-to-br from-primary/10 to-primary/5 border border-primary/20 group-hover:border-primary/40 transition-colors duration-300">
              <Icon size={24} className="text-primary" />
            </div>
            <div>
              <CardTitle className="text-lg font-semibold">{name}</CardTitle>
              <Badge variant="outline" className="mt-1.5 text-xs">
                {category}
              </Badge>
            </div>
          </div>
          {status === "connected" && (
            <Badge className="bg-green-500/10 text-green-600 border-green-500/20 hover:bg-green-500/20">
              Connected
            </Badge>
          )}
        </div>
        <CardDescription className="mt-3 text-sm leading-relaxed">
          {description}
        </CardDescription>
      </CardHeader>

      <CardContent className="relative">
        <Button
          onClick={onSetup}
          variant={status === "connected" ? "outline" : "default"}
          className="w-full group-hover:shadow-md transition-shadow duration-300"
        >
          {status === "connected" ? "Manage" : "Setup"}
        </Button>
      </CardContent>
    </Card>
  );
}
