import { Button } from '@/components/ui/button'
import { X, Trash2 } from 'lucide-react'
import type { ChangeAction } from '@/types/changeset'

const TYPE_LABELS: Record<string, string> = {
  move_service: 'Move Service',
  split_team: 'Split Team',
  merge_teams: 'Merge Teams',
  add_capability: 'Add Capability',
  remove_capability: 'Remove Capability',
  reassign_capability: 'Reassign Capability',
  add_interaction: 'Add Interaction',
  remove_interaction: 'Remove Interaction',
  update_team_size: 'Update Team Size',
  add_service: 'Add Service',
  remove_service: 'Remove Service',
  rename_service: 'Rename Service',
  add_team: 'Add Team',
  remove_team: 'Remove Team',
  update_team_type: 'Update Team Type',
  add_need: 'Add Need',
  remove_need: 'Remove Need',
  add_actor: 'Add Actor',
  remove_actor: 'Remove Actor',
  add_service_dependency: 'Add Service Dependency',
  remove_service_dependency: 'Remove Service Dependency',
  link_need_capability: 'Link Need → Capability',
  unlink_need_capability: 'Unlink Need → Capability',
  link_capability_service: 'Link Capability → Service',
  unlink_capability_service: 'Unlink Capability → Service',
  update_capability_visibility: 'Update Visibility',
  update_description: 'Update Description',
}

function summarizeAction(action: ChangeAction): string {
  const { type: _, ...fields } = action
  return Object.entries(fields)
    .filter(([, v]) => v !== undefined)
    .map(([k, v]) => `${k.replace(/_/g, ' ')}: ${v}`)
    .join(', ')
}

interface ActionListProps {
  actions: ChangeAction[]
  onRemove: (index: number) => void
  onClear: () => void
}

export function ActionList({ actions, onRemove, onClear }: ActionListProps) {
  if (actions.length === 0) {
    return (
      <div className="py-8 text-center">
        <p className="text-sm text-muted-foreground">No actions added yet</p>
        <p className="mt-1 text-xs text-muted-foreground/70">Use the form above to add actions</p>
      </div>
    )
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <span className="text-xs font-medium text-muted-foreground">
          {actions.length} action{actions.length !== 1 ? 's' : ''}
        </span>
        <Button variant="ghost" size="sm" onClick={onClear} className="h-6 px-2 text-xs">
          <Trash2 size={12} />
          Clear All
        </Button>
      </div>

      {actions.map((action, i) => (
        <div
          key={i}
          className="flex items-start gap-2 rounded-lg border border-border bg-muted p-3"
        >
          <div className="min-w-0 flex-1">
            <div className="text-xs font-semibold text-foreground">
              {TYPE_LABELS[action.type] ?? action.type}
            </div>
            <div className="mt-0.5 truncate text-xs text-muted-foreground">
              {summarizeAction(action)}
            </div>
          </div>
          <button
            type="button"
            className="mt-0.5 shrink-0 rounded p-0.5 transition-colors hover:bg-destructive/10"
            onClick={() => onRemove(i)}
          >
            <X size={14} className="text-muted-foreground" />
          </button>
        </div>
      ))}
    </div>
  )
}
