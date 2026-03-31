import { useState, useEffect, useMemo } from 'react'
import { Button } from '@/components/ui/button'
import { Plus } from 'lucide-react'
import { api } from '@/lib/api'
import { useModel } from '@/lib/model-context'
import { useChangeset } from '@/lib/changeset-context'
import type { ChangeAction } from '@/lib/api'

// Model entity lists fetched once and cached for dropdown population
export interface ModelEntities {
  teams: string[]
  services: string[]
  capabilities: string[]
  needs: string[]
  actors: string[]
}

export function useModelEntities(): ModelEntities {
  const { modelId } = useModel()
  const { actions } = useChangeset()
  const [fetched, setFetched] = useState<ModelEntities>({ teams: [], services: [], capabilities: [], needs: [], actors: [] })

  useEffect(() => {
    if (!modelId) return
    Promise.all([
      api.getTeams(modelId).then(r => r.teams.map(t => t.name)),
      api.getServices(modelId).then(r => r.services.map(s => s.name)),
      api.getCapabilities(modelId).then(r => r.capabilities.map(c => c.name)),
      api.getNeeds(modelId).then(r => r.needs.map(n => n.name)),
      api.getActors(modelId).then(r => r.actors.map(a => a.name)),
    ]).then(([teams, services, capabilities, needs, actors]) => {
      setFetched({ teams, services, capabilities, needs, actors })
    }).catch(() => {})
  }, [modelId])

  // Merge pending add_* actions so newly staged entities appear in dropdowns immediately
  return useMemo(() => {
    const pendingCaps: string[] = []
    const pendingTeams: string[] = []
    const pendingServices: string[] = []
    const pendingNeeds: string[] = []
    const pendingActors: string[] = []

    for (const a of actions) {
      const ac = a as unknown as Record<string, unknown>
      if (a.type === 'add_capability' && typeof ac.capability_name === 'string') pendingCaps.push(ac.capability_name)
      if (a.type === 'add_team'       && typeof ac.team_name        === 'string') pendingTeams.push(ac.team_name)
      if (a.type === 'add_service'    && typeof ac.service_name     === 'string') pendingServices.push(ac.service_name)
      if (a.type === 'add_need'       && typeof ac.need_name        === 'string') pendingNeeds.push(ac.need_name)
      if (a.type === 'add_actor'      && typeof ac.actor_name       === 'string') pendingActors.push(ac.actor_name)
    }

    return {
      teams:        [...new Set([...fetched.teams,        ...pendingTeams])],
      services:     [...new Set([...fetched.services,     ...pendingServices])],
      capabilities: [...new Set([...fetched.capabilities, ...pendingCaps])],
      needs:        [...new Set([...fetched.needs,        ...pendingNeeds])],
      actors:       [...new Set([...fetched.actors,       ...pendingActors])],
    }
  }, [fetched, actions])
}

type ActionType = ChangeAction['type']

interface ActionCategory {
  label: string
  actions: Array<{ value: ActionType; label: string }>
}

const ACTION_CATEGORIES: ActionCategory[] = [
  {
    label: 'Services',
    actions: [
      { value: 'move_service', label: 'Move Service' },
      { value: 'add_service', label: 'Add Service' },
      { value: 'remove_service', label: 'Remove Service' },
      { value: 'rename_service', label: 'Rename Service' },
      { value: 'add_service_dependency', label: 'Add Service Dependency' },
      { value: 'remove_service_dependency', label: 'Remove Service Dependency' },
      { value: 'link_capability_service', label: 'Link Capability to Service' },
      { value: 'unlink_capability_service', label: 'Unlink Capability from Service' },
    ],
  },
  {
    label: 'Teams',
    actions: [
      { value: 'add_team', label: 'Add Team' },
      { value: 'remove_team', label: 'Remove Team' },
      { value: 'update_team_type', label: 'Change Team Type' },
      { value: 'update_team_size', label: 'Change Team Size' },
      { value: 'split_team', label: 'Split Team' },
      { value: 'merge_teams', label: 'Merge Teams' },
    ],
  },
  {
    label: 'Capabilities',
    actions: [
      { value: 'add_capability', label: 'Add Capability' },
      { value: 'remove_capability', label: 'Remove Capability' },
      { value: 'reassign_capability', label: 'Reassign Capability' },
      { value: 'update_capability_visibility', label: 'Change Capability Visibility' },
    ],
  },
  {
    label: 'Needs & Actors',
    actions: [
      { value: 'add_need', label: 'Add Need' },
      { value: 'remove_need', label: 'Remove Need' },
      { value: 'link_need_capability', label: 'Link Need to Capability' },
      { value: 'unlink_need_capability', label: 'Unlink Need from Capability' },
      { value: 'add_actor', label: 'Add Actor' },
      { value: 'remove_actor', label: 'Remove Actor' },
    ],
  },
  {
    label: 'Interactions & Other',
    actions: [
      { value: 'add_interaction', label: 'Add Team Interaction' },
      { value: 'remove_interaction', label: 'Remove Team Interaction' },
      { value: 'update_description', label: 'Update Description' },
    ],
  },
]

