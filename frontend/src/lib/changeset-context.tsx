import { createContext, useContext, useState, useCallback } from 'react'
import type { ReactNode } from 'react'
import type { ChangeAction } from './api'

interface ChangesetContextValue {
  isEditMode: boolean
  actions: ChangeAction[]
  description: string
  refreshKey: number
  enterEditMode(): void
  exitEditMode(): void
  addAction(a: ChangeAction): void
  removeAction(index: number): void
  clearActions(): void
  setDescription(d: string): void
}

const ChangesetContext = createContext<ChangesetContextValue | null>(null)

export function ChangesetProvider({ children }: { children: ReactNode }) {
  const [isEditMode, setIsEditMode] = useState(false)
  const [actions, setActions] = useState<ChangeAction[]>([])
  const [description, setDescription] = useState('')
  const [refreshKey, setRefreshKey] = useState(0)

  const enterEditMode = useCallback(() => setIsEditMode(true), [])

  const exitEditMode = useCallback(() => {
    setIsEditMode(false)
    setActions([])
    setDescription('')
    setRefreshKey(k => k + 1)
  }, [])

  const addAction = useCallback((a: ChangeAction) => {
    setActions(prev => [...prev, a])
  }, [])

  const removeAction = useCallback((index: number) => {
    setActions(prev => prev.filter((_, i) => i !== index))
  }, [])

  const clearActions = useCallback(() => setActions([]), [])

  return (
    <ChangesetContext.Provider value={{
      isEditMode,
      actions,
      description,
      refreshKey,
      enterEditMode,
      exitEditMode,
      addAction,
      removeAction,
      clearActions,
      setDescription,
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
