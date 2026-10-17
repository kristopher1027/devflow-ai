import type {
  Answer,
  Explanation,
  GitHubConnection,
  Job,
  Project,
  Repository,
  Workspace,
} from './types'

export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

/** The request never got an HTTP response (server down, tunnel hiccup...). */
export class NetworkError extends Error {
  constructor() {
    super("Can't reach the DevFlow API. Is the server running?")
    this.name = 'NetworkError'
  }
}

export class JobFailedError extends Error {
  job: Job

  constructor(job: Job) {
    super(job.failureMessage || job.failureCode || 'The job failed')
    this.name = 'JobFailedError'
    this.job = job
  }
}

let unauthorizedHandler: (() => void) | null = null

/** Called when an authenticated request comes back 401 (session expired). */
export function setUnauthorizedHandler(handler: (() => void) | null): void {
  unauthorizedHandler = handler
}

async function request<T>(
  method: string,
  path: string,
  body?: unknown,
): Promise<T> {
  let response: Response
  try {
    response = await fetch(`/api${path}`, {
      method,
      credentials: 'same-origin',
      headers: body !== undefined ? { 'Content-Type': 'application/json' } : undefined,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    })
  } catch {
    throw new NetworkError()
  }

  if (!response.ok) {
    const text = (await response.text().catch(() => '')).trim()

    if (response.status === 401 && !path.startsWith('/auth/')) {
      unauthorizedHandler?.()
    }

    throw new ApiError(
      response.status,
      text || response.statusText || `Request failed (${response.status})`,
    )
  }

  if (response.status === 204) {
    return undefined as T
  }

  const text = await response.text()
  if (!text) {
    return undefined as T
  }

  return JSON.parse(text) as T
}

function normalizeJob(raw: Record<string, unknown>): Job {
  const pick = (a: string, b: string) => raw[a] ?? raw[b]

  return {
    id: String(pick('ID', 'id') ?? pick('job_id', 'jobId') ?? ''),
    status: String(pick('Status', 'status') ?? 'pending'),
    failureCode: (pick('FailureCode', 'failure_code') as string | null) ?? null,
    failureMessage:
      (pick('FailureMessage', 'failure_message') as string | null) ?? null,
  }
}

// ---- auth -----------------------------------------------------------------

export const register = (email: string, password: string) =>
  request<void>('POST', '/auth/register', { email, password })

export const login = (email: string, password: string) =>
  request<void>('POST', '/auth/login', { email, password })

export const logout = () => request<void>('POST', '/auth/logout')

// ---- workspaces & projects --------------------------------------------------

export const listWorkspaces = async () =>
  (await request<Workspace[] | null>('GET', '/workspaces')) ?? []

export const createWorkspace = (name: string) =>
  request<Workspace>('POST', '/workspaces', { name })

export const getWorkspace = (id: string) =>
  request<Workspace>('GET', `/workspaces/${encodeURIComponent(id)}`)

export const listProjects = async (workspaceId: string) =>
  (await request<Project[] | null>(
    'GET',
    `/workspaces/${encodeURIComponent(workspaceId)}/projects`,
  )) ?? []

export const createProject = (
  workspaceId: string,
  name: string,
  description: string,
) =>
  request<Project>(
    'POST',
    `/workspaces/${encodeURIComponent(workspaceId)}/projects`,
    { name, description: description.trim() === '' ? null : description },
  )

export const getProject = (id: string) =>
  request<Project>('GET', `/projects/${encodeURIComponent(id)}`)

// ---- GitHub connection ------------------------------------------------------

export async function getConnection(
  workspaceId: string,
): Promise<GitHubConnection | null> {
  try {
    return await request<GitHubConnection>(
      'GET',
      `/workspaces/${encodeURIComponent(workspaceId)}/github`,
    )
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      return null
    }
    throw error
  }
}

