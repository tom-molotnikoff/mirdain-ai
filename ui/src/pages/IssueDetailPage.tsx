import { useParams } from 'react-router-dom'

export default function IssueDetailPage() {
  const { id } = useParams()
  return (
    <main>
      <h1>Issue #{id}</h1>
      <p>Issue detail — TODO(#20): fetch from /api/issues/{id}.</p>
    </main>
  )
}
