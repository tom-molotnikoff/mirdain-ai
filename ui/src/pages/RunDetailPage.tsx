import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'

interface RunEvent {
  v: number
  type: string
  run_id: string
  ts: string
  text?: string
  exit_code?: number
}

export default function RunDetailPage() {
  const { id, runId } = useParams()
  const [events, setEvents] = useState<RunEvent[]>([])
  const [error, setError] = useState<string | null>(null)
  const [status, setStatus] = useState<'connecting' | 'connected' | 'closed'>('connecting')

  useEffect(() => {
    if (!runId) return

    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    const ws = new WebSocket(`${proto}://${location.host}/ws/${runId}`)

    ws.onopen = () => setStatus('connected')

    ws.onmessage = (e: MessageEvent<string>) => {
      try {
        const event = JSON.parse(e.data) as RunEvent
        setEvents((prev) => [...prev, event])
      } catch {
        // ignore malformed events
      }
    }

    ws.onerror = () => setError('WebSocket connection error')

    ws.onclose = () => setStatus('closed')

    return () => ws.close()
  }, [runId])

  if (error) {
    return (
      <main>
        <h1>Run {runId}</h1>
        <p role="alert">Error: {error}</p>
      </main>
    )
  }

  return (
    <main>
      <h1>Run {runId}</h1>
      <p>Issue #{id} — {status}</p>
      <ul>
        {events.map((ev, i) => (
          <li key={i}>
            <code>[{ev.ts}] {ev.type}</code>
            {ev.text != null && <span>: {ev.text}</span>}
            {ev.exit_code != null && <span> (exit {ev.exit_code})</span>}
          </li>
        ))}
      </ul>
    </main>
  )
}
