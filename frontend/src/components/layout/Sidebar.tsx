import { useState } from 'react'
import { NavLink } from 'react-router-dom'
import {
  Upload, LayoutDashboard, Map, Users, Layers, Flag, Network,
  Activity, GitBranch, AlertCircle, FlaskConical, Bot, FileText,
  ChevronLeft, ChevronRight,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useModel } from '@/lib/model-context'
import { useAIEnabled } from '@/hooks/useAIEnabled'

const NAV_SECTIONS = [
  {
    label: 'Model',
    items: [
      { to: '/', label: 'Upload', icon: Upload, always: true, ai: false },
      { to: '/dashboard', label: 'Dashboard', icon: LayoutDashboard, always: false, ai: false },
      { to: '/signals', label: 'Signals', icon: AlertCircle, always: false, ai: false },
    ],
  },
  {
    label: 'Views',
    items: [
      { to: '/unm-map', label: 'UNM Map', icon: Map, always: false, ai: false },
      { to: '/need', label: 'Need View', icon: Users, always: false, ai: false },
      { to: '/capability', label: 'Capability', icon: Layers, always: false, ai: false },
      { to: '/ownership', label: 'Ownership', icon: Flag, always: false, ai: false },
      { to: '/team-topology', label: 'Team Topology', icon: Network, always: false, ai: false },
      { to: '/cognitive-load', label: 'Cognitive Load', icon: Activity, always: false, ai: false },
      { to: '/realization', label: 'Realization', icon: GitBranch, always: false, ai: false },
    ],
  },
  {
    label: 'Editing',
    items: [
      { to: '/what-if', label: 'What-If', icon: FlaskConical, always: false, ai: false },
    ],
  },
  {
    label: 'AI',
    items: [
      { to: '/recommendations', label: 'Recommendations', icon: FileText, always: false, ai: true },
      { to: '/advisor', label: 'Advisor', icon: Bot, always: false, ai: true },
    ],
  },
]

export function Sidebar() {
  const { modelId } = useModel()
  const aiEnabled = useAIEnabled()
  const [collapsed, setCollapsed] = useState(false)

  return (
    <aside
      className={cn(
        'flex-shrink-0 flex flex-col bg-gray-50 border-r border-border transition-all duration-200',
        collapsed ? 'w-14' : 'w-56'
      )}
    >
      {/* Logo / Title */}
      <div className="flex items-center gap-2.5 px-4 border-b border-border h-14">
        {!collapsed && (
          <div className="flex flex-col min-w-0">
            <span className="text-sm font-semibold text-foreground truncate">UNM Platform</span>
            <span className="text-xs text-muted-foreground mt-0.5">Architecture Explorer</span>
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
                {visibleItems.map(({ to, label, icon: Icon, always }) => {
                  const disabled = !always && !modelId
                  if (disabled) {
                    return (
                      <span
                        key={to}
                        title={collapsed ? `${label} — load a model first` : 'Load a model first to access this view'}
                        aria-disabled="true"
                        className={cn(
                          'flex items-center gap-2.5 px-2 py-2 rounded-md text-sm opacity-40 cursor-not-allowed select-none text-muted-foreground',
                          collapsed && 'justify-center'
                        )}
                      >
                        <Icon className="w-4 h-4 shrink-0" />
                        {!collapsed && label}
                      </span>
                    )
                  }
                  return (
                    <NavLink
                      key={to}
                      to={to}
                      title={collapsed ? label : undefined}
                      className={({ isActive }) =>
                        cn(
                          'flex items-center gap-2.5 px-2 py-2 rounded-md text-sm transition-colors',
                          collapsed && 'justify-center',
                          isActive
                            ? 'bg-gray-200 text-gray-900 font-medium'
                            : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
                        )
                      }
                    >
                      <Icon className="w-4 h-4 shrink-0" />
                      {!collapsed && label}
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
