import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { cn } from '@/lib/utils'
import type { ComponentPropsWithoutRef } from 'react'

interface ProseProps {
  children: string
  className?: string
  compact?: boolean
}

const COMPONENTS: ComponentPropsWithoutRef<typeof ReactMarkdown>['components'] = {
  h1: ({ children }) => (
    <h1 className="text-lg font-bold tracking-tight mb-4 pb-2" style={{ color: '#111827', borderBottom: '1px solid #e5e7eb' }}>
      {children}
    </h1>
  ),
  h2: ({ children }) => (
    <h2 className="text-sm font-bold mt-6 mb-2 pb-1.5" style={{ color: '#111827', borderBottom: '1px solid #f3f4f6' }}>
      {children}
    </h2>
  ),
  h3: ({ children }) => (
    <h3 className="text-xs font-bold mt-4 mb-1.5" style={{ color: '#111827' }}>
      {children}
    </h3>
  ),
  h4: ({ children }) => (
    <h4 className="text-xs font-semibold mt-3 mb-1" style={{ color: '#374151' }}>
      {children}
    </h4>
  ),
  p: ({ children }) => (
    <p className="text-xs leading-relaxed mb-3" style={{ color: '#4b5563' }}>
      {children}
    </p>
  ),
  strong: ({ children }) => (
    <strong className="font-semibold" style={{ color: '#111827' }}>{children}</strong>
  ),
  em: ({ children }) => (
    <em className="italic" style={{ color: '#6b7280' }}>{children}</em>
  ),
  ul: ({ children }) => (
    <ul className="space-y-1 mb-3 ml-4 list-disc" style={{ color: '#4b5563' }}>
      {children}
    </ul>
  ),
  ol: ({ children }) => (
    <ol className="space-y-1 mb-3 ml-4 list-decimal" style={{ color: '#4b5563' }}>
      {children}
    </ol>
  ),
  li: ({ children }) => (
    <li className="text-xs leading-relaxed pl-0.5" style={{ color: '#4b5563' }}>
      {children}
    </li>
  ),
  blockquote: ({ children }) => (
    <blockquote className="my-3 pl-3 py-1 text-xs italic rounded-r" style={{ borderLeft: '3px solid #6366f1', background: '#f5f3ff', color: '#6b7280' }}>
      {children}
    </blockquote>
  ),
  code: ({ className, children }) => {
    const isBlock = className?.includes('language-')
    if (isBlock) {
      return (
        <code className="block overflow-x-auto rounded-lg px-4 py-3 text-[11px] font-mono leading-relaxed mb-3" style={{ background: '#1e293b', color: '#e2e8f0', border: '1px solid #334155' }}>
          {children}
        </code>
      )
    }
    return (
      <code className="px-1.5 py-0.5 rounded text-[11px] font-mono" style={{ background: '#f1f5f9', color: '#0f172a', border: '1px solid #e2e8f0' }}>
        {children}
      </code>
    )
  },
  pre: ({ children }) => (
    <pre className="my-3">{children}</pre>
  ),
  a: ({ href, children }) => (
    <a href={href} target="_blank" rel="noopener noreferrer" className="underline decoration-1 underline-offset-2 transition-colors hover:decoration-2" style={{ color: '#2563eb' }}>
      {children}
    </a>
  ),
  hr: () => (
    <hr className="my-4" style={{ borderColor: '#e5e7eb' }} />
  ),
  table: ({ children }) => (
    <div className="my-3 overflow-x-auto rounded-lg" style={{ border: '1px solid #e5e7eb' }}>
      <table className="w-full text-xs">{children}</table>
    </div>
  ),
  thead: ({ children }) => (
    <thead style={{ background: '#f9fafb', borderBottom: '1px solid #e5e7eb' }}>{children}</thead>
  ),
  th: ({ children }) => (
    <th className="px-3 py-2 text-left text-[10px] font-semibold uppercase tracking-wide" style={{ color: '#6b7280' }}>
      {children}
    </th>
  ),
  td: ({ children }) => (
    <td className="px-3 py-2 text-xs" style={{ color: '#374151', borderTop: '1px solid #f3f4f6' }}>
      {children}
    </td>
  ),
}

const COMPACT_COMPONENTS: ComponentPropsWithoutRef<typeof ReactMarkdown>['components'] = {
  ...COMPONENTS,
  h1: ({ children }) => (
    <h1 className="text-sm font-bold mb-2" style={{ color: '#111827' }}>{children}</h1>
  ),
  h2: ({ children }) => (
    <h2 className="text-xs font-bold mt-4 mb-1.5" style={{ color: '#111827' }}>{children}</h2>
  ),
  h3: ({ children }) => (
    <h3 className="text-[11px] font-bold mt-3 mb-1" style={{ color: '#111827' }}>{children}</h3>
  ),
  p: ({ children }) => (
    <p className="text-[11px] leading-relaxed mb-2" style={{ color: '#4b5563' }}>{children}</p>
  ),
  li: ({ children }) => (
    <li className="text-[11px] leading-relaxed" style={{ color: '#4b5563' }}>{children}</li>
  ),
}

export function Prose({ children, className, compact }: ProseProps) {
  return (
    <div className={cn('min-w-0', className)}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={compact ? COMPACT_COMPONENTS : COMPONENTS}
      >
        {children}
      </ReactMarkdown>
    </div>
  )
}
