import { useState } from 'react'
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

  async function importFromGitHub() {
    setImporting(true)
    setError(null)
    setNote('Importing repositories from GitHub…')

    try {
      const started = await importRepositories(id)
      await pollJob(() => getImportJob(started.id))
      setNote('Import finished.')
    } catch (err) {
      setNote(null)
      setError(friendlyError(err))
    } finally {
      repositories.reload()
      setImporting(false)
    }
  }

  async function sync(repositoryId: string) {
    setSyncingId(repositoryId)
    setError(null)
    setNote('Syncing… this can take a minute for a large repository.')

    try {
      const started = await requestSync(repositoryId)
      await pollJob(() => getSyncJob(started.id))
      setNote('Sync finished.')
    } catch (err) {
      setNote(null)
      setError(friendlyError(err))
    } finally {
      repositories.reload()
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
      <h1>{title}</h1>
      {project.data?.Description && <p className="muted">{project.data.Description}</p>}
      <ErrorText message={project.error} />

      <div className="row">
        <h2>Repositories</h2>
        <button type="button" disabled={busy} onClick={importFromGitHub}>
          {importing ? 'Importing…' : 'Import from GitHub'}
        </button>
      </div>

      <Notice message={note} />
      <ErrorText message={error ?? repositories.error} />
      {repositories.loading && !repositories.data && <Loading />}

      {repositories.data?.length === 0 && (
        <p className="muted">
          No repositories yet. Connect GitHub on the workspace page, then press “Import from
          GitHub”.
        </p>
      )}

      <div className="stack">
        {repositories.data?.map((repository) => (
          <div key={repository.ID} className="card row">
            <div>
              <a href={repositoryHref(repository.ID)}>
                <strong>{repository.FullName}</strong>
              </a>{' '}
              <Badge status={repository.SyncStatus} />
              <p className="muted">
                {repository.IsPrivate ? 'Private' : 'Public'} · branch{' '}
                {repository.DefaultBranch} · last synced {formatDate(repository.LastSyncedAt)}
              </p>
            </div>
            <div className="actions">
              <button
                type="button"
                className="secondary"
                disabled={busy}
                onClick={() => sync(repository.ID)}
              >
                {syncingId === repository.ID ? 'Syncing…' : 'Sync'}
              </button>
              <a className="button" href={repositoryHref(repository.ID)}>
                Open
              </a>
            </div>
          </div>
        ))}
      </div>
    </Layout>
  )
}
