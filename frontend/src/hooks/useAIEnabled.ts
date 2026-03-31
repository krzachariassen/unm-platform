import { useQuery } from '@tanstack/react-query'
import { getRuntimeConfig } from '@/lib/runtimeConfig'

export function useAIEnabled(): boolean {
  const { data } = useQuery({
    queryKey: ['runtimeConfig'],
    queryFn: getRuntimeConfig,
    staleTime: Infinity,
  })
  return data?.ai?.enabled ?? false
}
