import { useState, useRef, useEffect, useMemo } from 'react'
import { ZoomIn, ZoomOut, Maximize2, X, Info, Search, Zap, ArrowRight, Plus, Minus, RefreshCw, Pencil, GitBranch, Link2, Unlink, Users, Layers } from 'lucide-react'
import { useReactFlow } from '@xyflow/react'
import { VIS_ORDER, VIS } from '@/features/unm-map/constants'
import { useChangeset } from '@/lib/changeset-context'
import { ActionForm } from '@/components/changeset/ActionForm'
import type { ChangeAction } from '@/lib/api'

type ActionType = ChangeAction['type']

interface ActionDef {
  value: ActionType
  label: string
  icon: typeof Plus
  iconColor: string
}

interface ActionGroup {
  label: string
  actions: ActionDef[]
}

const ACTION_GROUPS: ActionGroup[] = [
  {
    label: 'Services',
    actions: [
      { value: 'move_service', label: 'Move Service', icon: ArrowRight, iconColor: '#2563eb' },
      { value: 'add_service', label: 'Add Service', icon: Plus, iconColor: '#059669' },
      { value: 'remove_service', label: 'Remove Service', icon: Minus, iconColor: '#dc2626' },
      { value: 'rename_service', label: 'Rename Service', icon: Pencil, iconColor: '#7c3aed' },
      { value: 'add_service_dependency', label: 'Add Dependency', icon: Link2, iconColor: '#0891b2' },
      { value: 'remove_service_dependency', label: 'Remove Dependency', icon: Unlink, iconColor: '#dc2626' },
      { value: 'link_capability_service', label: 'Link to Capability', icon: Link2, iconColor: '#059669' },
      { value: 'unlink_capability_service', label: 'Unlink from Capability', icon: Unlink, iconColor: '#dc2626' },
    ],
  },
  {
    label: 'Teams',
    actions: [
      { value: 'add_team', label: 'Add Team', icon: Plus, iconColor: '#059669' },
      { value: 'remove_team', label: 'Remove Team', icon: Minus, iconColor: '#dc2626' },
      { value: 'update_team_type', label: 'Change Type', icon: RefreshCw, iconColor: '#7c3aed' },
      { value: 'update_team_size', label: 'Change Size', icon: Users, iconColor: '#2563eb' },
      { value: 'split_team', label: 'Split Team', icon: GitBranch, iconColor: '#ea580c' },
      { value: 'merge_teams', label: 'Merge Teams', icon: Layers, iconColor: '#0891b2' },
    ],
  },
  {
    label: 'Capabilities',
    actions: [
      { value: 'add_capability', label: 'Add Capability', icon: Plus, iconColor: '#059669' },
      { value: 'remove_capability', label: 'Remove', icon: Minus, iconColor: '#dc2626' },
      { value: 'reassign_capability', label: 'Reassign', icon: ArrowRight, iconColor: '#2563eb' },
      { value: 'update_capability_visibility', label: 'Change Visibility', icon: Pencil, iconColor: '#7c3aed' },
    ],
  },
  {
    label: 'Needs & Actors',
    actions: [
      { value: 'add_need', label: 'Add Need', icon: Plus, iconColor: '#059669' },
      { value: 'remove_need', label: 'Remove Need', icon: Minus, iconColor: '#dc2626' },
      { value: 'link_need_capability', label: 'Link Need → Capability', icon: Link2, iconColor: '#059669' },
      { value: 'unlink_need_capability', label: 'Unlink Need', icon: Unlink, iconColor: '#dc2626' },
      { value: 'add_actor', label: 'Add Actor', icon: Plus, iconColor: '#059669' },
      { value: 'remove_actor', label: 'Remove Actor', icon: Minus, iconColor: '#dc2626' },
    ],
  },
  {
    label: 'Interactions',
    actions: [
      { value: 'add_interaction', label: 'Add Interaction', icon: Plus, iconColor: '#059669' },
      { value: 'remove_interaction', label: 'Remove Interaction', icon: Minus, iconColor: '#dc2626' },
      { value: 'update_description', label: 'Update Description', icon: Pencil, iconColor: '#7c3aed' },
    ],
  },
]

interface MapToolbarProps {
  highlighted: boolean
  onClearHighlight: () => void
}

