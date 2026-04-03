import { Routes, Route, Navigate } from 'react-router-dom'
import { ModelProvider } from '@/lib/model-context'
import { SearchProvider } from '@/lib/search-context'
import { ChangesetProvider } from '@/lib/changeset-context'
import { AuthProvider, useAuth } from '@/lib/auth-context'
import { AppShell } from '@/components/layout/AppShell'
import { LoginPage } from '@/pages/LoginPage'
import { UploadPage } from '@/pages/UploadPage'
import { DashboardPage } from '@/pages/DashboardPage'
import { NeedView } from '@/pages/views/NeedView'
import { RealizationView } from '@/pages/views/RealizationView'
import { GapsTab } from '@/features/needs/GapsTab'
import { CapabilityView } from '@/pages/views/CapabilityView'
import { DependenciesTab } from '@/features/capabilities/DependenciesTab'
import { TeamTopologyView } from '@/pages/views/TeamTopologyView'
import { OwnershipView } from '@/pages/views/OwnershipView'
import { CognitiveLoadView } from '@/pages/views/CognitiveLoadView'
import { InteractionsTab } from '@/features/teams/InteractionsTab'
import { UNMMapView } from '@/pages/views/UNMMapView'
import { SignalsView } from '@/pages/views/SignalsView'
import { WhatIfPage } from '@/pages/WhatIfPage'
import { AdvisorPage } from '@/pages/AdvisorPage'
import { RecommendationsPage } from '@/pages/RecommendationsPage'
import { ModelsPage } from '@/pages/ModelsPage'
import { ModelHistoryPage } from '@/pages/ModelHistoryPage'

// ProtectedRoute renders children when the user is authenticated.
// While loading, renders nothing to avoid flash.
// When auth.enabled=false, the backend injects a dev user so /api/me
// returns a valid user — no redirect occurs in development.
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth()
  if (loading) return null
  if (!user) return <Navigate to="/login" replace />
  return <>{children}</>
}

export default function App() {
  return (
    <AuthProvider>
      <ChangesetProvider>
        <ModelProvider>
          <SearchProvider>
            <Routes>
              <Route path="/login" element={<LoginPage />} />
              <Route
                element={
                  <ProtectedRoute>
                    <AppShell />
                  </ProtectedRoute>
                }
              >
                {/* Core pages */}
                <Route path="/" element={<UploadPage />} />
                <Route path="/models" element={<ModelsPage />} />
                <Route path="/dashboard" element={<DashboardPage />} />
                <Route path="/history" element={<ModelHistoryPage />} />

                {/* Architecture — Needs */}
                <Route path="/unm-map" element={<UNMMapView />} />
                <Route path="/needs" element={<NeedView />} />
                <Route path="/needs/traceability" element={<RealizationView />} />
                <Route path="/needs/gaps" element={<GapsTab />} />

                {/* Architecture — Capabilities */}
                <Route path="/capabilities" element={<CapabilityView />} />
                <Route path="/capabilities/services" element={<RealizationView />} />
                <Route path="/capabilities/dependencies" element={<DependenciesTab />} />

                {/* Architecture — Teams */}
                <Route path="/teams" element={<TeamTopologyView />} />
                <Route path="/teams/ownership" element={<OwnershipView />} />
                <Route path="/teams/cognitive-load" element={<CognitiveLoadView />} />
                <Route path="/teams/interactions" element={<InteractionsTab />} />

                <Route path="/signals" element={<SignalsView />} />

                {/* Editing */}
                <Route path="/what-if" element={<WhatIfPage />} />

                {/* AI */}
                <Route path="/recommendations" element={<RecommendationsPage />} />
                <Route path="/advisor" element={<AdvisorPage />} />

                {/* Backward-compat redirects */}
                <Route path="/need" element={<Navigate to="/needs" replace />} />
                <Route path="/capability" element={<Navigate to="/capabilities" replace />} />
                <Route path="/realization" element={<Navigate to="/needs/traceability" replace />} />
                <Route path="/ownership" element={<Navigate to="/teams/ownership" replace />} />
                <Route path="/team-topology" element={<Navigate to="/teams" replace />} />
                <Route path="/cognitive-load" element={<Navigate to="/teams/cognitive-load" replace />} />
                <Route path="/edit" element={<Navigate to="/unm-map" replace />} />

                <Route path="*" element={<Navigate to="/" replace />} />
              </Route>
            </Routes>
          </SearchProvider>
        </ModelProvider>
      </ChangesetProvider>
    </AuthProvider>
  )
}
