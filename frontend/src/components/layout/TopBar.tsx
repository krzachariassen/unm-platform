import { useState, useRef, useEffect } from 'react'
import { Search, X, Download, Loader2 } from 'lucide-react'
import { useModel } from '@/lib/model-context'
import { useSearch } from '@/lib/search-context'
import { useChangeset } from '@/lib/changeset-context'
import { ReviewDialog } from '@/components/changeset/ReviewDialog'
import { modelsApi } from '@/services/api'
import { cn } from '@/lib/utils'

export function TopBar() {
  const { modelId, parseResult } = useModel()
  const { query, setQuery } = useSearch()
  const { actions } = useChangeset()

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

  return (
    <header className="flex items-center h-14 px-5 gap-3 shrink-0" style={{ background: '#ffffff', borderBottom: '1px solid #e5e7eb' }}>
      {parseResult ? (
        <div className="flex items-center gap-2">
          <span className="text-sm font-semibold" style={{ color: '#111827' }}>{parseResult.system_name}</span>
          <span className={cn(
            'inline-flex items-center px-2 py-0.5 rounded text-[10px] font-semibold',
            parseResult.validation.is_valid ? 'text-green-700' : 'text-red-700'
          )} style={{ background: parseResult.validation.is_valid ? '#dcfce7' : '#fef2f2' }}>
            {parseResult.validation.is_valid ? 'Valid' : 'Invalid'}
          </span>
          {parseResult.validation.warnings.length > 0 && (
            <span className="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-semibold" style={{ background: '#fef3c7', color: '#92400e' }}>
              {parseResult.validation.warnings.length} {parseResult.validation.warnings.length === 1 ? 'warning' : 'warnings'}
            </span>
          )}
        </div>
      ) : (
        <span className="text-sm" style={{ color: '#9ca3af' }}>No model loaded</span>
      )}

      <div className="flex items-center gap-3 ml-auto">
        {actions.length > 0 && (
          <button
            type="button"
            onClick={() => setDialogOpen(true)}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors"
            style={{ background: '#111827', color: '#ffffff', border: '1px solid #111827' }}
          >
            <span className="inline-flex items-center justify-center w-[18px] h-[18px] rounded-full text-[10px] font-bold" style={{ background: 'rgba(255,255,255,0.2)' }}>
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
              className="flex items-center justify-center w-8 h-8 rounded-md hover:bg-gray-100 transition-colors"
              style={{ color: '#6b7280' }}
            >
              {exportingFormat ? <Loader2 className="w-4 h-4 animate-spin" /> : <Download className="w-4 h-4" />}
            </button>
            {exportOpen && (
              <div className="absolute right-0 top-full mt-1 w-44 rounded-lg shadow-lg py-1 z-50" style={{ background: '#ffffff', border: '1px solid #e5e7eb' }}>
                <button type="button" onClick={() => handleExport('yaml')}
                  className="w-full text-left px-3 py-2 text-sm hover:bg-gray-50 transition-colors"
                  style={{ color: '#111827' }}>
                  Download .unm.yaml
                </button>
                <button type="button" onClick={() => handleExport('dsl')}
                  className="w-full text-left px-3 py-2 text-sm hover:bg-gray-50 transition-colors"
                  style={{ color: '#111827' }}>
                  Download .unm
                </button>
              </div>
            )}
          </div>
        )}

        <div className="relative flex items-center">
          <Search className="absolute left-2.5 w-3.5 h-3.5 pointer-events-none" style={{ color: '#9ca3af' }} />
          <input
            type="text"
            value={query}
            onChange={e => setQuery(e.target.value)}
            placeholder="Search entities..."
            className="pl-8 pr-8 py-1.5 text-sm rounded-md w-52 focus:outline-none focus:ring-2"
            style={{ background: '#f9fafb', border: '1px solid #e5e7eb', color: '#111827' }}
          />
          {query && (
            <button onClick={() => setQuery('')} className="absolute right-2.5 hover:text-foreground" style={{ color: '#9ca3af' }} aria-label="Clear search">
              <X className="w-3 h-3" />
            </button>
          )}
        </div>
      </div>

      <ReviewDialog open={dialogOpen} onClose={() => setDialogOpen(false)} />
    </header>
  )
}
