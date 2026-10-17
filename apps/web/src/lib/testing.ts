// Helpers shared by the unit tests.
export interface FakeResponse {
  status?: number
  body?: unknown
  text?: string
}

/** Replaces fetch with a function that returns the given responses in order. */
export function queueFetch(vi: { stubGlobal: (name: string, value: unknown) => void; fn: () => any }, ...responses: FakeResponse[]) {
  const calls: { url: string; init?: RequestInit }[] = []
  const queue = [...responses]

  const fetchMock = vi.fn()
  fetchMock.mockImplementation(async (url: string, init?: RequestInit) => {
    calls.push({ url, init })
    const next = queue.shift() ?? { status: 200 }
    const text = next.text ?? (next.body === undefined ? '' : JSON.stringify(next.body))
    return new Response(text, { status: next.status ?? 200 })
  })

  vi.stubGlobal('fetch', fetchMock)
  return calls
}
