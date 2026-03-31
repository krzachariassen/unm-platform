import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { useModel } from '@/lib/model-context'

/**
 * TanStack Query wrapper for view data fetching.
 * Replaces the old useModelView hook pattern.
 *
 * Usage:
 *   const { data, isLoading, error } = useViewQuery(
 *     ['needView', modelId],
 *     (id) => viewsApi.getNeedView(id)
 *   )
 */
export function useViewQuery<T>(
  queryKey: readonly unknown[],
  fetcher: (modelId: string) => Promise<T>
): UseQueryResult<T, Error> {
  const { modelId } = useModel()
  return useQuery({
    queryKey,
    queryFn: () => fetcher(modelId!),
    enabled: !!modelId,
  })
}
