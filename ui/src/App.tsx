import { BrowserRouter, Routes, Route } from 'react-router-dom'
import HomePage from './pages/HomePage'
import IssuesPage from './pages/IssuesPage'
import IssueDetailPage from './pages/IssueDetailPage'
import RunDetailPage from './pages/RunDetailPage'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/issues" element={<IssuesPage />} />
        <Route path="/issues/:id" element={<IssueDetailPage />} />
        <Route path="/issues/:id/runs/:runId" element={<RunDetailPage />} />
      </Routes>
    </BrowserRouter>
  )
}
