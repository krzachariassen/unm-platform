import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { Database, GitBranch, Loader2, Plus, Calendar } from 'lucide-react'
import { useModel } from '@/lib/model-context'
import { modelsApi } from '@/services/api'
import { PageHeader } from '@/components/ui/page-header'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import type { ModelListItem } from '@/types/model'

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}

function ModelCard({ item, onLoad, isLoading }: {
  item: ModelListItem
  onLoad: () => void
  isLoading: boolean
}) {
  return (
    <div className="rounded-lg border border-border bg-card p-4 flex flex-col gap-3">
      <div className="flex items-start justify-between gap-2">
        <div className="flex items-center gap-2.5 min-w-0">
          <Database className="w-4 h-4 text-muted-foreground shrink-0" />
          <span className="text-sm font-semibold text-foreground truncate">{item.name || 'Untitled'}</span>
        </div>
        <span className="flex items-center gap-1 px-2 py-0.5 rounded-full bg-muted text-[10px] font-medium text-muted-foreground shrink-0">
          <GitBranch className="w-3 h-3" />
          {item.version_count}
        </span>
      </div>
      <div className="flex items-center gap-1 text-[11px] text-muted-foreground">
        <Calendar className="w-3 h-3" />
        {formatDate(item.created_at)}
      </div>
      <Button
        size="sm"
        className="w-full"
        onClick={onLoad}
        disabled={isLoading}
      >
        {isLoading ? <Loader2 className="w-3 h-3 animate-spin mr-1" /> : null}
        Load
      </Button>
    </div>
  )
}

export function ModelsPage() {
  const { setModel } = useModel()
  const navigate = useNavigate()

  const { data, isLoading, error } = useQuery({
    queryKey: ['modelList'],
    queryFn: () => modelsApi.listModels(),
  })

  async function handleLoad(item: ModelListItem) {
    const result = await modelsApi.loadStoredModel(item.id, item.version_count)
    setModel(item.id, result)
    navigate('/dashboard')
  }

  return (
    <div className="max-w-screen-xl mx-auto">
      <PageHeader
        title="Models"
        description="All stored architecture models"
        actions={
          <Button size="sm" variant="outline" onClick={() => navigate('/')}>
            <Plus className="w-3.5 h-3.5 mr-1.5" />
            Upload New
          </Button>
        }
      />

      {isLoading && (
        <div className="flex items-center gap-3 py-12 justify-center text-muted-foreground">
          <Loader2 className="w-4 h-4 animate-spin" />
          <span className="text-sm">Loading models…</span>
        </div>
      )}

      {error && (
        <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          Failed to load models: {(error as Error).message}
        </div>
      )}

      {data && data.models.length === 0 && (
        <div className={cn(
          'flex flex-col items-center justify-center py-16 gap-3',
          'rounded-lg border border-dashed border-border text-center mt-4'
        )}>
          <Database className="w-8 h-8 text-muted-foreground/40" />
          <p className="text-sm font-medium text-muted-foreground">No models stored yet</p>
          <p className="text-xs text-muted-foreground/70">Upload a model to get started</p>
          <Button size="sm" variant="outline" onClick={() => navigate('/')}>
            <Plus className="w-3.5 h-3.5 mr-1.5" /> Upload Model
          </Button>
        </div>
      )}

      {data && data.models.length > 0 && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 mt-4">
          {data.models.map(item => (
            <ModelCard
              key={item.id}
              item={item}
              onLoad={() => handleLoad(item)}
              isLoading={false}
            />
          ))}
        </div>
      )}
    </div>
  )
}
