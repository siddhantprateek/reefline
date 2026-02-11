import { Link } from "react-router-dom"
import { Home, Settings, Moon, Sun, Monitor, ChevronLeft, ChevronRight, Palette, Scan, Sparkles, Container, History, LogOut, User } from "lucide-react"

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
        className="absolute -right-3 top-6 z-50 h-6 w-6 rounded-full border bg-background text-foreground cursor-pointer hover:bg-muted"
        onClick={toggleSidebar}
      >
        {isCollapsed ? <ChevronRight className="h-3 w-3" /> : <ChevronLeft className="h-3 w-3" />}
      </Button>

      <div className="flex h-full flex-col justify-between overflow-x-hidden">
        {/* Main Navigation */}
        <nav className="flex flex-col gap-2 py-6 px-4">
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
            <Link to="/jobs">
              <History className="h-5 w-5 shrink-0" />
              {!isCollapsed && <span>Jobs</span>}
            </Link>
          </Button>
        </nav>

        {/* Bottom Section */}
        <div className="mt-auto">
          {/* Settings row with dotted background */}


          {/* Profile row */}
          <div className="border-t border-border">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button
                  className={cn(
                    "flex items-center gap-3 w-full px-4 py-3 text-left hover:bg-muted/50 transition-colors cursor-pointer",
                    isCollapsed && "justify-center px-2"
                  )}
                >
                  {/* Avatar */}
                  <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-primary/20 bg-gradient-to-br from-primary/10 to-primary/5 text-primary text-sm font-bold select-none">
                    S
                  </div>
                  {!isCollapsed && (
                    <>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium truncate">Siddhant</p>
                        <p className="text-xs text-muted-foreground truncate">siddhant@reefline.ai</p>
                      </div>
                      <Settings className="h-4 w-4 shrink-0 text-muted-foreground" />
                    </>
                  )}
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="center" side="top" className="w-56" sideOffset={8}>
                <div className="px-2 py-1.5">
                  <p className="text-sm font-medium">Siddhant Prateek</p>
                  <p className="text-xs text-muted-foreground">siddhant@reefline.ai</p>
                </div>
                <DropdownMenuSeparator />

                <DropdownMenuItem asChild>
                  <Link to="/settings" className="w-full cursor-pointer">
                    <User className="mr-2 h-4 w-4" />
                    <span>Profile</span>
                  </Link>
                </DropdownMenuItem>

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

                <DropdownMenuItem asChild>
                  <Link to="/settings" className="w-full cursor-pointer">
                    <Settings className="mr-2 h-4 w-4" />
                    <span>All Settings</span>
                  </Link>
                </DropdownMenuItem>

                <DropdownMenuSeparator />

                <DropdownMenuItem className="text-destructive focus:text-destructive cursor-pointer">
                  <LogOut className="mr-2 h-4 w-4" />
                  <span>Log out</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
      </div>
    </aside>
  )
}
