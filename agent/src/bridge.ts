import WebSocket from 'ws'

// Environment variables injected by the orchestrator at container start.
const runID = process.env.MIRDAIN_RUN_ID
const secret = process.env.MIRDAIN_RUN_SECRET
const orchestratorURL = process.env.MIRDAIN_ORCHESTRATOR_URL

if (!runID || !secret || !orchestratorURL) {
  console.error(
    'mirdain-bridge: missing required env vars ' +
      '(MIRDAIN_RUN_ID, MIRDAIN_RUN_SECRET, MIRDAIN_ORCHESTRATOR_URL)'
  )
  process.exit(1)
}

const wsURL = `${orchestratorURL}/internal/agent/${runID}?secret=${secret}`

const ws = new WebSocket(wsURL)

const now = () => new Date().toISOString()

const send = (payload: object) => ws.send(JSON.stringify(payload))

ws.on('open', () => {
  send({ v: 1, type: 'run.started', run_id: runID, ts: now() })
  send({ v: 1, type: 'text.delta', run_id: runID, ts: now(), text: 'Tracer skill running…' })
  send({ v: 1, type: 'run.completed', run_id: runID, ts: now(), exit_code: 0 })
  ws.close()
})

ws.on('error', (err) => {
  console.error('mirdain-bridge: WebSocket error:', err.message)
  process.exit(1)
})
