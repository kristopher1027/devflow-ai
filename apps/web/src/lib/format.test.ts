import { describe, expect, it } from 'vitest'
import { ApiError, JobFailedError, NetworkError } from './api'
import { explainGitHubFailure, formatDate, friendlyError, sourceUrl, statusTone } from './format'

describe('friendlyError', () => {
  it('explains AI-specific failures for the ask and explain actions', () => {
    expect(friendlyError(new ApiError(503, 'AI explanations are not configured'), 'ask')).toMatch(
      /isn't set up/,
    )
    expect(friendlyError(new ApiError(404, 'no code matched'), 'ask')).toMatch(
      /different keywords/,
    )
    expect(friendlyError(new ApiError(409, 'not synced'), 'explain')).toMatch(/Press Sync/)
    expect(friendlyError(new ApiError(502, 'AI provider error'), 'ask')).toMatch(/Try again/)
  })

  it('keeps the server message for generic failures', () => {
    expect(friendlyError(new ApiError(404, 'repository not found'))).toBe('repository not found')
    expect(friendlyError(new ApiError(409, 'duplicate'))).toBe('duplicate')
    expect(friendlyError(new ApiError(500, 'internal server error'))).toBe(
      'internal server error',
    )
  })

  it('handles permission, network and job failures', () => {
    expect(friendlyError(new ApiError(403, 'forbidden'))).toMatch(/don't have access/)
    expect(friendlyError(new NetworkError())).toMatch(/Can't reach/)
    // Programming errors must not be disguised as network problems.
    expect(friendlyError(new TypeError('x is undefined'))).toBe('x is undefined')
    expect(
      friendlyError(
        new JobFailedError({
          id: 'j',
          status: 'failed',
          failureCode: 'import_failed',
          failureMessage: 'github app ID is required',
        }),
      ),
    ).toMatch(/^github app ID is required\./)
  })
})

describe('sourceUrl', () => {
  const repository = { HTMLURL: 'https://github.com/me/app', DefaultBranch: 'main' }

  it('links to the synced commit and the exact line range', () => {
    expect(
      sourceUrl(repository, 'abc123', { file_path: 'src/a b.go', start_line: 3, end_line: 9 }),
    ).toBe('https://github.com/me/app/blob/abc123/src/a%20b.go#L3-L9')
  })

  it('uses a single line anchor and falls back to the branch', () => {
    expect(sourceUrl(repository, '', { file_path: 'x.go', start_line: 5, end_line: 5 })).toBe(
      'https://github.com/me/app/blob/main/x.go#L5',
    )
  })
})

describe('small helpers', () => {
  it('maps statuses to tones', () => {
    expect(statusTone('synced')).toBe('good')
    expect(statusTone('running')).toBe('warn')
    expect(statusTone('failed')).toBe('bad')
    expect(statusTone('whatever')).toBe('neutral')
  })

  it('formats dates defensively', () => {
    expect(formatDate(null)).toBe('never')
    expect(formatDate('not a date')).toBe('unknown')
    expect(formatDate('2026-10-14T10:00:00Z')).toMatch(/2026/)
  })
})

describe('GitHub setup hints', () => {
  it('explains the common GitHub failures', () => {
    expect(explainGitHubFailure('github API returned unexpected status: 404 Not Found')).toMatch(
      /installation ID is probably wrong/,
    )
    expect(explainGitHubFailure('github API returned unexpected status: 401 Unauthorized')).toMatch(
      /rejected the app credentials/,
    )
    expect(explainGitHubFailure('github app ID is required')).toMatch(/without the GitHub App/)
    expect(explainGitHubFailure('something else')).toBe('something else')
  })

  it('applies the hints to failed jobs', () => {
    const job = {
      id: 'j',
      status: 'failed',
      failureCode: 'import_failed',
      failureMessage: 'github API returned unexpected status: 404 Not Found',
    }

    expect(friendlyError(new JobFailedError(job))).toMatch(/installation ID is probably wrong/)
  })

  it('explains duplicate connections', () => {
    expect(friendlyError(new ApiError(409, 'github connection already exists'))).toMatch(
      /already used by another workspace/,
    )
  })
})
