import { useState, useEffect } from 'react'
import { getRuntimeConfig } from '@/lib/runtimeConfig'

let resolved: boolean | null = null

export function useAIEnabled(): boolean {
  const [enabled, setEnabled] = useState(resolved ?? false)

  useEffect(() => {
    if (resolved !== null) return
    getRuntimeConfig().then(cfg => {
      resolved = cfg.ai?.enabled ?? false
      setEnabled(resolved)
    })
  }, [])

  return enabled
}