export const createConnection = (
  workspaceId: string,
  installationId: string,
  accountLogin: string,
) =>
  request<GitHubConnection>(
    'POST',
    `/workspaces/${encodeURIComponent(workspaceId)}/github`,
    { installation_id: installationId, account_login: accountLogin },
  )

export const setConnectionStatus = (
  workspaceId: string,
  id: string,
  status: string,
) =>
  request<void>(
    'PATCH',
    `/workspaces/${encodeURIComponent(workspaceId)}/github`,
    { id, status },
  )

// ---- repositories -----------------------------------------------------------

export const importRepositories = async (projectId: string): Promise<Job> =>
  normalizeJob(
    await request<Record<string, unknown>>(
      'POST',
      `/projects/${encodeURIComponent(projectId)}/repositories/import`,
    ),
  )

export const getImportJob = async (id: string): Promise<Job> =>
  normalizeJob(
    await request<Record<string, unknown>>(
      'GET',
      `/github/repository-import-jobs/${encodeURIComponent(id)}`,
    ),
  )

export const listRepositories = async (projectId: string) =>
  (await request<Repository[] | null>(
    'GET',
    `/projects/${encodeURIComponent(projectId)}/repositories`,
  )) ?? []

export const getRepository = (id: string) =>
  request<Repository>('GET', `/repositories/${encodeURIComponent(id)}`)

export const requestSync = async (repositoryId: string): Promise<Job> =>
  normalizeJob(
    await request<Record<string, unknown>>(
      'POST',
      `/repositories/${encodeURIComponent(repositoryId)}/sync`,
    ),
  )

export const getSyncJob = async (id: string): Promise<Job> =>
  normalizeJob(
    await request<Record<string, unknown>>(
      'GET',
      `/repositories/sync-jobs/${encodeURIComponent(id)}`,
    ),
  )

// ---- AI ---------------------------------------------------------------------

export const askQuestion = (repositoryId: string, question: string) =>
  request<Answer>(
    'POST',
    `/repositories/${encodeURIComponent(repositoryId)}/ask`,
    { question },
  )

export const explainRepository = (repositoryId: string) =>
  request<Explanation>(
    'POST',
    `/repositories/${encodeURIComponent(repositoryId)}/explain`,
  )

// ---- job polling ------------------------------------------------------------

export interface PollOptions {
  intervalMs?: number
  timeoutMs?: number
  sleep?: (ms: number) => Promise<void>
  now?: () => number
  onUpdate?: (job: Job) => void
  /** Consecutive network failures tolerated before giving up (default 5). */
  maxNetworkErrors?: number
}

/** Polls a job until it succeeds. Throws JobFailedError if it fails. */
export async function pollJob(
  fetchJob: () => Promise<Job>,
  options: PollOptions = {},
): Promise<Job> {
  const intervalMs = options.intervalMs ?? 1500
  const timeoutMs = options.timeoutMs ?? 120_000
  const sleep =
    options.sleep ?? ((ms: number) => new Promise<void>((r) => setTimeout(r, ms)))
  const now = options.now ?? Date.now
  const maxNetworkErrors = options.maxNetworkErrors ?? 5
  const startedAt = now()
  let networkErrors = 0

  for (;;) {
    let job: Job

    try {
      job = await fetchJob()
      networkErrors = 0
    } catch (error) {
      // A single dropped request should not fail a job that is still running.
      if (
        error instanceof NetworkError &&
        ++networkErrors <= maxNetworkErrors &&
        now() - startedAt <= timeoutMs
      ) {
        await sleep(intervalMs)
        continue
      }
      throw error
    }

    options.onUpdate?.(job)

    if (job.status === 'succeeded') {
      return job
    }
    if (job.status === 'failed') {
      throw new JobFailedError(job)
    }
    if (now() - startedAt > timeoutMs) {
      throw new Error('Timed out waiting for the job to finish.')
    }

    await sleep(intervalMs)
  }
}
