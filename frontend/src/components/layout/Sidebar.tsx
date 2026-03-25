import { NavLink } from 'react-router-dom'
import { Upload, LayoutDashboard, Map, Users, Layers, Flag, Network, Activity, GitBranch, AlertCircle, FlaskConical, Bot, FileText, Pencil } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useModel } from '@/lib/model-context'
import { useAIEnabled } from '@/hooks/useAIEnabled'

const navItems = [
  { to: '/', label: 'Upload', icon: Upload, always: true, ai: false },
  { to: '/dashboard', label: 'Dashboard', icon: LayoutDashboard, always: false, ai: false },
  { to: '/signals', label: 'Signals', icon: AlertCircle, always: false, ai: false },
  { to: '/unm-map', label: 'UNM Map', icon: Map, always: false, ai: false },
  { to: '/need', label: 'Need View', icon: Users, always: false, ai: false },
  { to: '/capability', label: 'Capability View', icon: Layers, always: false, ai: false },
  { to: '/ownership', label: 'Ownership View', icon: Flag, always: false, ai: false },
  { to: '/team-topology', label: 'Team Topology', icon: Network, always: false, ai: false },
  { to: '/cognitive-load', label: 'Cognitive Load', icon: Activity, always: false, ai: false },
  { to: '/realization', label: 'Realization View', icon: GitBranch, always: false, ai: false },
  { to: '/edit', label: 'Edit Model', icon: Pencil, always: false, ai: false },
  { to: '/what-if', label: 'What-If Explorer', icon: FlaskConical, always: false, ai: false },
  { to: '/recommendations', label: 'AI Recommendations', icon: FileText, always: false, ai: true },
  { to: '/advisor', label: 'AI Advisor', icon: Bot, always: false, ai: true },
]

export function Sidebar() {
  const { modelId } = useModel()
  const aiEnabled = useAIEnabled()

  return (
    <aside className="w-56 flex-shrink-0 flex flex-col" style={{ background: '#f9fafb', borderRight: '1px solid #e5e7eb' }}>
      <div className="flex items-center gap-2.5 px-4 border-b" style={{ height: 56, borderColor: '#e5e7eb' }}>
        <div className="flex flex-col">
          <span className="text-sm font-semibold" style={{ color: '#111827', letterSpacing: '-0.01em' }}>UNM Platform</span>
          <span className="text-xs" style={{ color: '#9ca3af', marginTop: 1 }}>Architecture Explorer</span>
        </div>
      </div>

      <nav className="flex-1 px-3 py-3 space-y-0.5">
        {navItems.filter(item => !item.ai || aiEnabled).map(({ to, label, icon: Icon, always }) => {
          const disabled = !always && !modelId
          if (disabled) {
            return (
              <span
                key={to}
                title="Load a model first to access this view"
                aria-disabled="true"
                className="flex items-center gap-2.5 px-3 py-2 rounded-md text-sm opacity-40 cursor-not-allowed select-none"
                style={{ color: '#9ca3af' }}
              >
                <Icon size={15} style={{ flexShrink: 0 }} />
                {label}
              </span>
            )
          }
          return (
            <NavLink
              key={to}
              to={to}
              className={({ isActive }) =>
                cn(
                  'flex items-center gap-2.5 px-3 py-2 rounded-md text-sm transition-colors',
                  isActive ? 'font-medium' : '',
                )
              }
              style={({ isActive }) => ({
                color: isActive ? '#111827' : '#6b7280',
                background: isActive ? '#e5e7eb' : 'transparent',
              })}
              onMouseEnter={e => {
                const el = e.currentTarget
                if (!el.classList.contains('font-medium')) {
                  el.style.background = '#f3f4f6'
                  el.style.color = '#111827'
                }
              }}
              onMouseLeave={e => {
                const el = e.currentTarget
                if (!el.classList.contains('font-medium')) {
                  el.style.background = 'transparent'
                  el.style.color = '#6b7280'
                }
              }}
            >
              <Icon size={15} style={{ flexShrink: 0 }} />
              {label}
            </NavLink>
          )
        })}
      </nav>
    </aside>
  )
}
