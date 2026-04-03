import { ModelRequired } from '@/components/ui/ModelRequired'
import { useActiveTab } from '@/components/ui/url-tab-bar'
import { useRegisterTabs } from '@/lib/page-tabs-context'
import { NeedView } from '@/pages/views/NeedView'
import { RealizationView } from '@/pages/views/RealizationView'
import { GapsTab } from '@/features/needs/GapsTab'

const TABS = [
  { id: 'overview', label: 'Overview' },
  { id: 'traceability', label: 'Traceability' },
  { id: 'gaps', label: 'Gaps' },
]

export function NeedsPage() {
  useRegisterTabs(TABS)
  const activeTab = useActiveTab(TABS)

  return (
    <ModelRequired>
      <div className="[&_.page-header-root]:hidden">
        {activeTab === 'overview' && <NeedView />}
        {activeTab === 'traceability' && <RealizationView />}
        {activeTab === 'gaps' && <GapsTab />}
      </div>
    </ModelRequired>
  )
}
