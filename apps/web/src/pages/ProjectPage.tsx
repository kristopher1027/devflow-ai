import { useMemo, useState } from 'react'
import Layout from '../components/Layout'
import { Badge, ErrorText, Loading, Notice } from '../components/ui'
import {
  getImportJob,
  getProject,
  getSyncJob,
  getWorkspace,
  importRepositories,
  listRepositories,
  pollJob,
  requestSync,
} from '../lib/api'
import { formatDate, friendlyError } from '../lib/format'
import { useLoad } from '../lib/hooks'
import { repositoryHref, workspaceHref } from '../lib/route'

interface PageProps {
  id: string
  onLogout: () => void
}

export default function ProjectPage({ id, onLogout }: PageProps) {
  const project = useLoad(() => getProject(id), [id])
  const repositories = useLoad(() => listRepositories(id), [id])

  const workspaceId = project.data?.WorkspaceID
  const workspace = useLoad(
    () => (workspaceId ? getWorkspace(workspaceId) : Promise.resolve(null)),
    [workspaceId],
  )

  const [importing, setImporting] = useState(false)
  const [syncingId, setSyncingId] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [note, setNote] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [filter, setFilter] = useState('all')

  const repoList = repositories.data ?? []
  const syncedCount = repoList.filter((repo) => repo.LastSyncedAt !== null).length
  const privateCount = repoList.filter((repo) => repo.IsPrivate).length
  const publicCount = repoList.length - privateCount

  const filteredRepositories = useMemo(() => {
    const term = search.trim().toLowerCase()

    return repoList.filter((repo) => {
      const matchesSearch =
        !term ||
        repo.FullName.toLowerCase().includes(term) ||
        repo.DefaultBranch.toLowerCase().includes(term)

      const matchesFilter =
        filter === 'all' ||
        (filter === 'private' && repo.IsPrivate) ||
        (filter === 'public' && !repo.IsPrivate) ||
        (filter === 'synced' && repo.LastSyncedAt !== null) ||
        (filter === 'pending' && repo.LastSyncedAt === null)

      return matchesSearch && matchesFilter
    })
  }, [repoList, search, filter])

  async function importFromGitHub() {
    setImporting(true)
    setError(null)
    setNote('Importing repositories from GitHub. Please wait…')

    try {
      const started = await importRepositories(id)
      await pollJob(() => getImportJob(started.id))
      setNote('Repository import completed successfully.')
    } catch (err) {
      setNote(null)
      setError(friendlyError(err))
    } finally {
      await repositories.reload()
      setImporting(false)
    }
  }

  async function sync(repositoryId: string) {
    setSyncingId(repositoryId)
    setError(null)
    setNote('Repository sync is running. This may take a moment…')

    try {
      const started = await requestSync(repositoryId)
      await pollJob(() => getSyncJob(started.id))
      setNote('Repository sync completed successfully.')
    } catch (err) {
      setNote(null)
      setError(friendlyError(err))
    } finally {
      await repositories.reload()
      setSyncingId(null)
    }
  }

  const title = project.data?.Name ?? 'Project'
  const busy = importing || syncingId !== null

  return (
    <Layout
      crumbs={[
        { label: 'Workspaces', href: '#/' },
        ...(workspaceId
          ? [{ label: workspace.data?.Name ?? 'Workspace', href: workspaceHref(workspaceId) }]
          : []),
        { label: title },
      ]}
      onLogout={onLogout}
    >
      <div className="project-heading">
        <div>
          <div className="eyebrow">PROJECT OVERVIEW</div>
          <h1>{title}</h1>
          <p className="muted">
            {project.data?.Description || 'Manage repositories and keep your codebase intelligence up to date.'}
          </p>
        </div>
        <button type="button" disabled={busy} onClick={importFromGitHub}>
          {importing ? 'Importing…' : '＋ Import from GitHub'}
        </button>
      </div>

      <ErrorText message={project.error} />
      <Notice message={note} />
      <ErrorText message={error} />

      <div className="project-stats">
        <div className="project-stat">
          <span className="project-stat-icon">⌘</span>
          <div>
            <strong>{repositories.loading && !repositories.data ? '—' : repoList.length}</strong>
            <span>Repositories</span>
          </div>
        </div>
        <div className="project-stat">
          <span className="project-stat-icon">↻</span>
          <div>
            <strong>{repositories.loading && !repositories.data ? '—' : syncedCount}</strong>
            <span>Previously synced</span>
          </div>
        </div>
        <div className="project-stat">
          <span className="project-stat-icon">◈</span>
          <div>
            <strong>{repositories.loading && !repositories.data ? '—' : privateCount}</strong>
            <span>Private</span>
          </div>
        </div>
        <div className="project-stat">
          <span className="project-stat-icon">◎</span>
          <div>
            <strong>{repositories.loading && !repositories.data ? '—' : publicCount}</strong>
            <span>Public</span>
          </div>
        </div>
      </div>

      <section className="project-repositories">
        <div className="project-section-heading">
          <div>
            <div className="eyebrow">CODEBASE</div>
            <h2>Repositories</h2>
            <p className="muted">Browse connected repositories and manage their latest code snapshots.</p>
          </div>
          <span className="project-count">{repoList.length} total</span>
        </div>

        {repoList.length > 0 && (
          <div className="repository-toolbar">
            <input
              aria-label="Search repositories"
              type="search"
              placeholder="Search by repository or branch…"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
            />
            <select
              aria-label="Filter repositories"
              value={filter}
              onChange={(event) => setFilter(event.target.value)}
            >
              <option value="all">All repositories</option>
              <option value="synced">Previously synced</option>
              <option value="pending">Not yet synced</option>
              <option value="private">Private</option>
              <option value="public">Public</option>
            </select>
          </div>
        )}

        {repositories.loading && !repositories.data && <Loading label="Loading project repositories…" />}
        <ErrorText message={repositories.error} />

        {!repositories.loading && repoList.length === 0 && (
          <div className="project-empty">
            <div className="project-empty-icon">⌘</div>
            <h3>No repositories connected yet</h3>
            <p>
              Import repositories from your connected GitHub account to start exploring your codebase with DevFlow AI.
            </p>
            <button type="button" disabled={busy} onClick={importFromGitHub}>
              {importing ? 'Importing…' : 'Import repositories'}
            </button>
            <span className="muted">Make sure GitHub is connected in your workspace first.</span>
          </div>
        )}

        {repoList.length > 0 && filteredRepositories.length === 0 && (
          <div className="project-empty compact">
            <h3>No matching repositories</h3>
            <p>Try another search term or change the selected filter.</p>
            <button
              type="button"
              className="secondary"
              onClick={() => {
                setSearch('')
                setFilter('all')
              }}
            >
              Clear filters
            </button>
          </div>
        )}

        <div className="repository-grid">
          {filteredRepositories.map((repository) => (
            <article className="repository-card" key={repository.ID}>
              <div className="repository-card-top">
                <div className="repository-mark">GH</div>
                <Badge status={repository.SyncStatus} />
              </div>

              <a className="repository-name" href={repositoryHref(repository.ID)}>
                {repository.FullName}
              </a>

              <div className="repository-meta">
                <span>{repository.IsPrivate ? 'Private repository' : 'Public repository'}</span>
                <span>Branch: {repository.DefaultBranch || 'Not specified'}</span>
              </div>

              <div className="repository-sync-info">
                <span className="repository-sync-dot" />
                {repository.LastSyncedAt
                  ? `Last synced ${formatDate(repository.LastSyncedAt)}`
                  : 'Not synced yet'}
              </div>

              <div className="repository-card-actions">
                <a className="button" href={repositoryHref(repository.ID)}>
                  Explore repository <span>→</span>
                </a>
                <button
                  type="button"
                  className="secondary"
                  disabled={busy}
                  onClick={() => sync(repository.ID)}
                >
                  {syncingId === repository.ID ? 'Syncing…' : '↻ Sync'}
                </button>
              </div>
            </article>
          ))}
        </div>
      </section>
    </Layout>
  )
}
