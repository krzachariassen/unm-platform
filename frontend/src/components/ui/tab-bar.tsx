import { cn } from '@/lib/utils'

export interface TabItem {
  id: string
  label: string
  count?: number
}

export interface TabBarProps {
  tabs: TabItem[]
  activeTab: string
  onChange: (id: string) => void
  className?: string
}

export function TabBar({ tabs, activeTab, onChange, className }: TabBarProps) {
  return (
    <div className={cn('flex gap-1 border-b border-border', className)}>
      {tabs.map((tab) => (
        <button
          key={tab.id}
          onClick={() => onChange(tab.id)}
          className={cn(
            'px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors',
            activeTab === tab.id
              ? 'border-primary text-primary'
              : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
          )}
        >
          {tab.label}
          {tab.count !== undefined && (
            <span className={cn(
              'ml-2 px-1.5 py-0.5 rounded-full text-xs',
              activeTab === tab.id ? 'bg-primary/10 text-primary' : 'bg-muted text-muted-foreground'
            )}>
              {tab.count}
            </span>
          )}
        </button>
      ))}
    </div>
  )
}