const ACTION_DESCRIPTIONS: Record<ActionType, string> = {
  move_service: 'Move a service from one team to another',
  add_service: 'Add a new service to the model',
  remove_service: 'Remove a service and its relationships',
  link_capability_service: 'Link a capability to a service',
  unlink_capability_service: 'Unlink a capability from a service',
  add_team: 'Add a new team to the model',
  remove_team: 'Remove a team (must have no owned services)',
  split_team: 'Create a new team and move selected services to it',
  merge_teams: 'Merge two teams, combining all their services',
  update_team_type: 'Change the Team Topologies type of a team',
  update_team_size: 'Update the cognitive size of a team',
  add_capability: 'Add a new capability to the model',
  remove_capability: 'Remove a capability from the model',
  reassign_capability: "Move a capability's ownership to another team",
  update_capability_visibility: "Change a capability's visibility level",
  link_need_capability: 'Link a need to a supporting capability',
  unlink_need_capability: 'Unlink a need from a capability',
  add_need: 'Add a new need for an actor',
  remove_need: 'Remove a need from the model',
  add_actor: 'Add a new actor to the model',
  remove_actor: 'Remove an actor from the model',
  add_interaction: 'Add an interaction between two teams',
  remove_interaction: 'Remove an interaction between teams',
  update_description: 'Update the description of any entity',
  rename_service: 'Rename an existing service',
  add_service_dependency: 'Add a dependency between two services',
  remove_service_dependency: 'Remove a dependency between two services',
}

type FieldSource = 'teams' | 'services' | 'capabilities' | 'needs' | 'actors'

interface FieldDef {
  key: string
  label: string
  type: 'text' | 'number' | 'select' | 'entity'
  source?: FieldSource
  options?: string[]
  optional?: boolean
  placeholder?: string
}

