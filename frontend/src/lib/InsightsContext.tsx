import { createContext, useContext, useRef, useCallback, useEffect, type ReactNode } from 'react'
import { api, type InsightsResponse } from './api'
import { useModel } from './model-context'

interface InsightsContextValue {
  getInsights: (domain: string) => Promise<InsightsResponse>
}

const InsightsContext = createContext<InsightsContextValue | null>(null)

export function InsightsProvider({ children }: { children: ReactNode }) {
  const { modelId } = useModel()
  // Cache: key = "${modelId}:${domain}" → completed response
  const cache = useRef<Map<string, InsightsResponse>>(new Map())
  // In-flight: key = "${modelId}:${domain}" → pending promise (deduplicates concurrent callers)
  const inflight = useRef<Map<string, Promise<InsightsResponse>>>(new Map())

  // Invalidate both caches when modelId changes
  useEffect(() => {
    cache.current.clear()
    inflight.current.clear()
  }, [modelId])

  const getInsights = useCallback(async (domain: string): Promise<InsightsResponse> => {
    if (!modelId) return { domain, insights: {}, ai_configured: false }
    const key = `${modelId}:${domain}`
    if (cache.current.has(key)) return cache.current.get(key)!
    if (inflight.current.has(key)) return inflight.current.get(key)!
    const promise = api.getInsights(modelId, domain).then(result => {
      const resp = result as InsightsResponse & { status?: string }
      const hasInsights = resp.insights && Object.keys(resp.insights).length > 0
      const isReady = !resp.status || resp.status === 'ready' || resp.status === 'failed'
      if (hasInsights || isReady) {
        cache.current.set(key, result)
      }
      inflight.current.delete(key)
      return result
    }).catch(err => {
      inflight.current.delete(key)
      throw err
    })
    inflight.current.set(key, promise)
    return promise
  }, [modelId])

  return (
    <InsightsContext.Provider value={{ getInsights }}>
      {children}
    </InsightsContext.Provider>
  )
}

const noopInsights: InsightsContextValue = {
  getInsights: async (domain: string) => ({ domain, insights: {}, ai_configured: false }),
}

export function useInsightsContext() {
  const ctx = useContext(InsightsContext)
  return ctx ?? noopInsights
}
