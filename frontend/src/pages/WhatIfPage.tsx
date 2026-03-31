import { useState } from 'react'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { useModel } from '@/lib/model-context'
import { useAIEnabled } from '@/hooks/useAIEnabled'
import { Sparkles, Wrench } from 'lucide-react'
import { PageHeader } from '@/components/ui/page-header'
import { AIWhatIfTab } from '@/features/whatif/AIWhatIfTab'
import { ManualWhatIfTab } from '@/features/whatif/ManualWhatIfTab'
import { cn } from '@/lib/utils'

type Tab = 'ai' | 'manual'

export function WhatIfPage() {
  const { modelId } = useModel()
  const aiEnabled = useAIEnabled()
  const [tab, setTab] = useState<Tab>(aiEnabled ? 'ai' : 'manual')

  return (
    <ModelRequired>
      <div className="max-w-6xl space-y-6 h-full flex flex-col">
        <PageHeader
          title="What-If Explorer"
          description="Explore architectural changes with AI scenarios or manual changesets"
          actions={
            <div className="flex rounded-lg overflow-hidden border border-border">
              {aiEnabled && (
                <button
                  onClick={() => setTab('ai')}
                  className={cn('px-4 py-2 text-sm font-medium flex items-center gap-2 transition-colors',
                    tab === 'ai' ? 'bg-gray-900 text-white' : 'bg-white text-gray-700 hover:bg-gray-50')}
                >
                  <Sparkles className="w-3.5 h-3.5" /> AI Scenarios
                </button>
              )}
              <button
                onClick={() => setTab('manual')}
                className={cn('px-4 py-2 text-sm font-medium flex items-center gap-2 transition-colors',
                  aiEnabled && 'border-l border-border',
                  tab === 'manual' ? 'bg-gray-900 text-white' : 'bg-white text-gray-700 hover:bg-gray-50')}
              >
                <Wrench className="w-3.5 h-3.5" /> Manual Mode
              </button>
            </div>
          }
        />
        <div className="flex-1 min-h-0">
          {tab === 'ai' && modelId ? (
            <AIWhatIfTab modelId={modelId} />
          ) : modelId ? (
            <ManualWhatIfTab modelId={modelId} />
          ) : null}
        </div>
      </div>
    </ModelRequired>
  )
}
