import { useWorkspace } from '@/lib/workspace-context'
import { FolderOpen, Loader2 } from 'lucide-react'

export function WorkspaceSettingsPage() {
  const { workspace, org, workspaces, loading } = useWorkspace()

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (!workspace || !org) {
    return (
      <div className="max-w-xl mx-auto mt-16 text-center">
        <FolderOpen className="w-10 h-10 mx-auto text-muted-foreground mb-3" />
        <p className="text-sm text-muted-foreground">No workspace selected.</p>
      </div>
    )
  }

  return (
    <div className="max-w-2xl mx-auto">
      <div className="mb-8">
        <p className="text-xs text-muted-foreground uppercase tracking-wider font-medium mb-1">{org.name}</p>
        <h1 className="text-2xl font-bold text-foreground">Workspace Settings</h1>
        <p className="text-sm text-muted-foreground mt-1">Manage your workspace details and access.</p>
      </div>

      {/* Workspace details */}
      <section className="border border-border rounded-xl overflow-hidden mb-6">
        <div className="px-5 py-4 border-b border-border bg-muted/30">
          <h2 className="text-sm font-semibold text-foreground flex items-center gap-2">
            <FolderOpen className="w-4 h-4" />
            Workspace Details
          </h2>
        </div>
        <div className="px-5 py-4 space-y-4">
          <div>
            <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-1">Name</p>
            <p className="text-sm text-foreground font-medium">{workspace.name}</p>
          </div>
          <div>
            <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-1">Slug</p>
            <p className="text-sm text-foreground font-mono">{workspace.slug}</p>
          </div>
          <div>
            <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-1">Visibility</p>
            <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-muted text-muted-foreground">
              {workspace.visibility === 'org-visible' ? 'Visible to organisation' : 'Private'}
            </span>
          </div>
          {workspace.role && (
            <div>
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-1">Your Role</p>
              <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-muted text-muted-foreground capitalize">
                {workspace.role}
              </span>
            </div>
          )}
        </div>
      </section>

      {/* All workspaces in this org */}
      <section className="border border-border rounded-xl overflow-hidden">
        <div className="px-5 py-4 border-b border-border bg-muted/30">
          <h2 className="text-sm font-semibold text-foreground">
            All Workspaces in {org.name}
          </h2>
        </div>
        <div className="divide-y divide-border">
          {workspaces.map((ws) => (
            <div key={ws.id} className="flex items-center justify-between px-5 py-3">
              <div>
                <p className="text-sm font-medium text-foreground">
                  {ws.name}
                  {ws.slug === workspace.slug && (
                    <span className="ml-2 text-xs text-primary font-medium">(current)</span>
                  )}
                </p>
                <p className="text-xs text-muted-foreground font-mono">{ws.slug}</p>
              </div>
              <span className="text-xs text-muted-foreground">
                {ws.visibility === 'org-visible' ? 'Org-visible' : 'Private'}
              </span>
            </div>
          ))}
          {workspaces.length === 0 && (
            <div className="px-5 py-4 text-sm text-muted-foreground">No workspaces found.</div>
          )}
        </div>
      </section>
    </div>
  )
}
