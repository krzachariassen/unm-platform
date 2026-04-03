import { useState } from 'react'
import { authApi } from '@/services/api/auth'
import { useAuth } from '@/lib/auth-context'
import { useNavigate } from 'react-router-dom'

// Show "Continue as Dev User" in Vite dev builds (never in production bundles).
const IS_DEV = import.meta.env.DEV

export function LoginPage() {
  const { user } = useAuth()
  const navigate = useNavigate()
  const [devLoading, setDevLoading] = useState(false)
  const [devError, setDevError] = useState<string | null>(null)

  if (user) {
    navigate('/', { replace: true })
    return null
  }

  const handleDevLogin = async () => {
    setDevLoading(true)
    setDevError(null)
    try {
      const ok = await authApi.devLogin()
      if (ok) {
        // Session cookie is now set — reload so AuthProvider re-fetches /api/me.
        window.location.href = '/'
      } else {
        setDevError('Dev login not available. Set auth.dev_login=true in backend config.')
      }
    } catch {
      setDevError('Could not reach backend. Make sure it is running on localhost:8080.')
    } finally {
      setDevLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <div className="flex flex-col items-center gap-6 p-8 rounded-2xl border border-border bg-card shadow-sm max-w-sm w-full">
        <div className="flex flex-col items-center gap-2">
          <h1 className="text-2xl font-bold text-foreground">UNM Platform</h1>
          <p className="text-sm text-muted-foreground text-center">
            Sign in to access your architecture models and team insights.
          </p>
        </div>

        <a
          href={authApi.loginUrl()}
          className="flex items-center justify-center gap-3 w-full px-4 py-2.5 rounded-lg border border-border bg-background hover:bg-muted transition-colors text-sm font-medium text-foreground"
        >
          <GoogleIcon />
          Sign in with Google
        </a>

        {IS_DEV && (
          <>
            <div className="flex items-center gap-2 w-full">
              <div className="flex-1 h-px bg-border" />
              <span className="text-xs text-muted-foreground">local dev</span>
              <div className="flex-1 h-px bg-border" />
            </div>

            <button
              type="button"
              onClick={handleDevLogin}
              disabled={devLoading}
              className="flex items-center justify-center gap-2 w-full px-4 py-2.5 rounded-lg border border-dashed border-border bg-muted/30 hover:bg-muted transition-colors text-sm font-medium text-muted-foreground disabled:opacity-50"
            >
              {devLoading ? 'Connecting…' : 'Continue as Dev User'}
            </button>

            {devError && (
              <p className="text-xs text-red-600 text-center">{devError}</p>
            )}
          </>
        )}
      </div>
    </div>
  )
}

function GoogleIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 18 18" aria-hidden="true">
      <path
        d="M17.64 9.205c0-.639-.057-1.252-.164-1.841H9v3.481h4.844a4.14 4.14 0 0 1-1.796 2.716v2.259h2.908c1.702-1.567 2.684-3.875 2.684-6.615z"
        fill="#4285F4"
      />
      <path
        d="M9 18c2.43 0 4.467-.806 5.956-2.18l-2.908-2.259c-.806.54-1.837.86-3.048.86-2.344 0-4.328-1.584-5.036-3.711H.957v2.332A8.997 8.997 0 0 0 9 18z"
        fill="#34A853"
      />
      <path
        d="M3.964 10.71A5.41 5.41 0 0 1 3.682 9c0-.593.102-1.17.282-1.71V4.958H.957A8.996 8.996 0 0 0 0 9c0 1.452.348 2.827.957 4.042l3.007-2.332z"
        fill="#FBBC05"
      />
      <path
        d="M9 3.58c1.321 0 2.508.454 3.44 1.345l2.582-2.58C13.463.891 11.426 0 9 0A8.997 8.997 0 0 0 .957 4.958L3.964 7.29C4.672 5.163 6.656 3.58 9 3.58z"
        fill="#EA4335"
      />
    </svg>
  )
}
