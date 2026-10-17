import { afterEach, describe, expect, it, vi } from 'vitest'
import {
  ApiError,
  JobFailedError,
  NetworkError,
  askQuestion,
  createProject,
  getConnection,
  getSyncJob,
  importRepositories,
  listWorkspaces,
  login,
  pollJob,
  setUnauthorizedHandler,
} from './api'
import { queueFetch } from './testing'
import type { Job } from './types'

afterEach(() => {
  vi.unstubAllGlobals()
  setUnauthorizedHandler(null)
})

describe('request handling', () => {
  it('sends JSON to the /api prefix with the session cookie', async () => {
    const calls = queueFetch(vi, { status: 200 })

    await login('a@b.co', 'secret')

    expect(calls[0].url).toBe('/api/auth/login')
    expect(calls[0].init?.method).toBe('POST')
    expect(calls[0].init?.credentials).toBe('same-origin')
    expect(JSON.parse(String(calls[0].init?.body))).toEqual({
      email: 'a@b.co',
      password: 'secret',
    })
  })

  it('turns plain-text error bodies into ApiError messages', async () => {
    queueFetch(vi, { status: 409, text: 'repository has not been synced yet\n' })

    const failure = askQuestion('r1', 'q').catch((e: unknown) => e)

    expect(await failure).toMatchObject({
      name: 'ApiError',
      status: 409,
      message: 'repository has not been synced yet',
    })
  })

  it('reports an expired session on authenticated requests only', async () => {
    const onUnauthorized = vi.fn()
    setUnauthorizedHandler(onUnauthorized)

    queueFetch(vi, { status: 401, text: 'unauthorized' })
    await listWorkspaces().catch(() => undefined)
    expect(onUnauthorized).toHaveBeenCalledTimes(1)

    queueFetch(vi, { status: 401, text: 'invalid credentials' })
    await login('a@b.co', 'wrong').catch(() => undefined)
    expect(onUnauthorized).toHaveBeenCalledTimes(1)
  })

  it('treats a null list from Go as an empty list', async () => {
    queueFetch(vi, { text: 'null' })

    expect(await listWorkspaces()).toEqual([])
  })

  it('returns null when a workspace has no GitHub connection', async () => {
    queueFetch(vi, { status: 404, text: 'not found' })

    expect(await getConnection('w1')).toBeNull()
  })

  it('does not hide other connection errors', async () => {
    queueFetch(vi, { status: 500, text: 'internal server error' })

    await expect(getConnection('w1')).rejects.toBeInstanceOf(ApiError)
  })

  it('sends a null description when it is blank', async () => {
    const calls = queueFetch(vi, { body: {} })

    await createProject('w1', 'Demo', '   ')

    expect(JSON.parse(String(calls[0].init?.body))).toEqual({
      name: 'Demo',
      description: null,
    })
  })

  it('normalises job payloads from both the import and sync endpoints', async () => {
    queueFetch(
      vi,
      { body: { job_id: 'j1', status: 'queued' } },
      { body: { ID: 's1', Status: 'failed', FailureMessage: 'boom' } },
    )

    expect(await importRepositories('p1')).toMatchObject({ id: 'j1', status: 'queued' })
    expect(await getSyncJob('s1')).toMatchObject({
      id: 's1',
      status: 'failed',
      failureMessage: 'boom',
    })
  })
})

describe('network failures', () => {
  it('reports a request that never got a response as a NetworkError', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')))

    await expect(listWorkspaces()).rejects.toBeInstanceOf(NetworkError)
  })
})

describe('pollJob', () => {
  const job = (status: string, failureMessage: string | null = null): Job => ({
    id: 'j1',
    status,
    failureCode: null,
    failureMessage,
  })

  it('polls until the job succeeds', async () => {
    const statuses = ['pending', 'running', 'succeeded']
    const fetchJob = vi.fn(async () => job(statuses.shift()!))
    const updates: string[] = []

    const result = await pollJob(fetchJob, {
      sleep: async () => undefined,
      onUpdate: (j) => updates.push(j.status),
    })

    expect(result.status).toBe('succeeded')
    expect(fetchJob).toHaveBeenCalledTimes(3)
    expect(updates).toEqual(['pending', 'running', 'succeeded'])
  })

  it('throws JobFailedError with the failure message', async () => {
    const fetchJob = async () => job('failed', 'github app ID is required')

    const failure = await pollJob(fetchJob, { sleep: async () => undefined }).catch(
      (e: unknown) => e,
    )

    expect(failure).toBeInstanceOf(JobFailedError)
    expect((failure as Error).message).toBe('github app ID is required')
  })

  it('gives up after the timeout', async () => {
    let clock = 0
    const fetchJob = async () => job('running')

    await expect(
      pollJob(fetchJob, {
        timeoutMs: 1000,
        intervalMs: 400,
        now: () => clock,
        sleep: async (ms) => {
          clock += ms
        },
      }),
    ).rejects.toThrow('Timed out')
  })

  it('survives a few dropped requests while the job keeps running', async () => {
    const steps: Array<Job | Error> = [
      job('running'),
      new NetworkError(),
      new NetworkError(),
      job('succeeded'),
    ]
    const fetchJob = vi.fn(async () => {
      const next = steps.shift()!
      if (next instanceof Error) throw next
      return next
    })

    const result = await pollJob(fetchJob, { sleep: async () => undefined })

    expect(result.status).toBe('succeeded')
    expect(fetchJob).toHaveBeenCalledTimes(4)
  })

  it('gives up after too many consecutive network failures', async () => {
    const fetchJob = vi.fn(async (): Promise<Job> => {
      throw new NetworkError()
    })

    await expect(
      pollJob(fetchJob, { sleep: async () => undefined, maxNetworkErrors: 3 }),
    ).rejects.toBeInstanceOf(NetworkError)
    expect(fetchJob).toHaveBeenCalledTimes(4)
  })

  it('does not retry real API errors', async () => {
    const fetchJob = vi.fn(async (): Promise<Job> => {
      throw new ApiError(404, 'not found')
    })

    await expect(pollJob(fetchJob, { sleep: async () => undefined })).rejects.toBeInstanceOf(
      ApiError,
    )
    expect(fetchJob).toHaveBeenCalledTimes(1)
  })
})
