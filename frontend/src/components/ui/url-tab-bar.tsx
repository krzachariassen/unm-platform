import { useSearchParams } from 'react-router-dom'
import { cn } from '@/lib/utils'

export interface UrlTabItem {
  id: string
  label: string
}

export interface UrlTabBarProps {
  tabs: UrlTabItem[]
  searchParam?: string
  className?: string
}

/**
 * A tab bar that syncs the active tab with a URL query param (?tab=...).
 * Defaults to the first tab when the param is absent.
 */
export function UrlTabBar({ tabs, searchParam = 'tab', className }: UrlTabBarProps) {
  const [searchParams, setSearchParams] = useSearchParams()
  const current = searchParams.get(searchParam) ?? tabs[0]?.id

  function activate(id: string) {
    setSearchParams(prev => {
      const next = new URLSearchParams(prev)
      next.set(searchParam, id)
      return next
    }, { replace: true })
  }

  return (
    <div className={cn('sticky top-0 z-10 bg-background border-b border-border', className)}>
      <div className="flex px-6">
        {tabs.map((tab) => {
          const isActive = tab.id === current
          return (
            <button
              key={tab.id}
              onClick={() => activate(tab.id)}
              aria-label={tab.label}
              className={cn(
                'relative px-4 py-3 text-sm transition-colors whitespace-nowrap',
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
    </div>
  )
}

/** Returns the active tab id from the URL, defaulting to the first tab. */
export function useActiveTab(tabs: UrlTabItem[], searchParam = 'tab'): string {
  const [searchParams] = useSearchParams()
  return searchParams.get(searchParam) ?? tabs[0]?.id ?? ''
}
