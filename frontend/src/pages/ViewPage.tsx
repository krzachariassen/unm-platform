import { useParams } from 'react-router-dom'
import { useRequireModel } from '@/lib/model-context'

export function ViewPage() {
  const { viewType } = useParams<{ viewType: string }>()
  useRequireModel()

  return (
    <div className="h-full flex items-center justify-center text-muted-foreground">
      <div className="text-center">
        <p className="text-lg font-medium">{viewType} view</p>
        <p className="text-sm mt-1">Coming soon</p>
      </div>
    </div>
  )
}
