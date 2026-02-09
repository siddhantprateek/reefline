import { DotPattern } from "@/components/ui/dot-pattern"
import { cn } from "@/lib/utils"

interface DottedBackgroundProps {
  children: React.ReactNode
  className?: string
  patternClassName?: string
  width?: number
  height?: number
  cx?: number
  cy?: number
  cr?: number
  x?: number
  y?: number
}

export function DottedBackground({
  children,
  className,
  patternClassName,
  width = 20,
  height = 20,
  cx = 1,
  cy = 1,
  cr = 1,
  x = 0,
  y = 0
}: DottedBackgroundProps) {
  return (
    <div className={cn("relative w-full overflow-hidden bg-background", className)}>
      <DotPattern
        width={width}
        height={height}
        cx={cx}
        cy={cy}
        cr={cr}
        x={x}
        y={y}
        className={cn(
          "absolute inset-0 h-full w-full fill-neutral-500/50 dark:fill-neutral-400/20",
          patternClassName
        )}
      />
      <div className="relative z-10 flex w-full h-full">
        {children}
      </div>
    </div>
  )
}
