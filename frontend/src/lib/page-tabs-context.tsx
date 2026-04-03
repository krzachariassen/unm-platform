import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import type { UrlTabItem } from '@/components/ui/url-tab-bar'

interface PageTabsContextValue {
  tabs: UrlTabItem[]
  setTabs: (tabs: UrlTabItem[]) => void
}

const PageTabsContext = createContext<PageTabsContextValue>({
  tabs: [],
  setTabs: () => {},
})

export function PageTabsProvider({ children }: { children: ReactNode }) {
  const [tabs, setTabs] = useState<UrlTabItem[]>([])
  return <PageTabsContext.Provider value={{ tabs, setTabs }}>{children}</PageTabsContext.Provider>
}

export function usePageTabs() {
  return useContext(PageTabsContext)
}

/** Register tabs for the current page. Cleared automatically on unmount. */
export function useRegisterTabs(tabs: UrlTabItem[]) {
  const { setTabs } = usePageTabs()
  useEffect(() => {
    setTabs(tabs)
    return () => setTabs([])
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])
}
