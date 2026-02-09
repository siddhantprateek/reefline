import { Link } from "react-router-dom"
import { Button } from "@/components/ui/button"

export function NotFoundPage() {
  return (
    <div className="flex h-[80vh] bg-background flex-col items-center justify-center gap-2">
      <h1 className="text-4xl font-extrabold tracking-tight lg:text-5xl">404</h1>
      <p className="text-xl text-muted-foreground">Page not found</p>
      <Button asChild className="mt-4">
        <Link to="/">Go Home</Link>
      </Button>
    </div>
  )
}
