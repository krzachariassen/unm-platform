// This page is retired — the /edit route now redirects to /unm-map.
// Editing is done via the batch edit mode in the map view.
import { Navigate } from 'react-router-dom'

export function EditModelPage() {
  return <Navigate to="/unm-map" replace />
}
