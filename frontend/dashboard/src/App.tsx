import { Routes, Route } from "react-router-dom"
import Layout from "./layout"
import {
  OverviewPage,
  NotFoundPage,
  AnalysisPage,
  OptimizationPage,
  EditorPage,
  IntegrationsPage,
  HistoryPage,
  SettingsPage
} from "@/pages"

function App() {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<OverviewPage />} />
        <Route path="analysis" element={<AnalysisPage />} />
        <Route path="optimization" element={<OptimizationPage />} />
        <Route path="editor" element={<EditorPage />} />
        <Route path="integrations" element={<IntegrationsPage />} />
        <Route path="history" element={<HistoryPage />} />
        <Route path="settings" element={<SettingsPage />} />
        <Route path="*" element={<NotFoundPage />} />
      </Route>
    </Routes>
  )
}

export default App
