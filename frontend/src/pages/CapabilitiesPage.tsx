import { ModelRequired } from '@/components/ui/ModelRequired'
import { useActiveTab } from '@/components/ui/url-tab-bar'
import { useRegisterTabs } from '@/lib/page-tabs-context'
import { CapabilityView } from '@/pages/views/CapabilityView'
import { RealizationView } from '@/pages/views/RealizationView'
import { DependenciesTab } from '@/features/capabilities/DependenciesTab'

const TABS = [
  { id: 'hierarchy', label: 'Hierarchy' },
  { id: 'services', label: 'Services' },
  { id: 'dependencies', label: 'Dependencies' },
]

export function CapabilitiesPage() {
  useRegisterTabs(TABS)
  const activeTab = useActiveTab(TABS)

  return (
    <ModelRequired>
      <div className="[&_.page-header-root]:hidden">
        {activeTab === 'hierarchy' && <CapabilityView />}
        {activeTab === 'services' && <RealizationView />}
        {activeTab === 'dependencies' && <DependenciesTab />}
      </div>
    </ModelRequired>
  )
}
