import { ApiError, JobFailedError, NetworkError } from './api'
import type { AnswerSource, Repository } from './types'

export function formatDate(value: string | null | undefined): string {
  if (!value) return 'never'

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return 'unknown'

  return date.toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

/** Adds a hint to the raw GitHub errors people hit while setting up. */
export function explainGitHubFailure(message: string): string {
  if (/status: 404/.test(message)) {
    return `${message}. GitHub could not find the installation saved for this workspace, so its installation ID is probably wrong.`
  }
  if (/status: 401/.test(message)) {
    return `${message}. GitHub rejected the app credentials. Check the server's GITHUB_APP_ID and private key.`
  }
  if (/app ID is required/i.test(message)) {
    return `${message}. The server was started without the GitHub App settings.`
  }
  return message
}

export type ErrorContext = 'ask' | 'explain' | 'generic'

/** Turns API errors into messages a person can act on. */
export function friendlyError(
  error: unknown,
  context: ErrorContext = 'generic',
): string {
  if (error instanceof JobFailedError) {
    return explainGitHubFailure(error.message)
  }

  if (error instanceof ApiError) {
    switch (error.status) {
      case 403:
        return "You don't have access to this."
      case 409:
        if (/connection already exists/i.test(error.message)) {
          return 'A GitHub connection already exists for this workspace, or that installation is already used by another workspace.'
        }
        return context === 'generic'
          ? error.message
          : "This repository hasn't been synced yet. Press Sync first."
      case 404:
        return context === 'ask'
          ? 'No code matched that question. Try different keywords.'
          : error.message
      case 502:
        return 'The AI service returned an error. Try again in a moment.'
      case 503:
        return context === 'generic'
          ? error.message
          : "The AI isn't set up on the server yet (no API key). Everything else still works."
      default:
        return error.message
    }
  }

  if (error instanceof NetworkError) {
    return error.message
  }

  return error instanceof Error ? error.message : 'Something went wrong.'
}

/** Link to the exact lines of a file on GitHub, pinned to the synced commit. */
export function sourceUrl(
  repository: Pick<Repository, 'HTMLURL' | 'DefaultBranch'>,
  commitSha: string,
  source: AnswerSource,
): string {
  const ref = commitSha || repository.DefaultBranch
  const path = source.file_path.split('/').map(encodeURIComponent).join('/')
  const lines =
    source.end_line > source.start_line
      ? `#L${source.start_line}-L${source.end_line}`
      : `#L${source.start_line}`

  return `${repository.HTMLURL}/blob/${ref}/${path}${lines}`
}

export function statusTone(status: string): 'good' | 'warn' | 'bad' | 'neutral' {
  switch (status) {
    case 'synced':
    case 'succeeded':
    case 'active':
      return 'good'
    case 'syncing':
    case 'running':
    case 'pending':
      return 'warn'
    case 'failed':
    case 'error':
    case 'disconnected':
      return 'bad'
    default:
      return 'neutral'
  }
}