function buildFields(_entities: ModelEntities): Record<ActionType, FieldDef[]> {
  return {
    move_service: [
      { key: 'service_name', label: 'Service', type: 'entity', source: 'services' },
      { key: 'from_team_name', label: 'From Team', type: 'entity', source: 'teams' },
      { key: 'to_team_name', label: 'To Team', type: 'entity', source: 'teams' },
    ],
    split_team: [
      { key: 'original_team_name', label: 'Team to Split', type: 'entity', source: 'teams' },
      { key: 'new_team_a_name', label: 'New Team A', type: 'text', placeholder: 'New team name' },
      { key: 'new_team_b_name', label: 'New Team B', type: 'text', placeholder: 'New team name' },
    ],
    merge_teams: [
      { key: 'team_a_name', label: 'Team A', type: 'entity', source: 'teams' },
      { key: 'team_b_name', label: 'Team B', type: 'entity', source: 'teams' },
      { key: 'new_team_name', label: 'Merged Name', type: 'text', placeholder: 'New team name' },
    ],
    add_capability: [
      { key: 'capability_name', label: 'Name', type: 'text', placeholder: 'Capability name' },
      { key: 'owner_team_name', label: 'Owner Team', type: 'entity', source: 'teams' },
      { key: 'visibility', label: 'Visibility', type: 'select', options: ['user-facing', 'domain', 'foundational', 'infrastructure'] },
    ],
    remove_capability: [
      { key: 'capability_name', label: 'Capability', type: 'entity', source: 'capabilities' },
    ],
    reassign_capability: [
      { key: 'capability_name', label: 'Capability', type: 'entity', source: 'capabilities' },
      { key: 'from_team_name', label: 'From Team', type: 'entity', source: 'teams' },
      { key: 'to_team_name', label: 'To Team', type: 'entity', source: 'teams' },
    ],
    add_interaction: [
      { key: 'source_team_name', label: 'Source Team', type: 'entity', source: 'teams' },
      { key: 'target_team_name', label: 'Target Team', type: 'entity', source: 'teams' },
      { key: 'interaction_mode', label: 'Mode', type: 'select', options: ['collaboration', 'x-as-a-service', 'facilitating'] },
    ],
    remove_interaction: [
      { key: 'source_team_name', label: 'Source Team', type: 'entity', source: 'teams' },
      { key: 'target_team_name', label: 'Target Team', type: 'entity', source: 'teams' },
    ],
    update_team_size: [
      { key: 'team_name', label: 'Team', type: 'entity', source: 'teams' },
      { key: 'new_size', label: 'New Size', type: 'number', placeholder: 'e.g. 8' },
    ],
    add_service: [
      { key: 'service_name', label: 'Name', type: 'text', placeholder: 'Service name' },
      { key: 'owner_team_name', label: 'Owner Team', type: 'entity', source: 'teams' },
      { key: 'description', label: 'Description', type: 'text', optional: true, placeholder: 'Optional description' },
    ],
    remove_service: [
      { key: 'service_name', label: 'Service', type: 'entity', source: 'services' },
    ],
    rename_service: [
      { key: 'service_name', label: 'Service', type: 'entity', source: 'services' },
      { key: 'new_service_name', label: 'New Name', type: 'text', placeholder: 'New service name' },
    ],
    add_team: [
      { key: 'team_name', label: 'Name', type: 'text', placeholder: 'Team name' },
      { key: 'team_type', label: 'Type', type: 'select', options: ['stream-aligned', 'platform', 'enabling', 'complicated-subsystem'] },
      { key: 'description', label: 'Description', type: 'text', optional: true, placeholder: 'Optional description' },
      { key: 'new_size', label: 'Size', type: 'number', optional: true, placeholder: 'e.g. 5' },
    ],
    remove_team: [
      { key: 'team_name', label: 'Team', type: 'entity', source: 'teams' },
    ],
    update_team_type: [
      { key: 'team_name', label: 'Team', type: 'entity', source: 'teams' },
      { key: 'team_type', label: 'New Type', type: 'select', options: ['stream-aligned', 'platform', 'enabling', 'complicated-subsystem'] },
    ],
    add_need: [
      { key: 'need_name', label: 'Need Name', type: 'text', placeholder: 'Need name' },
      { key: 'actor_name', label: 'Actor', type: 'entity', source: 'actors' },
      { key: 'outcome', label: 'Outcome', type: 'text', optional: true, placeholder: 'Expected outcome' },
    ],
    remove_need: [
      { key: 'need_name', label: 'Need', type: 'entity', source: 'needs' },
    ],
    add_actor: [
      { key: 'actor_name', label: 'Name', type: 'text', placeholder: 'Actor name' },
      { key: 'description', label: 'Description', type: 'text', optional: true, placeholder: 'Optional description' },
    ],
    remove_actor: [
      { key: 'actor_name', label: 'Actor', type: 'entity', source: 'actors' },
    ],
    add_service_dependency: [
      { key: 'service_name', label: 'Service', type: 'entity', source: 'services' },
      { key: 'depends_on_service', label: 'Depends On', type: 'entity', source: 'services' },
    ],
    remove_service_dependency: [
      { key: 'service_name', label: 'Service', type: 'entity', source: 'services' },
      { key: 'depends_on_service', label: 'Depends On', type: 'entity', source: 'services' },
    ],
    link_need_capability: [
      { key: 'need_name', label: 'Need', type: 'entity', source: 'needs' },
      { key: 'capability_name', label: 'Capability', type: 'entity', source: 'capabilities' },
    ],
    unlink_need_capability: [
      { key: 'need_name', label: 'Need', type: 'entity', source: 'needs' },
      { key: 'capability_name', label: 'Capability', type: 'entity', source: 'capabilities' },
    ],
    link_capability_service: [
      { key: 'capability_name', label: 'Capability', type: 'entity', source: 'capabilities' },
      { key: 'service_name', label: 'Service', type: 'entity', source: 'services' },
      { key: 'role', label: 'Role', type: 'select', options: ['primary', 'supporting', 'consuming'], optional: true },
    ],
    unlink_capability_service: [
      { key: 'capability_name', label: 'Capability', type: 'entity', source: 'capabilities' },
      { key: 'service_name', label: 'Service', type: 'entity', source: 'services' },
    ],
    update_capability_visibility: [
      { key: 'capability_name', label: 'Capability', type: 'entity', source: 'capabilities' },
      { key: 'visibility', label: 'Visibility', type: 'select', options: ['user-facing', 'domain', 'foundational', 'infrastructure'] },
    ],
    update_description: [
      { key: 'entity_type', label: 'Entity Type', type: 'select', options: ['actor', 'need', 'capability', 'service', 'team'] },
      { key: 'entity_name', label: 'Entity', type: 'entity', source: 'services', placeholder: 'Select entity...' },
      { key: 'description', label: 'New Description', type: 'text', placeholder: 'Description' },
    ],
  }
}

