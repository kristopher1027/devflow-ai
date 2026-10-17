import { useCallback, useEffect, useState } from 'react'
import { ApiError, listWorkspaces, logout, setUnauthorizedHandler } from './lib/api'
import { friendlyError } from './lib/format'
import { useRoute } from './lib/hooks'
import AuthPage from './pages/AuthPage'
import ProjectPage from './pages/ProjectPage'
import RepositoryPage from './pages/RepositoryPage'
import WorkspacePage from './pages/WorkspacePage'
import WorkspacesPage from './pages/WorkspacesPage'

type AuthState = 'checking' | 'in' | 'out'

export default function App() {
  const route = useRoute()
  const [auth, setAuth] = useState<AuthState>('checking')
  const [startupError, setStartupError] = useState<string | null>(null)

  useEffect(() => {
    setUnauthorizedHandler(() => setAuth('out'))

    // There is no "who am I" endpoint, so a successful workspace list means
    // the session cookie is valid.
    listWorkspaces()
      .then(() => setAuth('in'))
      .catch((error: unknown) => {
        if (!(error instanceof ApiError && error.status === 401)) {
          setStartupError(friendlyError(error))
        }
        setAuth('out')
      })

    return () => setUnauthorizedHandler(null)
  }, [])

  const onLogout = useCallback(() => {
    logout()
      .catch(() => undefined)
      .finally(() => {
        window.location.hash = '#/'
        setAuth('out')
      })
  }, [])

  if (auth === 'checking') {
    return <p className="muted center">Loading…</p>
  }

  if (auth === 'out') {
    return (
      <AuthPage
        startupError={startupError}
        onAuthenticated={() => {
          setStartupError(null)
          setAuth('in')
        }}
      />
    )
  }

  switch (route.name) {
    case 'workspace':
      return <WorkspacePage id={route.id} onLogout={onLogout} />
    case 'project':
      return <ProjectPage id={route.id} onLogout={onLogout} />
    case 'repository':
      return <RepositoryPage id={route.id} onLogout={onLogout} />
    default:
      return <WorkspacesPage onLogout={onLogout} />
  }
}
