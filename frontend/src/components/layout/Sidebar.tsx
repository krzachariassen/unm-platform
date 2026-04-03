import { useState } from 'react'
import { NavLink, useLocation } from 'react-router-dom'
import {
  LayoutDashboard, Map, Users, Layers, AlertCircle, Network,
  FlaskConical, Bot, FileText, ChevronLeft, ChevronRight, ChevronDown,
  Database, History, Settings, ArrowLeftRight,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useModel } from '@/lib/model-context'
import { useAIEnabled } from '@/hooks/useAIEnabled'
import { useWorkspace } from '@/lib/workspace-context'
import type { LucideIcon } from 'lucide-react'

interface SubItem {
  to: string
  label: string
  end?: boolean
}

interface NavItem {
  to: string
  label: string
  icon: LucideIcon
  always: boolean
  ai: boolean
  children?: SubItem[]
}

interface NavSection {
  label: string
  items: NavItem[]
}

const NAV_SECTIONS: NavSection[] = [
  {
    label: 'Workspace',
    items: [
      { to: '/workspace', label: 'Dashboard', icon: LayoutDashboard, always: true, ai: false },
      { to: '/models', label: 'All Models', icon: Database, always: true, ai: false },
      { to: '/dashboard', label: 'Model View', icon: LayoutDashboard, always: false, ai: false },
    ],
  },
  {
    label: 'Architecture',
    items: [
      { to: '/unm-map', label: 'UNM Map', icon: Map, always: false, ai: false },
      {
        to: '/needs', label: 'Needs', icon: Users, always: false, ai: false,
        children: [
          { to: '/needs', label: 'Overview', end: true },
          { to: '/needs/traceability', label: 'Traceability' },
          { to: '/needs/gaps', label: 'Gaps' },
        ],
      },
      {
        to: '/capabilities', label: 'Capabilities', icon: Layers, always: false, ai: false,
        children: [
          { to: '/capabilities', label: 'Hierarchy', end: true },
          { to: '/capabilities/services', label: 'Services' },
          { to: '/capabilities/dependencies', label: 'Dependencies' },
        ],
      },
      {
        to: '/teams', label: 'Teams', icon: Network, always: false, ai: false,
        children: [
          { to: '/teams', label: 'Topology', end: true },
          { to: '/teams/ownership', label: 'Ownership' },
          { to: '/teams/cognitive-load', label: 'Cognitive Load' },
          { to: '/teams/interactions', label: 'Interactions' },
        ],
      },
      { to: '/signals', label: 'Signals', icon: AlertCircle, always: false, ai: false },
    ],
  },
  {
    label: 'Editing',
    items: [
      { to: '/what-if', label: 'What-If', icon: FlaskConical, always: false, ai: false },
      { to: '/history', label: 'History', icon: History, always: false, ai: false },
    ],
  },
  {
    label: 'AI',
    items: [
      { to: '/recommendations', label: 'Recommendations', icon: FileText, always: false, ai: true },
      { to: '/advisor', label: 'Advisor', icon: Bot, always: false, ai: true },
    ],
  },
  {
    label: 'Settings',
    items: [
      { to: '/settings/workspace', label: 'Workspace', icon: Settings, always: true, ai: false },
      { to: '/settings/org', label: 'Organisation', icon: Settings, always: true, ai: false },
    ],
  },
]

