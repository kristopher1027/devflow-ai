import ReactMarkdown from 'react-markdown'
import { statusTone } from '../lib/format'

export function ErrorText({ message }: { message: string | null }) {
  if (!message) return null
  return (
    <p className="error" role="alert">
      {message}
    </p>
  )
}

export function Notice({ message }: { message: string | null }) {
  if (!message) return null
  return (
    <p className="notice" role="status">
      {message}
    </p>
  )
}

export function Badge({ status }: { status: string }) {
  return <span className={`badge ${statusTone(status)}`}>{status}</span>
}

export function Loading({ label = 'Loading…' }: { label?: string }) {
  return <p className="muted">{label}</p>
}

/** Renders Markdown safely (raw HTML in the text is not interpreted). */
export function Markdown({ text }: { text: string }) {
  return (
    <div className="markdown">
      <ReactMarkdown>{text}</ReactMarkdown>
    </div>
  )
}
