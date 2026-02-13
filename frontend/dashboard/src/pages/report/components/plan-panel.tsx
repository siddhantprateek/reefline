import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import { Lightbulb, Shield, TrendingDown, Zap } from "lucide-react";
import { cn } from "@/lib/utils";

const EFFORT_COLORS: Record<string, string> = {
  low: "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400",
  medium: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400",
  high: "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400",
};

interface Recommendation {
  priority: number;
  category: string;
  title: string;
  description: string;
  effort: string;
  estimated_size_savings_mb?: number;
  estimated_cves_eliminated?: number;
  before?: string;
  after?: string;
}

interface Score {
  current: number;
  estimated_after: number;
  grade_current: string;
  grade_estimated: string;
}

interface PlanPanelProps {
  recommendations?: Recommendation[];
  score?: Score;
}

export function PlanPanel({ recommendations, score }: PlanPanelProps) {
  const improvement = score ? score.estimated_after - score.current : 0;

  return (
    <div className="p-6 space-y-6">
      {/* Score Overview */}
      {score && (
        <Card className="border-2 border-primary/20">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-lg">
              <Zap className="h-5 w-5 text-primary" />
              Optimization Score
            </CardTitle>
            <CardDescription>Potential improvement after applying recommendations</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-around">
              <div className="text-center">
                <div className="text-sm text-muted-foreground mb-2">Current</div>
                <div className="text-3xl font-bold">{score.current}</div>
                <Badge variant="outline" className="mt-2">
                  {score.grade_current}
                </Badge>
              </div>

              <div className="flex flex-col items-center">
                <TrendingDown className="h-6 w-6 text-green-600 dark:text-green-400 rotate-90" />
                <span className="text-sm text-green-600 dark:text-green-400 font-medium mt-1">
                  +{improvement}
                </span>
              </div>

              <div className="text-center">
                <div className="text-sm text-muted-foreground mb-2">Estimated</div>
                <div className="text-3xl font-bold text-primary">{score.estimated_after}</div>
                <Badge className="mt-2 bg-primary">
                  {score.grade_estimated}
                </Badge>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Recommendations */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Lightbulb className="h-5 w-5 text-amber-500" />
            Optimization Plan
          </CardTitle>
          <CardDescription>
            {recommendations?.length || 0} recommendations to improve your image
          </CardDescription>
        </CardHeader>
        <CardContent>
          {recommendations && recommendations.length > 0 ? (
            <Accordion type="single" collapsible className="w-full">
              {recommendations.map((rec, idx) => (
                <AccordionItem key={idx} value={`rec-${idx}`}>
                  <AccordionTrigger className="hover:no-underline">
                    <div className="flex items-center gap-3 w-full pr-4">
                      <Badge variant="outline" className="shrink-0 w-12 justify-center">
                        P{rec.priority}
                      </Badge>
                      <div className="flex-1 text-left">
                        <div className="font-medium text-sm">{rec.title}</div>
                        <div className="text-xs text-muted-foreground mt-1">
                          {rec.category}
                          {rec.estimated_size_savings_mb && (
                            <> • ~{rec.estimated_size_savings_mb}MB savings</>
                          )}
                        </div>
                      </div>
                      <Badge
                        className={cn("shrink-0", EFFORT_COLORS[rec.effort] || "")}
                        variant="outline"
                      >
                        {rec.effort}
                      </Badge>
                    </div>
                  </AccordionTrigger>
                  <AccordionContent className="pt-4 space-y-4">
                    <p className="text-sm text-muted-foreground">{rec.description}</p>

                    {rec.estimated_cves_eliminated && rec.estimated_cves_eliminated > 0 && (
                      <div className="text-sm flex items-center gap-2 text-green-600 dark:text-green-400">
                        <Shield className="h-4 w-4" />
                        Eliminates ~{rec.estimated_cves_eliminated} vulnerabilities
                      </div>
                    )}

                    {rec.before && rec.after && (
                      <div className="space-y-2">
                        <Separator />
                        <div className="grid md:grid-cols-2 gap-4">
                          <div>
                            <div className="text-xs font-medium mb-2 text-muted-foreground">Before</div>
                            <pre className="text-xs bg-muted p-3 rounded-md overflow-x-auto">
                              {rec.before}
                            </pre>
                          </div>
                          <div>
                            <div className="text-xs font-medium mb-2 text-primary">After</div>
                            <pre className="text-xs bg-primary/5 p-3 rounded-md overflow-x-auto border border-primary/20">
                              {rec.after}
                            </pre>
                          </div>
                        </div>
                      </div>
                    )}
                  </AccordionContent>
                </AccordionItem>
              ))}
            </Accordion>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              <Lightbulb className="h-12 w-12 mx-auto mb-3 opacity-50" />
              <p className="text-sm">No recommendations available yet</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
