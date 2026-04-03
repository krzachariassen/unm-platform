export interface OrgInfo {
  id: string
  name: string
  slug: string
  role: 'owner' | 'admin' | 'member'
}

export interface WorkspaceInfo {
  id: string
  name: string
  slug: string
  visibility: 'private' | 'org-visible'
  role: 'admin' | 'editor' | 'viewer' | ''
}
