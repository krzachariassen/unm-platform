import { AlertCircle, AlertTriangle, Sparkles, Lightbulb } from 'lucide-react'
import { SlidePanel, PanelSection } from '@/components/ui/slide-panel'

interface AntiPattern {
  code: string
  message: string
  severity: string
}

export interface EntityNode {
  id: string
  label: string
  nodeType: 'team' | 'capability' | 'service' | string
  data: Record<string, unknown>
}

interface EntityDetailPanelProps {
  entity: EntityNode | null
  insight?: { explanation: string; suggestion: string }
  onClose: () => void
}

const TEAM_TYPE_BADGE: Record<string, { bg: string; text: string }> = {
  'stream-aligned': { bg: '#eff6ff', text: '#1d4ed8' },
  platform:         { bg: '#f0fdf4', text: '#15803d' },
  'enabling':       { bg: '#fdf4ff', text: '#7e22ce' },
  'complicated-subsystem': { bg: '#fff7ed', text: '#c2410c' },
}

const VIS_BADGE: Record<string, { bg: string; text: string }> = {
  core:       { bg: '#eff6ff', text: '#1e40af' },
  supporting: { bg: '#f0fdf4', text: '#166534' },
  generic:    { bg: '#f9fafb', text: '#374151' },
}

function Badge({ label, bg, text }: { label: string; bg: string; text: string }) {
  return (
    <span
      className="inline-flex items-center rounded px-2 py-0.5 text-[10px] font-semibold"
      style={{ background: bg, color: text }}
    >
      {label}
    </span>
  )
}

function TeamContent({ data }: { data: Record<string, unknown> }) {
  const type = (data.type as string) ?? ''
  const desc = data.description ? String(data.description) : ''
  const badge = TEAM_TYPE_BADGE[type] ?? { bg: '#f3f4f6', text: '#374151' }
  return (
    <div className="space-y-2">
      <div className="flex flex-wrap gap-1.5">
        {type && <Badge label={type} bg={badge.bg} text={badge.text} />}
        {Boolean(data.is_overloaded) && <Badge label="Overloaded" bg="#fff7ed" text="#c2410c" />}
      </div>
      {desc && <p className="text-xs text-muted-foreground leading-relaxed">{desc}</p>}
    </div>
  )
}

function CapabilityContent({ data }: { data: Record<string, unknown> }) {
  const vis = (data.visibility as string) ?? ''
  const desc = data.description ? String(data.description) : ''
  const badge = VIS_BADGE[vis] ?? { bg: '#f3f4f6', text: '#374151' }
  return (
    <div className="space-y-2">
      <div className="flex flex-wrap gap-1.5">
        {vis && <Badge label={vis} bg={badge.bg} text={badge.text} />}
        {Boolean(data.is_fragmented) && <Badge label="Fragmented" bg="#fef2f2" text="#b91c1c" />}
      </div>
      {desc && <p className="text-xs text-muted-foreground leading-relaxed">{desc}</p>}
    </div>
  )
}

function ServiceContent({ data }: { data: Record<string, unknown> }) {
  const teamLabel = data.team_label ? String(data.team_label) : ''
  const desc = data.description ? String(data.description) : ''
  return (
    <div className="space-y-2">
      {teamLabel && (
        <p className="text-xs text-muted-foreground">
          Owned by <span className="font-medium text-foreground">{teamLabel}</span>
        </p>
      )}
      {desc && <p className="text-xs text-muted-foreground leading-relaxed">{desc}</p>}
    </div>
  )
}

export function EntityDetailPanel({ entity, insight, onClose }: EntityDetailPanelProps) {
  if (!entity) return null

  const antiPatterns = (entity.data.anti_patterns as AntiPattern[] | undefined) ?? []
  const nodeTypeBadges: Record<string, { bg: string; text: string }> = {
    team:        { bg: '#fef3c7', text: '#92400e' },
    capability:  { bg: '#d1fae5', text: '#065f46' },
    service:     { bg: '#f3f4f6', text: '#374151' },
  }
  const typeBadge = nodeTypeBadges[entity.nodeType] ?? { bg: '#f3f4f6', text: '#374151' }

  return (
    <SlidePanel
      open
      onClose={onClose}
      title={entity.label}
      badge={<Badge label={entity.nodeType} bg={typeBadge.bg} text={typeBadge.text} />}
    >
      <div className="space-y-3">
        {entity.nodeType === 'team' && <TeamContent data={entity.data} />}
        {entity.nodeType === 'capability' && <CapabilityContent data={entity.data} />}
        {entity.nodeType === 'service' && <ServiceContent data={entity.data} />}

        {insight && (
          <div className="rounded-lg p-3 border border-border bg-muted/50">
            <div className="flex items-center gap-1.5 mb-1.5">
              <Sparkles size={11} className="text-primary" />
              <span className="text-[10px] font-semibold text-primary uppercase tracking-wide">AI Insight</span>
            </div>
            <p className="text-xs text-foreground leading-relaxed">{insight.explanation}</p>
            {insight.suggestion && (
              <div className="flex items-start gap-1.5 mt-1.5">
                <Lightbulb size={11} className="text-primary shrink-0 mt-0.5" />
                <p className="text-xs text-primary leading-relaxed font-medium">{insight.suggestion}</p>
              </div>
            )}
          </div>
        )}

        {antiPatterns.length > 0 && (
          <PanelSection label={`Anti-patterns (${antiPatterns.length})`}>
            <div className="space-y-1.5 rounded-lg border border-destructive/25 bg-destructive/10 p-2.5">
              {antiPatterns.map((ap, i) => (
                <div key={i} className="flex items-start gap-2 text-xs" style={{ color: ap.severity === 'error' ? '#b91c1c' : '#c2410c' }}>
                  {ap.severity === 'error'
                    ? <AlertCircle size={12} className="mt-0.5 shrink-0" />
                    : <AlertTriangle size={12} className="mt-0.5 shrink-0" />}
                  <span className="leading-normal">{ap.message}</span>
                </div>
              ))}
            </div>
          </PanelSection>
        )}
      </div>
    </SlidePanel>
  )
}
