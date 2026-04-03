import { ModelRequired } from '@/components/ui/ModelRequired'
import { useActiveTab } from '@/components/ui/url-tab-bar'
import { useRegisterTabs } from '@/lib/page-tabs-context'
import { TeamTopologyView } from '@/pages/views/TeamTopologyView'
import { OwnershipView } from '@/pages/views/OwnershipView'
import { CognitiveLoadView } from '@/pages/views/CognitiveLoadView'
import { InteractionsTab } from '@/features/teams/InteractionsTab'

const TABS = [
  { id: 'topology', label: 'Topology' },
  { id: 'ownership', label: 'Ownership' },
  { id: 'cognitive-load', label: 'Cognitive Load' },
  { id: 'interactions', label: 'Interactions' },
]

export function TeamsPage() {
  useRegisterTabs(TABS)
  const activeTab = useActiveTab(TABS)

  return (
    <ModelRequired>
      <div className="[&_.page-header-root]:hidden">
        {activeTab === 'topology' && <TeamTopologyView />}
        {activeTab === 'ownership' && <OwnershipView />}
        {activeTab === 'cognitive-load' && <CognitiveLoadView />}
        {activeTab === 'interactions' && <InteractionsTab />}
      </div>
    </ModelRequired>
  )
}
