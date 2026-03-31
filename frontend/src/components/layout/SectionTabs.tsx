import { NavLink, useLocation } from 'react-router-dom'
import { cn } from '@/lib/utils'

export interface SectionTab {
  to: string
  label: string
}

export interface SectionTabsProps {
  tabs: SectionTab[]
}

/**
 * Horizontal section tabs rendered in the TopBar.
 * Uses React Router NavLink for active state detection.
 */
export function SectionTabs({ tabs }: SectionTabsProps) {
  const location = useLocation()

  return (
    <nav className="flex gap-1">
      {tabs.map(({ to, label }) => {
        const isActive = location.pathname === to
        return (
          <NavLink
            key={to}
            to={to}
            className={cn(
              'px-4 py-1.5 rounded-md text-sm font-medium transition-colors',
              isActive
                ? 'bg-primary/10 text-primary'
                : 'text-muted-foreground hover:text-foreground hover:bg-muted'
            )}
          >
            {label}
          </NavLink>
        )
      })}
    </nav>
  )
}
