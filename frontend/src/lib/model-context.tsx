import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react'
import type { ParseResponse } from './api'

interface ModelContextValue {
  modelId: string | null
  parseResult: ParseResponse | null
  loadedAt: Date | null
  isHydrating: boolean
  setModel: (id: string, result: ParseResponse) => void
  clearModel: () => void
}

const ModelContext = createContext<ModelContextValue | null>(null)

const LS_MODEL_ID = 'unm_model_id'
const LS_PARSE_RESULT = 'unm_parse_result'
const LS_LOADED_AT = 'unm_loaded_at'

export function ModelProvider({ children }: { children: ReactNode }) {
  const [modelId, setModelId] = useState<string | null>(null)
  const [parseResult, setParseResult] = useState<ParseResponse | null>(null)
  const [loadedAt, setLoadedAt] = useState<Date | null>(null)
  const [isHydrating, setIsHydrating] = useState(true)

  // On mount: restore from localStorage if available, then verify the model
  // still exists on the backend. If the backend returns 404, the model is
  // stale (e.g. server restarted) — clear localStorage so the user sees a
  // clean empty state instead of rendering stale data.
  useEffect(() => {
    const storedId = localStorage.getItem(LS_MODEL_ID)
    const storedResult = localStorage.getItem(LS_PARSE_RESULT)
    const storedLoadedAt = localStorage.getItem(LS_LOADED_AT)

    if (!storedId || !storedResult) {
      // Nothing stored — nothing to verify.
      setIsHydrating(false)
      return
    }

    let parsed: ParseResponse
    try {
      parsed = JSON.parse(storedResult) as ParseResponse
    } catch {
      // Corrupt localStorage entry — clear and bail out.
      localStorage.removeItem(LS_MODEL_ID)
      localStorage.removeItem(LS_PARSE_RESULT)
      localStorage.removeItem(LS_LOADED_AT)
      setIsHydrating(false)
      return
    }

    // Verify the model still exists on the backend before applying state.
    // Use a direct fetch (not api.ts) to avoid circular import issues.
    fetch(`/api/models/${storedId}/actors`)
      .then((res) => {
        if (res.status === 404) {
          // Model is gone — clear stale localStorage entry.
          localStorage.removeItem(LS_MODEL_ID)
          localStorage.removeItem(LS_PARSE_RESULT)
          localStorage.removeItem(LS_LOADED_AT)
        } else {
          // Model exists (any non-404 response, including server errors).
          // Trust the cached data and restore state.
          setModelId(storedId)
          setParseResult(parsed)
          setLoadedAt(storedLoadedAt ? new Date(storedLoadedAt) : null)
        }
      })
      .catch(() => {
        // Network error (e.g. backend unreachable) — keep the cached model
        // so the user isn't unexpectedly logged out on a flaky connection.
        setModelId(storedId)
        setParseResult(parsed)
        setLoadedAt(storedLoadedAt ? new Date(storedLoadedAt) : null)
      })
      .finally(() => {
        setIsHydrating(false)
      })
  }, [])

  const setModel = useCallback((id: string, result: ParseResponse) => {
    const now = new Date()
    setModelId(id)
    setParseResult(result)
    setLoadedAt(now)
    localStorage.setItem(LS_MODEL_ID, id)
    localStorage.setItem(LS_PARSE_RESULT, JSON.stringify(result))
    localStorage.setItem(LS_LOADED_AT, now.toISOString())
  }, [])

  const clearModel = useCallback(() => {
    setModelId(null)
    setParseResult(null)
    setLoadedAt(null)
    localStorage.removeItem(LS_MODEL_ID)
    localStorage.removeItem(LS_PARSE_RESULT)
    localStorage.removeItem(LS_LOADED_AT)
  }, [])

  return (
    <ModelContext.Provider value={{ modelId, parseResult, loadedAt, isHydrating, setModel, clearModel }}>
      {children}
    </ModelContext.Provider>
  )
}

export function useModel() {
  const ctx = useContext(ModelContext)
  if (!ctx) throw new Error('useModel must be used within ModelProvider')
  return ctx
}

/**
 * Use in protected pages that require a loaded model.
 * Returns model state. Pages should use <ModelRequired> to handle the guard UI
 * (spinner during hydration, empty state when no model is loaded).
 */
export function useRequireModel() {
  const { modelId, parseResult, loadedAt, isHydrating, setModel, clearModel } = useModel()

  return { modelId, parseResult, loadedAt, isHydrating, setModel, clearModel }
}
