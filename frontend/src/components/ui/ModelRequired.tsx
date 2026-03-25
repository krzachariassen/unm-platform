import { useEffect, type ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { useModel } from '@/lib/model-context'

interface ModelRequiredProps {
  children: ReactNode
}

/**
 * Guard component for pages that require a loaded model.
 * - While hydrating: renders nothing (avoids flash before localStorage is read).
 * - After hydration, when no model is loaded: redirects to "/" (Upload page).
 * - When a model is present: renders children.
 *
 * Use this at the top level of every protected page so the guard is consistent.
 */
export function ModelRequired({ children }: ModelRequiredProps) {
  const { modelId, isHydrating } = useModel()
  const navigate = useNavigate()

  useEffect(() => {
    if (!isHydrating && !modelId) {
      navigate('/', { replace: true })
    }
  }, [isHydrating, modelId, navigate])

  if (isHydrating || !modelId) return null

  return <>{children}</>
}
