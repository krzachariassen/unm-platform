import { Search, X } from 'lucide-react'
import { useModel } from '@/lib/model-context'
import { useSearch } from '@/lib/search-context'

export function TopBar() {
  const { parseResult } = useModel()
  const { query, setQuery } = useSearch()

  return (
    <header className="flex items-center px-5 gap-3" style={{ height: 56, background: '#ffffff', borderBottom: '1px solid #e5e7eb' }}>
      {parseResult ? (
        <div className="flex items-center gap-2.5">
          <span className="text-sm font-semibold" style={{ color: '#111827' }}>{parseResult.system_name}</span>
          <span
            className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium"
            style={parseResult.validation.is_valid
              ? { background: '#dcfce7', color: '#15803d' }
              : { background: '#fee2e2', color: '#b91c1c' }
            }
          >
            {parseResult.validation.is_valid ? 'Valid' : 'Invalid'}
          </span>
          {parseResult.validation.warnings.length > 0 && (
            <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium" style={{ background: '#fef3c7', color: '#92400e' }}>
              {parseResult.validation.warnings.length} {parseResult.validation.warnings.length === 1 ? 'warning' : 'warnings'}
            </span>
          )}
        </div>
      ) : (
        <span className="text-sm" style={{ color: '#9ca3af' }}>No model loaded</span>
      )}

      <div className="ml-auto flex items-center gap-2 relative">
        <Search size={13} className="absolute left-2.5 pointer-events-none" style={{ color: '#9ca3af' }} />
        <input
          type="text"
          value={query}
          onChange={e => setQuery(e.target.value)}
          placeholder="Search entities..."
          className="pl-8 pr-8 py-1.5 text-sm rounded-md w-52 focus:outline-none"
          style={{
            background: '#f9fafb',
            border: '1px solid #e5e7eb',
            color: '#111827',
          }}
        />
        {query && (
          <button onClick={() => setQuery('')} className="absolute right-2.5" style={{ color: '#9ca3af' }}>
            <X size={12} />
          </button>
        )}
      </div>
    </header>
  )
}
