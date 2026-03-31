import { createContext, useContext, useState, useCallback } from 'react'
import type { ReactNode } from 'react'
import type { ChangeAction } from './api'

interface ChangesetContextValue {
  actions: ChangeAction[]
  description: string
  refreshKey: number
  addAction(a: ChangeAction): void
  removeAction(index: number): void
  clearActions(): void
  setDescription(d: string): void
  discardAll(): void
}

const ChangesetContext = createContext<ChangesetContextValue | null>(null)

export function ChangesetProvider({ children }: { children: ReactNode }) {
  const [actions, setActions] = useState<ChangeAction[]>([])
  const [description, setDescription] = useState('')
  const [refreshKey, setRefreshKey] = useState(0)

  const addAction = useCallback((a: ChangeAction) => {
    setActions(prev => [...prev, a])
  }, [])

  const removeAction = useCallback((index: number) => {
    setActions(prev => prev.filter((_, i) => i !== index))
  }, [])

  const clearActions = useCallback(() => setActions([]), [])

  const discardAll = useCallback(() => {
    setActions([])
    setDescription('')
    setRefreshKey(k => k + 1)
  }, [])

  return (
    <ChangesetContext.Provider value={{
      actions, description, refreshKey,
      addAction, removeAction, clearActions, setDescription, discardAll,
    }}>
      {children}
    </ChangesetContext.Provider>
  )
}

export function useChangeset() {
  const ctx = useContext(ChangesetContext)
  if (!ctx) throw new Error('useChangeset must be used within ChangesetProvider')
  return ctx
}
