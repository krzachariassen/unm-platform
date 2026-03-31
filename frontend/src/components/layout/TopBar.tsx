import { Search, X, Pencil } from 'lucide-react'
import { useModel } from '@/lib/model-context'
import { useSearch } from '@/lib/search-context'
import { useChangeset } from '@/lib/changeset-context'

export function TopBar() {
  const { parseResult } = useModel()
  const { query, setQuery } = useSearch()
  const { isEditMode, actions, enterEditMode, exitEditMode } = useChangeset()

  const handleEditClick = () => {
    if (!parseResult) return
    if (!isEditMode) {
      enterEditMode()
      return
    }
    if (actions.length > 0) {
      if (!window.confirm(`Discard ${actions.length} pending change${actions.length === 1 ? '' : 's'} and exit edit mode?`)) return
    }
    exitEditMode()
  }

  return (
    <header
      className="flex items-center px-5 gap-3"
      style={{ height: 56, background: '#ffffff', borderBottom: '1px solid #e5e7eb', flexShrink: 0 }}
    >
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
            <span
              className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium"
              style={{ background: '#fef3c7', color: '#92400e' }}
            >
              {parseResult.validation.warnings.length} {parseResult.validation.warnings.length === 1 ? 'warning' : 'warnings'}
            </span>
          )}
        </div>
      ) : (
        <span className="text-sm" style={{ color: '#9ca3af' }}>No model loaded</span>
      )}

      <div className="ml-auto flex items-center gap-3">
        {/* Edit Model button */}
        {parseResult && (
          <button
            onClick={handleEditClick}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-all"
            style={isEditMode
              ? {
                  background: '#111827',
                  color: '#ffffff',
                  border: '1px solid #111827',
                }
              : {
                  background: '#f9fafb',
                  color: '#374151',
                  border: '1px solid #e5e7eb',
                }
            }
            title={isEditMode ? 'Exit edit mode' : 'Enter edit mode to make batch changes'}
          >
            <Pencil size={12} />
            {isEditMode
              ? actions.length > 0
                ? `${actions.length} change${actions.length === 1 ? '' : 's'}`
                : 'Editing…'
              : 'Edit Model'
            }
            {isEditMode && actions.length > 0 && (
              <span
                className="inline-flex items-center justify-center w-4 h-4 rounded-full text-[9px] font-bold"
                style={{ background: '#3b82f6', color: '#ffffff' }}
              >
                {actions.length}
              </span>
            )}
          </button>
        )}

        {/* Search */}
        <div className="relative flex items-center">
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
            <button
              onClick={() => setQuery('')}
              className="absolute right-2.5"
              style={{ color: '#9ca3af' }}
            >
              <X size={12} />
            </button>
          )}
        </div>
      </div>
    </header>
  )
}
