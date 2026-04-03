import { useState, useRef, useEffect } from 'react'
import { Search, X, Download, Loader2, CheckCircle2, AlertTriangle, LogOut } from 'lucide-react'
import { useModel } from '@/lib/model-context'
import { useSearch } from '@/lib/search-context'
import { useChangeset } from '@/lib/changeset-context'
import { useAuth } from '@/lib/auth-context'
import { ReviewDialog } from '@/components/changeset/ReviewDialog'
import { modelsApi } from '@/services/api'
import { cn } from '@/lib/utils'

export function TopBar() {
  const { modelId, parseResult } = useModel()
  const { query, setQuery } = useSearch()
  const { actions } = useChangeset()
  const { user, logout } = useAuth()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [exportOpen, setExportOpen] = useState(false)
  const [exportingFormat, setExportingFormat] = useState<'yaml' | 'dsl' | null>(null)
  const [userMenuOpen, setUserMenuOpen] = useState(false)
  const exportRef = useRef<HTMLDivElement>(null)
  const userMenuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!exportOpen) return
    const handler = (e: MouseEvent) => {
      if (exportRef.current && !exportRef.current.contains(e.target as Node)) setExportOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [exportOpen])

  useEffect(() => {
    if (!userMenuOpen) return
    const handler = (e: MouseEvent) => {
      if (userMenuRef.current && !userMenuRef.current.contains(e.target as Node)) setUserMenuOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [userMenuOpen])

  const handleExport = async (format: 'yaml' | 'dsl') => {
    if (!modelId) return
    setExportOpen(false)
    setExportingFormat(format)
    try {
      const content = await modelsApi.exportModel(modelId, format)
      const ext = format === 'dsl' ? '.unm' : '.unm.yaml'
      const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${modelId}${ext}`
      a.click()
      URL.revokeObjectURL(url)
    } finally {
      setExportingFormat(null)
    }
  }

  const summary = parseResult?.summary as Record<string, number> | undefined
  const hasErrors = parseResult ? parseResult.validation.errors.length > 0 : false
  const warningCount = parseResult?.validation.warnings.length ?? 0

  return (
    <header className="shrink-0 bg-background border-b border-border">
      <div className="flex items-center h-14 px-5 gap-3">
        {parseResult ? (
          <div className="flex items-center gap-3 min-w-0">
            <span className="text-base font-bold text-foreground truncate">{parseResult.system_name}</span>
            <span className="text-xs text-muted-foreground hidden sm:inline">
              {summary?.capabilities ?? 0} capabilities · {summary?.teams ?? 0} teams · {summary?.services ?? 0} services
            </span>
            <div className={cn(
              'flex items-center gap-1 px-2 py-0.5 rounded-md text-[11px] font-medium shrink-0',
              hasErrors ? 'bg-red-50 text-red-700' : 'bg-green-50 text-green-700'
            )}>
              {hasErrors
                ? <><AlertTriangle className="w-3 h-3" /> Invalid</>
                : <><CheckCircle2 className="w-3 h-3" /> Valid</>
              }
              {warningCount > 0 && (
                <span className="text-amber-600 ml-1">· {warningCount} warning{warningCount !== 1 ? 's' : ''}</span>
              )}
            </div>
          </div>
        ) : (
          <span className="text-sm text-muted-foreground">No model loaded</span>
        )}

        <div className="flex items-center gap-3 ml-auto">
          {actions.length > 0 && (
            <button
              type="button"
              onClick={() => setDialogOpen(true)}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-foreground text-background transition-colors hover:bg-foreground/90"
            >
              <span className="inline-flex items-center justify-center w-[18px] h-[18px] rounded-full text-[10px] font-bold bg-white/20">
                {actions.length}
              </span>
              {actions.length === 1 ? 'change' : 'changes'}
            </button>
          )}

          {modelId && (
            <div className="relative" ref={exportRef}>
              <button
                type="button"
                onClick={() => setExportOpen(o => !o)}
                title="Export model"
                className="flex items-center justify-center w-8 h-8 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
              >
                {exportingFormat ? <Loader2 className="w-4 h-4 animate-spin" /> : <Download className="w-4 h-4" />}
              </button>
              {exportOpen && (
                <div className="absolute right-0 top-full mt-1 w-44 rounded-lg shadow-lg border border-border bg-background py-1 z-50">
                  <button type="button" onClick={() => handleExport('yaml')}
                    className="w-full text-left px-3 py-2 text-sm text-foreground hover:bg-muted transition-colors">
                    Download .unm.yaml
                  </button>
                  <button type="button" onClick={() => handleExport('dsl')}
                    className="w-full text-left px-3 py-2 text-sm text-foreground hover:bg-muted transition-colors">
                    Download .unm
                  </button>
                </div>
              )}
            </div>
          )}

          <div className="relative flex items-center">
            <Search className="absolute left-2.5 w-3.5 h-3.5 text-muted-foreground pointer-events-none" />
            <input
              type="text"
              value={query}
              onChange={e => setQuery(e.target.value)}
              placeholder="Search entities..."
              className="pl-8 pr-8 py-1.5 text-sm rounded-md w-52 bg-muted/50 border border-border text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/30"
            />
            {query && (
              <button onClick={() => setQuery('')} className="absolute right-2.5 text-muted-foreground hover:text-foreground" aria-label="Clear search">
                <X className="w-3 h-3" />
              </button>
            )}
          </div>

          {user && (
            <div className="relative" ref={userMenuRef}>
              <button
                type="button"
                onClick={() => setUserMenuOpen(o => !o)}
                title={user.name}
                className="flex items-center justify-center w-8 h-8 rounded-full overflow-hidden border border-border hover:ring-2 hover:ring-primary/30 transition-all"
              >
                {user.avatar_url ? (
                  <img src={user.avatar_url} alt={user.name} className="w-full h-full object-cover" />
                ) : (
                  <span className="w-full h-full flex items-center justify-center bg-muted text-xs font-semibold text-muted-foreground">
                    {user.name.charAt(0).toUpperCase()}
                  </span>
                )}
              </button>
              {userMenuOpen && (
                <div className="absolute right-0 top-full mt-1 w-52 rounded-lg shadow-lg border border-border bg-background py-1 z-50">
                  <div className="px-3 py-2 border-b border-border">
                    <p className="text-sm font-medium text-foreground truncate">{user.name}</p>
                    <p className="text-xs text-muted-foreground truncate">{user.email}</p>
                  </div>
                  <button
                    type="button"
                    onClick={() => { setUserMenuOpen(false); logout() }}
                    className="w-full flex items-center gap-2 px-3 py-2 text-sm text-foreground hover:bg-muted transition-colors"
                  >
                    <LogOut className="w-3.5 h-3.5" />
                    Sign out
                  </button>
                </div>
              )}
            </div>
          )}
        </div>

        <ReviewDialog open={dialogOpen} onClose={() => setDialogOpen(false)} />
      </div>
    </header>
  )
}
