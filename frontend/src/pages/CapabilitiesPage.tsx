import { ModelRequired } from '@/components/ui/ModelRequired'
import { UrlTabBar, useActiveTab } from '@/components/ui/url-tab-bar'
import { CapabilityView } from '@/pages/views/CapabilityView'
import { RealizationView } from '@/pages/views/RealizationView'
import { DependenciesTab } from '@/features/capabilities/DependenciesTab'

const TABS = [
  { id: 'hierarchy', label: 'Hierarchy' },
  { id: 'services', label: 'Services' },
  { id: 'dependencies', label: 'Dependencies' },
]

export function CapabilitiesPage() {
  const activeTab = useActiveTab(TABS)

  return (
    <ModelRequired>
      <UrlTabBar tabs={TABS} />
      {activeTab === 'hierarchy' && <CapabilityView />}
      {activeTab === 'services' && <RealizationView />}
      {activeTab === 'dependencies' && <DependenciesTab />}
    </ModelRequired>
  )
}
