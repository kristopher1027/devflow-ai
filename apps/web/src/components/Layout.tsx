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
    <div className="app-shell">
      <aside className="sidebar">
        <a className="sidebar-brand" href={homeHref()}>
          <span className="brand-mark">D</span>
          <span>
            DevFlow<span className="brand-accent"> AI</span>
            <small>ENGINEERING INTELLIGENCE</small>
          </span>
        </a>

        <div className="sidebar-label">WORKSPACE</div>
        <nav className="sidebar-nav" aria-label="Main navigation">
          <a className="sidebar-link active" href={homeHref()}>
            <span className="nav-icon">◈</span>
            Overview
          </a>
        </nav>

        <div className="sidebar-bottom">
          <div className="sidebar-ai-card">
            <span className="ai-orb">✦</span>
            <strong>AI-powered development</strong>
            <p>Understand your codebase faster with DevFlow AI.</p>
          </div>
          <div className="sidebar-user">
            <div className="user-avatar">DF</div>
            <div className="user-details">
              <strong>DevFlow account</strong>
              <small>Workspace member</small>
            </div>
            <button
              type="button"
              className="logout-button"
              onClick={onLogout}
              aria-label="Log out"
              title="Log out"
            >
              ↗
            </button>
          </div>
        </div>
      </aside>

      <div className="app-main">
        <header className="app-topbar">
          <div className="breadcrumbs" aria-label="Breadcrumb">
            {crumbs.map((crumb, index) => (
              <Fragment key={`${crumb.label}-${index}`}>
                {index > 0 && <span className="breadcrumb-separator">/</span>}
                {crumb.href ? (
                  <a href={crumb.href}>{crumb.label}</a>
                ) : (
                  <span aria-current="page">{crumb.label}</span>
                )}
              </Fragment>
            ))}
          </div>
          <div className="topbar-right">
            <span className="system-status">
              <span className="status-dot" /> Platform workspace
            </span>
            <button type="button" className="topbar-logout" onClick={onLogout}>
              Log out
            </button>
          </div>
        </header>

        <main className="app-content">{children}</main>

        <footer className="app-footer">
          <span>DevFlow AI</span>
          <span>Build with clarity. Ship with confidence.</span>
        </footer>
      </div>
    </div>
  )
}
