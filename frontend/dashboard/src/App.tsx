import { Routes, Route } from "react-router-dom"
import Layout from "./layout"
import { OverviewPage } from "@/pages/overview/overview.page"
import { NotFoundPage } from "@/pages/notfound/notfound.page"

function App() {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<OverviewPage />} />
        <Route path="*" element={<NotFoundPage />} />
      </Route>
    </Routes>
  )
}

export default App
