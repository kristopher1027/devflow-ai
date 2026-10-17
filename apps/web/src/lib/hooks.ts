import { useEffect, useState, useSyncExternalStore } from 'react'
import { friendlyError } from './format'
import { parseRoute, type Route } from './route'

function subscribe(callback: () => void): () => void {
  window.addEventListener('hashchange', callback)
  return () => window.removeEventListener('hashchange', callback)
}

const currentHash = () => window.location.hash

/** The current page, read from the URL hash (e.g. #/w/<workspace id>). */
export function useRoute(): Route {
  return parseRoute(useSyncExternalStore(subscribe, currentHash, () => ''))
}

interface LoadState<T> {
  data: T | null
  error: string | null
  loading: boolean
}

/** Runs an async loader whenever `deps` change; call reload() to refetch. */
export function useLoad<T>(loader: () => Promise<T>, deps: unknown[]) {
  const [state, setState] = useState<LoadState<T>>({
    data: null,
    error: null,
    loading: true,
  })
  const [tick, setTick] = useState(0)

  useEffect(() => {
    let cancelled = false

    setState((previous) => ({ ...previous, loading: true, error: null }))

    loader()
      .then((data) => {
        if (!cancelled) setState({ data, error: null, loading: false })
      })
      .catch((error: unknown) => {
        if (!cancelled) {
          setState({ data: null, error: friendlyError(error), loading: false })
        }
      })

    return () => {
      cancelled = true
    }
  }, [...deps, tick])

  return { ...state, reload: () => setTick((value) => value + 1) }
}
