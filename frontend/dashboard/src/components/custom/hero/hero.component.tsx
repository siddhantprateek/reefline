import { Button } from "@/components/ui/button"

export function Hero() {
  return (
    <section className="container mx-auto grid items-center gap-6 pb-8 pt-6 md:py-10">
      <div className="flex max-w-[980px] flex-col items-start gap-2">
        <h1 className="text-3xl font-extrabold leading-tight tracking-tighter md:text-5xl lg:text-6xl lg:leading-[1.1]">
          Build your component library <br className="hidden sm:inline" />
          with Reefline
        </h1>
        <p className="max-w-[700px] text-lg text-muted-foreground sm:text-xl">
          Beautifully designed components built with Radix UI and Tailwind CSS.
        </p>
      </div>
      <div className="flex gap-4">
        <Button size="lg">Get Started</Button>
        <Button variant="outline" size="lg">Documentation</Button>
      </div>
    </section>
  )
}
