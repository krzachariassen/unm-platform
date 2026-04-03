import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { TopBar } from './TopBar'
import { AdvisorPanel } from '@/components/advisor/AdvisorPanel'
import { InsightsProvider } from '@/lib/InsightsContext'
import { PageTabsProvider } from '@/lib/page-tabs-context'
import { useAIEnabled } from '@/hooks/useAIEnabled'

export function AppShell() {
  const aiEnabled = useAIEnabled()

  const shell = (
    <div className="flex h-screen w-screen overflow-hidden bg-background text-foreground">
      <Sidebar />
      <div className="flex flex-col flex-1 overflow-hidden">
        <PageTabsProvider>
          <TopBar />
          <main className="flex-1 overflow-auto p-6">
            <Outlet />
          </main>
        </PageTabsProvider>
      </div>
      {aiEnabled && <AdvisorPanel />}
    </div>
  )

  return aiEnabled ? <InsightsProvider>{shell}</InsightsProvider> : shell
}
