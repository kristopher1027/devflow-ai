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
    setSyncNote('Syncing repository. This may take a moment…')

    try {
      const started = await requestSync(id)
      await pollJob(() => getSyncJob(started.id))
      setSyncNote('Repository sync completed successfully.')
    } catch (err) {
      setSyncNote(null)
      setSyncError(friendlyError(err))
    } finally {
      repository.reload()
      setSyncing(false)
    }
  }

  async function ask(event?: FormEvent) {
    event?.preventDefault()
    const text = question.trim()
    if (text === '' || asking) return

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
      event.preventDefault()
      void ask()
    }
  }

  async function explain() {
    if (explaining) return

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
      {repository.loading && !repo && <Loading label="Loading repository…" />}

      {repo && (
        <>
          <div className="repo-heading">
            <div className="repo-heading-main">
              <div className="eyebrow">REPOSITORY WORKSPACE</div>
              <h1>{repo.FullName}</h1>
              <p className="muted">
                Explore your codebase, ask questions, and understand how your software works.
              </p>
              <div className="repo-heading-tags">
                <Badge status={repo.SyncStatus} />
                <span>{repo.IsPrivate ? 'Private repository' : 'Public repository'}</span>
                <span>Default branch: {repo.DefaultBranch || 'Not specified'}</span>
              </div>
            </div>

            <div className="repo-heading-actions">
              <a href={repo.HTMLURL} target="_blank" rel="noreferrer" className="button secondary">
                View on GitHub ↗
              </a>
              <button type="button" disabled={syncing} onClick={sync}>
                {syncing ? 'Syncing…' : '↻ Sync repository'}
              </button>
            </div>
          </div>

          <Notice message={syncNote} />
          <ErrorText message={syncError} />

          <div className="repo-overview-stats">
            <div className="repo-overview-stat">
              <span className="repo-overview-icon">◈</span>
              <div>
                <span>Repository access</span>
                <strong>{repo.IsPrivate ? 'Private' : 'Public'}</strong>
              </div>
            </div>
            <div className="repo-overview-stat">
              <span className="repo-overview-icon">⌘</span>
              <div>
                <span>Default branch</span>
                <strong>{repo.DefaultBranch || 'Not specified'}</strong>
              </div>
            </div>
            <div className="repo-overview-stat">
              <span className="repo-overview-icon">↻</span>
              <div>
                <span>Last synced</span>
                <strong>{formatDate(repo.LastSyncedAt)}</strong>
              </div>
            </div>
          </div>

          <section className="repo-ai-section">
            <div className="repo-section-heading">
              <div>
                <div className="eyebrow">DEVFLOW AI ASSISTANT</div>
                <h2>Understand your codebase</h2>
                <p className="muted">
                  Ask questions about this repository and get answers grounded in its indexed code.
                </p>
              </div>
              <span className="repo-ai-indicator">✦ AI POWERED</span>
            </div>

            <form onSubmit={ask} className="repo-question-panel">
              <label htmlFor="repository-question">What would you like to know?</label>
              <textarea
                id="repository-question"
                rows={4}
                maxLength={MAX_QUESTION}
                placeholder="Ask about architecture, authentication, APIs, database access, background jobs…"
                value={question}
                onChange={(event) => setQuestion(event.target.value)}
                onKeyDown={onQuestionKeyDown}
              />
              <div className="repo-question-footer">
                <span className="muted">
                  {question.length}/{MAX_QUESTION} · Ctrl/⌘ + Enter to send
                </span>
                <button type="submit" disabled={asking || question.trim() === ''}>
                  {asking ? 'Analyzing code…' : 'Ask DevFlow AI →'}
                </button>
              </div>

              <div className="repo-prompt-area">
                <span className="muted">Try asking</span>
                <div className="chips">
                  {EXAMPLE_QUESTIONS.map((example) => (
                    <button
                      key={example}
                      type="button"
                      className="chip"
                      disabled={asking}
                      onClick={() => setQuestion(example)}
                    >
                      {example}
                    </button>
                  ))}
                </div>
              </div>
            </form>

            <ErrorText message={askError} />

            {asking && (
              <div className="repo-thinking">
                <span className="repo-thinking-dot" />
                DevFlow AI is analyzing the repository…
              </div>
            )}

            {answer && (
              <div className="repo-result-card">
                <div className="repo-result-heading">
                  <div>
                    <div className="eyebrow">AI RESPONSE</div>
                    <h3>Answer</h3>
                  </div>
                  <span className="repo-commit">Commit {answer.commit_sha.slice(0, 7)}</span>
                </div>
                <div className="repo-question-asked">
                  <span>YOUR QUESTION</span>
                  <p>{askedQuestion}</p>
                </div>
                <Markdown text={answer.answer} />

                {answer.sources.length > 0 && (
                  <div className="repo-sources">
                    <h3>Source references</h3>
                    <p className="muted">Open the referenced files on GitHub.</p>
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
                            <span>↗</span>
                          </a>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>
            )}
          </section>

          <section className="repo-explain-section">
            <div className="repo-section-heading">
              <div>
                <div className="eyebrow">CODEBASE OVERVIEW</div>
                <h2>Repository explanation</h2>
                <p className="muted">
                  Generate an AI explanation of the repository based on its indexed code.
                </p>
              </div>
              <button type="button" className="secondary" disabled={explaining} onClick={explain}>
                {explaining
                  ? 'Analyzing…'
                  : explanation
                    ? 'Regenerate explanation'
                    : '✦ Explain repository'}
              </button>
            </div>

            <ErrorText message={explainError} />

            {explaining && (
              <div className="repo-thinking">
                <span className="repo-thinking-dot" />
                Preparing your repository overview…
              </div>
            )}

            {explanation ? (
              <div className="repo-result-card">
                <div className="repo-result-heading">
                  <h3>About this codebase</h3>
                  <span className="repo-commit">Commit {explanation.commit_sha.slice(0, 7)}</span>
                </div>
                <Markdown text={explanation.explanation} />
              </div>
            ) : (
              !explaining && (
                <div className="repo-explain-empty">
                  <span className="repo-overview-icon">✦</span>
                  <div>
                    <strong>Get a guided overview of this repository</strong>
                    <p className="muted">
                      Use DevFlow AI to generate an explanation of the codebase and its structure.
                    </p>
                  </div>
                </div>
              )
            )}
          </section>
        </>
      )}
    </Layout>
  )
}
