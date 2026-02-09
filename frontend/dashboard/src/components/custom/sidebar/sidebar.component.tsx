import { Link } from "react-router-dom"
import { Home, Settings, Moon, Sun, Monitor, ChevronLeft, ChevronRight, Palette, Scan, Sparkles, FileCode, Container, History } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSub,
  DropdownMenuSubTrigger,
  DropdownMenuSubContent,
  DropdownMenuSeparator,
} from "@/components/ui/dropdown-menu"
import { useTheme } from "@/components/theme-provider"
import { cn } from "@/lib/utils"

interface SidebarProps {
  isCollapsed?: boolean
  toggleSidebar?: () => void
}

export function Sidebar({ isCollapsed = false, toggleSidebar }: SidebarProps) {
  const { theme, setTheme } = useTheme()

  return (
    <aside
      className={cn(
        "fixed top-14 left-0 hidden h-[calc(100vh-3.5rem)] border-r bg-background font-medium md:block transition-all duration-300",
        isCollapsed ? "w-16" : "w-60"
      )}
    >
      {/* Toggle Button */}
      <Button
        size="icon"
        className="absolute -right-3 top-6 z-20 h-6 w-6 rounded-full border bg-background text-foreground cursor-pointer hover:bg-muted"
        onClick={toggleSidebar}
      >
        {isCollapsed ? <ChevronRight className="h-3 w-3" /> : <ChevronLeft className="h-3 w-3" />}
      </Button>

      <div className="flex h-full flex-col justify-between py-6 px-4 overflow-x-hidden">
        <nav className="flex flex-col gap-2">
          <Button variant="ghost" className={cn("justify-start gap-3", isCollapsed && "justify-center px-2")} asChild>
            <Link to="/">
              <Home className="h-5 w-5 shrink-0" />
              {!isCollapsed && <span>Overview</span>}
            </Link>
          </Button>
          <Button variant="ghost" className={cn("justify-start gap-3", isCollapsed && "justify-center px-2")} asChild>
            <Link to="/analysis">
              <Scan className="h-5 w-5 shrink-0" />
              {!isCollapsed && <span>Analysis</span>}
            </Link>
          </Button>
          <Button variant="ghost" className={cn("justify-start gap-3", isCollapsed && "justify-center px-2")} asChild>
            <Link to="/optimization">
              <Sparkles className="h-5 w-5 shrink-0" />
              {!isCollapsed && <span>Optimization</span>}
            </Link>
          </Button>
          <Button variant="ghost" className={cn("justify-start gap-3", isCollapsed && "justify-center px-2")} asChild>
            <Link to="/integrations">
              <Container className="h-5 w-5 shrink-0" />
              {!isCollapsed && <span>Integrations</span>}
            </Link>
          </Button>
          <Button variant="ghost" className={cn("justify-start gap-3", isCollapsed && "justify-center px-2")} asChild>
            <Link to="/history">
              <History className="h-5 w-5 shrink-0" />
              {!isCollapsed && <span>History</span>}
            </Link>
          </Button>
        </nav>

        <div className="flex flex-col gap-2">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className={cn("justify-start gap-3 w-full", isCollapsed && "justify-center px-2")}>
                <Settings className="h-5 w-5 shrink-0" />
                {!isCollapsed && <span>Settings</span>}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start" side="top" className="w-56" sideOffset={10}>
              <DropdownMenuSub>
                <DropdownMenuSubTrigger>
                  <Palette className="mr-2 h-4 w-4" />
                  <span>Theme ({theme})</span>
                </DropdownMenuSubTrigger>
                <DropdownMenuSubContent>
                  <DropdownMenuItem onClick={() => setTheme("light")}>
                    <Sun className="mr-2 h-4 w-4" />
                    <span>Light</span>
                    {theme === "light" && <span className="ml-auto text-xs">✓</span>}
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => setTheme("dark")}>
                    <Moon className="mr-2 h-4 w-4" />
                    <span>Dark</span>
                    {theme === "dark" && <span className="ml-auto text-xs">✓</span>}
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => setTheme("system")}>
                    <Monitor className="mr-2 h-4 w-4" />
                    <span>System</span>
                    {theme === "system" && <span className="ml-auto text-xs">✓</span>}
                  </DropdownMenuItem>
                </DropdownMenuSubContent>
              </DropdownMenuSub>

              <DropdownMenuSeparator />

              <DropdownMenuItem asChild>
                <Link to="/settings" className="w-full cursor-pointer">
                  <span>All Settings</span>
                </Link>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </aside>
  )
}
