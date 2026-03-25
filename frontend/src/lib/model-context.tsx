import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
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

  // On mount: restore from localStorage if available.
  useEffect(() => {
    const storedId = localStorage.getItem(LS_MODEL_ID)
    const storedResult = localStorage.getItem(LS_PARSE_RESULT)
    const storedLoadedAt = localStorage.getItem(LS_LOADED_AT)
    if (storedId && storedResult) {
      try {
        const parsed = JSON.parse(storedResult) as ParseResponse
        setModelId(storedId)
        setParseResult(parsed)
        setLoadedAt(storedLoadedAt ? new Date(storedLoadedAt) : null)
      } catch {
        localStorage.removeItem(LS_MODEL_ID)
        localStorage.removeItem(LS_PARSE_RESULT)
        localStorage.removeItem(LS_LOADED_AT)
      }
    }
    setIsHydrating(false)
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
 * Returns model state. Redirects to "/" after hydration if no model is loaded.
 * During hydration, returns isHydrating=true so the page can show a spinner
 * instead of triggering a premature redirect.
 */
export function useRequireModel() {
  const { modelId, parseResult, loadedAt, isHydrating, setModel, clearModel } = useModel()
  const navigate = useNavigate()

  useEffect(() => {
    if (!isHydrating && !modelId) {
      navigate('/')
    }
  }, [isHydrating, modelId, navigate])

  return { modelId, parseResult, loadedAt, isHydrating, setModel, clearModel }
}
