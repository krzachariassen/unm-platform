import { type ReactNode } from 'react'
import { Search, X, Pencil } from 'lucide-react'
import { useModel } from '@/lib/model-context'
import { useSearch } from '@/lib/search-context'
import { useChangeset } from '@/lib/changeset-context'
import { cn } from '@/lib/utils'

export interface TopBarProps {
  sectionTabs?: ReactNode
}

export function TopBar({ sectionTabs }: TopBarProps) {
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
    <header className="flex items-center h-14 px-5 gap-3 bg-white border-b border-border shrink-0">
      {/* Model status */}
      {parseResult ? (
        <div className="flex items-center gap-2">
          <span className="text-sm font-semibold text-foreground">{parseResult.system_name}</span>
          <span className={cn(
            'inline-flex items-center px-2 py-0.5 rounded text-xs font-medium',
            parseResult.validation.is_valid
              ? 'bg-green-100 text-green-700'
              : 'bg-red-100 text-red-700'
          )}>
            {parseResult.validation.is_valid ? 'Valid' : 'Invalid'}
          </span>
          {parseResult.validation.warnings.length > 0 && (
            <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-amber-100 text-amber-700">
              {parseResult.validation.warnings.length}{' '}
              {parseResult.validation.warnings.length === 1 ? 'warning' : 'warnings'}
            </span>
          )}
        </div>
      ) : (
        <span className="text-sm text-muted-foreground">No model loaded</span>
      )}

      {/* Section tabs slot */}
      {sectionTabs && <div className="flex-1 flex justify-center">{sectionTabs}</div>}

      {/* Right side controls */}
      <div className={cn('flex items-center gap-3', !sectionTabs && 'ml-auto')}>
        {/* Edit Model button */}
        {parseResult && (
          <button
            onClick={handleEditClick}
            className={cn(
              'flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors border',
              isEditMode
                ? 'bg-gray-900 text-white border-gray-900'
                : 'bg-gray-50 text-gray-700 border-gray-200 hover:bg-gray-100'
            )}
            title={isEditMode ? 'Exit edit mode' : 'Enter edit mode to make batch changes'}
          >
            <Pencil className="w-3 h-3" />
            {isEditMode
              ? actions.length > 0
                ? `${actions.length} change${actions.length === 1 ? '' : 's'}`
                : 'Edit mode on'
              : 'Edit Model'
            }
            {isEditMode && actions.length > 0 && (
              <span className="inline-flex items-center justify-center w-4 h-4 rounded-full text-[9px] font-bold bg-blue-500 text-white">
                {actions.length}
              </span>
            )}
          </button>
        )}

        {/* Search */}
        <div className="relative flex items-center">
          <Search className="absolute left-2.5 w-3.5 h-3.5 text-muted-foreground pointer-events-none" />
          <input
            type="text"
            value={query}
            onChange={e => setQuery(e.target.value)}
            placeholder="Search entities..."
            className="pl-8 pr-8 py-1.5 text-sm rounded-md w-52 bg-gray-50 border border-gray-200 text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
          />
          {query && (
            <button
              onClick={() => setQuery('')}
              className="absolute right-2.5 text-muted-foreground hover:text-foreground"
              aria-label="Clear search"
            >
              <X className="w-3 h-3" />
            </button>
          )}
        </div>
      </div>
    </header>
  )
}
