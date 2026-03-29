import { AlertTriangle } from 'lucide-react'

/**
 * Shared loading spinner shown while a view is fetching data.
 */
export function LoadingState({ message = 'Loading…' }: { message?: string }) {
  return (
    <div className="flex flex-col items-center justify-center gap-3 h-full min-h-[200px]">
      <div
        className="rounded-full animate-spin"
        style={{
          width: 36,
          height: 36,
          border: '2px solid #e2e8f0',
          borderTopColor: '#6366f1',
        }}
      />
      <span style={{ fontSize: 14, color: '#94a3b8' }}>{message}</span>
    </div>
  )
}

/**
 * Shared error display shown when a view fetch fails.
 */
export function ErrorState({ message }: { message: string }) {
  return (
    <div className="flex flex-col items-center justify-center gap-3 h-full min-h-[200px] px-4">
      <div
        className="flex items-center gap-3 rounded-2xl px-5 py-4 max-w-md w-full"
        style={{
          borderRadius: 20,
          background: 'linear-gradient(135deg, #ffffff 0%, #f8fafc 100%)',
          border: '1px solid #fecaca',
          boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
        }}
      >
        <div className="rounded-xl p-2 flex-shrink-0" style={{ background: '#fee2e2' }}>
          <span title="Error loading view" aria-label="Error">
            <AlertTriangle size={20} style={{ color: '#dc2626' }} />
          </span>
        </div>
        <span className="text-sm font-medium" style={{ color: '#b91c1c' }}>{message}</span>
      </div>
    </div>
  )
}
