import { ModelRequired } from '@/components/ui/ModelRequired'
import { UrlTabBar, useActiveTab } from '@/components/ui/url-tab-bar'
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
  const activeTab = useActiveTab(TABS)

  return (
    <ModelRequired>
      <UrlTabBar tabs={TABS} />
      {activeTab === 'topology' && <TeamTopologyView />}
      {activeTab === 'ownership' && <OwnershipView />}
      {activeTab === 'cognitive-load' && <CognitiveLoadView />}
      {activeTab === 'interactions' && <InteractionsTab />}
    </ModelRequired>
  )
}
