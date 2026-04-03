import type {
  NeedViewResponse,
  CapabilityViewResponse,
  OwnershipViewResponse,
  TeamTopologyViewResponse,
  CognitiveLoadViewResponse,
  SignalsViewResponse,
  RealizationViewResponse,
  UNMMapViewResponse,
  GapsView,
  DependenciesView,
  InteractionsView,
} from '@/types/views'
import { apiFetch } from './client'

export const viewsApi = {
  getNeedView: (modelId: string, signal?: AbortSignal): Promise<NeedViewResponse> =>
    apiFetch(`/models/${encodeURIComponent(modelId)}/views/need`, { signal }),

  getCapabilityView: (modelId: string, signal?: AbortSignal): Promise<CapabilityViewResponse> =>
    apiFetch(`/models/${encodeURIComponent(modelId)}/views/capability`, { signal }),

  getOwnershipView: (modelId: string, signal?: AbortSignal): Promise<OwnershipViewResponse> =>
    apiFetch(`/models/${encodeURIComponent(modelId)}/views/ownership`, { signal }),

  getTeamTopologyView: (modelId: string, signal?: AbortSignal): Promise<TeamTopologyViewResponse> =>
    apiFetch(`/models/${encodeURIComponent(modelId)}/views/team-topology`, { signal }),

  getCognitiveLoadView: (modelId: string, signal?: AbortSignal): Promise<CognitiveLoadViewResponse> =>
    apiFetch(`/models/${encodeURIComponent(modelId)}/views/cognitive-load`, { signal }),

  getSignalsView: (modelId: string, signal?: AbortSignal): Promise<SignalsViewResponse> =>
    apiFetch(`/models/${encodeURIComponent(modelId)}/views/signals`, { signal }),

  getRealizationView: (modelId: string, signal?: AbortSignal): Promise<RealizationViewResponse> =>
    apiFetch(`/models/${encodeURIComponent(modelId)}/views/realization`, { signal }),

  getUNMMapView: (modelId: string, signal?: AbortSignal): Promise<UNMMapViewResponse> =>
    apiFetch(`/models/${encodeURIComponent(modelId)}/views/unm-map`, { signal }),

  getGaps: (modelId: string, signal?: AbortSignal): Promise<GapsView> =>
    apiFetch(`/views/gaps?model_id=${encodeURIComponent(modelId)}`, { signal }),

  getDependencies: (modelId: string, signal?: AbortSignal): Promise<DependenciesView> =>
    apiFetch(`/views/dependencies?model_id=${encodeURIComponent(modelId)}`, { signal }),

  getInteractions: (modelId: string, signal?: AbortSignal): Promise<InteractionsView> =>
    apiFetch(`/views/interactions?model_id=${encodeURIComponent(modelId)}`, { signal }),
}
