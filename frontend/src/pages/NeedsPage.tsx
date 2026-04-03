import { ModelRequired } from '@/components/ui/ModelRequired'
import { UrlTabBar, useActiveTab } from '@/components/ui/url-tab-bar'
import { NeedView } from '@/pages/views/NeedView'
import { RealizationView } from '@/pages/views/RealizationView'

const TABS = [
  { id: 'overview', label: 'Overview' },
  { id: 'traceability', label: 'Traceability' },
]

export function NeedsPage() {
  const activeTab = useActiveTab(TABS)

  return (
    <ModelRequired>
      <div className="px-6 pt-4">
        <UrlTabBar tabs={TABS} />
      </div>
      {activeTab === 'overview' && <NeedView />}
      {activeTab === 'traceability' && <RealizationView />}
    </ModelRequired>
  )
}
