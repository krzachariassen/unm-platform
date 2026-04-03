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
    <div className={cn('flex gap-1 border-b border-border mb-4', className)}>
      {tabs.map((tab) => {
        const isActive = tab.id === current
        return (
          <button
            key={tab.id}
            onClick={() => activate(tab.id)}
            aria-label={tab.label}
            className={cn(
              'px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors',
              isActive
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
            )}
          >
            {tab.label}
          </button>
        )
      })}
    </div>
  )
}

/** Returns the active tab id from the URL, defaulting to the first tab. */
export function useActiveTab(tabs: UrlTabItem[], searchParam = 'tab'): string {
  const [searchParams] = useSearchParams()
  return searchParams.get(searchParam) ?? tabs[0]?.id ?? ''
}
