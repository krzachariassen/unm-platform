import { type ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { Upload } from 'lucide-react'
import { useModel } from '@/lib/model-context'

interface ModelRequiredProps {
  children: ReactNode
}

/**
 * Guard component for pages that require a loaded model.
 * - While hydrating: shows a centered spinner.
 * - After hydration, when no model is loaded: shows a polished empty state.
 * - When a model is present: renders children.
 */
export function ModelRequired({ children }: ModelRequiredProps) {
  const { modelId, isHydrating } = useModel()
  const navigate = useNavigate()

  if (isHydrating) {
    return (
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          minHeight: 320,
          width: '100%',
        }}
      >
        <span
          style={{
            width: 28,
            height: 28,
            border: '3px solid #e2e8f0',
            borderTopColor: '#6366f1',
            borderRadius: '50%',
            display: 'inline-block',
            animation: 'model-required-spin 0.8s linear infinite',
          }}
          aria-label="Loading..."
          role="status"
        />
        <style>{`@keyframes model-required-spin { to { transform: rotate(360deg); } }`}</style>
      </div>
    )
  }

  if (!modelId) {
    return (
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          minHeight: 320,
          width: '100%',
          padding: '40px 16px',
        }}
      >
        <div
          style={{
            background: '#ffffff',
            border: '1px solid #e2e8f0',
            borderRadius: 20,
            boxShadow: '0 1px 6px rgba(0,0,0,0.05)',
            padding: '48px 40px',
            textAlign: 'center',
            maxWidth: 400,
            width: '100%',
          }}
        >
          <div
            style={{
              width: 56,
              height: 56,
              borderRadius: 16,
              background: 'linear-gradient(135deg, #ede9fe 0%, #dbeafe 100%)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              margin: '0 auto 20px',
            }}
          >
            <Upload size={24} style={{ color: '#6366f1' }} />
          </div>
          <h2
            style={{
              fontSize: 18,
              fontWeight: 700,
              color: '#1e293b',
              margin: '0 0 8px 0',
              letterSpacing: '-0.01em',
            }}
          >
            No model loaded
          </h2>
          <p
            style={{
              fontSize: 14,
              color: '#64748b',
              margin: '0 0 28px 0',
              lineHeight: 1.5,
            }}
          >
            Upload a .unm.yaml file to get started
          </p>
          <button
            onClick={() => navigate('/')}
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 8,
              padding: '10px 24px',
              background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
              color: '#ffffff',
              border: 'none',
              borderRadius: 10,
              fontSize: 14,
              fontWeight: 600,
              cursor: 'pointer',
              letterSpacing: '-0.01em',
            }}
            onMouseEnter={e => {
              e.currentTarget.style.opacity = '0.9'
            }}
            onMouseLeave={e => {
              e.currentTarget.style.opacity = '1'
            }}
          >
            <Upload size={15} />
            Load a model
          </button>
        </div>
      </div>
    )
  }

  return <>{children}</>
}
