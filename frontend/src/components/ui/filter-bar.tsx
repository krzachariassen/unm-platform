import { Search } from 'lucide-react'
import { cn } from '@/lib/utils'
import { type ReactNode } from 'react'

export interface FilterBarProps {
  query: string
  onQueryChange: (q: string) => void
  placeholder?: string
  filters?: ReactNode
  className?: string
}

export function FilterBar({ query, onQueryChange, placeholder = 'Search…', filters, className }: FilterBarProps) {
  return (
    <div className={cn('flex items-center gap-3', className)}>
      <div className="relative flex-1 max-w-sm">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground pointer-events-none" />
        <input
          type="text"
          value={query}
          onChange={(e) => onQueryChange(e.target.value)}
          placeholder={placeholder}
          className="w-full pl-9 pr-3 py-2 text-sm border border-border rounded-lg bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
        />
      </div>
      {filters}
    </div>
  )
}
