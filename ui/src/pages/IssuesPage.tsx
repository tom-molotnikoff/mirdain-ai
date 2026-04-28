import { useEffect, useState } from 'react'

interface Issue {
  number: number
  title: string
  workflow_label: string
}

export default function IssuesPage() {
  const [issues, setIssues] = useState<Issue[]>([])
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch('/api/issues')
      .then(async (res) => {
        if (!res.ok) {
          const body = await res.json().catch(() => ({}))
          throw new Error(body.error ?? `HTTP ${res.status}`)
        }
        return res.json() as Promise<Issue[]>
      })
      .then((data) => {
        setIssues(data)
        setLoading(false)
      })
      .catch((err: Error) => {
        setError(err.message)
        setLoading(false)
      })
  }, [])

  if (loading) {
    return (
      <main>
        <h1>Issues</h1>
        <p>Loading…</p>
      </main>
    )
  }

  if (error) {
    return (
      <main>
        <h1>Issues</h1>
        <p role="alert">Error: {error}</p>
      </main>
    )
  }

  return (
    <main>
      <h1>Issues</h1>
      {issues.length === 0 ? (
        <p>No mirdain-managed issues found.</p>
      ) : (
        <ul>
          {issues.map((iss) => (
            <li key={iss.number}>
              <a href={`/issues/${iss.number}`}>
                <strong>#{iss.number}</strong> {iss.title}
              </a>{' '}
              <span>{iss.workflow_label}</span>
            </li>
          ))}
        </ul>
      )}
    </main>
  )
}

