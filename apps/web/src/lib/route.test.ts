import { describe, expect, it } from 'vitest'
import { parseRoute, projectHref, repositoryHref, workspaceHref } from './route'

describe('parseRoute', () => {
  it('parses each page type', () => {
    expect(parseRoute('#/w/abc')).toEqual({ name: 'workspace', id: 'abc' })
    expect(parseRoute('#/p/abc')).toEqual({ name: 'project', id: 'abc' })
    expect(parseRoute('#/r/abc')).toEqual({ name: 'repository', id: 'abc' })
  })

  it('falls back to home for anything unrecognised', () => {
    for (const hash of ['', '#', '#/', '#/x/abc', '#/w', '#/w/a/b', '#/w/%E0%A4%A']) {
      expect(parseRoute(hash)).toEqual({ name: 'home' })
    }
  })

  it('round-trips ids through the link helpers', () => {
    const id = 'a b/c'

    expect(parseRoute(workspaceHref(id))).toEqual({ name: 'workspace', id })
    expect(parseRoute(projectHref(id))).toEqual({ name: 'project', id })
    expect(parseRoute(repositoryHref(id))).toEqual({ name: 'repository', id })
  })
})
