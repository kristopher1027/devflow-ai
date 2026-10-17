import { useState, type FormEvent, type KeyboardEvent } from 'react'
import Layout from '../components/Layout'
import { Badge, ErrorText, Loading, Markdown, Notice } from '../components/ui'
import {
  askQuestion,
  explainRepository,
  getProject,
  getRepository,
  getSyncJob,
  pollJob,
  requestSync,
} from '../lib/api'
import { formatDate, friendlyError, sourceUrl } from '../lib/format'
import { useLoad } from '../lib/hooks'
import { projectHref } from '../lib/route'
import type { Answer, Explanation } from '../lib/types'

const MAX_QUESTION = 500

const EXAMPLE_QUESTIONS = [
  'How does authentication work?',
  'Where are background jobs handled?',
  'How is the database accessed?',
]

interface PageProps {
  id: string
  onLogout: () => void
}

export default function RepositoryPage({ id, onLogout }: PageProps) {
  const repository = useLoad(() => getRepository(id), [id])

  const projectId = repository.data?.ProjectID
  const project = useLoad(
    () => (projectId ? getProject(projectId) : Promise.resolve(null)),
    [projectId],
  )

  const [syncing, setSyncing] = useState(false)
  const [syncNote, setSyncNote] = useState<string | null>(null)
  const [syncError, setSyncError] = useState<string | null>(null)

  const [question, setQuestion] = useState('')
  const [asking, setAsking] = useState(false)
  const [answer, setAnswer] = useState<Answer | null>(null)
  const [askedQuestion, setAskedQuestion] = useState('')
  const [askError, setAskError] = useState<string | null>(null)

  const [explaining, setExplaining] = useState(false)
  const [explanation, setExplanation] = useState<Explanation | null>(null)
  const [explainError, setExplainError] = useState<string | null>(null)

  async function sync() {
    setSyncing(true)
    setSyncError(null)
    setSyncNote('Syncing… this can take a minute for a large repository.')

    try {
      const started = await requestSync(id)
      await pollJob(() => getSyncJob(started.id))
      setSyncNote('Sync finished. Ask away!')
    } catch (err) {
      setSyncNote(null)
      setSyncError(friendlyError(err))
    } finally {
      // Refresh even after a failure: the sync may have finished on the server.
      repository.reload()
      setSyncing(false)
    }
  }

  async function ask(event?: FormEvent) {
    event?.preventDefault()
    const text = question.trim()
    if (text === '') return

    setAsking(true)
    setAskError(null)

    try {
      const result = await askQuestion(id, text)
      setAnswer(result)
      setAskedQuestion(text)
    } catch (err) {
      setAnswer(null)
      setAskError(friendlyError(err, 'ask'))
    } finally {
      setAsking(false)
    }
  }

  function onQuestionKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
    if (event.key === 'Enter' && (event.metaKey || event.ctrlKey)) {
      void ask()
    }
  }

  async function explain() {
    setExplaining(true)
    setExplainError(null)

    try {
      setExplanation(await explainRepository(id))
    } catch (err) {
      setExplanation(null)
      setExplainError(friendlyError(err, 'explain'))
    } finally {
      setExplaining(false)
    }
  }

  const repo = repository.data
  const title = repo?.FullName ?? 'Repository'

  return (
    <Layout
      crumbs={[
        { label: 'Workspaces', href: '#/' },
        ...(projectId
          ? [{ label: project.data?.Name ?? 'Project', href: projectHref(projectId) }]
          : []),
        { label: title },
      ]}
      onLogout={onLogout}
    >
      <ErrorText message={repository.error} />
      {repository.loading && !repo && <Loading />}

      {repo && (
        <>
          <div className="row">
            <div>
              <h1>{repo.FullName}</h1>
              <p className="muted">
                <Badge status={repo.SyncStatus} /> · branch {repo.DefaultBranch} · last synced{' '}
                {formatDate(repo.LastSyncedAt)} ·{' '}
                <a href={repo.HTMLURL} target="_blank" rel="noreferrer">
                  View on GitHub
                </a>
              </p>
            </div>
            <button type="button" className="secondary" disabled={syncing} onClick={sync}>
              {syncing ? 'Syncing…' : 'Sync now'}
            </button>
          </div>
          <Notice message={syncNote} />
          <ErrorText message={syncError} />

          <section>
            <h2>Ask about this code</h2>
            <form onSubmit={ask} className="card stack">
              <label>
                Your question
                <textarea
                  rows={3}
                  maxLength={MAX_QUESTION}
                  placeholder="e.g. How are interrupted jobs recovered after a restart?"
                  value={question}
                  onChange={(e) => setQuestion(e.target.value)}
                  onKeyDown={onQuestionKeyDown}
                />
              </label>
              <div className="row">
                <span className="muted">
                  {question.length}/{MAX_QUESTION} · Ctrl/⌘ + Enter to send
                </span>
                <button type="submit" disabled={asking || question.trim() === ''}>
                  {asking ? 'Thinking…' : 'Ask'}
                </button>
              </div>
              <div className="chips">
                {EXAMPLE_QUESTIONS.map((example) => (
                  <button
                    key={example}
                    type="button"
                    className="chip"
                    onClick={() => setQuestion(example)}
                  >
                    {example}
                  </button>
                ))}
              </div>
            </form>
            <ErrorText message={askError} />

            {answer && (
              <div className="card answer">
                <p className="muted">Q: {askedQuestion}</p>
                <Markdown text={answer.answer} />
                {answer.sources.length > 0 && (
                  <>
                    <h3>Sources</h3>
                    <ul className="sources">
                      {answer.sources.map((source, index) => (
                        <li key={`${source.file_path}-${source.start_line}-${index}`}>
                          <a
                            href={sourceUrl(repo, answer.commit_sha, source)}
                            target="_blank"
                            rel="noreferrer"
                          >
                            <code>
                              {source.file_path}:{source.start_line}-{source.end_line}
                            </code>
                          </a>
                        </li>
                      ))}
                    </ul>
                  </>
                )}
                <p className="muted">Commit {answer.commit_sha.slice(0, 7)}</p>
              </div>
            )}
          </section>

          <section>
            <div className="row">
              <h2>Explain this repository</h2>
              <button type="button" disabled={explaining} onClick={explain}>
                {explaining ? 'Reading the code…' : explanation ? 'Explain again' : 'Explain'}
              </button>
            </div>
            <ErrorText message={explainError} />
            {explanation && (
              <div className="card answer">
                <Markdown text={explanation.explanation} />
                <p className="muted">Commit {explanation.commit_sha.slice(0, 7)}</p>
              </div>
            )}
          </section>
        </>
      )}
    </Layout>
  )
}
