import { Routes, Route, Navigate } from 'react-router-dom'
import { ModelProvider } from '@/lib/model-context'
import { SearchProvider } from '@/lib/search-context'
import { ChangesetProvider } from '@/lib/changeset-context'
import { AppShell } from '@/components/layout/AppShell'
import { UploadPage } from '@/pages/UploadPage'
import { DashboardPage } from '@/pages/DashboardPage'
import { NeedView } from '@/pages/views/NeedView'
import { CapabilityView } from '@/pages/views/CapabilityView'
import { RealizationView } from '@/pages/views/RealizationView'
import { OwnershipView } from '@/pages/views/OwnershipView'
import { TeamTopologyView } from '@/pages/views/TeamTopologyView'
import { CognitiveLoadView } from '@/pages/views/CognitiveLoadView'
import { UNMMapView } from '@/pages/views/UNMMapView'
import { SignalsView } from '@/pages/views/SignalsView'
import { WhatIfPage } from '@/pages/WhatIfPage'
import { AdvisorPage } from '@/pages/AdvisorPage'
import { RecommendationsPage } from '@/pages/RecommendationsPage'
import { ModelsPage } from '@/pages/ModelsPage'
import { ModelHistoryPage } from '@/pages/ModelHistoryPage'

export default function App() {
  return (
    <ChangesetProvider>
      <ModelProvider>
        <SearchProvider>
          <Routes>
            <Route element={<AppShell />}>
              <Route path="/" element={<UploadPage />} />
              <Route path="/models" element={<ModelsPage />} />
              <Route path="/dashboard" element={<DashboardPage />} />
              <Route path="/history" element={<ModelHistoryPage />} />
              <Route path="/unm-map" element={<UNMMapView />} />
              <Route path="/need" element={<NeedView />} />
              <Route path="/capability" element={<CapabilityView />} />
              <Route path="/realization" element={<RealizationView />} />
              <Route path="/ownership" element={<OwnershipView />} />
              <Route path="/team-topology" element={<TeamTopologyView />} />
              <Route path="/cognitive-load" element={<CognitiveLoadView />} />
              <Route path="/signals" element={<SignalsView />} />
              <Route path="/edit" element={<Navigate to="/unm-map" replace />} />
              <Route path="/what-if" element={<WhatIfPage />} />
              <Route path="/recommendations" element={<RecommendationsPage />} />
              <Route path="/advisor" element={<AdvisorPage />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </Route>
          </Routes>
        </SearchProvider>
      </ModelProvider>
    </ChangesetProvider>
  )
}
