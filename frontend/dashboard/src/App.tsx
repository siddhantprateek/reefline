import { Routes, Route } from "react-router-dom"
import Layout from "./layout"
import {
  OverviewPage,
  NotFoundPage,
  AnalysisPage,
  OptimizationPage,
  EditorPage,
  IntegrationsPage,
  JobsPage,
  ReportPage,
  SettingsPage
} from "@/pages"
import { Toaster } from "@/components/ui/toaster"

function App() {
  return (
    <>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<OverviewPage />} />
          <Route path="analysis" element={<AnalysisPage />} />
          <Route path="optimization" element={<OptimizationPage />} />
          <Route path="editor" element={<EditorPage />} />
          <Route path="integrations" element={<IntegrationsPage />} />
          <Route path="jobs" element={<JobsPage />} />
          <Route path="jobs/:jobId" element={<ReportPage />} />
          <Route path="settings" element={<SettingsPage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Route>
      </Routes>
      <Toaster />
    </>
  )
}

export default App
