import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { TopBar } from './TopBar'
import { AdvisorPanel } from '@/components/advisor/AdvisorPanel'
import { PendingChangesBar } from '@/components/changeset/PendingChangesBar'
import { InsightsProvider } from '@/lib/InsightsContext'
import { useAIEnabled } from '@/hooks/useAIEnabled'
import { useChangeset } from '@/lib/changeset-context'

export function AppShell() {
  const aiEnabled = useAIEnabled()
  const { isEditMode, actions } = useChangeset()
  const showBar = isEditMode

  const shell = (
    <div className="flex h-screen w-screen overflow-hidden bg-background text-foreground">
      <Sidebar />
      <div className="flex flex-col flex-1 overflow-hidden">
        <TopBar />
        <main
          className="flex-1 overflow-auto p-6"
          style={showBar && actions.length > 0 ? { paddingBottom: 72 } : undefined}
        >
          <Outlet />
        </main>
      </div>
      {aiEnabled && <AdvisorPanel />}
      <PendingChangesBar />
    </div>
  )

  return aiEnabled ? <InsightsProvider>{shell}</InsightsProvider> : shell
}