export function MapToolbar({ highlighted, onClearHighlight }: MapToolbarProps) {
  const { zoomIn, zoomOut, fitView } = useReactFlow()
  const { addAction } = useChangeset()
  const [showActions, setShowActions] = useState(false)
  const [selectedAction, setSelectedAction] = useState<ActionType | null>(null)
  const [showLegend, setShowLegend] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const actionsRef = useRef<HTMLDivElement>(null)
  const legendRef = useRef<HTMLDivElement>(null)
  const searchInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (!showActions) return
    const handler = (e: MouseEvent) => {
      if (actionsRef.current && !actionsRef.current.contains(e.target as Node)) {
        setShowActions(false); setSelectedAction(null); setSearchQuery('')
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [showActions])

  useEffect(() => {
    if (showActions && !selectedAction && searchInputRef.current) {
      searchInputRef.current.focus()
    }
  }, [showActions, selectedAction])

  useEffect(() => {
    if (!showLegend) return
    const handler = (e: MouseEvent) => {
      if (legendRef.current && !legendRef.current.contains(e.target as Node)) setShowLegend(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [showLegend])

  const filteredGroups = useMemo(() => {
    if (!searchQuery.trim()) return ACTION_GROUPS
    const q = searchQuery.toLowerCase()
    return ACTION_GROUPS
      .map(g => ({
        ...g,
        actions: g.actions.filter(a => a.label.toLowerCase().includes(q) || a.value.toLowerCase().includes(q)),
      }))
      .filter(g => g.actions.length > 0)
  }, [searchQuery])

  const handleSelectAction = (type: ActionType) => {
    setSelectedAction(type); setSearchQuery('')
  }

  const handleAdd = (action: ChangeAction) => {
    addAction(action); setShowActions(false); setSelectedAction(null); setSearchQuery('')
  }

  return (
    <div className="mb-2 flex items-center gap-2 flex-shrink-0">
      <h2 className="text-base font-semibold text-foreground">UNM Map</h2>

      {highlighted && (
        <button
          onClick={onClearHighlight}
          className="rounded px-2 py-0.5 text-[10px] text-muted-foreground border border-border hover:bg-muted hover:text-foreground transition-colors"
        >
          Clear highlight
        </button>
      )}

      <div className="ml-auto flex items-center gap-0.5">
        {/* Legend toggle */}
        <div ref={legendRef} className="relative">
          <button
            type="button"
            onClick={() => setShowLegend(s => !s)}
            className="rounded p-1.5 hover:bg-muted text-muted-foreground transition-colors"
            title="Legend"
          >
            <Info size={14} />
          </button>
          {showLegend && (
            <div className="absolute right-0 top-full mt-1 z-50 rounded-lg border px-3 py-2 shadow-lg whitespace-nowrap" style={{ background: '#ffffff', borderColor: '#e5e7eb' }}>
              <div className="flex flex-col gap-1.5 text-[11px]">
                <span className="flex items-center gap-1.5" style={{ color: '#2563eb' }}>
                  <span className="w-2 h-2 rounded-full" style={{ background: '#2563eb' }} /> Actor
                </span>
                <span className="flex items-center gap-1.5" style={{ color: '#3b82f6' }}>
                  <span className="w-2 h-2 rounded-sm" style={{ background: '#3b82f6' }} /> Need
                </span>
                {VIS_ORDER.map(v => (
                  <span key={v} className="flex items-center gap-1.5" style={{ color: VIS[v].border }}>
                    <span className="w-2 h-2 rounded-sm" style={{ background: VIS[v].border }} /> {VIS[v].label}
                  </span>
                ))}
                <span className="flex items-center gap-1.5 text-red-500">
                  <svg width="14" height="6"><line x1="0" y1="3" x2="14" y2="3" stroke="#ef4444" strokeWidth="1.5" strokeDasharray="3 2"/></svg>
                  External dep.
                </span>
              </div>
            </div>
          )}
        </div>

        <div className="h-4 w-px bg-border mx-0.5" />

        {/* Zoom controls */}
        <button type="button" onClick={() => zoomIn()} className="rounded p-1.5 hover:bg-muted text-muted-foreground" title="Zoom in"><ZoomIn size={14} /></button>
        <button type="button" onClick={() => zoomOut()} className="rounded p-1.5 hover:bg-muted text-muted-foreground" title="Zoom out"><ZoomOut size={14} /></button>
        <button type="button" onClick={() => fitView({ padding: 0.05 })} className="rounded p-1.5 hover:bg-muted text-muted-foreground" title="Fit view"><Maximize2 size={14} /></button>

        <div className="h-4 w-px bg-border mx-0.5" />

        {/* Actions menu */}
        <div ref={actionsRef} className="relative">
          <button
            type="button"
            onClick={() => { setShowActions(s => !s); setSelectedAction(null); setSearchQuery('') }}
            className="flex items-center gap-1.5 rounded-md px-2.5 py-1 text-[11px] font-medium transition-colors"
            style={{ background: '#111827', color: '#ffffff' }}
          >
            <Zap size={12} /> Actions
          </button>
          {showActions && (
            <div
              className="absolute right-0 top-full mt-1 z-50 w-[360px] rounded-xl border overflow-hidden"
              style={{ background: '#ffffff', borderColor: '#e5e7eb', boxShadow: '0 20px 40px -12px rgba(0,0,0,0.2)' }}
            >
              {!selectedAction ? (
                <>
                  {/* Search bar */}
                  <div className="px-3 py-2.5" style={{ borderBottom: '1px solid #f3f4f6' }}>
                    <div className="relative">
                      <Search size={13} className="absolute left-2.5 top-1/2 -translate-y-1/2" style={{ color: '#9ca3af' }} />
                      <input
                        ref={searchInputRef}
                        type="text"
                        value={searchQuery}
                        onChange={e => setSearchQuery(e.target.value)}
                        placeholder="Search actions…"
                        className="w-full pl-8 pr-3 py-1.5 text-xs rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500/30"
                        style={{ background: '#f9fafb', border: '1px solid #e5e7eb', color: '#111827' }}
                      />
                    </div>
                  </div>

                  {/* Action list */}
                  <div className="max-h-[400px] overflow-y-auto py-1">
                    {filteredGroups.length === 0 && (
                      <p className="text-center py-6 text-xs" style={{ color: '#9ca3af' }}>No matching actions</p>
                    )}
                    {filteredGroups.map((group, gi) => (
                      <div key={group.label}>
                        {gi > 0 && <div className="mx-3 my-1" style={{ borderTop: '1px solid #f3f4f6' }} />}
                        <div className="px-3 pt-2 pb-1">
                          <span className="text-[10px] font-semibold uppercase tracking-wide" style={{ color: '#9ca3af' }}>{group.label}</span>
                        </div>
                        {group.actions.map(action => {
                          const Icon = action.icon
                          return (
                            <button
                              key={action.value}
                              type="button"
                              onClick={() => handleSelectAction(action.value)}
                              className="w-full flex items-center gap-2.5 px-3 py-1.5 text-left transition-colors hover:bg-gray-50"
                            >
                              <span className="shrink-0 w-5 h-5 rounded flex items-center justify-center" style={{ background: `${action.iconColor}12` }}>
                                <Icon size={11} style={{ color: action.iconColor }} />
                              </span>
                              <span className="text-xs" style={{ color: '#374151' }}>{action.label}</span>
                            </button>
                          )
                        })}
                      </div>
                    ))}
                  </div>
                </>
              ) : (
                <>
                  {/* Selected action — show form */}
                  <div className="px-4 py-3 flex items-center justify-between" style={{ borderBottom: '1px solid #f3f4f6' }}>
                    <button
                      type="button"
                      onClick={() => setSelectedAction(null)}
                      className="text-[11px] transition-colors hover:underline" style={{ color: '#6b7280' }}
                    >
                      ← Back
                    </button>
                    <button type="button" onClick={() => { setShowActions(false); setSelectedAction(null) }}
                      className="rounded p-0.5 hover:bg-muted" style={{ color: '#9ca3af' }}>
                      <X size={12} />
                    </button>
                  </div>
                  <div className="px-4 py-3">
                    <ActionForm
                      onAdd={handleAdd}
                      compact
                      initialAction={{ type: selectedAction } as ChangeAction}
                    />
                  </div>
                </>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
