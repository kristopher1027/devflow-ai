// Field names match what the Go API sends: domain objects are encoded with
// Go's default (capitalised) names, while request/response structs use
// snake_case.

export interface Workspace {
  ID: string
  OwnerID: string
  Name: string
  CreatedAt: string
}

export interface Project {
  ID: string
  WorkspaceID: string
  Name: string
  Description: string | null
  CreatedAt: string
}

export interface Repository {
  ID: string
  ProjectID: string
  FullName: string
  Name: string
  Owner: string
  DefaultBranch: string
  HTMLURL: string
  IsPrivate: boolean
  SyncStatus: string
  LastSyncedAt: string | null
}

export interface GitHubConnection {
  ID: string
  WorkspaceID: string
  InstallationID: string
  AccountLogin: string
  Status: string
}

export type JobStatus = 'pending' | 'running' | 'succeeded' | 'failed'

export interface Job {
  id: string
  status: JobStatus | string
  failureCode: string | null
  failureMessage: string | null
}

export interface AnswerSource {
  file_path: string
  start_line: number
  end_line: number
}

export interface Answer {
  snapshot_id: string
  commit_sha: string
  answer: string
  sources: AnswerSource[]
}

export interface Explanation {
  snapshot_id: string
  commit_sha: string
  explanation: string
}