const SELECT_STYLE = {
  borderColor: '#d1d5db', background: '#ffffff', color: '#111827',
} as const

const INPUT_STYLE = SELECT_STYLE

interface ActionFormProps {
  onAdd: (action: ChangeAction) => void
  entities?: ModelEntities
  compact?: boolean
  initialAction?: ChangeAction | null
}

export function ActionForm({ onAdd, entities: entityProp, compact, initialAction }: ActionFormProps) {
  const defaultEntities = useModelEntities()
  const entities = entityProp ?? defaultEntities

  const [actionType, setActionType] = useState<ActionType | null>(initialAction?.type ?? null)
  const [fieldValues, setFieldValues] = useState<Record<string, string>>({})
  const [fieldErrors, setFieldErrors] = useState<Record<string, boolean>>({})
  const [submitted, setSubmitted] = useState(false)

  const fieldsMap = buildFields(entities)
  const fields = actionType ? fieldsMap[actionType] : []

  useEffect(() => {
    if (!initialAction) return
    setActionType(initialAction.type)
    const vals: Record<string, string> = {}
    for (const [k, v] of Object.entries(initialAction)) {
      if (k !== 'type' && v !== undefined) vals[k] = String(v)
    }
    setFieldValues(vals)
  }, [initialAction])

  const handleTypeChange = (type: ActionType) => {
    setActionType(type)
    setFieldValues({})
    setFieldErrors({})
    setSubmitted(false)
  }

  const handleSubmit = () => {
    if (!actionType) return
    setSubmitted(true)

    // Validate required fields
    const errors: Record<string, boolean> = {}
    let hasError = false
    for (const field of fields) {
      const val = fieldValues[field.key]?.trim()
      if (!val && !field.optional) {
        errors[field.key] = true
        hasError = true
      }
    }
    setFieldErrors(errors)
    if (hasError) return

    const action: Record<string, unknown> = { type: actionType }
    for (const field of fields) {
      const val = fieldValues[field.key]?.trim()
      if (!val) {
        if (field.optional) continue
        return
      }
      action[field.key] = field.type === 'number' ? Number(val) : val
    }
    onAdd(action as unknown as ChangeAction)
    setFieldValues({})
    setFieldErrors({})
    setSubmitted(false)
  }

  const handleFieldChange = (key: string, value: string) => {
    setFieldValues(p => ({ ...p, [key]: value }))
    if (fieldErrors[key]) {
      setFieldErrors(p => ({ ...p, [key]: false }))
    }
  }

  const allFilled = fields.every(f => f.optional || fieldValues[f.key]?.trim())

  const resolveSource = (field: FieldDef): FieldSource | undefined => {
    if (actionType === 'update_description' && field.key === 'entity_name') {
      const entityTypeMap: Record<string, FieldSource> = {
        actor: 'actors', need: 'needs', capability: 'capabilities', service: 'services', team: 'teams',
      }
      return entityTypeMap[fieldValues['entity_type'] ?? ''] ?? undefined
    }
    return field.source
  }

  const sortedEntities = (source: FieldSource | undefined): string[] => {
    if (!source) return []
    return [...entities[source]].sort((a, b) => a.localeCompare(b))
  }

  const renderField = (field: FieldDef) => {
    const hasError = submitted && fieldErrors[field.key]
    const errorBorder = hasError ? '#ef4444' : '#d1d5db'

    if (field.type === 'entity') {
      const source = resolveSource(field)
      const options = sortedEntities(source)
      return (
        <>
          <select
            className="w-full rounded-md border px-2.5 py-1.5 text-sm"
            style={{ ...SELECT_STYLE, borderColor: errorBorder }}
            value={fieldValues[field.key] ?? ''}
            onChange={e => handleFieldChange(field.key, e.target.value)}
          >
            <option value="">Select {field.label.toLowerCase()}...</option>
            {options.map(o => <option key={o} value={o}>{o}</option>)}
          </select>
          {hasError && <p className="text-xs mt-0.5" style={{ color: '#ef4444' }}>This field is required</p>}
        </>
      )
    }
    if (field.type === 'select') {
      return (
        <>
          <select
            className="w-full rounded-md border px-2.5 py-1.5 text-sm"
            style={{ ...SELECT_STYLE, borderColor: errorBorder }}
            value={fieldValues[field.key] ?? ''}
            onChange={e => handleFieldChange(field.key, e.target.value)}
          >
            <option value="">{field.optional ? '(none)' : 'Select...'}</option>
            {field.options?.map(o => <option key={o} value={o}>{o}</option>)}
          </select>
          {hasError && <p className="text-xs mt-0.5" style={{ color: '#ef4444' }}>This field is required</p>}
        </>
      )
    }
    return (
      <>
        <input
          className="w-full rounded-md border px-2.5 py-1.5 text-sm"
          style={{ ...INPUT_STYLE, borderColor: errorBorder }}
          type={field.type}
          placeholder={field.placeholder ?? field.label}
          value={fieldValues[field.key] ?? ''}
          onChange={e => handleFieldChange(field.key, e.target.value)}
        />
        {hasError && <p className="text-xs mt-0.5" style={{ color: '#ef4444' }}>This field is required</p>}
      </>
    )
  }

  // Find the label for the currently selected action type
  const selectedLabel = actionType
    ? ACTION_CATEGORIES.flatMap(c => c.actions).find(a => a.value === actionType)?.label ?? actionType
    : null

  return (
    <div className="space-y-2.5">
      {/* Action type: grouped select dropdown */}
      <div>
        <label className="text-xs font-medium block mb-1.5" style={{ color: '#6b7280' }}>Action Type</label>
        <select
          className="w-full rounded-md border px-2.5 py-1.5 text-sm"
          style={{ ...SELECT_STYLE, color: actionType ? '#111827' : '#9ca3af' }}
          value={actionType ?? ''}
          onChange={e => handleTypeChange(e.target.value as ActionType)}
        >
          <option value="">Select action type…</option>
          {ACTION_CATEGORIES.map(cat => (
            <optgroup key={cat.label} label={cat.label}>
              {cat.actions.map(a => (
                <option key={a.value} value={a.value}>{a.label}</option>
              ))}
            </optgroup>
          ))}
        </select>
      </div>

      {/* Action description + form fields */}
      {actionType && (
        <>
          <div>
            <p className="text-sm font-medium" style={{ color: '#111827' }}>{selectedLabel}</p>
            <p className="text-xs" style={{ color: '#9ca3af' }}>{ACTION_DESCRIPTIONS[actionType]}</p>
          </div>

          {fields.map(field => (
            <div key={field.key}>
              <label className="text-xs font-medium block mb-1" style={{ color: '#6b7280' }}>
                {field.label}
                {field.optional && <span style={{ color: '#d1d5db', fontWeight: 400 }}> (optional)</span>}
              </label>
              {renderField(field)}
            </div>
          ))}

          <Button size={compact ? 'sm' : 'default'} disabled={!allFilled} onClick={handleSubmit} className="w-full gap-1.5">
            <Plus size={14} />
            Add
          </Button>
        </>
      )}
    </div>
  )
}