export function Sidebar() {
  const { modelId } = useModel()
  const aiEnabled = useAIEnabled()
  const location = useLocation()
  const [collapsed, setCollapsed] = useState(false)
  const { workspace, org, workspaces } = useWorkspace()
  const [expandedSections, setExpandedSections] = useState<Set<string>>(() => {
    const initial = new Set<string>()
    for (const section of NAV_SECTIONS) {
      for (const item of section.items) {
        if (item.children && location.pathname.startsWith(item.to)) {
          initial.add(item.to)
        }
      }
    }
    return initial
  })

  const hasMultipleWorkspaces = workspaces.length > 1

  const toggleSection = (path: string) => {
    setExpandedSections(prev => {
      const next = new Set(prev)
      if (next.has(path)) next.delete(path)
      else next.add(path)
      return next
    })
  }

  const isPathActive = (path: string) => location.pathname === path || location.pathname.startsWith(path + '/')

  return (
    <aside
      className={cn(
        'flex-shrink-0 flex flex-col bg-muted/40 border-r border-border transition-all duration-200',
        collapsed ? 'w-14' : 'w-56'
      )}
    >
      <div className="flex items-center gap-2.5 px-4 border-b border-border h-14">
        {!collapsed && (
          <div className="flex flex-col min-w-0">
            <span className="text-sm font-semibold text-foreground truncate">UNM Platform</span>
            {workspace ? (
              <span className="text-xs text-muted-foreground mt-0.5 truncate">{workspace.name}</span>
            ) : (
              <span className="text-xs text-muted-foreground mt-0.5">Architecture Explorer</span>
            )}
          </div>
        )}
        <button
          onClick={() => setCollapsed(c => !c)}
          className={cn(
            'p-1 rounded text-muted-foreground hover:text-foreground hover:bg-muted transition-colors shrink-0',
            collapsed ? 'mx-auto' : 'ml-auto'
          )}
          title={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
          aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
        >
          {collapsed ? <ChevronRight className="w-4 h-4" /> : <ChevronLeft className="w-4 h-4" />}
        </button>
      </div>

      {/* Workspace label + switch link */}
      {!collapsed && workspace && org && (
        <div className="px-4 py-2.5 border-b border-border/60 bg-muted/20">
          <p className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground mb-0.5">
            {org.name}
          </p>
          <div className="flex items-center justify-between">
            <p className="text-xs font-medium text-foreground truncate">{workspace.name}</p>
            {hasMultipleWorkspaces && (
              <NavLink
                to="/settings/workspace"
                title="Switch workspace"
                className="ml-2 shrink-0 text-muted-foreground hover:text-foreground transition-colors"
              >
                <ArrowLeftRight className="w-3 h-3" />
              </NavLink>
            )}
          </div>
        </div>
      )}

      {/* Navigation */}
      <nav className="flex-1 px-2 py-3 overflow-y-auto space-y-4">
        {NAV_SECTIONS.map((section) => {
          const visibleItems = section.items.filter(item => !item.ai || aiEnabled)
          if (visibleItems.length === 0) return null

          return (
            <div key={section.label}>
              {!collapsed && (
                <p className="px-2 pb-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
                  {section.label}
                </p>
              )}
              <div className="space-y-0.5">
                {visibleItems.map((item) => {
                  const disabled = !item.always && !modelId

                  if (disabled) {
                    return (
                      <span
                        key={item.to}
                        title={collapsed ? `${item.label} — load a model first` : 'Load a model first'}
                        aria-disabled="true"
                        className={cn(
                          'flex items-center gap-2.5 px-2 py-2 rounded-md text-sm opacity-40 cursor-not-allowed select-none text-muted-foreground',
                          collapsed && 'justify-center'
                        )}
                      >
                        <item.icon className="w-4 h-4 shrink-0" />
                        {!collapsed && item.label}
                      </span>
                    )
                  }

                  if (item.children && !collapsed) {
                    const sectionActive = isPathActive(item.to)
                    const isOpen = expandedSections.has(item.to) || sectionActive

                    return (
                      <div key={item.to}>
                        <button
                          onClick={() => toggleSection(item.to)}
                          className={cn(
                            'flex items-center gap-2.5 px-2 py-2 rounded-md text-sm w-full text-left transition-colors',
                            sectionActive
                              ? 'text-foreground font-medium'
                              : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                          )}
                        >
                          <item.icon className="w-4 h-4 shrink-0" />
                          <span className="flex-1">{item.label}</span>
                          <ChevronDown className={cn(
                            'w-3 h-3 transition-transform duration-200',
                            isOpen && 'rotate-180'
                          )} />
                        </button>
                        {isOpen && (
                          <div className="ml-[22px] mt-0.5 space-y-0.5 border-l border-border pl-2.5">
                            {item.children.map(child => (
                              <NavLink
                                key={child.to}
                                to={child.to}
                                end={child.end}
                                className={({ isActive }) => cn(
                                  'block px-2 py-1.5 rounded-md text-[13px] transition-colors',
                                  isActive
                                    ? 'bg-muted text-foreground font-medium'
                                    : 'text-muted-foreground hover:text-foreground hover:bg-muted/60'
                                )}
                              >
                                {child.label}
                              </NavLink>
                            ))}
                          </div>
                        )}
                      </div>
                    )
                  }

                  return (
                    <NavLink
                      key={item.to}
                      to={item.to}
                      end={item.to === '/workspace' || item.to === '/'}
                      title={collapsed ? item.label : undefined}
                      className={({ isActive }) =>
                        cn(
                          'flex items-center gap-2.5 px-2 py-2 rounded-md text-sm transition-colors',
                          collapsed && 'justify-center',
                          isActive
                            ? 'bg-muted text-foreground font-medium'
                            : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                        )
                      }
                    >
                      <item.icon className="w-4 h-4 shrink-0" />
                      {!collapsed && item.label}
                    </NavLink>
                  )
                })}
              </div>
            </div>
          )
        })}
      </nav>
    </aside>
  )
}
