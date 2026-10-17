import { useState, type FormEvent } from 'react'
import Layout from '../components/Layout'
import { ErrorText, Loading } from '../components/ui'
import { createWorkspace, listWorkspaces } from '../lib/api'
import { formatDate, friendlyError } from '../lib/format'
import { useLoad } from '../lib/hooks'
import { workspaceHref } from '../lib/route'

export default function WorkspacesPage({ onLogout }: { onLogout: () => void }) {
  const workspaces = useLoad(listWorkspaces, [])
  const [name, setName] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function create(event: FormEvent) {
    event.preventDefault()
    setBusy(true)
    setError(null)

    try {
      await createWorkspace(name.trim())
      setName('')
      workspaces.reload()
    } catch (err) {
      setError(friendlyError(err))
    } finally {
      setBusy(false)
    }
  }

  return (
    <Layout crumbs={[{ label: 'Workspaces' }]} onLogout={onLogout}>
      <h1>Workspaces</h1>
      <p className="muted">A workspace groups your projects and one GitHub connection.</p>

      {workspaces.loading && !workspaces.data && <Loading />}
      <ErrorText message={workspaces.error} />

      <div className="grid">
        {workspaces.data?.map((workspace) => (
          <a key={workspace.ID} className="card link-card" href={workspaceHref(workspace.ID)}>
            <h2>{workspace.Name}</h2>
            <p className="muted">Created {formatDate(workspace.CreatedAt)}</p>
          </a>
        ))}
      </div>

      {workspaces.data?.length === 0 && (
        <p className="muted">No workspaces yet. Create your first one below.</p>
      )}

      <form onSubmit={create} className="card inline-form">
        <label>
          New workspace
          <input
            required
            maxLength={100}
            placeholder="e.g. My projects"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
        </label>
        <button type="submit" disabled={busy || name.trim() === ''}>
          {busy ? 'Creating…' : 'Create'}
        </button>
      </form>
      <ErrorText message={error} />
    </Layout>
  )
}
