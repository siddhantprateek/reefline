import { Button } from "@/components/ui/button"
import { Bell, User } from "lucide-react"
import { DottedBackground } from "@/components/custom/header/dotted-background"

export function Header() {
  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <DottedBackground
        className="h-14 px-6 items-center bg-transparent"
        y={6}
      >
        <div className="flex w-full items-center justify-between">
          <div className="mr-4 hidden md:flex">
            <a className="mr-6 flex items-center space-x-2" href="/">
              <span className="hidden font-bold sm:inline-block">Reefline</span>
            </a>
          </div>
          <div className="flex flex-1 items-center justify-between space-x-2 md:justify-end">
            <div className="w-full flex-1 md:w-auto md:flex-none">
              <Button variant="outline" className="relative h-8 w-full justify-start rounded-[0.5rem] bg-background text-sm font-normal text-muted-foreground shadow-none sm:pr-12 md:w-40 lg:w-64">
                <span className="hidden lg:inline-flex">Search documentation...</span>
                <span className="inline-flex lg:hidden">Search...</span>
                <kbd className="pointer-events-none absolute right-[0.3rem] top-[0.3rem] hidden h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium opacity-100 sm:flex">
                  <span className="text-xs">⌘</span>K
                </kbd>
              </Button>
            </div>
            <nav className="flex items-center gap-2">
              <Button variant="ghost" size="icon" className="h-8 w-8 px-0">
                <Bell className="h-4 w-4" />
                <span className="sr-only">Notifications</span>
              </Button>
              <Button variant="ghost" size="icon" className="h-8 w-8 px-0">
                <User className="h-4 w-4" />
                <span className="sr-only">Profile</span>
              </Button>
            </nav>
          </div>
        </div>
      </DottedBackground>
    </header>
  )
}
