import { Card, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Plus, MoreVertical, Power, Trash2 } from "lucide-react";

export interface IntegrationCardProps {
  name: string;
  description: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
  category: string;
  status?: "connected" | "disconnected";
  onSetup: () => void;
  onDisable?: () => void;
  onRemove?: () => void;
}

export function IntegrationCard({
  name,
  description,
  icon: Icon,
  category,
  status = "disconnected",
  onSetup,
  onDisable,
  onRemove,
}: IntegrationCardProps) {
  return (
    <Card className="group rounded-none border-0 border-r border-b border-border relative overflow-hidden transition-all duration-300">
      <div className="absolute inset-0 bg-gradient-to-t from-primary/5 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />

      <CardHeader className="relative">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-gradient-to-br from-primary/10 to-primary/5 border border-primary/20 group-hover:border-primary/40 transition-colors duration-300">
              <Icon size={24} className="text-primary" />
            </div>
            <div>
              <CardTitle className="text-lg font-medium">{name}</CardTitle>
              <Badge variant="outline" className="mt-1.5 text-xs">
                {category}
              </Badge>
            </div>
          </div>

          {/* Right side - Plus icon or Configured with dropdown */}
          {status === "disconnected" ? (
            <Button
              variant="ghost"
              size="icon"
              onClick={onSetup}
              className="h-8 w-8 hover:bg-primary/10"
            >
              <Plus className="h-5 w-5 text-primary" />
            </Button>
          ) : (
            <div className="flex items-center gap-2">
              <Badge className="bg-green-500/10 text-green-600 border-green-500/20">
                Configured
              </Badge>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" size="icon" className="h-8 w-8">
                    <MoreVertical className="h-4 w-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem onClick={onDisable}>
                    <Power className="h-4 w-4 mr-2" />
                    Disable
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={onRemove} className="text-destructive">
                    <Trash2 className="h-4 w-4 mr-2" />
                    Remove
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          )}
        </div>
        <CardDescription className="mt-3 text-sm leading-relaxed">
          {description}
        </CardDescription>
      </CardHeader>
    </Card>
  );
}
