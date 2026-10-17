export type Route =
  | { name: 'home' }
  | { name: 'workspace'; id: string }
  | { name: 'project'; id: string }
  | { name: 'repository'; id: string }

const kinds: Record<string, 'workspace' | 'project' | 'repository'> = {
  w: 'workspace',
  p: 'project',
  r: 'repository',
}

/** Turns "#/w/abc" into a route. Anything unrecognised is the home page. */
export function parseRoute(hash: string): Route {
  const parts = hash.replace(/^#\/?/, '').split('/').filter(Boolean)

  if (parts.length === 2) {
    const kind = kinds[parts[0]]
    if (kind) {
      try {
        return { name: kind, id: decodeURIComponent(parts[1]) }
      } catch {
        return { name: 'home' }
      }
    }
  }

  return { name: 'home' }
}

export const homeHref = () => '#/'
export const workspaceHref = (id: string) => `#/w/${encodeURIComponent(id)}`
export const projectHref = (id: string) => `#/p/${encodeURIComponent(id)}`
export const repositoryHref = (id: string) => `#/r/${encodeURIComponent(id)}`
