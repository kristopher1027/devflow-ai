import { useEffect, useState, type FormEvent } from 'react'
import Layout from '../components/Layout'
import { Badge, ErrorText, Loading, Notice } from '../components/ui'
import {
  createConnection,
  createProject,
  getConnection,
  getWorkspace,
  listProjects,
  listRepositories,
  setConnectionStatus,
} from '../lib/api'
import { formatDate, friendlyError } from '../lib/format'
import { useLoad } from '../lib/hooks'
import { projectHref } from '../lib/route'
import type { Repository } from '../lib/types'

interface PageProps {
  id: string
  onLogout: () => void
}

export default function WorkspacePage({ id, onLogout }: PageProps) {
  const workspace = useLoad(() => getWorkspace(id), [id])
  const projects = useLoad(() => listProjects(id), [id])

  const [repositories, setRepositories] = useState<Repository[]>([])
  const [repositoriesLoading, setRepositoriesLoading] = useState(false)
  const [repositoriesError, setRepositoriesError] = useState<string | null>(null)

  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [note, setNote] = useState<string | null>(null)

  useEffect(() => {
    let active = true

    async function loadRepositories() {
      if (!projects.data) return

      if (projects.data.length === 0) {
        setRepositories([])
        setRepositoriesError(null)
        return
      }

      setRepositoriesLoading(true)
      setRepositoriesError(null)

      try {
        const results = await Promise.all(
          projects.data.map((project) => listRepositories(project.ID)),
        )

        if (active) {
          setRepositories(results.flat())
        }
      } catch {
        if (active) {
          setRepositoriesError('Could not load repository statistics.')
        }
      } finally {
        if (active) setRepositoriesLoading(false)
      }
    }

    void loadRepositories()

    return () => {
      active = false
    }
  }, [projects.data])

  async function create(event: FormEvent) {
    event.preventDefault()
    setBusy(true)
    setError(null)
    setNote(null)

    try {
      await createProject(id, name.trim(), description)
      setName('')
      setDescription('')
      setNote('Project created successfully.')
      projects.reload()
    } catch (err) {
      setError(friendlyError(err))
    } finally {
      setBusy(false)
    }
  }

  const title = workspace.data?.Name ?? 'Workspace'
  const projectList = projects.data ?? []
  const recentProjects = [...projectList]
    .sort((a, b) => new Date(b.CreatedAt).getTime() - new Date(a.CreatedAt).getTime())
    .slice(0, 4)

  const syncedCount = repositories.filter(
    (repository) => repository.SyncStatus.toLowerCase() === 'synced',
  ).length

  return (
    <Layout
      crumbs={[{ label: 'Workspaces', href: '#/' }, { label: title }]}
      onLogout={onLogout}
    >
      <div className="dashboard-heading">
        <div>
          <p className="eyebrow">WORKSPACE OVERVIEW</p>
          <h1>{title}</h1>
          <p className="dashboard-subtitle">
            Your engineering workspace at a glance.
          </p>
        </div>
        <a className="button dashboard-primary-action" href="#create-project">
          + New project
        </a>
      </div>

      <ErrorText message={workspace.error} />
      <ErrorText message={projects.error} />

      <section className="dashboard-stats" aria-label="Workspace statistics">
        <article className="stat-card">
          <div className="stat-icon purple">◈</div>
          <p className="stat-label">Total projects</p>
          <strong className="stat-value">
            {projects.loading && !projects.data ? '—' : projectList.length}
          </strong>
          <span className="stat-footnote">In this workspace</span>
        </article>

        <article className="stat-card">
          <div className="stat-icon blue">⌘</div>
          <p className="stat-label">Repositories</p>
          <strong className="stat-value">
            {repositoriesLoading ? '—' : repositories.length}
          </strong>
          <span className="stat-footnote">Across all projects</span>
        </article>

        <article className="stat-card">
          <div className="stat-icon green">✓</div>
          <p className="stat-label">Synced repositories</p>
          <strong className="stat-value">
            {repositoriesLoading ? '—' : syncedCount}
          </strong>
          <span className="stat-footnote">Successfully synced</span>
        </article>

        <article className="stat-card">
          <div className="stat-icon orange">◷</div>
          <p className="stat-label">Workspace created</p>
          <strong className="stat-date">
            {workspace.data ? formatDate(workspace.data.CreatedAt) : '—'}
          </strong>
          <span className="stat-footnote">Workspace creation date</span>
        </article>
      </section>

      <ErrorText message={repositoriesError} />

      <div className="dashboard-main-grid">
        <section className="dashboard-panel">
          <div className="panel-heading">
            <div>
              <p className="eyebrow">YOUR WORK</p>
              <h2>Recent projects</h2>
            </div>
            <span className="panel-count">{projectList.length} total</span>
          </div>

          {projects.loading && !projects.data && <Loading />}

          {projectList.length === 0 && !projects.loading && (
            <div className="dashboard-empty">
              <span className="empty-icon">◈</span>
              <h3>No projects yet</h3>
              <p>Create your first project to start importing repositories.</p>
              <a className="button" href="#create-project">Create project</a>
            </div>
          )}

          <div className="dashboard-project-list">
            {recentProjects.map((project) => {
              const count = repositories.filter(
                (repository) => repository.ProjectID === project.ID,
              ).length

              return (
                <a
                  key={project.ID}
                  className="dashboard-project"
                  href={projectHref(project.ID)}
                >
                  <div className="project-avatar">
                    {project.Name.slice(0, 1).toUpperCase()}
                  </div>
                  <div className="project-info">
                    <strong>{project.Name}</strong>
                    <span>{project.Description || 'No description provided'}</span>
                    <small>Created {formatDate(project.CreatedAt)}</small>
                  </div>
                  <div className="project-meta">
                    <span>{repositoriesLoading ? '…' : count}</span>
                    <small>Repos</small>
                  </div>
                  <span className="project-arrow">→</span>
                </a>
              )
            })}
          </div>
        </section>

        <div className="dashboard-side-column">
          <ConnectionPanel workspaceId={id} />

          <section className="dashboard-tip">
            <span className="tip-icon">✦</span>
            <p className="eyebrow">DEVFLOW AI</p>
            <h3>Understand your codebase.</h3>
            <p>
              Import a repository, sync its code, and use AI to explore its
              architecture and implementation.
            </p>
          </section>
        </div>
      </div>

      <section className="dashboard-create-section" id="create-project">
        <div>
          <p className="eyebrow">GET STARTED</p>
          <h2>Create a project</h2>
          <p className="muted">
            Organize repositories and development work under this workspace.
          </p>
        </div>

        <form onSubmit={create} className="dashboard-create-form">
          <label>
            Project name
            <input
              required
              maxLength={100}
              placeholder="e.g. Customer Portal"
              value={name}
              onChange={(event) => setName(event.target.value)}
            />
          </label>

          <label>
            Description
            <input
              maxLength={300}
              placeholder="What are you building?"
              value={description}
              onChange={(event) => setDescription(event.target.value)}
            />
          </label>

          <div className="row">
            <ErrorText message={error} />
            <Notice message={note} />
            <button type="submit" disabled={busy || name.trim() === ''}>
              {busy ? 'Creating…' : 'Create project'}
            </button>
          </div>
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

  return (
    <section className="dashboard-panel github-panel">
      <div className="panel-heading">
        <div>
          <p className="eyebrow">INTEGRATIONS</p>
          <h2>GitHub connection</h2>
        </div>
        <span className="github-mark">GH</span>
      </div>

      {connection.loading && !connection.data && !connection.error && <Loading />}

      {connection.data && (
        <div className="github-connected">
          <div className="connection-indicator">
            <span className="connection-dot" />
            <strong>GitHub App connected</strong>
          </div>
          <Badge status={connection.data.Status} />
          <p className="muted">
            Account: <strong>{connection.data.AccountLogin}</strong>
          </p>
          <p className="muted">
            Installation ID: {connection.data.InstallationID}
          </p>
          {connection.data.Status !== 'active' && (
            <button
              type="button"
              disabled={busy}
              onClick={() => activate(connection.data!.ID)}
            >
              {busy ? 'Activating…' : 'Activate connection'}
            </button>
          )}
        </div>
      )}

      {!connection.data && !connection.loading && (
        <form onSubmit={connect} className="github-form">
          <p className="muted">
            Connect your GitHub App to import repositories into your projects.
          </p>
          <label>
            Installation ID
            <input
              required
              inputMode="numeric"
              pattern="[0-9]+"
              placeholder="GitHub installation ID"
              value={installationId}
              onChange={(event) => setInstallationId(event.target.value)}
            />
          </label>
          <label>
            GitHub account
            <input
              required
              placeholder="your-github-username"
              value={accountLogin}
              onChange={(event) => setAccountLogin(event.target.value)}
            />
          </label>
          <button type="submit" disabled={busy}>
            {busy ? 'Connecting…' : 'Connect GitHub'}
          </button>
        </form>
      )}

      <ErrorText message={error ?? connection.error} />
      <Notice message={note} />
    </section>
  )
}
