import { useState, useRef, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Search, X, Download, Loader2, ChevronRight } from 'lucide-react'
import { useModel } from '@/lib/model-context'
import { useSearch } from '@/lib/search-context'
import { useChangeset } from '@/lib/changeset-context'
import { usePageTabs } from '@/lib/page-tabs-context'
import { useWorkspace } from '@/lib/workspace-context'
import { ReviewDialog } from '@/components/changeset/ReviewDialog'
import { modelsApi } from '@/services/api'
import { cn } from '@/lib/utils'

export function TopBar() {
  const { modelId, parseResult } = useModel()
  const { query, setQuery } = useSearch()
  const { actions } = useChangeset()
  const { tabs } = usePageTabs()
  const { org, workspace } = useWorkspace()

  const [searchParams, setSearchParams] = useSearchParams()
  const [dialogOpen, setDialogOpen] = useState(false)
  const [exportOpen, setExportOpen] = useState(false)
  const [exportingFormat, setExportingFormat] = useState<'yaml' | 'dsl' | null>(null)
  const exportRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!exportOpen) return
    const handler = (e: MouseEvent) => {
      if (exportRef.current && !exportRef.current.contains(e.target as Node)) setExportOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [exportOpen])

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

  const currentTab = searchParams.get('tab') ?? tabs[0]?.id
  function activateTab(id: string) {
    setSearchParams(prev => {
      const next = new URLSearchParams(prev)
      next.set('tab', id)
      return next
    }, { replace: true })
  }

  return (
    <header className="shrink-0 bg-background border-b border-border">
      {/* Row 1 — breadcrumb + model identity + actions */}
      <div className="flex items-center h-14 px-5 gap-3">
        {/* Breadcrumb: OrgName › WorkspaceName */}
        {org && workspace && (
          <div className="flex items-center gap-1.5 text-sm text-muted-foreground min-w-0 mr-1">
            <span className="truncate max-w-[120px]">{org.name}</span>
            <ChevronRight className="w-3.5 h-3.5 shrink-0" />
            <span className="truncate max-w-[120px] font-medium text-foreground">{workspace.name}</span>
            {parseResult && <ChevronRight className="w-3.5 h-3.5 shrink-0" />}
          </div>
        )}

        {parseResult ? (
          <div className="flex items-center gap-2.5 min-w-0">
            <span className="text-base font-bold text-foreground truncate">{parseResult.system_name}</span>
            <span className={cn(
              'inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-semibold shrink-0',
              parseResult.validation.is_valid
                ? 'bg-green-100 text-green-700'
                : 'bg-red-100 text-red-700'
            )}>
              {parseResult.validation.is_valid ? 'Valid' : 'Invalid'}
            </span>
            {parseResult.validation.warnings.length > 0 && (
              <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-semibold bg-amber-100 text-amber-700 shrink-0">
                {parseResult.validation.warnings.length}{' '}
                {parseResult.validation.warnings.length === 1 ? 'warning' : 'warnings'}
              </span>
            )}
          </div>
        ) : (
          !org && !workspace && (
            <span className="text-sm text-muted-foreground">No model loaded</span>
          )
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
        </div>

        <ReviewDialog open={dialogOpen} onClose={() => setDialogOpen(false)} />
      </div>

      {/* Row 2 — page tabs (only when a page registers them) */}
      {tabs.length > 0 && (
        <div className="flex px-5 border-t border-border/40">
          {tabs.map(tab => {
            const isActive = tab.id === currentTab
            return (
              <button
                key={tab.id}
                onClick={() => activateTab(tab.id)}
                className={cn(
                  'relative px-4 py-2.5 text-sm whitespace-nowrap transition-colors',
                  'after:absolute after:bottom-0 after:left-0 after:right-0 after:h-0.5 after:rounded-full after:transition-colors',
                  isActive
                    ? 'font-semibold text-foreground after:bg-primary'
                    : 'font-medium text-muted-foreground hover:text-foreground after:bg-transparent'
                )}
              >
                {tab.label}
              </button>
            )
          })}
        </div>
      )}
    </header>
  )
}
