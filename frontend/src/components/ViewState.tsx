import { AlertTriangle } from 'lucide-react'

/**
 * Shared loading spinner shown while a view is fetching data.
 */
export function LoadingState({ message = 'Loading…' }: { message?: string }) {
  return (
    <div className="flex h-full min-h-[200px] flex-col items-center justify-center gap-3">
      <div
        className="size-9 animate-spin rounded-full border-2 border-muted border-t-primary"
      />
      <span className="text-sm text-muted-foreground">{message}</span>
    </div>
  )
}

/**
 * Shared error display shown when a view fetch fails.
 */
export function ErrorState({ message }: { message: string }) {
  return (
    <div className="flex h-full min-h-[200px] flex-col items-center justify-center gap-3 px-4">
      <div className="flex w-full max-w-md items-center gap-3 rounded-lg border border-destructive/30 bg-card px-5 py-4">
        <div className="shrink-0 rounded-lg bg-destructive/15 p-2">
          <span title="Error loading view" aria-label="Error">
            <AlertTriangle size={20} className="text-destructive" />
          </span>
        </div>
        <span className="text-sm font-medium text-destructive">{message}</span>
      </div>
    </div>
  )
}
