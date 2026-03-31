import { useState, useRef, useEffect } from 'react'
import { User, Bot, Download, Sparkles, Zap, Brain, Gauge, RefreshCw } from 'lucide-react'
import { Prose } from '@/components/ui/prose'
import { exportToPdf } from '@/lib/export-pdf'
import type { RoutingInfo } from '@/types/insights'

export interface ChatEntry {
  question: string
  answer: string
  aiConfigured: boolean
  routing?: RoutingInfo
}

type TierKey = 'simple' | 'medium' | 'complex'

const TIER_CONFIG: Record<TierKey, { label: string; desc: string; icon: typeof Zap; color: string; bg: string }> = {
  simple: { label: 'Quick', desc: 'Fast model, no reasoning', icon: Zap, color: '#22c55e', bg: '#f0fdf4' },
  medium: { label: 'Standard', desc: 'Balanced model, low reasoning', icon: Gauge, color: '#6366f1', bg: '#eef2ff' },
  complex: { label: 'Deep analysis', desc: 'Best model, high reasoning', icon: Brain, color: '#f59e0b', bg: '#fffbeb' },
}

const TIER_ORDER: TierKey[] = ['simple', 'medium', 'complex']

interface ChatMessageProps {
  entry: ChatEntry
  onApply?: (answer: string) => void
  onRetryWithTier?: (question: string, tier: TierKey) => void
}

export function ChatMessage({ entry, onApply, onRetryWithTier }: ChatMessageProps) {
  const [showTierPicker, setShowTierPicker] = useState(false)
  const pickerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!showTierPicker) return
    const handler = (e: MouseEvent) => {
      if (pickerRef.current && !pickerRef.current.contains(e.target as Node)) {
        setShowTierPicker(false)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [showTierPicker])

  const currentTier = entry.routing?.tier as TierKey | undefined

  return (
    <div className="space-y-4">
      {/* User question */}
      <div className="flex justify-end gap-2">
        <div
          className="max-w-[min(100%,28rem)] px-4 py-3 text-sm leading-relaxed"
          style={{
            background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
            color: '#ffffff',
            borderRadius: '20px 20px 4px 20px',
            boxShadow: '0 4px 14px rgba(99, 102, 241, 0.25)',
          }}
        >
          <div className="text-[10px] font-bold uppercase tracking-wider opacity-90 mb-1.5 flex items-center gap-1.5">
            <User size={12} strokeWidth={2.5} className="opacity-90" aria-hidden />
            You
          </div>
          {entry.question}
        </div>
      </div>

      {/* AI answer */}
      <div className="flex justify-start gap-2">
        <div
          className="flex max-w-full min-w-0 flex-1 flex-col px-4 py-3"
          style={{
            background: '#ffffff',
            border: '1px solid #e2e8f0',
            borderRadius: '20px 20px 20px 4px',
            boxShadow: '0 1px 3px rgba(0,0,0,0.06)',
          }}
        >
          <div className="flex items-center gap-2 mb-2">
            <div
              className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg"
              style={{ background: 'linear-gradient(135deg, #eef2ff 0%, #f5f3ff 100%)' }}
            >
              <Bot size={14} style={{ color: '#6366f1' }} />
            </div>
            <span className="text-xs font-bold" style={{ color: '#6366f1' }}>Advisor</span>
            {currentTier && (() => {
              const tier = TIER_CONFIG[currentTier]
              const Icon = tier.icon
              return (
                <div className="relative" ref={pickerRef}>
                  <button
                    type="button"
                    onClick={() => setShowTierPicker(s => !s)}
                    className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[9px] font-semibold cursor-pointer transition-opacity hover:opacity-80"
                    style={{ background: tier.bg, color: tier.color }}
                    title="Click to re-run with a different model"
                  >
                    <Icon size={9} />
                    {tier.label}
                  </button>
                  {showTierPicker && onRetryWithTier && (
                    <div
                      className="absolute left-0 top-full mt-1 z-50 w-56 rounded-xl border overflow-hidden"
                      style={{ background: '#ffffff', borderColor: '#e5e7eb', boxShadow: '0 12px 32px -8px rgba(0,0,0,0.18)' }}
                    >
                      <div className="px-3 py-2" style={{ borderBottom: '1px solid #f3f4f6' }}>
                        <span className="text-[10px] font-semibold uppercase tracking-wide" style={{ color: '#9ca3af' }}>
                          Re-run with different model
                        </span>
                      </div>
                      {TIER_ORDER.map(key => {
                        const t = TIER_CONFIG[key]
                        const TIcon = t.icon
                        const isCurrent = key === currentTier
                        return (
                          <button
                            key={key}
                            type="button"
                            disabled={isCurrent}
                            onClick={() => {
                              setShowTierPicker(false)
                              onRetryWithTier(entry.question, key)
                            }}
                            className="w-full flex items-center gap-2.5 px-3 py-2 text-left transition-colors hover:bg-gray-50 disabled:opacity-40 disabled:pointer-events-none"
                          >
                            <span
                              className="shrink-0 flex items-center justify-center w-6 h-6 rounded-md"
                              style={{ background: t.bg }}
                            >
                              <TIcon size={12} style={{ color: t.color }} />
                            </span>
                            <div className="min-w-0">
                              <div className="flex items-center gap-1.5">
                                <span className="text-xs font-semibold" style={{ color: '#111827' }}>{t.label}</span>
                                {isCurrent && (
                                  <span className="text-[8px] font-bold px-1 py-0.5 rounded" style={{ background: '#f3f4f6', color: '#9ca3af' }}>
                                    current
                                  </span>
                                )}
                              </div>
                              <span className="text-[10px]" style={{ color: '#9ca3af' }}>{t.desc}</span>
                            </div>
                            {!isCurrent && <RefreshCw size={10} className="ml-auto shrink-0" style={{ color: '#d1d5db' }} />}
                          </button>
                        )
                      })}
                    </div>
                  )}
                </div>
              )
            })()}
            <div className="ml-auto flex items-center gap-1">
              {onApply && entry.aiConfigured && (
                <button
                  type="button"
                  onClick={() => onApply(entry.answer)}
                  title="Apply recommendations as changes"
                  className="flex items-center gap-1.5 rounded-md px-2.5 py-1 text-[11px] font-medium transition-all hover:opacity-90"
                  style={{ background: '#111827', color: '#ffffff' }}
                >
                  <Sparkles size={11} />
                  Apply
                </button>
              )}
              <button
                type="button"
                onClick={() => exportToPdf(entry.answer, `AI Advisor — ${entry.question.slice(0, 60)}`)}
                title="Save as PDF"
                className="rounded-md p-1.5 transition-colors hover:bg-gray-100"
                style={{ color: '#9ca3af' }}
              >
                <Download size={12} />
              </button>
            </div>
          </div>
          <div className="pl-0.5">
            <Prose>{entry.answer}</Prose>
          </div>
        </div>
      </div>
    </div>
  )
}
