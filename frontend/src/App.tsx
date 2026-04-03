import { Routes, Route, Navigate } from 'react-router-dom'
import { ModelProvider } from '@/lib/model-context'
import { SearchProvider } from '@/lib/search-context'
import { ChangesetProvider } from '@/lib/changeset-context'
import { AppShell } from '@/components/layout/AppShell'
import { UploadPage } from '@/pages/UploadPage'
import { DashboardPage } from '@/pages/DashboardPage'
import { NeedsPage } from '@/pages/NeedsPage'
import { CapabilitiesPage } from '@/pages/CapabilitiesPage'
import { TeamsPage } from '@/pages/TeamsPage'
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
              {/* Core pages */}
              <Route path="/" element={<UploadPage />} />
              <Route path="/models" element={<ModelsPage />} />
              <Route path="/dashboard" element={<DashboardPage />} />
              <Route path="/history" element={<ModelHistoryPage />} />

              {/* Architecture views */}
              <Route path="/unm-map" element={<UNMMapView />} />
              <Route path="/needs" element={<NeedsPage />} />
              <Route path="/capabilities" element={<CapabilitiesPage />} />
              <Route path="/teams" element={<TeamsPage />} />
              <Route path="/signals" element={<SignalsView />} />

              {/* Editing */}
              <Route path="/what-if" element={<WhatIfPage />} />

              {/* AI */}
              <Route path="/recommendations" element={<RecommendationsPage />} />
              <Route path="/advisor" element={<AdvisorPage />} />

              {/* Backward-compat redirects for old routes */}
              <Route path="/need" element={<Navigate to="/needs" replace />} />
              <Route path="/capability" element={<Navigate to="/capabilities" replace />} />
              <Route path="/realization" element={<Navigate to="/needs?tab=traceability" replace />} />
              <Route path="/ownership" element={<Navigate to="/teams?tab=ownership" replace />} />
              <Route path="/team-topology" element={<Navigate to="/teams" replace />} />
              <Route path="/cognitive-load" element={<Navigate to="/teams?tab=cognitive-load" replace />} />
              <Route path="/edit" element={<Navigate to="/unm-map" replace />} />

              <Route path="*" element={<Navigate to="/" replace />} />
            </Route>
          </Routes>
        </SearchProvider>
      </ModelProvider>
    </ChangesetProvider>
  )
}
