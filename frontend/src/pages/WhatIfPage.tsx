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
      <div className="max-w-screen-xl mx-auto space-y-4 h-full flex flex-col">
        <PageHeader
          title="What-If Explorer"
          description="Explore architectural changes with AI scenarios or manual changesets"
          actions={
            <div className="flex rounded-md overflow-hidden border border-border">
              {aiEnabled && (
                <button
                  onClick={() => setTab('ai')}
                  className={cn('px-3 py-1.5 text-xs font-medium flex items-center gap-1.5 transition-colors',
                    tab === 'ai' ? 'bg-foreground text-background' : 'bg-card text-muted-foreground hover:bg-muted')}
                >
                  <Sparkles className="w-3 h-3" /> AI Scenarios
                </button>
              )}
              <button
                onClick={() => setTab('manual')}
                className={cn('px-3 py-1.5 text-xs font-medium flex items-center gap-1.5 transition-colors',
                  aiEnabled && 'border-l border-border',
                  tab === 'manual' ? 'bg-foreground text-background' : 'bg-card text-muted-foreground hover:bg-muted')}
              >
                <Wrench className="w-3 h-3" /> Manual Mode
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
