import { useState } from "react"
import { Outlet } from "react-router-dom"
import { Header } from "@/components/custom/header"
import { Sidebar } from "@/components/custom/sidebar"
// import { Footer } from "@/components/custom/footer"
import { cn } from "@/lib/utils"

export default function Layout() {
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false)

  return (
    <div className="flex min-h-screen w-full flex-col">
      <Header />
      <div className="flex flex-1">
        <Sidebar
          isCollapsed={isSidebarCollapsed}
          toggleSidebar={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
        />
        <main className={cn(
          "flex w-full flex-col transition-all duration-300",
          isSidebarCollapsed ? "md:ml-16" : "md:ml-60"
        )}>
          <div className="flex-1">
            <Outlet />
          </div>
          {/* <Footer /> */}
        </main>
      </div>
    </div>
  )
}
