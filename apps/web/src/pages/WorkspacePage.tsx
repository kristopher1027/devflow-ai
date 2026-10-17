import { useState, type FormEvent } from 'react'
import Layout from '../components/Layout'
import { Badge, ErrorText, Loading, Notice } from '../components/ui'
import {
  createConnection,
  createProject,
  getConnection,
  getWorkspace,
  listProjects,
  setConnectionStatus,
} from '../lib/api'
import { formatDate, friendlyError } from '../lib/format'
import { useLoad } from '../lib/hooks'
import { projectHref } from '../lib/route'

interface PageProps {
  id: string
  onLogout: () => void
}

export default function WorkspacePage({ id, onLogout }: PageProps) {
  const workspace = useLoad(() => getWorkspace(id), [id])
  const projects = useLoad(() => listProjects(id), [id])

  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function create(event: FormEvent) {
    event.preventDefault()
    setBusy(true)
    setError(null)

    try {
      await createProject(id, name.trim(), description)
      setName('')
      setDescription('')
      projects.reload()
    } catch (err) {
      setError(friendlyError(err))
    } finally {
      setBusy(false)
    }
  }

  const title = workspace.data?.Name ?? 'Workspace'

  return (
    <Layout
      crumbs={[{ label: 'Workspaces', href: '#/' }, { label: title }]}
      onLogout={onLogout}
    >
      <h1>{title}</h1>
      <ErrorText message={workspace.error} />

      <section>
        <h2>GitHub connection</h2>
        <ConnectionPanel workspaceId={id} />
      </section>

      <section>
        <h2>Projects</h2>
        {projects.loading && !projects.data && <Loading />}
        <ErrorText message={projects.error} />

        <div className="grid">
          {projects.data?.map((project) => (
            <a key={project.ID} className="card link-card" href={projectHref(project.ID)}>
              <h3>{project.Name}</h3>
              {project.Description && <p>{project.Description}</p>}
              <p className="muted">Created {formatDate(project.CreatedAt)}</p>
            </a>
          ))}
        </div>

        {projects.data?.length === 0 && (
          <p className="muted">No projects yet. Create one to import repositories into.</p>
        )}

        <form onSubmit={create} className="card stack">
          <label>
            Project name
            <input
              required
              maxLength={100}
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </label>
          <label>
            Description (optional)
            <input
              maxLength={300}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </label>
          <button type="submit" disabled={busy || name.trim() === ''}>
            {busy ? 'Creating…' : 'Create project'}
          </button>
          <ErrorText message={error} />
        </form>
      </section>
    </Layout>
  )
}

function ConnectionPanel({ workspaceId }: { workspaceId: string }) {
  const connection = useLoad(() => getConnection(workspaceId), [workspaceId])
  const [installationId, setInstallationId] = useState('')
  const [accountLogin, setAccountLogin] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [note, setNote] = useState<string | null>(null)

  async function connect(event: FormEvent) {
    event.preventDefault()
    setBusy(true)
    setError(null)
    setNote(null)

    try {
      const created = await createConnection(
        workspaceId,
        installationId.trim(),
        accountLogin.trim(),
      )
      await setConnectionStatus(workspaceId, created.ID, 'active')
      setNote('GitHub connected.')
    } catch (err) {
      setError(friendlyError(err))
    } finally {
      // Always refresh: the connection may exist even if a later step failed.
      connection.reload()
      setBusy(false)
    }
  }

  async function activate(id: string) {
    setBusy(true)
    setError(null)

    try {
      await setConnectionStatus(workspaceId, id, 'active')
      connection.reload()
    } catch (err) {
      setError(friendlyError(err))
    } finally {
      setBusy(false)
    }
  }

  if (connection.loading && !connection.data && !connection.error) {
    return <Loading />
  }

  if (connection.data) {
    const saved = connection.data

    return (
      <div className="card stack">
        <div className="row">
          <div>
            <strong>GitHub App connected</strong> <Badge status={saved.Status} />
            <p className="muted">
              Installation {saved.InstallationID} · account {saved.AccountLogin}
            </p>
          </div>
          {saved.Status !== 'active' && (
            <button type="button" disabled={busy} onClick={() => activate(saved.ID)}>
              Activate
            </button>
          )}
        </div>
        <ErrorText message={error} />
        <Notice message={note} />
      </div>
    )
  }

  return (
    <form onSubmit={connect} className="card stack">
      <p className="muted">
        Install your GitHub App on your account, then enter its installation ID. You can
        find it in the address of the page GitHub shows after installing:
        github.com/settings/installations/<strong>&lt;number&gt;</strong>
      </p>
      <label>
        Installation ID
        <input
          required
          inputMode="numeric"
          pattern="[0-9]+"
          value={installationId}
          onChange={(e) => setInstallationId(e.target.value)}
        />
      </label>
      <label>
        GitHub account
        <input
          required
          placeholder="your-github-username"
          value={accountLogin}
          onChange={(e) => setAccountLogin(e.target.value)}
        />
      </label>
      <button type="submit" disabled={busy}>
        {busy ? 'Connecting…' : 'Connect GitHub'}
      </button>
      <ErrorText message={error ?? connection.error} />
      <Notice message={note} />
    </form>
  )
}
