import { useState, useEffect, useRef } from 'react'
import { useInsightsContext } from '@/lib/InsightsContext'
import { type InsightItem } from '@/lib/api'
import { useAIEnabled } from '@/hooks/useAIEnabled'

interface InsightsResponseWithStatus {
  insights: Record<string, InsightItem>
  ai_configured: boolean
  status?: string
  error?: string
}

const POLL_INTERVAL_MS = 3000
const TIMEOUT_MS = 10000

export type InsightStatus = 'idle' | 'loading' | 'loaded' | 'error' | 'unavailable'

export function usePageInsights(domain: string): {
  insights: Record<string, InsightItem>
  loading: boolean
  aiError: boolean
  status: InsightStatus
} {
  const aiEnabled = useAIEnabled()
  const { getInsights } = useInsightsContext()
  const [insights, setInsights] = useState<Record<string, InsightItem>>({})
  const [loading, setLoading] = useState(false)
  const [aiError, setAiError] = useState(false)
  const [status, setStatus] = useState<InsightStatus>('idle')
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    if (!aiEnabled) {
      setStatus('unavailable')
      setLoading(false)
      return
    }

    let cancelled = false

    const clearPoll = () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current)
        timerRef.current = null
      }
    }

    const clearTimeoutTimer = () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
        timeoutRef.current = null
      }
    }

    const fetchInsight = () => {
      setLoading(true)
      setAiError(false)
      setStatus('loading')
      getInsights(domain)
        .then(resp => {
          if (cancelled) return
          const r = resp as InsightsResponseWithStatus
          setInsights(r.insights)

          if (r.status === 'computing') {
            timerRef.current = setTimeout(fetchInsight, POLL_INTERVAL_MS)
            return
          }

          clearTimeoutTimer()
          setLoading(false)
          if (r.error === 'ai_unavailable' || r.error === 'ai_parse_error') {
            setAiError(true)
            setStatus('error')
          } else {
            setStatus('loaded')
          }
        })
        .catch(() => {
          if (cancelled) return
          clearTimeoutTimer()
          setInsights({})
          setAiError(true)
          setLoading(false)
          setStatus('error')
        })
    }

    // Set a global timeout — if still loading after TIMEOUT_MS, force error state
    timeoutRef.current = setTimeout(() => {
      if (cancelled) return
      clearPoll()
      setLoading(false)
      setAiError(true)
      setStatus('error')
    }, TIMEOUT_MS)

    fetchInsight()

    return () => {
      cancelled = true
      clearPoll()
      clearTimeoutTimer()
    }
  }, [domain, getInsights, aiEnabled])

  return { insights, loading, aiError, status }
}
