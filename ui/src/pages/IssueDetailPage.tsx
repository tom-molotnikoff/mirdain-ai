import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

export default function IssueDetailPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const startAgent = async () => {
    setLoading(true)
    setError(null)
    try {
      const res = await fetch('/api/runs', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ issue_id: id }),
      })
      if (!res.ok) {
        const body = await res.json().catch(() => ({})) as { error?: string }
        throw new Error(body.error ?? `HTTP ${res.status}`)
      }
      const data = await res.json() as { run_id: string }
      navigate(`/issues/${id}/runs/${data.run_id}`)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <main>
      <h1>Issue #{id}</h1>
      {error && <p role="alert">Error: {error}</p>}
      <button onClick={startAgent} disabled={loading}>
        {loading ? 'Starting…' : 'Start agent'}
      </button>
    </main>
  )
}
