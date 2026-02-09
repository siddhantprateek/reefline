import { Hero } from "@/components/custom/hero"

export function OverviewPage() {
  return (
    <div className="space-y-6">
      <Hero />
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {/* Further content can go here */}
      </div>
    </div>
  )
}
