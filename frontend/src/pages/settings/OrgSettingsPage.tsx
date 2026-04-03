import { useWorkspace } from '@/lib/workspace-context'
import { Building2, Users, Loader2 } from 'lucide-react'

export function OrgSettingsPage() {
  const { org, orgs, loading } = useWorkspace()

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (!org) {
    return (
      <div className="max-w-xl mx-auto mt-16 text-center">
        <Building2 className="w-10 h-10 mx-auto text-muted-foreground mb-3" />
        <p className="text-sm text-muted-foreground">No organisation selected.</p>
      </div>
    )
  }

  return (
    <div className="max-w-2xl mx-auto">
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-foreground">Organisation Settings</h1>
        <p className="text-sm text-muted-foreground mt-1">Manage your organisation details.</p>
      </div>

      {/* Org details */}
      <section className="border border-border rounded-xl overflow-hidden mb-6">
        <div className="px-5 py-4 border-b border-border bg-muted/30">
          <h2 className="text-sm font-semibold text-foreground flex items-center gap-2">
            <Building2 className="w-4 h-4" />
            Organisation Details
          </h2>
        </div>
        <div className="px-5 py-4 space-y-4">
          <div>
            <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-1">Name</p>
            <p className="text-sm text-foreground font-medium">{org.name}</p>
          </div>
          <div>
            <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-1">Slug</p>
            <p className="text-sm text-foreground font-mono">{org.slug}</p>
          </div>
          <div>
            <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-1">Your Role</p>
            <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-muted text-muted-foreground capitalize">
              {org.role}
            </span>
          </div>
        </div>
      </section>

      {/* Members (read-only) */}
      <section className="border border-border rounded-xl overflow-hidden">
        <div className="px-5 py-4 border-b border-border bg-muted/30">
          <h2 className="text-sm font-semibold text-foreground flex items-center gap-2">
            <Users className="w-4 h-4" />
            Your Organisations
          </h2>
        </div>
        <div className="divide-y divide-border">
          {orgs.map((o) => (
            <div key={o.id} className="flex items-center justify-between px-5 py-3">
              <div>
                <p className="text-sm font-medium text-foreground">{o.name}</p>
                <p className="text-xs text-muted-foreground font-mono">{o.slug}</p>
              </div>
              <span className="text-xs text-muted-foreground capitalize">{o.role}</span>
            </div>
          ))}
          {orgs.length === 0 && (
            <div className="px-5 py-4 text-sm text-muted-foreground">No organisations found.</div>
          )}
        </div>
      </section>
    </div>
  )
}
