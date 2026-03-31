import { type ReactNode } from 'react'
import { cn } from '@/lib/utils'

interface ContentContainerProps {
  children: ReactNode
  className?: string
  fullWidth?: boolean
}

export function ContentContainer({ children, className, fullWidth = false }: ContentContainerProps) {
  return (
    <div className={cn(fullWidth ? 'w-full' : 'max-w-screen-xl mx-auto', className)}>
      {children}
    </div>
  )
}
