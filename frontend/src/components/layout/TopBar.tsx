import { useState } from 'react'
import { Search, X } from 'lucide-react'
import { useModel } from '@/lib/model-context'
import { useSearch } from '@/lib/search-context'
import { useChangeset } from '@/lib/changeset-context'
import { ReviewDialog } from '@/components/changeset/ReviewDialog'
import { cn } from '@/lib/utils'

export function TopBar() {
  const { parseResult } = useModel()
  const { query, setQuery } = useSearch()
  const { actions } = useChangeset()

  const [dialogOpen, setDialogOpen] = useState(false)

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
