import { useParams } from 'react-router-dom'

export default function RunDetailPage() {
  const { id, runId } = useParams()
  return (
    <main>
      <h1>Run {runId}</h1>
      <p>Issue #{id} — run console TODO(#2): stream events from WS /ws/{runId}.</p>
    </main>
  )
}
