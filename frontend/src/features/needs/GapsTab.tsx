import { useQuery } from '@tanstack/react-query'
import { useModel } from '@/lib/model-context'
import { viewsApi } from '@/services/api'
import { LoadingState, ErrorState } from '@/components/ViewState'
import { EmptyState } from '@/components/ui/empty-state'
import { ContentContainer } from '@/components/ui/content-container'
import { CheckCircle, ShieldAlert } from 'lucide-react'
import { cn } from '@/lib/utils'

interface GapSectionProps {
  title: string
  items: string[]
  color: 'amber' | 'red' | 'orange' | 'gray'
  description: string
}

const colorMap = {
  amber: { badge: 'bg-amber-100 text-amber-800 border-amber-200', dot: 'bg-amber-500' },
  red:   { badge: 'bg-red-100 text-red-800 border-red-200',       dot: 'bg-red-500'   },
  orange:{ badge: 'bg-orange-100 text-orange-800 border-orange-200', dot: 'bg-orange-500' },
  gray:  { badge: 'bg-gray-100 text-gray-700 border-gray-200',    dot: 'bg-gray-400'  },
}

function GapSection({ title, items, color, description }: GapSectionProps) {
  const c = colorMap[color]
  return (
    <div className="rounded-lg border border-border bg-card overflow-hidden">
      <div className={cn('flex items-center gap-2 px-4 py-2.5 border-b border-border', c.badge)}>
        <span className={cn('w-2 h-2 rounded-full shrink-0', c.dot)} />
        <span className="text-sm font-semibold">{title}</span>
        <span className="ml-auto text-xs font-medium">{items.length}</span>
      </div>
      <div className="px-4 py-3">
        {items.length === 0 ? (
          <p className="text-sm text-muted-foreground flex items-center gap-1.5">
            <CheckCircle className="w-4 h-4 text-green-500 shrink-0" />
            None — all clear ✓
          </p>
        ) : (
          <>
            <p className="text-xs text-muted-foreground mb-2">{description}</p>
            <ul className="flex flex-wrap gap-1.5">
              {items.map(item => (
                <li key={item} className="text-xs font-mono rounded-md px-2.5 py-1 bg-muted border border-border text-foreground">
                  {item}
                </li>
              ))}
            </ul>
          </>
        )}
      </div>
    </div>
  )
}

export function GapsTab() {
  const { modelId } = useModel()
  const { data, isLoading, error } = useQuery({
    queryKey: ['gaps', modelId],
    queryFn: () => viewsApi.getGaps(modelId!),
    enabled: !!modelId,
  })

  if (isLoading) return <LoadingState />
  if (error) return <ErrorState message={(error as Error).message} />
  if (!data) return null

  const totalGaps = data.unmapped_needs.length + data.unrealized_capabilities.length +
    data.unowned_services.length + data.unneeded_capabilities.length + data.orphan_services.length

  if (totalGaps === 0) {
    return (
      <ContentContainer>
        <EmptyState
          title="No gaps detected"
          description="All needs are mapped, all leaf capabilities are realized, all services are owned and realizing capabilities. Good architecture hygiene."
          icon={<CheckCircle className="w-12 h-12 text-green-500" />}
        />
      </ContentContainer>
    )
  }

  return (
    <ContentContainer className="space-y-4">
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <ShieldAlert className="w-4 h-4 text-amber-500 shrink-0" />
        <span>{totalGaps} structural gap{totalGaps !== 1 ? 's' : ''} detected across the model</span>
      </div>
      <GapSection
        title="Unmapped Needs"
        items={data.unmapped_needs}
        color="amber"
        description="Needs with no SupportedBy relationships — not linked to any capability."
      />
      <GapSection
        title="Unrealized Capabilities"
        items={data.unrealized_capabilities}
        color="red"
        description="Leaf capabilities with no services realizing them."
      />
      <GapSection
        title="Unowned Services"
        items={data.unowned_services}
        color="red"
        description="Services with no team assigned as owner."
      />
      <GapSection
        title="Unneeded Capabilities"
        items={data.unneeded_capabilities}
        color="gray"
        description="Capabilities not referenced by any need's SupportedBy (directly or via ancestry)."
      />
      <GapSection
        title="Orphan Services"
        items={data.orphan_services}
        color="orange"
        description="Services that do not realize any capability."
      />
    </ContentContainer>
  )
}
