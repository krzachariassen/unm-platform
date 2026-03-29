import { useEffect, useState } from 'react'
import { useRequireModel } from '@/lib/model-context'

export interface ModelViewState<T> {
  data: T | null
  loading: boolean
  error: string | null
}

/**
 * Shared hook for the repeated fetch+loading+error pattern in every view.
 *
 * Handles:
 * - Waiting for model hydration before fetching
 * - Loading state (starts true, set false when fetch resolves or rejects)
 * - Error state (set to message string on fetch failure)
 * - Data state (null until fetch succeeds)
 *
 * Usage:
 *   const { data, loading, error } = useModelView(modelId, (id) => api.getNeedView(id))
 *
 * Note: Views that need a custom callback (e.g. UNMMapView that reloads after edits)
 * should manage their own state rather than using this hook.
 */
export function useModelView<T>(
  fetcher: (id: string) => Promise<T>
): ModelViewState<T> {
  const { modelId, isHydrating } = useRequireModel()
  const [data, setData] = useState<T | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (isHydrating || !modelId) return
    setLoading(true)
    setError(null)
    fetcher(modelId)
      .then(setData)
      .catch((e: unknown) => setError((e as Error).message))
      .finally(() => setLoading(false))
    // fetcher is intentionally excluded from deps — callers pass a stable api.* method
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isHydrating, modelId])

  return { data, loading, error }
}
