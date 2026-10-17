import { Fragment, type ReactNode } from 'react'
import { homeHref } from '../lib/route'

export interface Crumb {
  label: string
  href?: string
}

interface LayoutProps {
  crumbs: Crumb[]
  onLogout: () => void
  children: ReactNode
}

export default function Layout({ crumbs, onLogout, children }: LayoutProps) {
  return (
    <div className="shell">
      <header className="topbar">
        <a className="brand" href={homeHref()}>
          DevFlow <span>AI</span>
        </a>
        <nav className="crumbs" aria-label="Breadcrumb">
          {crumbs.map((crumb, index) => (
            <Fragment key={`${crumb.label}-${index}`}>
              <span className="sep">/</span>
              {crumb.href ? (
                <a href={crumb.href}>{crumb.label}</a>
              ) : (
                <span aria-current="page">{crumb.label}</span>
              )}
            </Fragment>
          ))}
        </nav>
        <button type="button" className="ghost" onClick={onLogout}>
          Log out
        </button>
      </header>
      <main className="content">{children}</main>
      <footer className="footer">DevFlow AI — ask questions about any repository</footer>
    </div>
  )
}
